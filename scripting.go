package main

// ─────────────────────────────────────────────────────────────────────────────
//  KatSH — scripting.go
//
//  Core scripting engine:
//   • evalScript — top-level dispatch
//   • tryVarAssign / evalRHS / evalExpr
//   • evalCond  (rich operators: in, contains, starts/ends with, is empty…)
//   • Control flow: if/unless/for/while/do/repeat/match
//   • Function def + call (with _return capture)
//   • Subshell / backtick expansion
//   • execBodyLines / execBodyLinesWithGoto
//   • Range expression (inline — no scripting3.go required)
//
//  NEW FEATURES (v2.1):
//   1. loop / forever  — infinite loop  { loop { ...; break } }
//   2. format          — sprintf-style:  format "Hello %s, age %d" $name $age
//   3. pipe select (nth) — x = "a b c" | field 2      → "b"
//   4. |? filter expr  — $arr |? $_ > 5               → filtered array
// ─────────────────────────────────────────────────────────────────────────────

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

// ─── Core types ───────────────────────────────────────────────────────────────

// UserFunc is a user-defined function.
type UserFunc struct {
	Name     string
	Params   []string
	Body     []string
	Exported bool
}

// ScriptArray holds array items internally.
type ScriptArray struct {
	Items []string
}

const arraySep = "\x00" // internal separator for array values

// ─────────────────────────────────────────────────────────────────────────────
//  evalScript — scripting entry point (called before builtins in execLine)
// ─────────────────────────────────────────────────────────────────────────────

func (sh *Shell) evalScript(raw string) (bool, int) {
	raw = strings.TrimSpace(raw)
	if raw == "" { return true, 0 }

	// Comments
	if (strings.HasPrefix(raw, "#") && !strings.HasPrefix(raw, "#=")) ||
		strings.HasPrefix(raw, "//") {
		return true, 0
	}

	// Standalone background: ?(cmd)
	if strings.HasPrefix(raw, "?(") && strings.HasSuffix(raw, ")") {
		inner := sh.expandVars(strings.TrimSpace(raw[2 : len(raw)-1]))
		job   := runBgJob(inner, sh.cwd)
		out, err := waitBgJob(job)
		if err != nil && sh.errHandlerDepth == 0 {
			PrintError(&ShellError{Code: "E012", Kind: "BackgroundError",
				Message: fmt.Sprintf("background command failed: %v", err),
				Source:  raw, Line: sh.currentLine})
		} else if out != "" {
			fmt.Println("\n  " + strings.ReplaceAll(strings.TrimRight(out, "\n"), "\n", "\n  ") + "\n")
		}
		return true, 0
	}

	// ── Delegate to evalScript2 (handles switch/enum/struct/defer/with/
	//   throw/assert/typeof/when/pipe-expr/func-call-capture + scripting3) ──
	if handled, code := sh.evalScript2(raw); handled {
		return true, code
	}

	// Build a cheap lowercase prefix for dispatch (max 20 chars)
	lp := 20
	if len(raw) < lp { lp = len(raw) }
	lower := strings.ToLower(raw[:lp])

	switch {
	case strings.HasPrefix(lower, "if ") || strings.HasPrefix(lower, "if("):
		return true, sh.evalIf(raw)
	case strings.HasPrefix(lower, "unless "):
		return true, sh.evalUnless(raw)
	case strings.HasPrefix(lower, "for "):
		return true, sh.evalFor(raw)
	case strings.HasPrefix(lower, "while "):
		return true, sh.evalWhile(raw)
	case strings.HasPrefix(lower, "do "), strings.HasPrefix(lower, "do{"):
		return true, sh.evalDo(raw)
	case strings.HasPrefix(lower, "repeat "):
		return true, sh.evalRepeat(raw)
	case strings.HasPrefix(lower, "match "):
		return true, sh.evalMatch(raw)
	case strings.HasPrefix(lower, "try "), strings.HasPrefix(lower, "try{"):
		return true, sh.evalTry(raw)
	case strings.HasPrefix(lower, "func "):
		return true, sh.evalFuncDef(raw)

	// ── NEW: loop / forever ── infinite loop, use break to exit ──────────
	case strings.HasPrefix(lower, "loop "), strings.HasPrefix(lower, "loop{"),
		lower == "loop":
		return true, sh.evalLoop(raw)
	case strings.HasPrefix(lower, "forever "), strings.HasPrefix(lower, "forever{"):
		return true, sh.evalLoop(raw)

	// ── NEW: format — sprintf-style output ───────────────────────────────
	//   format "Hello %s, you are %d years old" $name $age
	case strings.HasPrefix(lower, "format "):
		return true, sh.evalFormat(raw)

	case strings.HasPrefix(lower, "print "), strings.HasPrefix(lower, "println "):
		prefix := 6
		if strings.HasPrefix(lower, "println ") { prefix = 8 }
		text := sh.expandBackticks(sh.expandVars(raw[prefix:]))
		text  = expandStringExpr(sh, text)
		fmt.Println("  " + text)
		return true, 0

	case lower == "pass":
		return true, 0

	case strings.HasPrefix(lower, "local "):
		inner := strings.TrimSpace(raw[6:])
		if code, ok := sh.tryVarAssign(inner); ok { return true, code }
	}

	// Top-level control-flow keywords — show helpful errors
	switch lower {
	case "return", "break", "continue":
		PrintError(&ShellError{
			Code: "E002", Kind: "SyntaxError",
			Message: fmt.Sprintf("'%s' can only be used inside a func, loop, or block body", lower),
			Source:  raw, Col: 0, Span: len(lower),
			Hint:    "use 'return' inside a func { } body, 'break'/'continue' inside a loop",
		})
		return true, 1
	}
	if strings.HasPrefix(lower, "return ") {
		PrintError(&ShellError{
			Code: "E002", Kind: "SyntaxError",
			Message: "'return' can only be used inside a func body",
			Source:  raw, Col: 0, Span: 6,
			Hint:    "define a function with:  func myFunc() { return value }",
		})
		return true, 1
	}

	// && / ||
	if code, ok := sh.tryAndOr(raw); ok { return true, code }

	// Variable assignment
	if code, ok := sh.tryVarAssign(raw); ok { return true, code }

	// Increment/decrement
	if code, ok := sh.tryIncrDecr(raw); ok { return true, code }

	// User function call (bare: myfunc arg1 arg2)
	parts := tokenize(raw)
	if len(parts) > 0 {
		if fn, ok := sh.funcs[parts[0]]; ok {
			sh.vars["_return"] = ""
			code := sh.callUserFunc(fn, parts[1:], raw)
			if ret := strings.TrimSpace(sh.vars["_return"]); ret != "" {
				fmt.Println("\n  " + ret + "\n")
			}
			return true, code
		}
	}
	return false, 0
}

// ─── NEW FEATURE 1: loop / forever ───────────────────────────────────────────
//
//   loop { println "hi"; break }
//   loop: println "hi"; break
//   forever { ... }
//
// Identical to  while true { }  but cleaner syntax.
// The loop variable _i counts iterations (0-based).
// Hard-capped at 10 million iterations to prevent accidental infinite hang.

func (sh *Shell) evalLoop(raw string) int {
	// strip keyword
	s := strings.TrimSpace(raw)
	for _, kw := range []string{"forever", "loop"} {
		if strings.HasPrefix(strings.ToLower(s), kw) {
			s = strings.TrimSpace(s[len(kw):])
			break
		}
	}
	body := extractBody(s)
	const maxIter = 10_000_000
	for i := 0; i < maxIter; i++ {
		sh.vars["_i"] = strconv.Itoa(i)
		code := sh.execBodyLines(body)
		if code == codeBreak { break }
		if code == codeContinue { continue }
		if code != 0 { sh.delVar("_i"); return code }
	}
	sh.delVar("_i")
	return 0
}

// ─── NEW FEATURE 2: format ────────────────────────────────────────────────────
//
//   format "Hello %s, you are %d years old" $name $age
//   x = format "%.2f" $pi
//   format "%05d" 42   → "00042"
//
// Supported verbs: %s %d %f %g %x %o %b %05d %.2f %q %%

