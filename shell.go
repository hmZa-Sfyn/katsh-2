package main

import (
	"strconv"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
//  Background job
// ─────────────────────────────────────────────────────────────────────────────

// BgJob represents a command running asynchronously via ?(cmd) syntax.
type BgJob struct {
	ID     int
	Cmd    string
	Done   chan struct{}
	Result string
	Err    error
}

// runBgJob executes cmd in a goroutine and stores the result.
func runBgJob(cmd, cwd string) *BgJob {
	job := &BgJob{Cmd: cmd, Done: make(chan struct{})}
	go func() {
		defer close(job.Done)
		out, err := RunCaptureShell("", cmd, cwd)
		job.Result = strings.TrimRight(out, "\n ")
		job.Err = err
	}()
	return job
}

// waitBgJob blocks until the job finishes and returns its output.
func waitBgJob(job *BgJob) (string, error) {
	<-job.Done
	return job.Result, job.Err
}



type HistoryEntry struct {
	Raw      string    `json:"raw"`
	At       time.Time `json:"at"`
	ExitCode int       `json:"exit_code"`
}

type Shell struct {
	cwd      string
	prevDir  string
	dirStack []string
	box      *Box
	history  []HistoryEntry
	aliases  map[string]Alias
	vars     map[string]string
	funcs    map[string]*UserFunc
	lastCode int

	// Scripting extras
	readonlyVars    map[string]bool
	arrays          map[string]*ShArray
	errHandlerDepth int
	lastErrMsg      string
	deferStack      []string // deferred commands (LIFO)

	// Output capture (for backtick subshells)
	captureMode bool
	captureOut  bytes.Buffer

	// Error location tracking — updated by runner.go before each execLine call
	currentLine int    // 1-based physical line number in the running script
	currentFile string // basename of the running script (e.g. "myscript.ksh")
	currentSrc  string // the raw source line being executed

	// Background job tracking
	bgJobs []*BgJob

	// Throw signal — carry the thrown message through non-zero return codes
	// so try/catch can retrieve the exact thrown string, not just "exit code N"
	thrownMsg string // set by throw/raise, cleared by catch

	// pasteQueue holds lines from a multi-line paste to be executed one by one
	pasteQueue []string
}

func NewShell() *Shell {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "/"
	}
	sh := &Shell{
		cwd:     cwd,
		box:     NewBox(),
		aliases: make(map[string]Alias),
		vars:    make(map[string]string),
		funcs:        make(map[string]*UserFunc),
		readonlyVars: make(map[string]bool),
		arrays:       make(map[string]*ShArray),
	}
	sh.loadHistory()
	return sh
}

// ─────────────────────────────────────────────────────────────────────────────
//  REPL
// ─────────────────────────────────────────────────────────────────────────────

func (sh *Shell) Run() {
	fmt.Print(banner())

	execAndRecord := func(raw string) {
		raw = strings.TrimSpace(raw)
		if raw == "" { return }
		sh.history = append(sh.history, HistoryEntry{Raw: raw, At: time.Now()})
		code := sh.execLine(raw)
		sh.lastCode = code
		sh.history[len(sh.history)-1].ExitCode = code
		sh.saveHistory()
	}

	for {
		// Drain any lines queued by a multi-line paste before prompting again
		for len(sh.pasteQueue) > 0 {
			line := sh.pasteQueue[0]
			sh.pasteQueue = sh.pasteQueue[1:]
			line = strings.TrimSpace(line)
			if line == "" { continue }
			// Echo the line so the user can see what's being executed
			prompt := renderPrompt(sh.cwd, os.Getenv("USER"), sh.lastCode)
			fmt.Printf("%s%s%s\n", prompt, line, ansiReset)
			execAndRecord(line)
		}

		prompt := renderPrompt(sh.cwd, os.Getenv("USER"), sh.lastCode)
		line, eof := sh.Readline(prompt)
		if eof { break }

		execAndRecord(line)
	}

	fmt.Println(c(ansiGrey, "\nbye."))
}

