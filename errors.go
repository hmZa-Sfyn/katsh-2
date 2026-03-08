package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// ─────────────────────────────────────────────────────────────────────────────
//  Rich error reporting — shows the full source line with the bad span
//  underlined using ~~~ and pointed at with a caret + label.
//
//  Example output:
//
//    SyntaxError[E002] expected ':' in while statement
//    ╭─ line 1
//    │  while x < 100 echo $x
//    │         ~~~~~~~
//    │         ^── error here
//    ╰─ 💡 hint: Check for missing colons, brackets, or mismatched quotes
//       fix :  while x < 100: echo $x
// ─────────────────────────────────────────────────────────────────────────────

// ShellError is a structured error with context, hint, and fix suggestion.
type ShellError struct {
	Code     string // e.g. "E001"
	Kind     string // e.g. "SyntaxError", "TypeError", "CommandNotFound"
	Message  string // short message
	Detail   string // longer explanation
	Source   string // the FULL input line that caused the error
	Col      int    // column offset of error start (0-based, -1 = unknown)
	Span     int    // length of the error token in runes (0 = use ^, >0 = use ~~~)
	Line     int    // line number (1-based, 0 = unknown)
	Hint     string // what to check
	Fix      string // suggested fix / example
	Trace    []TraceFrame
}

type TraceFrame struct {
	At  string // e.g. "line 3 in func bogo_sort"
	Src string  // source snippet
}

func (e *ShellError) Error() string { return e.Message }

// PrintError renders a rich, coloured error block to stderr.
func PrintError(err *ShellError) {
	if err == nil { return }

	// ── Header ──────────────────────────────────────────────────────────────
	locLabel := ""
	if err.Line > 0 {
		locLabel = fmt.Sprintf(" at line %d", err.Line)
		if err.Col >= 0 {
			locLabel += fmt.Sprintf(", col %d", err.Col+1)
		}
	}
	fmt.Printf("\n  %s%s[%s]%s%s%s%s %s%s%s\n",
		ansiBold+ansiRed, err.Kind, err.Code, ansiReset,
		ansiGrey, locLabel, ansiReset,
		ansiBold+ansiWhite, err.Message, ansiReset)

	// ── Source block ─────────────────────────────────────────────────────────
	if err.Source != "" {
		// Location badge
		if err.Line > 0 {
			fmt.Printf("  %s╭─ line %d%s\n", ansiGrey, err.Line, ansiReset)
		} else {
			fmt.Printf("  %s╭─%s\n", ansiGrey, ansiReset)
		}

		// The full source line, with the error span highlighted in red+bold
		displayLine := buildHighlightedLine(err.Source, err.Col, err.Span)
		fmt.Printf("  %s│%s  %s\n", ansiGrey, ansiReset, displayLine)

		// Underline row  ───  spaces + ~~~ (or ^)
		if err.Col >= 0 {
			pre := runePrefix(err.Source, err.Col)
			pad := strings.Repeat(" ", pre+3) // +3 for "│  " prefix

			spanLen := err.Span
			if spanLen <= 0 {
				spanLen = tokenLenAt(err.Source, err.Col)
			}
			if spanLen <= 0 { spanLen = 1 }

			underline := strings.Repeat("~", spanLen)
			fmt.Printf("  %s│%s  %s%s%s%s\n",
				ansiGrey, ansiReset,
				pad, ansiRed+ansiBold, underline, ansiReset)

			fmt.Printf("  %s│%s  %s%s^── here%s\n",
				ansiGrey, ansiReset,
				pad, ansiRed, ansiReset)
		}

		fmt.Printf("  %s│%s\n", ansiGrey, ansiReset)
	}

	// ── Detail ───────────────────────────────────────────────────────────────
	if err.Detail != "" {
		fmt.Printf("  %s│%s  %s%s%s\n", ansiGrey, ansiReset, ansiDim+ansiWhite, err.Detail, ansiReset)
		fmt.Printf("  %s│%s\n", ansiGrey, ansiReset)
	}

	// ── Stack trace ──────────────────────────────────────────────────────────
	if len(err.Trace) > 0 {
		fmt.Printf("  %s│  Traceback:%s\n", ansiGrey, ansiReset)
		for i, f := range err.Trace {
			fmt.Printf("  %s│  %s%d: %s%s  %s%s%s\n",
				ansiGrey, ansiDim, i+1, ansiCyan, f.At, ansiGrey, f.Src, ansiReset)
		}
		fmt.Printf("  %s│%s\n", ansiGrey, ansiReset)
	}

	// ── Hint ─────────────────────────────────────────────────────────────────
	if err.Hint != "" {
		fmt.Printf("  %s╰─ 💡 hint:%s %s\n", ansiYellow, ansiReset, err.Hint)
	} else {
		fmt.Printf("  %s╰─%s\n", ansiGrey, ansiReset)
	}

	// ── Fix ──────────────────────────────────────────────────────────────────
	if err.Fix != "" {
		fmt.Printf("  %s   fix :%s  %s%s%s\n\n",
			ansiGreen, ansiReset, ansiDarkCyan, err.Fix, ansiReset)
	} else {
		fmt.Println()
	}
}