func (sh *Shell) evalFormat(raw string) int {
	rest := strings.TrimSpace(raw[7:]) // strip "format "

	// Check for assignment:  x = format "..." args
	eqIdx := strings.Index(rest, "=")
	lhsVar := ""
	if eqIdx > 0 {
		lhs := strings.TrimSpace(rest[:eqIdx])
		if isIdent(lhs) {
			lhsVar = lhs
			rest = strings.TrimSpace(rest[eqIdx+1:])
			if strings.HasPrefix(strings.ToLower(rest), "format ") {
				rest = strings.TrimSpace(rest[7:])
			}
		}
	}

	// First token is the format string (must be quoted)
	if rest == "" {
		PrintError(&ShellError{Code: "E007", Kind: "ArgumentError",
			Message: "format requires a format string", Source: raw, Col: 7})
		return 1
	}

	// Find the format string — handle quoted and unquoted
	fmtStr, argsRaw := parseFormatArgs(rest)
	fmtStr = sh.interpolate(fmtStr) // expand $vars inside the format string

	// Collect remaining args
	var args []interface{}
	for _, tok := range tokenize(argsRaw) {
		v := sh.evalExpr(tok)
		// Try numeric first, then string
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			args = append(args, f)
		} else {
			args = append(args, v)
		}
	}

	result := fmt.Sprintf(fmtStr, args...)

	if lhsVar != "" {
		sh.setVar(lhsVar, result)
	} else {
		fmt.Println("  " + result)
	}
	return 0
}

// parseFormatArgs splits a format call rest into (fmtString, argsRaw).
func parseFormatArgs(s string) (string, string) {
	s = strings.TrimSpace(s)
	if len(s) == 0 { return "", "" }
	// Quoted string
	if s[0] == '"' || s[0] == '\'' {
		q := rune(s[0])
		for i := 1; i < len(s); i++ {
			if rune(s[i]) == q && s[i-1] != '\\' {
				return s[1:i], strings.TrimSpace(s[i+1:])
			}
		}
		return s[1:], ""
	}
	// Unquoted — take first word
	idx := strings.IndexByte(s, ' ')
	if idx < 0 { return s, "" }
	return s[:idx], strings.TrimSpace(s[idx+1:])
}

// ─── && / || ─────────────────────────────────────────────────────────────────

func (sh *Shell) tryAndOr(raw string) (int, bool) {
	if idx := findOutside(raw, "&&"); idx >= 0 {
		code := sh.execLine(strings.TrimSpace(raw[:idx]))
		if code == 0 { return sh.execLine(strings.TrimSpace(raw[idx+2:])), true }
		return code, true
	}
	if idx := findOutside(raw, "||"); idx >= 0 {
		code := sh.execLine(strings.TrimSpace(raw[:idx]))
		if code != 0 { return sh.execLine(strings.TrimSpace(raw[idx+2:])), true }
		return 0, true
	}
	return 0, false
}

// findOutside finds needle in s, skipping quoted regions and backticks.
func findOutside(s, needle string) int {
	inQ, inBt := false, false
	qCh := rune(0)
	nl := len(needle)
	for i := 0; i <= len(s)-nl; i++ {
		ch := rune(s[i])
		if inBt { if ch == '`' { inBt = false }; continue }
		if inQ  { if ch == qCh { inQ = false }; continue }
		if ch == '`' { inBt = true; continue }
		if ch == '"' || ch == '\'' { inQ = true; qCh = ch; continue }
		if s[i:i+nl] == needle { return i }
	}
	return -1
}

// ─── Variable assignment ──────────────────────────────────────────────────────

func (sh *Shell) tryVarAssign(raw string) (int, bool) {
	// Compound operators  **=  +=  -=  *=  /=  %=
	for _, op := range []string{"**=", "+=", "-=", "*=", "/=", "%="} {
		if idx := strings.Index(raw, op); idx > 0 {
			name := strings.TrimSpace(raw[:idx])
			if !isIdent(name) { continue }
			rhs := strings.TrimSpace(raw[idx+len(op):])
			curF, _ := strconv.ParseFloat(sh.getVar(name), 64)
			rhsF, err := strconv.ParseFloat(sh.evalExpr(rhs), 64)
			if err != nil {
				_, col, src := FindErrorPos(raw, rhs)
				PrintError(&ShellError{Code: "E003", Kind: "TypeError",
					Message: fmt.Sprintf("cannot apply %s: %q is not a number", op, rhs),
					Source: src, Col: col, Hint: "Use a numeric value on the right-hand side"})
				return 1, true
			}
			var result float64
			switch op {
			case "+=":  result = curF + rhsF
			case "-=":  result = curF - rhsF
			case "*=":  result = curF * rhsF
			case "/=":
				if rhsF == 0 { PrintError(errDivZero(raw)); return 1, true }
				result = curF / rhsF
			case "%=":
				if rhsF == 0 { PrintError(errDivZero(raw)); return 1, true }
				result = math.Mod(curF, rhsF)
			case "**=": result = math.Pow(curF, rhsF)
			}
			sh.setVar(name, fmtNum(result)); return 0, true
		}
	}

	// Array index assign:  arr[N] = val  or  arr[] = val (append)
	if lbIdx := strings.Index(raw, "["); lbIdx > 0 {
		rbIdx := strings.Index(raw, "]")
		if rbIdx > lbIdx && len(raw) > rbIdx+1 && raw[rbIdx+1] == '=' {
			arrName := strings.TrimSpace(raw[:lbIdx])
			if isIdent(arrName) {
				idxStr := strings.TrimSpace(raw[lbIdx+1 : rbIdx])
				val := sh.evalRHS(strings.TrimSpace(raw[rbIdx+2:]), raw)
				if idxStr == "" {
					sh.arrayAppend(arrName, val)
				} else {
					sh.arraySet(arrName, sh.evalExpr(idxStr), val)
				}
				return 0, true
			}
		}
	}

	// Simple assignment  name = value
	idx := strings.Index(raw, "=")
	if idx <= 0 { return 0, false }
	if idx > 0 && (raw[idx-1] == '!' || raw[idx-1] == '<' || raw[idx-1] == '>') { return 0, false }
	if idx+1 < len(raw) && raw[idx+1] == '=' { return 0, false }
	name := strings.TrimSpace(raw[:idx])
	if !isIdent(name) { return 0, false }
	sh.setVar(name, sh.evalRHS(strings.TrimSpace(raw[idx+1:]), raw))
	return 0, true
}

func (sh *Shell) tryIncrDecr(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if strings.HasSuffix(raw, "++") { n := strings.TrimSpace(raw[:len(raw)-2]); if isIdent(n) { sh.incrVar(n, 1); return 0, true } }
	if strings.HasSuffix(raw, "--") { n := strings.TrimSpace(raw[:len(raw)-2]); if isIdent(n) { sh.incrVar(n, -1); return 0, true } }
	if strings.HasPrefix(raw, "++") { n := strings.TrimSpace(raw[2:]); if isIdent(n) { sh.incrVar(n, 1); return 0, true } }
	if strings.HasPrefix(raw, "--") { n := strings.TrimSpace(raw[2:]); if isIdent(n) { sh.incrVar(n, -1); return 0, true } }
	return 0, false
}
func (sh *Shell) incrVar(name string, d float64) {
	v, _ := strconv.ParseFloat(sh.getVar(name), 64)
	sh.setVar(name, fmtNum(v+d))
}

// ─── evalRHS — right-hand side of an assignment ───────────────────────────────