// ─────────────────────────────────────────────────────────────────────────────
//  execLine — parse and execute one raw input line
// ─────────────────────────────────────────────────────────────────────────────

func (sh *Shell) execLine(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" { return 0 }

	// Track current source for error reporting
	if sh.currentSrc == "" { sh.currentSrc = raw }

	rawLow := strings.ToLower(raw)

	// ── Control-flow keywords typed at the REPL top-level ────────────────
	switch {
	case rawLow == "return" || (strings.HasPrefix(rawLow, "return ") && !strings.Contains(raw, "{")):
		PrintError(&ShellError{Code:"E002", Kind:"SyntaxError",
			Message: "'return' can only be used inside a function body",
			Source: raw, Col: 0, Span: 6, Line: sh.currentLine,
			Hint: "define a function:  func myFunc() { ... return $val }"})
		return 1
	case rawLow == "break":
		PrintError(&ShellError{Code:"E002", Kind:"SyntaxError",
			Message: "'break' can only be used inside a loop body",
			Source: raw, Col: 0, Span: 5, Line: sh.currentLine,
			Hint: "use 'break' inside a for/while loop"})
		return 1
	case rawLow == "continue":
		PrintError(&ShellError{Code:"E002", Kind:"SyntaxError",
			Message: "'continue' can only be used inside a loop body",
			Source: raw, Col: 0, Span: 8, Line: sh.currentLine,
			Hint: "use 'continue' inside a for/while loop"})
		return 1
	}

	// ── EARLY OS-pipe detection — on the ORIGINAL raw line, BEFORE expansion ─
	// If the line contains a | and any right-hand segment is NOT a katsh op or
	// user-func, route the ENTIRE original line to the OS shell unchanged.
	// This handles:  figlet hello | lolcat
	//               ls -la | grep go | wc -l
	//               echo hi | upper   ← katsh op, stays in katsh
	if hasPipeOutsideQuotes(raw) && !isKatshPipeLine(raw, sh) {
		return RunPassthrough("", raw, sh.cwd)
	}

	// ── Expand backticks ──────────────────────────────────────────────────
	raw = sh.expandBackticks(raw)

	// ── $(...) and ?(...) expansion ───────────────────────────────────────
	// Strategy: if the line contains $(...) and the OUTER command is an
	// external tool (not a katsh builtin, not a script op), route the whole
	// original line to the OS shell — it handles $() natively and won't
	// misparse figlet art (which contains |) as a katsh pipe.
	if strings.Contains(raw, "$(") {
		// Determine the command name (first token) before expansion
		outerCmd := strings.ToLower(strings.Fields(raw)[0])
		outerIsScript := isKatshScriptStart(outerCmd)
		if !outerIsScript {
			if !isBuiltin(outerCmd) {
				// External command with $(...) — let the OS shell handle it
				return RunPassthrough("", raw, sh.cwd)
			}
		}
		// Safe to expand inline (katsh builtin or scripting op)
		raw = sh.expandDollarParens(raw)
	}
	if strings.Contains(raw, "?(") {
		raw = sh.expandDollarParens(raw)
	}

	// ── Passthrough prefixes (bash! zsh! run etc.) ────────────────────────
	if code, handled := sh.tryPassthrough(raw); handled { return code }

	// ── Scripting engine (if/for/while/func/variables etc.) ───────────────
	if handled, code := sh.evalScript(raw); handled { return code }

	// ── Standard command path ─────────────────────────────────────────────
	raw = sh.expandAliases(raw)
	raw = sh.expandVars(raw)
	pc  := Parse(raw)
	if len(pc.Args) == 0 { return 0 }

	command := pc.Args[0]
	args    := pc.Args[1:]

	// ── Literal value pipe: "hello" | upper  /  42 | add 8 ──────────────
	if command == "__literal__" && len(args) > 0 {
		litRaw := args[0]
		var result *Result
		if (strings.HasPrefix(litRaw, `"`) && strings.HasSuffix(litRaw, `"`)) ||
			(strings.HasPrefix(litRaw, `'`) && strings.HasSuffix(litRaw, `'`)) {
			result = NewTyped(litRaw[1:len(litRaw)-1], "string")
		} else {
			result = NewTyped(litRaw, "number")
		}
		if len(pc.Pipes) == 0 { sh.printResult(result); return 0 }
		var err error
		result, err = ApplyPipes(result, pc.Pipes, sh)
		if err != nil { sh.printErr(wrapErr(err, strings.Join(pc.Args, " "))); return 1 }
		sh.printResult(result)
		if pc.ShouldStore && result != nil { sh.storeResult(pc, "literal", result) }
		return 0
	}

	// ── Built-ins ─────────────────────────────────────────────────────────
	result, wasBuiltin, err := handleBuiltin(sh, command, args)
	if err == errExit { sh.saveHistory(); fmt.Println(c(ansiGrey, "bye.")); os.Exit(0) }
	if err == errClear { cmd := exec.Command("clear"); cmd.Stdout = os.Stdout; _ = cmd.Run(); return 0 }
	if wasBuiltin {
		if err != nil { sh.printErr(wrapErr(err, raw)); return 1 }
		if result != nil && (result.IsTable || strings.TrimSpace(result.Text) != "") {
			result, err = ApplyPipes(result, pc.Pipes, sh)
			if err != nil { sh.printErr(wrapErr(err, raw)); return 1 }
			sh.printResult(result)
		}
		if pc.ShouldStore && result != nil { sh.storeResult(pc, command, result) }
		return 0
	}

	// ── User-defined function call (bare: double 234 or double) ──────────
	if fn, ok := sh.funcs[command]; ok {
		sh.vars["_return"] = "" // clear before call
		code := sh.callUserFunc(fn, args, raw)
		ret  := sh.vars["_return"]
		if code == 0 {
			if len(pc.Pipes) > 0 && ret != "" {
				result = NewTyped(ret, KindString)
				result, _ = ApplyPipes(result, pc.Pipes, sh)
				sh.printResult(result)
			} else if ret != "" && sh.currentFile == "" {
				// At the interactive REPL (no script file), print return value
				sh.printResult(NewTyped(ret, KindString))
			}
		}
		return code
	}

	// ── Interactive/TUI commands — auto passthrough ───────────────────────
	if needsPassthrough(command) { return RunPassthroughArgs(command, args, sh.cwd) }

	// ── External command not in PATH ──────────────────────────────────────
	if findInPath(command) == "" {
		sh.printErr(errUnknownCmd(command, raw))
		return 127
	}

	// ── External command in PATH ──────────────────────────────────────────
	if len(pc.Pipes) == 0 { return RunPassthroughArgs(command, args, sh.cwd) }
	result, err = RunExternal(command, args, sh.cwd)
	if err != nil { sh.printErr(wrapErr(err, raw)); return 1 }
	result, err = ApplyPipes(result, pc.Pipes, sh)
	if err != nil { sh.printErr(wrapErr(err, raw)); return 1 }
	sh.printResult(result)
	if pc.ShouldStore { sh.storeResult(pc, command, result) }
	return 0
}


