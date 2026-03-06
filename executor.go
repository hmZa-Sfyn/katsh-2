package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
//  Executor — two modes of running external commands
//
//  CAPTURE mode  (default)
//    Collects output into a Result so katsh can render, pipe, store it.
//    Uses CombinedOutput() — process does NOT see a real TTY.
//
//  PASSTHROUGH mode
//    Connects real stdin/stdout/stderr directly to the child process.
//    Colours, prompts, pagers, interactive tools all work correctly.
//    Output is NOT captured — it flows straight to the terminal.
//
//  Passthrough syntax (all equivalent, pick your style):
//
//    bash! git log --oneline | head -20
//    zsh!  autoload -U compinit && compinit
//    sh!   for f in *.go; do echo $f; done
//    run   git log --oneline | head -20      (uses $SHELL or bash)
//    !     git log --oneline | head -20      (bare ! prefix)
//
//  Capture-into-variable:
//    x = $(git rev-parse HEAD)              (POSIX $() — works anywhere)
//    x = `git branch --show-current`        (backtick — already existed)
//    capture x git log --oneline | head -5  (explicit capture builtin)
//
//  Auto-passthrough (no prefix needed):
//    vim  nvim  nano  emacs  less  man  htop  ssh  mysql  psql  ...
// ─────────────────────────────────────────────────────────────────────────────

// RunExternal executes a command in CAPTURE mode → returns a structured Result.
func RunExternal(command string, args []string, cwd string) (*Result, error) {
	out, _, err := rawCapture(command, args, cwd)
	if err != nil {
		return nil, err
	}
	if autoDetectTable(out) {
		if r := autoParseTable(out); r != nil {
			return r, nil
		}
	}
	return NewText(out), nil
}

// RunPassthrough runs a shell command string with the real TTY attached.
// shell="" picks the user's $SHELL (or bash).
// cmdStr="" launches an interactive session in that shell.
// Returns the exit code. All output goes directly to the terminal.
func RunPassthrough(shell, cmdStr, cwd string) int {
	shellPath := resolveShell(shell)
	if shellPath == "" {
		fmt.Fprintf(os.Stderr, "  katsh: shell %q not found\n", shell)
		return 127
	}

	var cmd *exec.Cmd
	if cmdStr == "" {
		// Interactive session — no -c, just launch the shell
		cmd = exec.Command(shellPath)
	} else {
		cmd = exec.Command(shellPath, "-c", cmdStr)
	}
	cmd.Stdin  = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if cwd != "" { cmd.Dir = cwd }
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}

// RunPassthroughArgs runs a program + args with real TTY (for vim, htop, etc.).
func RunPassthroughArgs(command string, args []string, cwd string) int {
	path, err := exec.LookPath(command)
	if err != nil {
		path = command
	}
	cmd := exec.Command(path, args...)
	cmd.Stdin  = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if cwd != "" { cmd.Dir = cwd }
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode()
		}
		return 1
	}
	return 0
}

// RunCaptureShell runs a shell command string in capture mode.
// Used by $() expansion and backtick subshells that need shell features
// (pipes, globs, redirects) rather than katsh builtins.
// Returns trimmed stdout; stderr is discarded (matches bash $() behaviour).
func RunCaptureShell(shell, cmdStr, cwd string) (string, error) {
	shellPath := resolveShell(shell)
	if shellPath == "" {
		shellPath = "/bin/sh"
	}
	cmd := exec.Command(shellPath, "-c", cmdStr)
	if cwd != "" { cmd.Dir = cwd }
	cmd.Env = os.Environ()
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil // discard — matches $() semantics
	_ = cmd.Run()
	return strings.TrimRight(stdout.String(), "\n"), nil
}

// CaptureOutput runs a command string and returns its stdout as a string.
// Used by the `capture` builtin.
func CaptureOutput(shell, cmdStr, cwd string) (string, int) {
	shellPath := resolveShell(shell)
	if shellPath == "" { shellPath = "/bin/sh" }
	cmd := exec.Command(shellPath, "-c", cmdStr)
	if cwd != "" { cmd.Dir = cwd }
	cmd.Env = os.Environ()
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			code = 1
		}
	}
	return strings.TrimRight(out.String(), "\n"), code
}