func (sh *Shell) evalRHS(rhs, src string) string {
	rhs = strings.TrimSpace(rhs)
	if rhs == "" { return "" }
	switch strings.ToLower(rhs) {
	case "null", "nil", "none", "undefined": return ""
	}

	// Background command: ?(cmd)
	if strings.HasPrefix(rhs, "?(") && strings.HasSuffix(rhs, ")") {
		inner := sh.expandVars(strings.TrimSpace(rhs[2 : len(rhs)-1]))
		job   := runBgJob(inner, sh.cwd)
		out, err := waitBgJob(job)
		if err != nil && sh.errHandlerDepth == 0 {
			PrintError(&ShellError{Code: "E012", Kind: "BackgroundError",
				Message: fmt.Sprintf("background command failed: %v", err),
				Source: src, Line: sh.currentLine,
				Col:    strings.Index(src, "?("),
				Hint:   "check the command inside ?(...) is valid and in $PATH"})
		}
		return strings.TrimRight(out, "\n ")
	}

	// Ternary: cond ? then : else
	if idx := findTernaryQ(rhs); idx >= 0 {
		rest := rhs[idx+1:]
		if colonIdx := findTernaryColon(rest); colonIdx >= 0 {
			condPart := strings.TrimSpace(rhs[:idx])
			thenPart := strings.TrimSpace(rest[:colonIdx])
			elsePart := strings.TrimSpace(rest[colonIdx+1:])
			if sh.evalCond(condPart) { return sh.evalRHS(thenPart, src) }
			return sh.evalRHS(elsePart, src)
		}
	}

	// Range expression: range 5 / range 1 10 / range 1 10 2
	if v, ok := sh.evalRangeExprEval(rhs); ok { return v }

	// ?? null-coalescing
	if qIdx := strings.Index(rhs, "??"); qIdx > 0 {
		primary  := strings.TrimSpace(rhs[:qIdx])
		fallback := strings.TrimSpace(rhs[qIdx+2:])
		val := sh.evalExpr(primary)
		if val == "" || val == "null" || val == "nil" || val == "undefined" {
			return sh.evalExpr(stripQuotes(fallback))
		}
		return val
	}

	if strings.HasPrefix(strings.ToLower(rhs), "if ") { return sh.evalInlineIf(rhs, src) }

	// ── NEW FEATURE 3: field N — select Nth whitespace-separated token ────
	//   x = "one two three" | field 2   → "two"
	//   (also works as standalone pipe op in shell pipeline)
	if strings.Contains(rhs, "| field ") || strings.Contains(rhs, "|field ") {
		if v, ok := sh.evalFieldPipe(rhs); ok { return v }
	}

	// ── NEW FEATURE 4: |? inline filter ──────────────────────────────────
	//   filtered = $arr |? $_ > 5
	//   result   = $names |? $_ starts with "A"
	if strings.Contains(rhs, "|?") {
		if v, ok := sh.evalFilterPipe(rhs); ok { return v }
	}

	// Data type literals
	if v, ok := sh.evalDataTypeLiteral(rhs); ok { return v }

	// Array literal
	if strings.HasPrefix(rhs, "[") && strings.HasSuffix(rhs, "]") {
		return sh.makeArray(sh.parseArrayLiteral(rhs))
	}

	// Quoted strings
	if strings.HasPrefix(rhs, "`") && strings.HasSuffix(rhs, "`") { return sh.runSubshell(rhs[1:len(rhs)-1]) }
	if strings.HasPrefix(rhs, `"`) && strings.HasSuffix(rhs, `"`) { return sh.interpolate(rhs[1:len(rhs)-1]) }
	if strings.HasPrefix(rhs, "'") && strings.HasSuffix(rhs, "'") { return rhs[1 : len(rhs)-1] }

	return sh.evalExpr(rhs)
}

// ─── NEW FEATURE 3: field N pipe ─────────────────────────────────────────────
//
//   x = "hello world foo" | field 2   → "world"
//   x = "a:b:c" | field 2 ":"         → "b"  (custom delimiter)

func (sh *Shell) evalFieldPipe(rhs string) (string, bool) {
	// Split on  | field  or  |field
	sep := "| field "
	idx := strings.Index(rhs, sep)
	if idx < 0 {
		sep = "|field "
		idx = strings.Index(rhs, sep)
	}
	if idx < 0 { return "", false }

	left  := strings.TrimSpace(rhs[:idx])
	right := strings.TrimSpace(rhs[idx+len(sep):])

	src := sh.evalExpr(left)

	// Parse:  N  or  N "delim"
	parts := strings.Fields(right)
	if len(parts) == 0 { return "", false }
	n, err := strconv.Atoi(parts[0])
	if err != nil { return "", false }

	delim := ""
	if len(parts) >= 2 { delim = stripQuotes(parts[1]) }

	var fields []string
	if delim != "" {
		fields = strings.Split(src, delim)
	} else {
		fields = strings.Fields(src)
	}

	if n < 1 || n > len(fields) { return "", true }
	return fields[n-1], true
}

// ─── NEW FEATURE 4: |? inline array filter ───────────────────────────────────
//
//   filtered = $nums |? $_ > 5
//   names    = $people |? $_ starts with "A"
//   evens    = $nums |? $_ % 2 == 0

func (sh *Shell) evalFilterPipe(rhs string) (string, bool) {
	idx := strings.Index(rhs, "|?")
	if idx < 0 { return "", false }

	left  := strings.TrimSpace(rhs[:idx])
	cond  := strings.TrimSpace(rhs[idx+2:])
	if cond == "" { return "", false }

	// Evaluate LHS — must resolve to an array
	left = sh.expandVars(left)
	var items []string
	if strings.HasPrefix(left, "$") {
		items = sh.arrayItems(left[1:])
	} else {
		val := sh.evalExpr(left)
		if strings.HasPrefix(val, "[") {
			items = sh.arrayItems2(val)
		} else {
			items = strings.Fields(val)
		}
	}

	var filtered []string
	savedItem := sh.vars["_"]
	savedUnder := sh.vars["_item"]
	for _, it := range items {
		sh.vars["_"]     = it
		sh.vars["_item"] = it
		if sh.evalCond(cond) { filtered = append(filtered, it) }
	}
	sh.vars["_"]     = savedItem
	sh.vars["_item"] = savedUnder

	return sh.makeArray(filtered), true
}

// ─── findTernaryQ / findTernaryColon ─────────────────────────────────────────

func findTernaryQ(s string) int {
	depth, inQ := 0, false
	qCh := rune(0)
	for i, ch := range s {
		if inQ { if ch == qCh { inQ = false }; continue }
		if ch == '"' || ch == '\'' { inQ = true; qCh = ch; continue }
		if ch == '(' || ch == '[' || ch == '{' { depth++; continue }
		if ch == ')' || ch == ']' || ch == '}' { depth--; continue }
		if ch == '?' && depth == 0 { return i }
	}
	return -1
}

func findTernaryColon(s string) int {
	depth, inQ, nested := 0, false, 0
	qCh := rune(0)
	for i, ch := range s {
		if inQ { if ch == qCh { inQ = false }; continue }
		if ch == '"' || ch == '\'' { inQ = true; qCh = ch; continue }
		if ch == '(' || ch == '[' || ch == '{' { depth++; continue }
		if ch == ')' || ch == ']' || ch == '}' { depth--; continue }
		if ch == '?' && depth == 0 { nested++; continue }
		if ch == ':' && depth == 0 {
			if nested > 0 { nested--; continue }
			return i
		}
	}
	return -1
}

// ─── interpolate ──────────────────────────────────────────────────────────────

func (sh *Shell) interpolate(s string) string {
	if strings.Contains(s, "$(") { s = sh.expandDollarParens(s) }
	s = sh.expandBackticks(s)
	return os.Expand(s, func(key string) string {
		if strings.HasPrefix(key, "#") { return strconv.Itoa(len(sh.getVar(key[1:]))) }
		if strings.Contains(key, ":-") {
			p := strings.SplitN(key, ":-", 2)
			if v := sh.getVar(p[0]); v != "" { return v }
			return p[1]
		}
		if strings.Contains(key, ":+") {
			p := strings.SplitN(key, ":+", 2)
			if v := sh.getVar(p[0]); v != "" { return p[1] }
			return ""
		}
		return sh.getVar(key)
	})
}

// ─── Arrays ───────────────────────────────────────────────────────────────────

func (sh *Shell) makeArray(items []string) string {
	return "[" + strings.Join(items, arraySep) + "]"
}

func (sh *Shell) parseArrayLiteral(s string) []string {
	inner := strings.TrimSpace(s[1 : len(s)-1])
	if inner == "" { return nil }
	var out []string
	for _, item := range strings.Split(inner, ",") {
		out = append(out, sh.evalExpr(stripQuotes(strings.TrimSpace(item))))
	}
	return out
}

func (sh *Shell) arrayItems(name string) []string {
	raw := sh.vars[name]
	if !strings.HasPrefix(raw, "[") { return strings.Fields(raw) }
	content := raw[1 : len(raw)-1]
	if content == "" { return nil }
	return strings.Split(content, arraySep)
}

// arrayItems2 extracts items from an inline array string (not a variable lookup).
func (sh *Shell) arrayItems2(raw string) []string {
	if raw == "" { return nil }
	if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
		inner := raw[1 : len(raw)-1]
		if strings.TrimSpace(inner) == "" { return nil }
		if strings.Contains(inner, arraySep) {
			var out []string
			for _, p := range strings.Split(inner, arraySep) {
				if t := strings.TrimSpace(p); t != "" { out = append(out, t) }
			}
			return out
		}
		return []string{strings.TrimSpace(inner)}
	}
	return strings.Fields(raw)
}