// ─────────────────────────────────────────────────────────────────────────────
//  Output rendering  (respects capture mode for backtick subshells)
// ─────────────────────────────────────────────────────────────────────────────

func (sh *Shell) printResult(r *Result) {
	if r == nil {
		return
	}
	if r.IsTable {
		lines := RenderTable(r.Cols, r.Rows)
		if sh.captureMode {
			// Return rows as newline-separated column values
			for _, row := range r.Rows {
				var parts []string
				for _, col := range r.Cols {
					parts = append(parts, row[col])
				}
				sh.captureOut.WriteString(strings.Join(parts, "\t") + "\n")
			}
			return
		}
		fmt.Println()
		for _, line := range lines {
			fmt.Println(line)
		}
		fmt.Println()
	} else {
		text := strings.TrimRight(r.Text, "\n")
		if text == "" {
			return
		}

		// Array result: render as a table with index column
		if r.ValueKind == KindArray {
			items := splitArrayResult(text)
			cols := []string{"index", "value"}
			rows := make([]Row, len(items))
			for i, it := range items { rows[i] = Row{"index": strconv.Itoa(i), "value": it} }
			if sh.captureMode {
				for _, it := range items { sh.captureOut.WriteString(it + "\n") }
				return
			}
			fmt.Println()
			for _, line := range RenderTable(cols, rows) { fmt.Println(line) }
			fmt.Println()
			return
		}

		// Typed scalar (string/number): show with kind badge
		if r.ValueKind == KindString || r.ValueKind == KindNumber {
			if sh.captureMode {
				sh.captureOut.WriteString(text)
				return
			}
			badge := c(ansiGrey, "("+r.ValueKind+")")
			fmt.Println()
			fmt.Println("  " + text + "  " + badge)
			fmt.Println()
			return
		}

		if sh.captureMode {
			sh.captureOut.WriteString(text)
			return
		}
		fmt.Println()
		for _, line := range strings.Split(text, "\n") {
			fmt.Println("  " + line)
		}
		fmt.Println()
	}
}