// buildHighlightedLine returns the source line with the error span
// printed in bold red, and the rest in normal white.
func buildHighlightedLine(src string, col, span int) string {
	if col < 0 {
		return ansiWhite + src + ansiReset
	}
	runes := []rune(src)
	if col >= len(runes) {
		return ansiWhite + src + ansiReset
	}

	if span <= 0 {
		span = tokenLenAt(src, col)
	}
	if span <= 0 { span = 1 }
	end := col + span
	if end > len(runes) { end = len(runes) }

	before := string(runes[:col])
	bad    := string(runes[col:end])
	after  := string(runes[end:])

	return ansiWhite + before +
		ansiBold + ansiRed + bad + ansiReset +
		ansiWhite + after + ansiReset
}

// runePrefix returns the number of runes before byte index col in s.
// (We store Col as byte offset from the lexer, but display in rune columns.)
func runePrefix(s string, byteCol int) int {
	if byteCol <= 0 { return 0 }
	if byteCol > len(s) { byteCol = len(s) }
	return utf8.RuneCountInString(s[:byteCol])
}

// tokenLenAt returns the length (in runes) of the identifier/word
// starting at byte offset col in s.
func tokenLenAt(s string, col int) int {
	if col < 0 || col >= len(s) { return 1 }
	runes := []rune(s)
	// Find rune index from byte col
	runeIdx := utf8.RuneCountInString(s[:col])
	if runeIdx >= len(runes) { return 1 }
	// Scan forward while it looks like the same token
	start := runeIdx
	ch := runes[start]
	isOp := strings.ContainsRune("=<>!+-*/%&|^~", ch)
	i := start + 1
	for i < len(runes) {
		c := runes[i]
		if isOp {
			if !strings.ContainsRune("=<>!+-*/%&|^~", c) { break }
		} else {
			if c == ' ' || c == '\t' || c == '(' || c == ')' ||
				c == '{' || c == '}' || c == ';' || c == ':' { break }
		}
		i++
	}
	n := i - start
	if n < 1 { n = 1 }
	return n
}

// ── Constructor helpers ──────────────────────────────────────────────────────

func errUnknownCmd(cmd, src string) *ShellError {
	similar := findSimilarCmd(cmd)
	fix := ""
	hint := fmt.Sprintf("%q is not a built-in or executable in $PATH", cmd)
	if similar != "" {
		fix = fmt.Sprintf("did you mean: %s", similar)
	}
	col := strings.Index(src, cmd)
	return &ShellError{
		Code:    "E001",
		Kind:    "CommandNotFound",
		Message: fmt.Sprintf("command not found: %q", cmd),
		Source:  src,
		Col:     col,
		Span:    len(cmd),
		Hint:    hint,
		Fix:     fix,
	}
}

func errSyntax(msg, src string, col int) *ShellError {
	return &ShellError{
		Code:    "E002",
		Kind:    "SyntaxError",
		Message: msg,
		Source:  src,
		Col:     col,
		Span:    0,
		Hint:    "Check for missing colons, brackets, or mismatched quotes",
	}
}

func errType(msg, src string) *ShellError {
	return &ShellError{
		Code:    "E003",
		Kind:    "TypeError",
		Message: msg,
		Source:  src,
		Col:     -1,
	}
}