func (sh *Shell) arrayGet(name, idx string) string {
	raw := sh.vars[name]
	if strings.HasPrefix(raw, mapPfx) { return mapGet(raw, idx) }
	if strings.HasPrefix(raw, tupPfx) {
		n, err := strconv.Atoi(idx)
		if err != nil {
			PrintError(&ShellError{Code: "E013", Kind: "IndexError",
				Message: fmt.Sprintf("tuple index must be an integer, got %q", idx),
				Source: sh.currentSrc, Line: sh.currentLine,
				Hint: "tuple indices are 0-based integers: t[0], t[1], ..."})
			return ""
		}
		return tupleGet(raw, n)
	}
	items := sh.arrayItems(name)
	if idx == "len" || idx == "#" { return strconv.Itoa(len(items)) }
	i, err := strconv.Atoi(idx)
	if err != nil {
		PrintError(&ShellError{Code: "E013", Kind: "IndexError",
			Message: fmt.Sprintf("array index must be an integer, got %q in %s[%s]", idx, name, idx),
			Source: sh.currentSrc, Line: sh.currentLine,
			Col: strings.Index(sh.currentSrc, name), Span: len(name),
			Hint: fmt.Sprintf("use an integer index like %s[0], or arr_len to get the length", name)})
		return ""
	}
	if i < 0 { i = len(items) + i }
	if i < 0 || i >= len(items) {
		if len(items) == 0 {
			PrintError(&ShellError{Code: "E013", Kind: "IndexError",
				Message: fmt.Sprintf("%s[%d]: array is empty", name, i),
				Source: sh.currentSrc, Line: sh.currentLine,
				Col: strings.Index(sh.currentSrc, name), Span: len(name),
				Hint: fmt.Sprintf("'%s' has no elements — append with:  %s[] = value", name, name)})
		} else {
			PrintError(&ShellError{Code: "E013", Kind: "IndexError",
				Message: fmt.Sprintf("%s[%d]: out of range (has %d element(s), valid: 0..%d)", name, i, len(items), len(items)-1),
				Source: sh.currentSrc, Line: sh.currentLine,
				Col: strings.Index(sh.currentSrc, name), Span: len(name),
				Hint: fmt.Sprintf("use negative index to count from end: %s[-1] is last", name)})
		}
		return ""
	}
	return items[i]
}

func (sh *Shell) arraySet(name, idx, val string) {
	items := sh.arrayItems(name)
	i, err := strconv.Atoi(idx)
	if err != nil { return }
	for len(items) <= i { items = append(items, "") }
	items[i] = val
	sh.vars[name] = sh.makeArray(items)
}

func (sh *Shell) arrayAppend(name, val string) {
	sh.vars[name] = sh.makeArray(append(sh.arrayItems(name), val))
}

// ─── evalExpr — expression evaluator ─────────────────────────────────────────
//
// Performance notes:
//   • Early returns on empty string
//   • $( and ?( expansion only when the substring is present (fast check)
//   • Backtick only when ` present
//   • Type-cast builtins (int/float/str/bool/len/abs/min/max) use HasPrefix only
//   • Arithmetic: single pass, last operator wins for correct precedence

func (sh *Shell) evalExpr(expr string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" { return "" }

	// ── Expand $(...) / ?(...) before any other evaluation ────────────────
	if strings.Contains(expr, "$(") || strings.Contains(expr, "?(") {
		expr = sh.expandDollarParens(sh.expandVars(expr))
		expr = strings.TrimSpace(expr)
		if expr == "" { return "" }
	}

	// ── Backtick subshell ─────────────────────────────────────────────────
	if strings.Contains(expr, "`") {
		expr = sh.expandBackticks(expr)
		expr = strings.TrimSpace(expr)
	}

	// ── Ternary  cond ? true : false ──────────────────────────────────────
	if qIdx := findOutside(expr, " ? "); qIdx >= 0 {
		if cIdx := findOutside(expr, " : "); cIdx > qIdx {
			condPart  := strings.TrimSpace(expr[:qIdx])
			truePart  := strings.TrimSpace(expr[qIdx+3 : cIdx])
			falsePart := strings.TrimSpace(expr[cIdx+3:])
			if sh.evalCond(condPart) { return sh.evalExpr(truePart) }
			return sh.evalExpr(falsePart)
		}
	}

	// ── Type-cast builtins ────────────────────────────────────────────────
	// Using a single pass through prefix checks — only allocates when matched
	if len(expr) > 4 {
		switch {
		case strings.HasPrefix(expr, "int(") && expr[len(expr)-1] == ')':
			f, _ := strconv.ParseFloat(sh.evalExpr(expr[4:len(expr)-1]), 64)
			return strconv.Itoa(int(f))
		case strings.HasPrefix(expr, "str(") && expr[len(expr)-1] == ')':
			return sh.evalExpr(expr[4 : len(expr)-1])
		case strings.HasPrefix(expr, "abs(") && expr[len(expr)-1] == ')':
			f, _ := strconv.ParseFloat(sh.evalExpr(expr[4:len(expr)-1]), 64)
			return fmtNum(math.Abs(f))
		case strings.HasPrefix(expr, "len(") && expr[len(expr)-1] == ')':
			return strconv.Itoa(len([]rune(sh.evalExpr(expr[4 : len(expr)-1]))))
		case strings.HasPrefix(expr, "bool(") && expr[len(expr)-1] == ')':
			inner := strings.ToLower(sh.evalExpr(expr[5 : len(expr)-1]))
			if inner == "true" || inner == "1" || inner == "yes" { return "true" }
			return "false"
		case strings.HasPrefix(expr, "float(") && expr[len(expr)-1] == ')':
			f, _ := strconv.ParseFloat(sh.evalExpr(expr[6:len(expr)-1]), 64)
			return strconv.FormatFloat(f, 'f', -1, 64)
		case strings.HasPrefix(expr, "min(") && expr[len(expr)-1] == ')':
			if parts := strings.SplitN(expr[4:len(expr)-1], ",", 2); len(parts) == 2 {
				a, _ := strconv.ParseFloat(sh.evalExpr(parts[0]), 64)
				b, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64)
				if a < b { return fmtNum(a) }; return fmtNum(b)
			}
		case strings.HasPrefix(expr, "max(") && expr[len(expr)-1] == ')':
			if parts := strings.SplitN(expr[4:len(expr)-1], ",", 2); len(parts) == 2 {
				a, _ := strconv.ParseFloat(sh.evalExpr(parts[0]), 64)
				b, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64)
				if a > b { return fmtNum(a) }; return fmtNum(b)
			}
		}
	}

	// ── ${VAR...} ─────────────────────────────────────────────────────────
	if strings.HasPrefix(expr, "${") && strings.HasSuffix(expr, "}") {
		return sh.interpolate(expr)
	}

	// ── $VAR or $arr[N] ───────────────────────────────────────────────────
	if strings.HasPrefix(expr, "$") {
		name := expr[1:]
		if lbIdx := strings.IndexByte(name, '['); lbIdx >= 0 {
			arrName := name[:lbIdx]
			idxStr  := strings.TrimSuffix(name[lbIdx+1:], "]")
			return sh.arrayGet(arrName, sh.evalExpr(idxStr))
		}
		return sh.getVar(name)
	}

	// ── arr[N] or arr.len (no $ prefix) ──────────────────────────────────
	if lbIdx := strings.IndexByte(expr, '['); lbIdx > 0 && strings.HasSuffix(expr, "]") {
		arrName := expr[:lbIdx]
		if isIdent(arrName) {
			return sh.arrayGet(arrName, sh.evalExpr(expr[lbIdx+1:len(expr)-1]))
		}
	}
	if strings.HasSuffix(expr, ".len") {
		arrName := expr[:len(expr)-4]
		if isIdent(arrName) { return sh.arrayGet(arrName, "len") }
	}

	// ── Bare identifier ───────────────────────────────────────────────────
	if isIdent(expr) {
		if v, ok := sh.vars[expr]; ok { return v }
	}

	// ── Backtick literal ──────────────────────────────────────────────────
	if strings.HasPrefix(expr, "`") && strings.HasSuffix(expr, "`") {
		return sh.runSubshell(expr[1 : len(expr)-1])
	}

	// ── String concatenation / quoted expressions ─────────────────────────
	if strings.ContainsAny(expr, `"'`) {
		return expandStringExpr(sh, expr)
	}

	// ── Power ─────────────────────────────────────────────────────────────
	if idx := strings.LastIndex(expr, " ** "); idx >= 0 {
		base, _ := strconv.ParseFloat(sh.evalExpr(expr[:idx]), 64)
		exp, _  := strconv.ParseFloat(sh.evalExpr(expr[idx+4:]), 64)
		return fmtNum(math.Pow(base, exp))
	}

	// ── String repeat: "ha" * 3 → "hahaha" ───────────────────────────────
	if idx := strings.LastIndex(expr, " * "); idx >= 0 {
		lhs := strings.TrimSpace(expr[:idx])
		rhs := strings.TrimSpace(expr[idx+3:])
		if strings.HasPrefix(lhs, `"`) || strings.HasPrefix(lhs, "'") {
			n, _ := strconv.Atoi(sh.evalExpr(rhs))
			if n > 0 { return strings.Repeat(stripQuotes(sh.evalExpr(lhs)), n) }
		}
	}

	// ── Arithmetic ────────────────────────────────────────────────────────
	if r, ok := tryArith(sh, expr); ok { return r }

	return expr
}