// resolveShell returns the full path to the requested shell.
// shell="" → uses $SHELL env var, then tries bash, zsh, sh in order.
func resolveShell(shell string) string {
	if shell == "" {
		if s := os.Getenv("SHELL"); s != "" {
			shell = s // may already be a full path
		}
	}
	// If it looks like a full path, verify it exists
	if strings.HasPrefix(shell, "/") {
		if _, err := os.Stat(shell); err == nil {
			return shell
		}
	}
	// Try LookPath (handles bare names like "bash", "zsh")
	if shell != "" {
		// strip leading slash for LookPath
		name := shell
		if idx := strings.LastIndex(shell, "/"); idx >= 0 {
			name = shell[idx+1:]
		}
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	// Fallback candidates
	for _, p := range []string{"/bin/bash", "/usr/bin/bash", "/bin/zsh", "/usr/bin/zsh", "/bin/sh"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// needsPassthrough returns true for commands that require a real TTY.
var interactiveCommands = map[string]bool{
	// Editors
	"vim": true, "vi": true, "nvim": true, "nano": true, "emacs": true, "micro": true,
	"hx": true, "helix": true, "kak": true,
	// Pagers
	"less": true, "more": true, "most": true,
	// System monitors
	"htop": true, "top": true, "btop": true, "glances": true, "bpytop": true,
	"iotop": true, "iftop": true, "nethogs": true, "nload": true,
	// Remote / network
	"ssh": true, "telnet": true, "ftp": true, "sftp": true, "mosh": true,
	// Databases
	"mysql": true, "psql": true, "sqlite3": true, "mongo": true, "redis-cli": true,
	"pgcli": true, "mycli": true,
	// REPLs
	"python": true, "python3": true, "node": true, "ruby": true, "irb": true,
	"julia": true, "R": true, "ghci": true, "iex": true, "scala": true,
	"lua": true, "php": true, "perl": true,
	// Debuggers
	"gdb": true, "lldb": true, "pdb": true,
	// Multiplexers / terminals
	"tmux": true, "screen": true, "byobu": true, "zellij": true,
	// Other interactive TUIs
	"ranger": true, "nnn": true, "mc": true, "vifm": true,
	"tig": true, "lazygit": true, "gitui": true,
	"cmus": true, "ncmpcpp": true, "mutt": true, "neomutt": true,
	"calcurse": true, "taskwarrior": true, "w3m": true, "lynx": true,
	"fzf": true,
	// man pages
	"man": true,
}

func needsPassthrough(command string) bool {
	return interactiveCommands[strings.ToLower(command)]
}

// ─── rawCapture ───────────────────────────────────────────────────────────────

func rawCapture(command string, args []string, cwd string) (string, int, error) {
	cmd := exec.Command(command, args...)
	if cwd != "" { cmd.Dir = cwd }
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	text := strings.TrimRight(string(out), "\n")
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return text, exitErr.ExitCode(), fmt.Errorf("%s", text)
		}
		return "", 1, fmt.Errorf("%s: %w", command, err)
	}
	return text, 0, nil
}

// ─── Table auto-parsing ───────────────────────────────────────────────────────

func autoDetectTable(s string) bool {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) < 2 { return false }
	words := strings.Fields(lines[0])
	if len(words) < 2 { return false }
	upper := 0
	for _, w := range words {
		if w == strings.ToUpper(w) && len(w) > 1 { upper++ }
	}
	return upper >= len(words)/2+1
}

func autoParseTable(s string) *Result {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) < 2 { return nil }
	cols, starts := detectColumns(lines[0])
	if len(cols) < 2 { return nil }
	var rows []Row
	for _, line := range lines[1:] {
		if strings.TrimSpace(line) == "" { continue }
		rows = append(rows, extractCells(line, cols, starts))
	}
	low := make([]string, len(cols))
	for i, col := range cols { low[i] = strings.ToLower(col) }
	return NewTable(low, rows)
}

func detectColumns(header string) ([]string, []int) {
	var cols []string
	var starts []int
	inWord := false
	start := 0
	for i, ch := range header {
		if ch != ' ' && !inWord {
			inWord = true
			start = i
		} else if ch == ' ' && inWord {
			rest := header[i:]
			if strings.TrimLeft(rest, " ") != "" {
				cols = append(cols, strings.TrimSpace(header[start:i]))
				starts = append(starts, start)
				inWord = false
			}
		}
	}
	if inWord {
		cols = append(cols, strings.TrimSpace(header[start:]))
		starts = append(starts, start)
	}
	return cols, starts
}

func extractCells(line string, cols []string, starts []int) Row {
	row := make(Row, len(cols))
	for i, col := range cols {
		s := starts[i]
		var val string
		if s >= len(line) {
			val = ""
		} else if i == len(cols)-1 {
			val = strings.TrimSpace(line[s:])
		} else {
			end := starts[i+1]
			if end > len(line) { end = len(line) }
			val = strings.TrimSpace(line[s:end])
		}
		row[strings.ToLower(col)] = val
	}
	return row
}