func (sh *Shell) printErr(e *ShellError) {
	if e == nil {
		return
	}
	// Inject file/line if not already set
	if e.Line == 0 && sh.currentLine > 0 {
		e.Line = sh.currentLine
	}
	if e.Source == "" && sh.currentSrc != "" {
		e.Source = sh.currentSrc
	}
	if sh.captureMode {
		// In capture mode just write to stderr
		fmt.Fprintln(os.Stderr, e.Message)
		return
	}
	PrintError(e)
}

// ─────────────────────────────────────────────────────────────────────────────
//  Box store helpers
// ─────────────────────────────────────────────────────────────────────────────

func (sh *Shell) storeResult(pc *ParsedCommand, command string, r *Result) {
	key := pc.StoreKey
	src := strings.Join(pc.Args, " ")
	if key == "" {
		key = sh.box.autoKey()
	}
	var e *BoxEntry
	if r.IsTable {
		e = sh.box.StoreTable(key, src, r.Cols, r.Rows)
	} else {
		e = sh.box.StoreText(key, src, r.Text)
	}
	fmt.Println(boxMsg(e.Key, e.ID, e.Size()))
	fmt.Println()
}

func (sh *Shell) printBoxList(entries []*BoxEntry) {
	fmt.Println()
	fmt.Printf("  %s%s◈ BOX  (%d entries)%s\n\n", ansiBold, ansiCyan, len(entries), ansiReset)
	cols := []string{"id", "key", "type", "size", "tags", "created", "source"}
	var rows []Row
	for _, e := range entries {
		tags := strings.Join(e.Tags, " #")
		if tags != "" {
			tags = "#" + tags
		}
		rows = append(rows, Row{
			"id":      fmt.Sprintf("%d", e.ID),
			"key":     e.Key,
			"type":    string(e.Type),
			"size":    e.Size(),
			"tags":    tags,
			"created": e.Created.Format("15:04:05"),
			"source":  truncStr(e.Source, 36),
		})
	}
	for _, line := range RenderTable(cols, rows) {
		fmt.Println(line)
	}
	fmt.Println()
	fmt.Println(c(ansiGrey, "  box get <key>  ·  box rm <key>  ·  box rename old new  ·  box tag <key> <tag>  ·  box export <file>"))
	fmt.Println()
}

func (sh *Shell) printBoxEntry(e *BoxEntry) {
	tags := ""
	if len(e.Tags) > 0 {
		tags = "  " + c(ansiMagenta, "#"+strings.Join(e.Tags, " #"))
	}
	fmt.Printf("\n  %s%s◈ box[%q]%s  %sid:%d%s%s%s\n\n",
		ansiBold, ansiCyan, e.Key, ansiReset,
		ansiGrey, e.ID, ansiReset,
		tags,
		c(ansiGrey, "  "+e.Created.Format("15:04:05")),
	)
	if e.Source != "" {
		fmt.Printf("  %s$ %s%s\n\n", ansiGrey, e.Source, ansiReset)
	}
	switch e.Type {
	case TypeTable:
		for _, line := range RenderTable(e.Cols, e.Rows) {
			fmt.Println(line)
		}
	case TypeText:
		for _, line := range strings.Split(strings.TrimRight(e.Text, "\n"), "\n") {
			fmt.Println("  " + line)
		}
	}
	fmt.Println()
}