// tryArith tries to evaluate expr as  LHS op RHS  arithmetic.
// Uses LastIndex so operator precedence follows conventional order (+/- before */%):
// The operators are tried in reverse precedence — last matched wins.
func tryArith(sh *Shell, expr string) (string, bool) {
	for _, op := range []string{"+", "-", "*", "/", "%"} {
		idx := strings.LastIndex(expr, " "+op+" ")
		if idx < 0 { continue }
		lv, _ := strconv.ParseFloat(sh.evalExpr(expr[:idx]), 64)
		rv, _ := strconv.ParseFloat(sh.evalExpr(expr[idx+3:]), 64)
		var r float64
		switch op {
		case "+": r = lv + rv
		case "-": r = lv - rv
		case "*": r = lv * rv
		case "/":
			if rv == 0 {
				PrintError(&ShellError{Code: "E006", Kind: "DivisionByZero",
					Message: fmt.Sprintf("cannot divide %s by zero", fmtNum(lv)),
					Source: sh.currentSrc, Line: sh.currentLine,
					Col: strings.IndexByte(sh.currentSrc, '/'), Span: 1,
					Hint: "check the divisor is non-zero before dividing",
					Fix:  "if denom != 0 { result = num / denom }"})
				return "0", true
			}
			r = lv / rv
		case "%":
			if rv == 0 {
				PrintError(&ShellError{Code: "E006", Kind: "DivisionByZero",
					Message: "modulo by zero", Source: sh.currentSrc, Line: sh.currentLine,
					Col: strings.IndexByte(sh.currentSrc, '%'), Span: 1,
					Hint: "the right operand of % must not be zero"})
				return "0", true
			}
			r = math.Mod(lv, rv)
		}
		return fmtNum(r), true
	}
	return "", false
}

// ─── String expression helpers ───────────────────────────────────────────────

func expandStringExpr(sh *Shell, expr string) string {
	var sb strings.Builder
	for _, p := range splitOnDot(expr) {
		sb.WriteString(evalStringPart(sh, strings.TrimSpace(p)))
	}
	return sb.String()
}

func evalStringPart(sh *Shell, p string) string {
	if xIdx := strings.LastIndex(p, `"x`); xIdx > 0 {
		n := 0; fmt.Sscanf(p[xIdx+2:], "%d", &n)
		return strings.Repeat(stripQuotes(p[:xIdx+1]), n)
	}
	if strings.HasPrefix(p, "`") && strings.HasSuffix(p, "`") { return strings.TrimSpace(sh.runSubshell(p[1:len(p)-1])) }
	if strings.HasPrefix(p, "$") { return sh.getVar(p[1:]) }
	if strings.HasPrefix(p, `"`) || strings.HasPrefix(p, "'") { return sh.interpolate(stripQuotes(p)) }
	return sh.expandVars(p)
}

func splitOnDot(s string) []string {
	var parts []string
	var cur strings.Builder
	inQ, inBt := false, false
	qCh := rune(0)
	for _, ch := range s {
		switch {
		case inBt:
			cur.WriteRune(ch)
			if ch == '`' { inBt = false }
		case inQ:
			cur.WriteRune(ch)
			if ch == qCh { inQ = false }
		case ch == '`':
			inBt = true; cur.WriteRune(ch)
		case ch == '"' || ch == '\'':
			inQ = true; qCh = ch; cur.WriteRune(ch)
		case ch == '.':
			parts = append(parts, cur.String()); cur.Reset()
		default:
			cur.WriteRune(ch)
		}
	}
	if cur.Len() > 0 { parts = append(parts, cur.String()) }
	return parts
}

func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		return s[1 : len(s)-1]
	}
	return s
}

// ─── evalCond — condition evaluator ──────────────────────────────────────────

func (sh *Shell) evalCond(cond string) bool {
	cond  = strings.TrimSpace(cond)
	lower := strings.ToLower(cond)

	// not / !
	if strings.HasPrefix(lower, "not ") { return !sh.evalCond(cond[4:]) }
	if strings.HasPrefix(cond, "!") && !strings.HasPrefix(cond, "!=") { return !sh.evalCond(cond[1:]) }

	// Logical connectives  (and/or short-circuit)
	if idx := findOutside(cond, " and "); idx >= 0 { return sh.evalCond(cond[:idx]) && sh.evalCond(cond[idx+5:]) }
	if idx := findOutside(cond, " or ");  idx >= 0 { return sh.evalCond(cond[:idx]) || sh.evalCond(cond[idx+4:]) }
	if idx := findOutside(cond, " && ");  idx >= 0 { return sh.evalCond(cond[:idx]) && sh.evalCond(cond[idx+4:]) }
	if idx := findOutside(cond, " || ");  idx >= 0 { return sh.evalCond(cond[:idx]) || sh.evalCond(cond[idx+4:]) }

	// "not in" / "in"
	if idx := findOutside(lower, " not in "); idx >= 0 {
		lv  := sh.evalExpr(cond[:idx])
		col := strings.TrimSpace(cond[idx+8:])
		return !sh.evalInCond(lv, col)
	}
	if idx := findOutside(lower, " in "); idx >= 0 {
		lv  := sh.evalExpr(cond[:idx])
		col := strings.TrimSpace(cond[idx+4:])
		return sh.evalInCond(lv, col)
	}

	// "starts with" / "ends with" / "contains"
	if idx := findOutside(lower, " starts with "); idx >= 0 {
		return strings.HasPrefix(sh.evalExpr(cond[:idx]), stripQuotes(sh.evalExpr(strings.TrimSpace(cond[idx+13:]))))
	}
	if idx := findOutside(lower, " ends with "); idx >= 0 {
		return strings.HasSuffix(sh.evalExpr(cond[:idx]), stripQuotes(sh.evalExpr(strings.TrimSpace(cond[idx+11:]))))
	}
	if idx := findOutside(lower, " contains "); idx >= 0 {
		return strings.Contains(sh.evalExpr(cond[:idx]), stripQuotes(sh.evalExpr(strings.TrimSpace(cond[idx+10:]))))
	}

	// "is empty" / "is not empty"
	if strings.HasSuffix(lower, " is empty") {
		return strings.TrimSpace(sh.evalExpr(strings.TrimSpace(cond[:len(cond)-9]))) == ""
	}
	if strings.HasSuffix(lower, " is not empty") {
		return strings.TrimSpace(sh.evalExpr(strings.TrimSpace(cond[:len(cond)-13]))) != ""
	}

	// Comparison operators — ordered longest-first to avoid prefix conflicts
	for _, op := range []string{"!=", ">=", "<=", "==", "!~", "~=", ">", "<", "~"} {
		idx := strings.Index(cond, op)
		if idx <= 0 { continue }
		lv := strings.TrimSpace(sh.evalExpr(cond[:idx]))
		rv := strings.TrimSpace(stripQuotes(sh.evalExpr(cond[idx+len(op):])))
		lf, lNum := parseNum(lv)
		rf, rNum := parseNum(rv)
		switch op {
		case "==": return lv == rv
		case "!=": return lv != rv
		case "~", "~=": return strings.Contains(lv, rv)
		case "!~": return !strings.Contains(lv, rv)
		case ">":  if lNum && rNum { return lf > rf }; return lv > rv
		case "<":  if lNum && rNum { return lf < rf }; return lv < rv
		case ">=": if lNum && rNum { return lf >= rf }; return lv >= rv
		case "<=": if lNum && rNum { return lf <= rf }; return lv <= rv
		}
	}

	// Test flags: -z -n -f -d -e -r -s -w
	if strings.HasPrefix(cond, "-") {
		parts := strings.Fields(cond)
		if len(parts) == 2 { return evalTestFlag(parts[0], sh.evalExpr(parts[1])) }
	}

	v := strings.ToLower(strings.TrimSpace(sh.evalExpr(cond)))
	switch v {
	case "true", "1", "yes":                              return true
	case "false", "0", "no", "", "null", "nil", "none", "undefined": return false
	}
	return v != ""
}