func errRuntime(msg, src string, trace []TraceFrame) *ShellError {
	return &ShellError{
		Code:    "E004",
		Kind:    "RuntimeError",
		Message: msg,
		Source:  src,
		Col:     -1,
		Trace:   trace,
		Hint:    "Use 'vars' to inspect current variable values",
	}
}

func errUndefined(varName, src string) *ShellError {
	col := strings.Index(src, varName)
	if col < 0 { col = strings.Index(src, "$"+varName) }
	return &ShellError{
		Code:    "E005",
		Kind:    "UndefinedVariable",
		Message: fmt.Sprintf("variable %q is not defined", varName),
		Source:  src,
		Col:     col,
		Span:    len(varName),
		Hint:    "Declare it first with:  " + varName + " = <value>",
		Fix:     varName + " = \"\"",
	}
}

func errDivZero(src string) *ShellError {
	col := strings.Index(src, "/")
	return &ShellError{
		Code:    "E006",
		Kind:    "DivisionByZero",
		Message: "division by zero",
		Source:  src,
		Col:     col,
		Span:    1,
		Hint:    "Check the denominator before dividing",
		Fix:     "if denom != 0: result = num / denom",
	}
}

func errArgCount(funcName string, want, got int, src string) *ShellError {
	col := strings.Index(src, funcName)
	return &ShellError{
		Code:    "E007",
		Kind:    "ArgumentError",
		Message: fmt.Sprintf("%s() expects %d argument(s), got %d", funcName, want, got),
		Source:  src,
		Col:     col,
		Span:    len(funcName),
		Hint:    fmt.Sprintf("func %s takes %d param(s)", funcName, want),
	}
}

func errSimple(msg string) *ShellError {
	return &ShellError{
		Code:    "E000",
		Kind:    "Error",
		Message: msg,
		Col:     -1,
	}
}

// wrapErr wraps a plain Go error as a ShellError for nice display.
func wrapErr(err error, src string) *ShellError {
	if err == nil { return nil }
	msg := err.Error()
	if strings.Contains(msg, "no such file") {
		return &ShellError{
			Code:    "E010",
			Kind:    "FileNotFound",
			Message: msg,
			Source:  src,
			Col:     -1,
			Hint:    "Check the path exists with: stat <path>",
			Fix:     "ls " + extractPath(msg),
		}
	}
	if strings.Contains(msg, "permission denied") {
		return &ShellError{
			Code:    "E011",
			Kind:    "PermissionDenied",
			Message: msg,
			Source:  src,
			Col:     -1,
			Hint:    "You may need elevated privileges",
			Fix:     "chmod +r <file>   or   sudo <command>",
		}
	}
	if strings.Contains(msg, "not found") {
		parts := strings.SplitN(msg, ":", 2)
		cmd := strings.TrimSpace(parts[0])
		return errUnknownCmd(cmd, src)
	}
	// Try to locate a word from the error message inside src
	col := -1
	span := 0
	msgWords := strings.Fields(msg)
	for _, w := range msgWords {
		if len(w) > 3 {
			if idx := strings.Index(src, w); idx >= 0 {
				col = idx
				span = len(w)
				break
			}
		}
	}
	return &ShellError{
		Code:    "E000",
		Kind:    "Error",
		Message: msg,
		Source:  src,
		Col:     col,
		Span:    span,
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func findSimilarCmd(cmd string) string {
	best := ""
	bestDist := 4
	for _, b := range allBuiltinNames() {
		d := editDistance(cmd, b)
		if d < bestDist {
			bestDist = d
			best = b
		}
	}
	return best
}

func editDistance(a, b string) int {
	la, lb := len(a), len(b)
	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
		dp[i][0] = i
	}
	for j := 0; j <= lb; j++ { dp[0][j] = j }
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1]
			} else {
				dp[i][j] = 1 + min3(dp[i-1][j], dp[i][j-1], dp[i-1][j-1])
			}
		}
	}
	return dp[la][lb]
}

func min3(a, b, c int) int {
	if a < b { if a < c { return a }; return c }
	if b < c { return b }
	return c
}