// ─────────────────────────────────────────────────────────────────────────────
//  Alias & variable expansion
// ─────────────────────────────────────────────────────────────────────────────

func (sh *Shell) expandAliases(raw string) string {
	parts := strings.SplitN(raw, " ", 2)
	name := parts[0]
	if a, ok := sh.aliases[name]; ok {
		if len(parts) > 1 {
			return a.Expand + " " + parts[1]
		}
		return a.Expand
	}
	return raw
}

func (sh *Shell) expandVars(raw string) string {
	return os.Expand(raw, func(key string) string {
		if v, ok := sh.vars[key]; ok {
			return v
		}
		return os.Getenv(key)
	})
}

// ─────────────────────────────────────────────────────────────────────────────
//  Shell passthrough — run commands in bash/zsh/sh with a real TTY
// ─────────────────────────────────────────────────────────────────────────────

// tryPassthrough checks for passthrough syntax and runs it immediately.
// Returns (exitCode, true) if matched; (0, false) if not a passthrough command.
//
// Syntax summary:
//   bash! git log | grep fix          explicit bash — rest of line is the cmd
//   zsh!  print -P "%F{red}hi%f"      explicit zsh
//   sh!   for f in *.go; do wc $f; done
//   run   git log | grep fix          uses $SHELL or bash
//   !     git log | grep fix          bare ! prefix (same as run)
//   bash                              bare name → interactive session
//   zsh                               bare name → interactive session
func (sh *Shell) tryPassthrough(raw string) (int, bool) {
	trimmed := strings.TrimSpace(raw)

	// run / shell prefix
	for _, prefix := range []string{"run ", "run! ", "shell ", "shell! "} {
		if strings.HasPrefix(trimmed, prefix) {
			cmdStr := sh.shellExpand(strings.TrimSpace(trimmed[len(prefix):]))
			return RunPassthrough("", cmdStr, sh.cwd), true
		}
	}

	// ! prefix  (but not !=  !~  used as operators)
	if len(trimmed) > 1 && trimmed[0] == '!' && trimmed[1] != '=' && trimmed[1] != '~' {
		cmdStr := sh.shellExpand(strings.TrimSpace(trimmed[1:]))
		return RunPassthrough("", cmdStr, sh.cwd), true
	}

	// bash! cmd  /  zsh! cmd  /  sh! cmd
	for _, shell := range []string{"bash", "zsh", "sh", "fish", "ksh", "dash", "tcsh"} {
		bangPrefix := shell + "! "
		if strings.HasPrefix(trimmed, bangPrefix) {
			cmdStr := sh.shellExpand(strings.TrimSpace(trimmed[len(bangPrefix):]))
			return RunPassthrough(shell, cmdStr, sh.cwd), true
		}
		// bare  bash!  → interactive session
		if trimmed == shell+"!" {
			return RunPassthrough(shell, "", sh.cwd), true
		}
	}

	// bare shell name → interactive session  (e.g. user types just "bash")
	for _, shell := range []string{"bash", "zsh", "fish", "ksh", "tcsh"} {
		if trimmed == shell {
			return RunPassthrough(shell, "", sh.cwd), true
		}
	}

	// bash -c "..."  /  zsh -c "..."  — native syntax, routed to passthrough
	for _, shell := range []string{"bash", "zsh", "sh", "fish"} {
		dashC := shell + " -c "
		if strings.HasPrefix(trimmed, dashC) {
			cmdStr := stripOuterQuotes(sh.shellExpand(strings.TrimSpace(trimmed[len(dashC):])))
			return RunPassthrough(shell, cmdStr, sh.cwd), true
		}
		// bash ./script.sh  /  zsh script.zsh
		shellSpace := shell + " "
		if strings.HasPrefix(trimmed, shellSpace) {
			rest := strings.TrimSpace(trimmed[len(shellSpace):])
			for _, ext := range []string{".sh", ".bash", ".zsh", ".ksh"} {
				if strings.Contains(rest, ext) {
					return RunPassthrough(shell, trimmed, sh.cwd), true
				}
			}
		}
	}

	return 0, false
}