func (sh *Shell) evalInCond(lv, expr string) bool {
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(expr, "[") && strings.HasSuffix(expr, "]") {
		for _, it := range sh.parseArrayLiteral(expr) { if it == lv { return true } }
		return false
	}
	val := sh.evalExpr(expr)
	if strings.HasPrefix(val, "[") {
		name := expr
		if strings.HasPrefix(expr, "$") { name = expr[1:] }
		for _, it := range sh.arrayItems(name) { if it == lv { return true } }
		return false
	}
	for _, w := range strings.Fields(val) { if w == lv { return true } }
	return false
}

func evalTestFlag(flag, val string) bool {
	switch flag {
	case "-z": return val == ""
	case "-n": return val != ""
	case "-f": info, err := os.Stat(val); return err == nil && !info.IsDir()
	case "-d": info, err := os.Stat(val); return err == nil && info.IsDir()
	case "-e": _, err := os.Stat(val); return err == nil
	case "-r": f, err := os.Open(val); if err != nil { return false }; f.Close(); return true
	case "-s": info, err := os.Stat(val); return err == nil && info.Size() > 0
	case "-w": f, err := os.OpenFile(val, os.O_WRONLY, 0); if err != nil { return false }; f.Close(); return true
	}
	return false
}

// ─── if / elif / else ─────────────────────────────────────────────────────────

func (sh *Shell) evalIf(raw string) int {
	rest := strings.TrimSpace(raw[3:])
	type branch struct{ cond, body string }
	var branches []branch
	var elsebody string

	for _, cl := range splitSemicolon(rest) {
		cl = strings.TrimSpace(cl)
		low := strings.ToLower(cl)
		switch {
		case strings.HasPrefix(low, "elif ") || strings.HasPrefix(low, "else if "):
			off := 5
			if strings.HasPrefix(low, "else if ") { off = 8 }
			cond, body := splitColon(cl[off:])
			branches = append(branches, branch{cond, extractBody(body)})
		case strings.HasPrefix(low, "else:"), strings.HasPrefix(low, "else "),
			low == "else", strings.HasPrefix(low, "else{"):
			after := strings.TrimSpace(cl[4:])
			after = strings.TrimPrefix(after, ":")
			elsebody = extractBody(strings.TrimSpace(after))
		default:
			cond, body := splitColon(cl)
			branches = append(branches, branch{cond, extractBody(body)})
		}
	}
	for _, br := range branches {
		if sh.evalCond(br.cond) { return sh.execBodyLines(br.body) }
	}
	if elsebody != "" { return sh.execBodyLines(elsebody) }
	return 0
}

func (sh *Shell) evalInlineIf(raw, src string) string {
	rest := strings.TrimSpace(raw[3:])
	var cond, thenVal, elseVal string
	for _, p := range splitSemicolon(rest) {
		p = strings.TrimSpace(p)
		low := strings.ToLower(p)
		if strings.HasPrefix(low, "else:") || strings.HasPrefix(low, "else ") {
			elseVal = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(p[4:]), ":"))
		} else {
			c, body := splitColon(p)
			if cond == "" { cond = c; thenVal = body }
		}
	}
	if sh.evalCond(cond) { return sh.evalRHS(thenVal, src) }
	return sh.evalRHS(elseVal, src)
}

// ─── unless ───────────────────────────────────────────────────────────────────

func (sh *Shell) evalUnless(raw string) int {
	rest := strings.TrimSpace(raw[7:])
	cond, body := splitColon(rest)
	elsebody := ""
	if idx := strings.Index(body, "; else"); idx >= 0 {
		parts := splitSemicolon(body)
		body = strings.TrimSpace(parts[0])
		if len(parts) > 1 {
			after := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(parts[1]), "else:"), "else ")
			elsebody = extractBody(strings.TrimSpace(after))
		}
	}
	if !sh.evalCond(cond) { return sh.execBodyLines(extractBody(body)) }
	if elsebody != "" { return sh.execBodyLines(elsebody) }
	return 0
}

// ─── match / case ─────────────────────────────────────────────────────────────

func (sh *Shell) evalMatch(raw string) int {
	rest := strings.TrimSpace(raw[6:])
	braceIdx := strings.Index(rest, "{")
	if braceIdx < 0 {
		colonIdx := colonOutsideBraces(rest)
		if colonIdx < 0 { PrintError(errSyntax("expected '{' in match", raw, 6)); return 1 }
		return sh.runMatch(strings.TrimSpace(rest[:colonIdx]), rest[colonIdx+1:], raw)
	}
	return sh.runMatch(strings.TrimSpace(rest[:braceIdx]), rest[braceIdx:], raw)
}

func (sh *Shell) runMatch(subject, body, raw string) int {
	val   := sh.evalExpr(subject)
	inner := extractBody(body)
	if strings.HasPrefix(inner, "{") { inner = inner[1 : len(inner)-1] }

	for _, cl := range splitSemicolon(inner) {
		cl = strings.TrimSpace(cl)
		low := strings.ToLower(cl)
		if strings.HasPrefix(low, "default") {
			after := strings.TrimPrefix(strings.TrimSpace(cl[7:]), ":")
			return sh.execBodyLines(extractBody(strings.TrimSpace(after)))
		}
		if !strings.HasPrefix(low, "case ") { continue }
		caseExpr := cl[5:]
		ci := colonOutsideBraces(caseExpr)
		if ci < 0 { continue }
		pattern  := strings.TrimSpace(caseExpr[:ci])
		casebody := strings.TrimSpace(caseExpr[ci+1:])
		if sh.matchPattern(val, pattern) { return sh.execBodyLines(extractBody(casebody)) }
	}
	return 0
}

func (sh *Shell) matchPattern(val, pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	if strings.Contains(pattern, "|") {
		for _, alt := range strings.Split(pattern, "|") {
			if sh.matchPattern(val, strings.TrimSpace(alt)) { return true }
		}
		return false
	}
	if pattern == "*" || pattern == "_" { return true }
	for _, op := range []string{">=", "<=", ">", "<", "!="} {
		if strings.HasPrefix(pattern, op) {
			return sh.evalCond(val + " " + op + " " + strings.TrimSpace(pattern[len(op):]))
		}
	}
	p := stripQuotes(pattern)
	if strings.Contains(p, "*") { return globMatch(p, val) }
	return val == p || val == pattern
}

func globMatch(pattern, s string) bool {
	if pattern == "*" { return true }
	parts := strings.Split(pattern, "*")
	pos := 0
	for i, p := range parts {
		if p == "" { continue }
		idx := strings.Index(s[pos:], p)
		if idx < 0 { return false }
		if i == 0 && idx != 0 { return false }
		pos += idx + len(p)
	}
	return true
}

// ─── for loop ─────────────────────────────────────────────────────────────────

func (sh *Shell) evalFor(raw string) int {
	rest  := strings.TrimSpace(raw[4:])
	inIdx := strings.Index(strings.ToLower(rest), " in ")
	if inIdx < 0 { PrintError(errSyntax("expected 'for <var> in <iterable>: <body>'", raw, 0)); return 1 }
	varN  := strings.TrimSpace(rest[:inIdx])
	after := strings.TrimSpace(rest[inIdx+4:])
	colonIdx := colonOutsideBraces(after)

	var iterExpr, body string
	if colonIdx >= 0 {
		iterExpr = strings.TrimSpace(after[:colonIdx])
		body     = extractBody(strings.TrimSpace(after[colonIdx+1:]))
	} else if strings.HasPrefix(strings.TrimSpace(after), "[") {
		rb := strings.Index(after, "]")
		if rb < 0 { PrintError(errSyntax("unclosed '[' in for loop", raw, 0)); return 1 }
		iterExpr = strings.TrimSpace(after[:rb+1])
		body     = extractBody(strings.TrimSpace(after[rb+1:]))
	} else {
		PrintError(errSyntax("expected ':' or '{' after iterable in for", raw, 0)); return 1
	}

	for _, item := range sh.evalIterable(iterExpr, raw) {
		sh.setVar(varN, item)
		code := sh.execBodyLines(body)
		if code == codeBreak { break }
		if code == codeContinue { continue }
		if code != 0 { return code }
	}
	sh.delVar(varN)
	return 0
}