func extractPath(msg string) string {
	parts := strings.Fields(msg)
	for _, p := range parts {
		if strings.HasPrefix(p, "/") || strings.HasPrefix(p, "./") {
			return strings.TrimSuffix(p, ":")
		}
	}
	return ""
}

// ── Position-aware constructors (use lexer line+col) ─────────────────────────

func errSyntaxAt(msg, src, excerpt string, line, col int) *ShellError {
	if line == 0 {
		lx := Lex(src)
		if len(lx.Tokens) > 0 { line = lx.Tokens[0].Line; col = lx.Tokens[0].Col }
	}
	return &ShellError{
		Code:    "E002",
		Kind:    "SyntaxError",
		Message: msg,
		Source:  excerpt,
		Line:    line,
		Col:     col,
		Hint:    "Check for missing colons, brackets, or mismatched quotes",
	}
}

func errTypeAt(msg, src string, line, col int) *ShellError {
	if line == 0 {
		lx := Lex(src)
		if len(lx.Tokens) > 0 { line = lx.Tokens[0].Line }
	}
	return &ShellError{Code:"E003", Kind:"TypeError", Message:msg, Source:src, Line:line, Col:col}
}

func errDivZeroAt(src string, col int) *ShellError {
	lx := Lex(src)
	line := 1
	for _, tok := range lx.Tokens {
		if tok.Text == "/" { line = tok.Line; col = tok.Col; break }
	}
	return &ShellError{
		Code:"E006", Kind:"DivisionByZero", Message:"division by zero",
		Source:src, Line:line, Col:col, Span:1,
		Hint:"Check the denominator before dividing",
		Fix:"if denom != 0: result = num / denom",
	}
}

func errReadonly(name, src string) *ShellError {
	lx := Lex(src)
	line, col := 1, 0
	for _, tok := range lx.Tokens {
		if tok.Text == name { line = tok.Line; col = tok.Col; break }
	}
	return &ShellError{
		Code:"E008", Kind:"ReadonlyError",
		Message:fmt.Sprintf("cannot assign to readonly variable %q", name),
		Source:src, Line:line, Col:col, Span:len(name),
		Hint:"Remove the 'readonly' declaration or use a different name",
	}
}

// errUnhandledThrow creates an E009 for a throw not caught by any try block.
func errUnhandledThrow(msg, src string, line int) *ShellError {
	return &ShellError{
		Code:    "E009",
		Kind:    "UnhandledThrow",
		Message: msg,
		Source:  src,
		Line:    line,
		Col:     -1,
		Hint:    "wrap the throwing code in:  try { ... } catch e { println $e }",
		Fix:     "try { ... } catch e { println \"caught: $e\" }",
	}
}

// errIndexOOB creates an E013 for array/tuple index out of range.
func errIndexOOB(name string, i, length int, src string, line int) *ShellError {
	col := strings.Index(src, name)
	var msg, hint string
	if length == 0 {
		msg  = fmt.Sprintf("%s[%d]: index out of range — array is empty", name, i)
		hint = fmt.Sprintf("append items first:  %s[] = value", name)
	} else {
		msg  = fmt.Sprintf("%s[%d]: index out of range (length %d, valid indices: 0..%d or -%d..-1)", name, i, length, length-1, length)
		hint = fmt.Sprintf("use a negative index to count from the end: %s[-1] = last item", name)
	}
	return &ShellError{
		Code:    "E013",
		Kind:    "IndexError",
		Message: msg,
		Source:  src,
		Line:    line,
		Col:     col,
		Span:    len(name),
		Hint:    hint,
	}
}

// errBackgroundFailed creates an E012 for a failed ?(cmd) background command.
func errBackgroundFailed(cmd string, err error, src string, line, col int) *ShellError {
	return &ShellError{
		Code:    "E012",
		Kind:    "BackgroundError",
		Message: fmt.Sprintf("background command failed: %v", err),
		Detail:  fmt.Sprintf("command was: %s", cmd),
		Source:  src,
		Line:    line,
		Col:     col,
		Hint:    "check the command inside ?(...) is valid and in $PATH",
		Fix:     "use $(...) for synchronous capture, or wrap in try/catch",
	}
}