// ─────────────────────────────────────────────────────────────────────────────
//  OS-pipe detection helpers
// ─────────────────────────────────────────────────────────────────────────────

// hasPipeOutsideQuotes returns true if s contains a | outside any quotes.
func hasPipeOutsideQuotes(s string) bool {
	inQ := false; qCh := rune(0)
	for _, ch := range s {
		if inQ { if ch == qCh { inQ = false }; continue }
		if ch == '"' || ch == '\'' { inQ = true; qCh = ch; continue }
		if ch == '|' { return true }
	}
	return false
}

// countPipesOutsideQuotes counts | characters outside quotes.
func countPipesOutsideQuotes(s string) int {
	n := 0; inQ := false; qCh := rune(0)
	for _, ch := range s {
		if inQ { if ch == qCh { inQ = false }; continue }
		if ch == '"' || ch == '\'' { inQ = true; qCh = ch; continue }
		if ch == '|' { n++ }
	}
	return n
}

// isKatshPipeLine returns true when the line's pipe stages are ALL
// katsh-native ops or user-defined functions (so katsh can handle them).
// Returns false if any segment is an external OS command (route to shell).
func isKatshPipeLine(raw string, sh *Shell) bool {
	segments := splitOnPipes(raw)
	if len(segments) < 2 { return true } // no pipe at all — katsh handles
	for _, seg := range segments[1:] {
		fields := strings.Fields(strings.TrimSpace(seg))
		if len(fields) == 0 { continue }
		op := strings.ToLower(fields[0])
		if isKatshPipeOp(op) { continue }
		if isStringOp(op) { continue }
		if sh != nil {
			if _, ok := sh.funcs[op]; ok { continue }
		}
		return false // this segment is an OS command → route everything to shell
	}
	return true
}

// isKatshScriptStart returns true for keywords that begin katsh scripting
// constructs — these need $() expansion done inline by katsh (not the OS shell).
func isKatshScriptStart(cmd string) bool {
	switch cmd {
	case "if","unless","for","while","until","func","match","try","switch",
		"enum","struct","defer","with","when","repeat","do","echo","println",
		"print","let","set","export","readonly","unset","source","alias":
		return true
	}
	return false
}

// looksLikeOSPipe is kept for compatibility.
func looksLikeOSPipe(raw string) bool { return !isKatshPipeLine(raw, nil) }

// allKatshPipes is kept for compatibility.
func allKatshPipes(pipes []PipeStage, sh *Shell) bool {
	for _, p := range pipes {
		if isKatshPipeOp(p.Op) || isStringOp(p.Op) { continue }
		if sh != nil { if _, ok := sh.funcs[p.Op]; ok { continue } }
		return false
	}
	return true
}

// isKatshPipeOp returns true for built-in katsh pipe operators.
func isKatshPipeOp(op string) bool {
	switch op {
	case "select","cols","where","filter","grep","search",
		"sort","orderby","order","limit","head","top",
		"skip","offset","tail","count","unique","distinct",
		"reverse","fmt","format","add","addcol","rename","renamecol":
		return true
	}
	return isStringOp(op)
}


// shellExpand expands $VAR and $() in a string before passing to the OS shell.
func (sh *Shell) shellExpand(s string) string {
	s = sh.expandVars(s)
	s = sh.expandDollarParens(s)
	return s
}