func (sh *Shell) evalIterable(expr, src string) []string {
	expr = strings.TrimSpace(expr)
	low  := strings.ToLower(expr)

	// range N / range A B / range A B S / range(A,B)
	if strings.HasPrefix(low, "range") {
		inner := strings.TrimSpace(expr[5:])
		inner  = strings.Trim(inner, "()")
		if strings.Contains(inner, "..") {
			p := strings.SplitN(inner, "..", 2)
			a, b := 0, 0
			fmt.Sscanf(sh.evalExpr(strings.TrimSpace(p[0])), "%d", &a)
			fmt.Sscanf(sh.evalExpr(strings.TrimSpace(p[1])), "%d", &b)
			return makeRange(a, b)
		}
		inner = strings.ReplaceAll(inner, ",", " ")
		return sh.arrayItems2(sh.evalRangeExpr(inner))
	}
	if strings.HasPrefix(expr, "[") && strings.HasSuffix(expr, "]") { return sh.parseArrayLiteral(expr) }
	if strings.HasPrefix(expr, "`") && strings.HasSuffix(expr, "`") {
		var lines []string
		for _, l := range strings.Split(sh.runSubshell(expr[1:len(expr)-1]), "\n") {
			if t := strings.TrimSpace(l); t != "" { lines = append(lines, t) }
		}
		return lines
	}
	if strings.HasPrefix(expr, "$") {
		val := sh.getVar(expr[1:])
		if strings.HasPrefix(val, "[") { return sh.arrayItems(expr[1:]) }
		return strings.Fields(val)
	}
	if isIdent(expr) {
		val := sh.getVar(expr)
		if strings.HasPrefix(val, "[") { return sh.arrayItems(expr) }
		if val != "" { return strings.Fields(val) }
	}
	return strings.Fields(expr)
}

func makeRange(a, b int) []string {
	if a > b {
		out := make([]string, a-b)
		for i := range out { out[i] = strconv.Itoa(a - i) }
		return out
	}
	out := make([]string, b-a)
	for i := range out { out[i] = strconv.Itoa(a + i) }
	return out
}

// ─── while ────────────────────────────────────────────────────────────────────

func (sh *Shell) evalWhile(raw string) int {
	rest := strings.TrimSpace(raw[6:])
	ci   := colonOutsideBraces(rest)
	if ci < 0 { PrintError(errSyntax("expected ':' in while", raw, 6)); return 1 }
	cond := strings.TrimSpace(rest[:ci])
	body := extractBody(strings.TrimSpace(rest[ci+1:]))
	for i := 0; i < 1_000_000; i++ {
		if !sh.evalCond(cond) { break }
		code := sh.execBodyLines(body)
		if code == codeBreak { break }
		if code == codeContinue { continue }
		if code != 0 { return code }
	}
	return 0
}

// ─── do { } while / until ─────────────────────────────────────────────────────

func (sh *Shell) evalDo(raw string) int {
	rest := strings.TrimSpace(raw[3:])
	body := extractBody(rest)
	remaining := strings.TrimSpace(rest[len(body):])
	low := strings.ToLower(remaining)
	isUntil := strings.HasPrefix(low, "until ")
	isWhile  := strings.HasPrefix(low, "while ")
	if !isUntil && !isWhile { return sh.execBodyLines(body) }
	cond := strings.TrimSpace(remaining[6:])
	for i := 0; i < 1_000_000; i++ {
		code := sh.execBodyLines(body)
		if code == codeBreak { break }
		if code != 0 && code != codeContinue { return code }
		c := sh.evalCond(cond)
		if isWhile && !c { break }
		if isUntil && c  { break }
	}
	return 0
}

// ─── repeat N: body ───────────────────────────────────────────────────────────

func (sh *Shell) evalRepeat(raw string) int {
	rest := strings.TrimSpace(raw[7:])
	ci   := colonOutsideBraces(rest)
	if ci < 0 { PrintError(errSyntax("expected ':' in repeat", raw, 7)); return 1 }
	n := 0
	fmt.Sscanf(sh.evalExpr(strings.TrimSpace(rest[:ci])), "%d", &n)
	body := extractBody(strings.TrimSpace(rest[ci+1:]))
	for i := 0; i < n; i++ {
		sh.vars["_i"] = strconv.Itoa(i)
		code := sh.execBodyLines(body)
		if code == codeBreak { break }
		if code == codeContinue { continue }
		if code != 0 { return code }
	}
	sh.delVar("_i")
	return 0
}

// ─── try / catch / finally ────────────────────────────────────────────────────

func (sh *Shell) evalTry(raw string) int {
	rest := strings.TrimSpace(raw[4:])
	if !strings.HasPrefix(rest, "{") { rest = "{" + rest + "}" }
	body := extractBody(rest)
	remaining := strings.TrimSpace(rest[len(body):])
	catchBody, finallyBody, catchVar := "", "", ""

	low := strings.ToLower(remaining)
	if strings.HasPrefix(low, "catch") {
		after := strings.TrimSpace(remaining[5:])
		if after != "" && after[0] != '{' {
			p := strings.Fields(after)
			if isIdent(p[0]) {
				catchVar = p[0]
				if len(p) > 1 { after = strings.Join(p[1:], " ") }
			}
		}
		catchBody = extractBody(after)
		remaining = strings.TrimSpace(after[len(catchBody):])
	}
	low = strings.ToLower(remaining)
	if strings.HasPrefix(low, "finally") {
		finallyBody = extractBody(strings.TrimSpace(remaining[7:]))
	}

	sh.errHandlerDepth++
	savedThrown := sh.thrownMsg
	sh.thrownMsg = ""
	code := sh.execBodyLines(body)
	sh.errHandlerDepth--
	thrown := sh.thrownMsg
	if thrown == "" && code != 0 {
		if sh.lastErrMsg != "" { thrown = sh.lastErrMsg } else {
			thrown = fmt.Sprintf("command failed with exit code %d", code)
		}
	}

	if code != 0 && catchBody != "" {
		if catchVar != "" { sh.vars[catchVar] = thrown }
		sh.thrownMsg  = savedThrown
		sh.lastErrMsg = ""
		code = sh.execBodyLines(catchBody)
	} else {
		sh.thrownMsg = savedThrown
	}
	if finallyBody != "" { sh.execBodyLines(finallyBody) }
	return code
}

// ─── func def + call ─────────────────────────────────────────────────────────

func (sh *Shell) evalFuncDef(raw string) int {
	rest     := strings.TrimSpace(raw[5:])
	parenIdx := strings.IndexAny(rest, "( {")
	if parenIdx < 0 { PrintError(errSyntax("expected '(' after func name", raw, 5)); return 1 }
	name := strings.TrimSpace(rest[:parenIdx])
	exported := strings.HasSuffix(name, "!")
	if exported { name = name[:len(name)-1] }

	var params []string
	afterParams := rest
	if rest[parenIdx] == '(' {
		ci := strings.Index(rest, ")")
		if ci < 0 { PrintError(errSyntax("unclosed '(' in func", raw, parenIdx)); return 1 }
		for _, p := range strings.Split(rest[parenIdx+1:ci], ",") {
			p = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(p), "[]"), "...")
			if p != "" { params = append(params, p) }
		}
		afterParams = strings.TrimSpace(rest[ci+1:])
	}
	body := extractBody(afterParams)
	sh.funcs[name] = &UserFunc{Name: name, Params: params, Body: bodyLines(body), Exported: exported}
	fmt.Printf("  %s✔ func %s%s%s(%s)%s defined\n",
		ansiGreen, ansiBold+ansiCyan, name, ansiReset, strings.Join(params, ", "), ansiReset)
	return 0
}

func (sh *Shell) callUserFunc(fn *UserFunc, args []string, src string) int {
	if len(fn.Params) > 0 && len(args) < len(fn.Params) {
		PrintError(&ShellError{
			Code: "E007", Kind: "ArgumentError",
			Message: fmt.Sprintf("%s() expects %d argument(s), got %d", fn.Name, len(fn.Params), len(args)),
			Source:  src, Line: sh.currentLine,
			Col:     strings.Index(src, fn.Name), Span: len(fn.Name),
			Hint:    fmt.Sprintf("func %s takes: (%s)", fn.Name, strings.Join(fn.Params, ", ")),
			Fix:     fmt.Sprintf("%s %s", fn.Name, strings.Repeat(`"value" `, len(fn.Params))),
		})
	}

	saved := make(map[string]string)
	for i, p := range fn.Params {
		saved[p] = sh.vars[p]
		if i < len(args) { sh.vars[p] = sh.evalExpr(args[i]) } else { sh.vars[p] = "" }
	}
	if len(args) > len(fn.Params) { sh.vars["_args"] = strings.Join(args[len(fn.Params):], " ") }
	sh.vars["_argc"]   = strconv.Itoa(len(args))
	sh.vars["_return"] = ""
	outerDefer := sh.deferStack
	sh.deferStack = nil
	body := strings.Join(fn.Body, "\n")
	code := sh.execBodyLinesWithGoto(body)
	if code == codeReturn { code = 0 }
	for i := len(sh.deferStack) - 1; i >= 0; i-- { sh.execLine(sh.deferStack[i]) }
	sh.deferStack = outerDefer
	for p, v := range saved { sh.vars[p] = v }

	// Surface _return into captureOut when called inside backtick/subshell
	if sh.captureMode && sh.captureOut.Len() == 0 && sh.vars["_return"] != "" {
		sh.captureOut.WriteString(sh.vars["_return"])
	}
	return code
}

// ─── Subshell / backtick ─────────────────────────────────────────────────────

func (sh *Shell) runSubshell(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" { return "" }

	oldMode := sh.captureMode
	oldOut  := sh.captureOut.String()
	sh.captureMode = true
	sh.captureOut.Reset()

	sh.execLine(cmd)
	captured := strings.TrimRight(sh.captureOut.String(), "\n ")
	// Fallback: if nothing printed but a user func ran, use its _return
	if captured == "" {
		if ret := sh.vars["_return"]; ret != "" { captured = ret }
	}

	sh.captureMode = oldMode
	sh.captureOut.Reset()
	if oldOut != "" { sh.captureOut.WriteString(oldOut) }
	return captured
}

func (sh *Shell) expandBackticks(s string) string {
	for {
		start := strings.IndexByte(s, '`')
		if start < 0 { break }
		end := strings.IndexByte(s[start+1:], '`')
		if end < 0 { break }
		end += start + 1
		s = s[:start] + sh.runSubshell(s[start+1:end]) + s[end+1:]
	}
	return s
}

// ─── Body execution ───────────────────────────────────────────────────────────

const (codeBreak = -1; codeContinue = -2; codeReturn = -3)

func (sh *Shell) execBodyLines(body string) int {
	for _, line := range bodyLines(body) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") { continue }
		low := strings.ToLower(line)
		if low == "break"    { return codeBreak }
		if low == "continue" { return codeContinue }
		if low == "pass"     { continue }
		if strings.HasPrefix(low, "return") {
			val := strings.TrimSpace(line[6:])
			if val != "" { sh.vars["_return"] = sh.evalExpr(val) }
			return codeReturn
		}
		prevSrc := sh.currentSrc
		sh.currentSrc = line
		code := sh.execLine(line)
		sh.currentSrc = prevSrc
		if code == codeBreak || code == codeContinue || code == codeReturn { return code }
		if code != 0 { return code }
	}
	return 0
}

func bodyLines(body string) []string {
	body = strings.TrimSpace(body)
	if strings.HasPrefix(body, "{") && strings.HasSuffix(body, "}") { body = body[1 : len(body)-1] }
	var lines []string
	for _, l := range strings.Split(body, ";") {
		for _, s := range strings.Split(l, "\n") {
			if t := strings.TrimSpace(s); t != "" { lines = append(lines, t) }
		}
	}
	return lines
}

func extractBody(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") { return s }
	depth := 0
	for i, ch := range s {
		if ch == '{' { depth++ }
		if ch == '}' { depth--; if depth == 0 { return s[:i+1] } }
	}
	return s
}

func splitColon(s string) (string, string) {
	idx := colonOutsideBraces(s)
	if idx < 0 { return s, "" }
	return strings.TrimSpace(s[:idx]), strings.TrimSpace(s[idx+1:])
}

func colonOutsideBraces(s string) int {
	depth, inQ := 0, false
	qCh := rune(0)
	for i, ch := range s {
		if inQ { if ch == qCh { inQ = false }; continue }
		if ch == '"' || ch == '\'' { inQ = true; qCh = ch; continue }
		if ch == '{' || ch == '(' || ch == '[' { depth++; continue }
		if ch == '}' || ch == ')' || ch == ']' { depth--; continue }
		if ch == ':' && depth == 0 { return i }
	}
	return -1
}

func splitSemicolon(s string) []string {
	var parts []string
	var cur strings.Builder
	depth, inQ := 0, false
	qCh := rune(0)
	for _, ch := range s {
		if inQ { cur.WriteRune(ch); if ch == qCh { inQ = false }; continue }
		if ch == '"' || ch == '\'' { inQ = true; qCh = ch; cur.WriteRune(ch); continue }
		if ch == '{' || ch == '(' { depth++; cur.WriteRune(ch); continue }
		if ch == '}' || ch == ')' { depth--; cur.WriteRune(ch); continue }
		if ch == ';' && depth == 0 {
			if t := strings.TrimSpace(cur.String()); t != "" { parts = append(parts, t) }
			cur.Reset()
		} else {
			cur.WriteRune(ch)
		}
	}
	if t := strings.TrimSpace(cur.String()); t != "" { parts = append(parts, t) }
	return parts
}

// ─── Range expression (inlined from scripting3.go) ───────────────────────────
//
//   range 5        → [0,1,2,3,4]
//   range 1 10     → [1..9]
//   range 1 10 2   → [1,3,5,7,9]
//   range 10 0 -1  → [10,9,...,1]

func (sh *Shell) evalRangeExpr(args string) string {
	parts := strings.Fields(sh.expandVars(args))
	var start, end, step float64
	switch len(parts) {
	case 0: return sh.makeArray(nil)
	case 1:
		end, _  = strconv.ParseFloat(sh.evalExpr(parts[0]), 64)
		start, step = 0, 1
	case 2:
		start, _ = strconv.ParseFloat(sh.evalExpr(parts[0]), 64)
		end, _   = strconv.ParseFloat(sh.evalExpr(parts[1]), 64)
		if end >= start { step = 1 } else { step = -1 }
	default:
		start, _ = strconv.ParseFloat(sh.evalExpr(parts[0]), 64)
		end, _   = strconv.ParseFloat(sh.evalExpr(parts[1]), 64)
		step, _  = strconv.ParseFloat(sh.evalExpr(parts[2]), 64)
	}
	if step == 0 { step = 1 }
	var items []string
	for i := 0; i < 100_000; i++ {
		v := start + float64(i)*step
		if step > 0 && v >= end { break }
		if step < 0 && v <= end { break }
		items = append(items, fmtNum(v))
	}
	return sh.makeArray(items)
}

// evalRangeExprEval wraps evalRangeExpr for use in evalRHS.
func (sh *Shell) evalRangeExprEval(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	rawLow := strings.ToLower(raw)
	if !strings.HasPrefix(rawLow, "range ") && rawLow != "range" { return "", false }
	args := ""
	if len(raw) > 6 { args = raw[6:] }
	return sh.evalRangeExpr(args), true
}

// ─── Variable helpers ─────────────────────────────────────────────────────────

func (sh *Shell) getVar(name string) string {
	if v, ok := sh.vars[name]; ok { return v }
	return os.Getenv(name)
}

func (sh *Shell) setVar(name, val string) {
	if sh.readonlyVars[name] {
		locStr := ""
		if sh.currentFile != "" {
			locStr = fmt.Sprintf(" (at %s", sh.currentFile)
			if sh.currentLine > 0 { locStr += fmt.Sprintf(":%d", sh.currentLine) }
			locStr += ")"
		}
		PrintError(&ShellError{
			Code: "E008", Kind: "ReadonlyError",
			Message: fmt.Sprintf("cannot reassign readonly variable %q%s", name, locStr),
			Source:  sh.currentSrc, Col: strings.Index(sh.currentSrc, name), Span: len(name),
			Line:    sh.currentLine,
			Hint:    fmt.Sprintf("'%s' was declared readonly — use a different name", name),
			Fix:     "unset " + name + "  # removes readonly",
		})
		return
	}
	sh.vars[name] = val
}

func (sh *Shell) delVar(name string) { delete(sh.vars, name) }

// ─── Misc helpers ─────────────────────────────────────────────────────────────

func isIdent(s string) bool {
	if s == "" { return false }
	for i, ch := range s {
		if i == 0 && !unicode.IsLetter(ch) && ch != '_' { return false }
		if i > 0 && !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' { return false }
	}
	return true
}

func parseNum(s string) (float64, bool) { f, err := strconv.ParseFloat(s, 64); return f, err == nil }

func fmtNum(f float64) string {
	if f == math.Trunc(f) { return strconv.FormatInt(int64(f), 10) }
	return strconv.FormatFloat(f, 'f', -1, 64)
}