// expandDollarParens replaces $(...) with captured shell output.
// Also handles ?(cmd) — background execution; all ?(cmd) on a line
// are launched in parallel, then collected left-to-right.
// Supports nesting:  x = $(echo $(whoami))
func (sh *Shell) expandDollarParens(s string) string {
	// ── First pass: launch ?(cmd) goroutines in parallel ─────────────────
	type bgSlot struct {
		start, end int    // byte offsets in original s
		job        *BgJob
	}
	var bgSlots []bgSlot
	for i := 0; i < len(s)-1; {
		if s[i] != '?' || s[i+1] != '(' {
			i++
			continue
		}
		depth, end := 0, -1
		for j := i + 2; j < len(s); j++ {
			switch s[j] {
			case '(':  depth++
			case ')':
				if depth == 0 { end = j } else { depth-- }
			}
			if end >= 0 { break }
		}
		if end < 0 { i++; continue }
		inner := strings.TrimSpace(s[i+2 : end])
		inner  = sh.expandVars(inner)
		job    := runBgJob(inner, sh.cwd)
		bgSlots = append(bgSlots, bgSlot{i, end, job})
		i = end + 1 // skip past this ?(...)
	}
	// Collect bg results right-to-left so byte offsets stay valid
	for idx := len(bgSlots) - 1; idx >= 0; idx-- {
		slot := bgSlots[idx]
		out, err := waitBgJob(slot.job)
		if err != nil && sh.errHandlerDepth == 0 {
			PrintError(&ShellError{
				Code:    "E012",
				Kind:    "BackgroundError",
				Message: fmt.Sprintf("background command failed: %v", err),
				Source:  sh.currentSrc,
				Line:    sh.currentLine,
				Col:     slot.start,
				Hint:    "check the command inside ?(...) is valid and in $PATH",
				Fix:     "use $(...) for synchronous capture, or wrap in try/catch",
			})
		}
		s = s[:slot.start] + strings.TrimRight(out, "\n ") + s[slot.end+1:]
	}

	// ── Second pass: expand $(...) inline (iterative, handles nesting) ───
	for {
		start := strings.Index(s, "$(")
		if start < 0 { break }
		depth, end := 0, -1
		for i := start + 2; i < len(s); i++ {
			switch s[i] {
			case '(': depth++
			case ')':
				if depth == 0 { end = i } else { depth-- }
			}
			if end >= 0 { break }
		}
		if end < 0 { break }
		inner := s[start+2 : end]
		inner  = sh.expandDollarParens(inner)
		inner  = sh.expandVars(inner)
		out, _ := RunCaptureShell("", inner, sh.cwd)
		s = s[:start] + strings.TrimRight(out, "\n ") + s[end+1:]
	}
	return s
}

// stripOuterQuotes removes one layer of surrounding quotes.
func stripOuterQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}


// ─────────────────────────────────────────────────────────────────────────────
//  History persistence
// ─────────────────────────────────────────────────────────────────────────────

// historyPaths returns (tempPath, permanentPath) for the current OS.
func historyPaths() (string, string) {
	var tmpDir, permDir string

	if runtime.GOOS == "windows" {
		tmpDir = filepath.Join(os.Getenv("TEMP"), "Katsh")
		permDir = filepath.Join(os.Getenv("APPDATA"), "Katsh")
	} else {
		tmpDir = filepath.Join("/tmp", "Katsh")
		home := os.Getenv("HOME")
		if home == "" {
			home = "/"
		}
		permDir = filepath.Join(home, ".config", "Katsh")
	}

	_ = os.MkdirAll(tmpDir, 0755)
	_ = os.MkdirAll(permDir, 0755)

	return filepath.Join(tmpDir, "history.json"),
		filepath.Join(permDir, "history.json")
}

func (sh *Shell) loadHistory() {
	_, permPath := historyPaths()
	data, err := os.ReadFile(permPath)
	if err != nil {
		return // no history yet — that's fine
	}
	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return
	}
	sh.history = entries
}

func (sh *Shell) saveHistory() {
	tmpPath, permPath := historyPaths()

	// Keep last 10000 entries
	entries := sh.history
	if len(entries) > 10000 {
		entries = entries[len(entries)-10000:]
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return
	}

	_ = os.WriteFile(tmpPath, data, 0600)
	_ = os.WriteFile(permPath, data, 0600)
}
