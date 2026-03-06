package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
//  Katsh — 10 new scripting features
//
//  1.  return values from functions  — result = add(3, 4)
//  2.  switch statement              — switch $x { "a": ... "b": ... }
//  3.  enum declaration              — enum Color { Red Green Blue }
//  4.  struct declaration            — struct Point { x y }
//  5.  pipe expression  |>           — x = "hello" |> upper |> trim
//  6.  when guard modifier           — echo "big" when $n > 100
//  7.  defer statement               — defer echo "cleanup"
//  8.  with scoped binding           — with x = expr: body
//  9.  goto / label                  — label top:  /  goto top
// 10.  throw / raise                 — throw "oops"  (works inside try/catch)
//
//  Turing completeness notes:
//   • Arbitrary loops (while + infinite loop + break) ✓
//   • Conditional branching (if/unless/match/switch/when) ✓
//   • Mutable state (variables, arrays, structs) ✓
//   • Unbounded recursion with proper return values ✓  (fixed here)
//   • goto provides arbitrary control flow ✓
// ─────────────────────────────────────────────────────────────────────────────

// ─── 1. Function return values ────────────────────────────────────────────────
//
// Functions now return a value that can be captured:
//   result = myFunc(args)
//   result = myFunc arg1 arg2
//
// Inside a func:  return $x  stores $x in _return and exits the func.
// The caller reads _return automatically.
//
// Patched in callUserFunc2 (replaces callUserFunc for value-capturing calls).

// callUserFuncReturning calls a user function and returns its return value as a string.
func (sh *Shell) callUserFuncReturning(fn *UserFunc, args []string, src string) string {
	// Save and bind parameters
	saved := make(map[string]string)
	for i, p := range fn.Params {
		saved[p] = sh.vars[p]
		if i < len(args) {
			sh.vars[p] = sh.evalExpr(args[i])
		} else {
			sh.vars[p] = ""
		}
	}
	if len(args) > len(fn.Params) {
		sh.vars["_args"] = strings.Join(args[len(fn.Params):], " ")
	}
	sh.vars["_argc"] = strconv.Itoa(len(args))
	sh.vars["_return"] = "" // clear previous return value

	for _, line := range fn.Body {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		code := sh.execLine(line)
		if code == codeReturn {
			break
		}
		if code != 0 && code != codeBreak && code != codeContinue {
			break
		}
	}

	retVal := sh.vars["_return"]
	// Restore parameters
	for p, v := range saved {
		sh.vars[p] = v
	}
	return retVal
}

// tryFuncCallAssign detects patterns like:  result = funcName(args)
// or:  result = funcName arg1 arg2
// Returns (value, true) if matched.
func (sh *Shell) tryFuncCallAssign(raw string) (string, bool) {
	// Must have = sign
	eqIdx := strings.Index(raw, "=")
	if eqIdx <= 0 {
		return "", false
	}
	if eqIdx > 0 && (raw[eqIdx-1] == '!' || raw[eqIdx-1] == '<' || raw[eqIdx-1] == '>') {
		return "", false
	}
	if eqIdx+1 < len(raw) && raw[eqIdx+1] == '=' {
		return "", false
	}

	lhs := strings.TrimSpace(raw[:eqIdx])
	if !isIdent(lhs) {
		return "", false
	}

	rhs := strings.TrimSpace(raw[eqIdx+1:])
	// rhs must start with a known func name optionally followed by ( or space+args
	firstWord := strings.FieldsFunc(rhs, func(r rune) bool { return r == '(' || r == ' ' || r == '\t' })[0]
	fn, ok := sh.funcs[firstWord]
	if !ok {
		return "", false
	}

	// Parse args: either funcName(a, b, c) or funcName a b c
	var argStr string
	after := strings.TrimSpace(rhs[len(firstWord):])
	if strings.HasPrefix(after, "(") {
		end := strings.LastIndex(after, ")")
		if end < 0 {
			end = len(after)
		}
		argStr = after[1:end]
	} else {
		argStr = after
	}

	// Split args respecting quotes
	var args []string
	for _, a := range tokenizeUnquoted(argStr) {
		if t := strings.TrimSpace(a); t != "" {
			args = append(args, sh.evalExpr(t))
		}
	}

	retVal := sh.callUserFuncReturning(fn, args, raw)
	sh.setVar(lhs, retVal)
	return retVal, true
}

// ─── 2. switch statement ──────────────────────────────────────────────────────
//
//   switch $x {
//     "hello": echo "greeting"
//     "bye":   echo "farewell"
//     default: echo "unknown"
//   }
//
// Unlike match, switch uses strict equality only (no operators/globs).
// Falls through to next case only with explicit `fallthrough`.

func (sh *Shell) evalSwitch(raw string) int {
	// switch <expr> { cases }  or  switch <expr>: cases
	rest := strings.TrimSpace(raw[7:]) // strip "switch "
	braceIdx := strings.Index(rest, "{")
	colonIdx := colonOutsideBraces(rest)

	var subject, body string
	if braceIdx >= 0 && (colonIdx < 0 || braceIdx < colonIdx) {
		subject = strings.TrimSpace(rest[:braceIdx])
		body = extractBody(rest[braceIdx:])
	} else if colonIdx >= 0 {
		subject = strings.TrimSpace(rest[:colonIdx])
		body = strings.TrimSpace(rest[colonIdx+1:])
	} else {
		PrintError(errSyntax("expected '{' or ':' in switch", raw, 7))
		return 1
	}

	val := sh.evalExpr(subject)
	inner := body
	if strings.HasPrefix(inner, "{") {
		inner = inner[1 : len(inner)-1]
	}

	cases := splitSemicolon(inner)
	matched := false
	fallthru := false

	for _, cl := range cases {
		cl = strings.TrimSpace(cl)
		low := strings.ToLower(cl)

		if strings.HasPrefix(low, "default") {
			after := strings.TrimSpace(cl[7:])
			after = strings.TrimPrefix(after, ":")
			if matched || fallthru {
				return sh.execBodyLines(extractBody(strings.TrimSpace(after)))
			}
			// default always runs if nothing matched
			matched = true
			code := sh.execBodyLines(extractBody(strings.TrimSpace(after)))
			if code == codeBreak {
				return 0
			}
			return code
		}

		// case expr: body
		ci := colonOutsideBraces(cl)
		if ci < 0 {
			continue
		}
		pattern := strings.TrimSpace(cl[:ci])
		casebody := strings.TrimSpace(cl[ci+1:])

		// Evaluate pattern (strip quotes for comparison)
		patVal := sh.evalExpr(stripQuotes(pattern))

		if matched || fallthru || patVal == val {
			matched = true
			fallthru = false
			// Check for fallthrough keyword
			if strings.TrimSpace(casebody) == "fallthrough" {
				fallthru = true
				continue
			}
			code := sh.execBodyLines(extractBody(casebody))
			if code == codeBreak {
				return 0
			}
			if code != 0 {
				return code
			}
			return 0 // switch exits after first match (no implicit fallthrough)
		}
	}
	return 0
}

// ─── 3. enum declaration ──────────────────────────────────────────────────────
//
//   enum Color { Red Green Blue }
//   enum Direction { North=0 South East West }
//
// Creates variables: Color_Red=0, Color_Green=1, Color_Blue=2
// Also stores the enum in sh.vars as "Color" = "[Red,Green,Blue]"

func (sh *Shell) evalEnum(raw string) int {
	rest := strings.TrimSpace(raw[5:]) // strip "enum "
	spIdx := strings.IndexAny(rest, " {")
	if spIdx < 0 {
		PrintError(errSyntax("expected enum name followed by { members }", raw, 5))
		return 1
	}
	enumName := strings.TrimSpace(rest[:spIdx])
	body := extractBody(strings.TrimSpace(rest[spIdx:]))
	inner := body
	if strings.HasPrefix(inner, "{") {
		inner = inner[1 : len(inner)-1]
	}

	var members []string
	counter := 0
	for _, tok := range strings.Fields(inner) {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if idx := strings.Index(tok, "="); idx > 0 {
			name := tok[:idx]
			val := tok[idx+1:]
			n, err := strconv.Atoi(val)
			if err == nil {
				counter = n
			}
			sh.setVar(enumName+"_"+name, strconv.Itoa(counter))
			members = append(members, name)
		} else {
			sh.setVar(enumName+"_"+tok, strconv.Itoa(counter))
			members = append(members, tok)
		}
		counter++
	}
	// Store enum member list
	sh.setVar(enumName, "["+strings.Join(members, arraySep)+"]")
	fmt.Printf("  %s✔ enum %s%s%s { %s }%s defined\n",
		ansiGreen, ansiBold+ansiCyan, enumName, ansiReset,
		strings.Join(members, ", "), ansiReset)
	return 0
}

// ─── 4. struct declaration + instantiation ───────────────────────────────────
//
//   struct Point { x y }
//   p = Point(10, 20)       → creates p_x=10, p_y=20, p__type=Point
//   echo $p_x               → 10
//   p_x = 99                → update field directly
//   struct Person { name age="unknown" }   → default values

// StructDef holds a struct type definition.
type StructDef struct {
	Name     string
	Fields   []string
	Defaults map[string]string
}

func (sh *Shell) evalStruct(raw string) int {
	rest := strings.TrimSpace(raw[7:]) // strip "struct "
	spIdx := strings.IndexAny(rest, " {")
	if spIdx < 0 {
		PrintError(errSyntax("expected struct name followed by { fields }", raw, 7))
		return 1
	}
	name := strings.TrimSpace(rest[:spIdx])
	body := extractBody(strings.TrimSpace(rest[spIdx:]))
	inner := body
	if strings.HasPrefix(inner, "{") {
		inner = inner[1 : len(inner)-1]
	}

	def := &StructDef{Name: name, Fields: nil, Defaults: make(map[string]string)}
	for _, tok := range strings.Fields(inner) {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if idx := strings.Index(tok, "="); idx > 0 {
			field := tok[:idx]
			defaultVal := stripQuotes(tok[idx+1:])
			def.Fields = append(def.Fields, field)
			def.Defaults[field] = defaultVal
		} else {
			def.Fields = append(def.Fields, tok)
		}
	}

	// Store struct def as a special variable: __struct_Point = "x,y"
	sh.setVar("__struct_"+name, strings.Join(def.Fields, ","))
	for f, d := range def.Defaults {
		sh.setVar("__struct_"+name+"_default_"+f, d)
	}

	fmt.Printf("  %s✔ struct %s%s%s { %s }%s defined\n",
		ansiGreen, ansiBold+ansiCyan, name, ansiReset,
		strings.Join(def.Fields, ", "), ansiReset)
	return 0
}

// instantiateStruct handles:  p = Point(10, 20)  or  p = Point(x=10 y=20)
func (sh *Shell) instantiateStruct(varName, typeName, argStr string) bool {
	fieldList := sh.vars["__struct_"+typeName]
	if fieldList == "" {
		return false
	}

	fields := strings.Split(fieldList, ",")
	args := tokenizeUnquoted(argStr)

	// Set type tag
	sh.setVar(varName+"__type", typeName)

	// Apply defaults first
	for _, f := range fields {
		defKey := "__struct_" + typeName + "_default_" + f
		if d, ok := sh.vars[defKey]; ok {
			sh.setVar(varName+"_"+f, d)
		} else {
			sh.setVar(varName+"_"+f, "")
		}
	}

	// Apply provided args
	for i, arg := range args {
		arg = strings.TrimSpace(arg)
		if kv := strings.SplitN(arg, "=", 2); len(kv) == 2 {
			// named: x=10
			sh.setVar(varName+"_"+kv[0], sh.evalExpr(kv[1]))
		} else if i < len(fields) {
			// positional
			sh.setVar(varName+"_"+fields[i], sh.evalExpr(arg))
		}
	}
	return true
}

// ─── 5. Pipe expression |> ────────────────────────────────────────────────────
//
//   x = "hello world" |> upper |> trim |> split " "
//   result = 42 |> add 8 |> mul 2
//
// Each |> stage is a string/number pipe op (from stringops.go).
// Can also appear as a standalone expression:  "hello" |> upper |> echo

func (sh *Shell) evalPipeExpr(raw string) (string, bool) {
	if !strings.Contains(raw, "|>") {
		return "", false
	}

	// Split on |> outside quotes
	stages := splitPipeArrow(raw)
	if len(stages) < 2 {
		return "", false
	}

	// Evaluate first stage as an expression
	first := strings.TrimSpace(stages[0])
	val := sh.evalExpr(stripQuotes(first))

	// Detect kind
	kind := KindString
	if _, err := strconv.ParseFloat(val, 64); err == nil {
		kind = KindNumber
	}
	r := NewTyped(val, kind)

	// Apply each subsequent stage as a pipe op
	for _, stage := range stages[1:] {
		stage = strings.TrimSpace(stage)
		parts := tokenizeUnquoted(stage)
		if len(parts) == 0 {
			continue
		}
		op := strings.ToLower(parts[0])
		opArgs := parts[1:]

		if isStringOp(op) {
			result, err := applyStringOp(r, op, opArgs)
			if err != nil {
				if se, ok := err.(*ShellError); ok {
					PrintError(se)
				}
				return r.Text, true
			}
			if result != nil {
				r = result
			}
		} else {
			// Try as a user function call
			if fn, ok := sh.funcs[op]; ok {
				var fnArgs []string
				fnArgs = append(fnArgs, r.Text) // pipe value as first arg
				fnArgs = append(fnArgs, opArgs...)
				retVal := sh.callUserFuncReturning(fn, fnArgs, raw)
				r = NewTyped(retVal, KindString)
			}
		}
	}

	return r.Text, true
}

func splitPipeArrow(s string) []string {
	var parts []string
	var cur strings.Builder
	inQ, inBt := false, false
	qCh := rune(0)

	runes := []rune(s)
	for i := 0; i < len(runes); i++ {
		ch := runes[i]
		switch {
		case inBt:
			cur.WriteRune(ch)
			if ch == '`' {
				inBt = false
			}
		case inQ:
			cur.WriteRune(ch)
			if ch == qCh {
				inQ = false
			}
		case ch == '`':
			inBt = true
			cur.WriteRune(ch)
		case ch == '"' || ch == '\'':
			inQ = true
			qCh = ch
			cur.WriteRune(ch)
		case ch == '|' && i+1 < len(runes) && runes[i+1] == '>':
			parts = append(parts, cur.String())
			cur.Reset()
			i++ // skip >
		default:
			cur.WriteRune(ch)
		}
	}
	if cur.Len() > 0 {
		parts = append(parts, cur.String())
	}
	return parts
}

// ─── 6. when guard modifier ───────────────────────────────────────────────────
//
//   echo "big" when $n > 100
//   rm tmpfile when -f tmpfile
//   x++ when $x < 10
//
// The command runs only if the condition is true.

func (sh *Shell) tryWhenGuard(raw string) (int, bool) {
	// Find " when " outside quotes
	idx := findOutside(raw, " when ")
	if idx < 0 {
		return 0, false
	}

	cmd := strings.TrimSpace(raw[:idx])
	cond := strings.TrimSpace(raw[idx+6:])

	if !sh.evalCond(cond) {
		return 0, true
	}
	return sh.execLine(cmd), true
}

// ─── defer implementation using sh.deferStack field ─────────────────────────

func (sh *Shell) evalDefer(raw string) int {
	cmd := strings.TrimSpace(raw[6:]) // strip "defer "
	if cmd == "" {
		PrintError(errSyntax("defer requires a command", raw, 6))
		return 1
	}
	sh.deferStack = append(sh.deferStack, cmd)
	return 0
}

func (sh *Shell) runDeferred() {
	stack := sh.deferStack
	sh.deferStack = nil
	for i := len(stack) - 1; i >= 0; i-- {
		sh.execLine(stack[i])
	}
}

// ─── 8. with scoped binding ───────────────────────────────────────────────────
//
//   with x = 42: echo $x        → x exists only for this block
//   with name = "Alice" {
//     echo "Hello $name"
//   }
//   # $name is unset after the block

func (sh *Shell) evalWith(raw string) int {
	rest := strings.TrimSpace(raw[5:]) // strip "with "
	ci := colonOutsideBraces(rest)
	braceIdx := strings.Index(rest, "{")

	var binding, body string
	if braceIdx >= 0 && (ci < 0 || braceIdx < ci) {
		binding = strings.TrimSpace(rest[:braceIdx])
		body = extractBody(rest[braceIdx:])
	} else if ci >= 0 {
		binding = strings.TrimSpace(rest[:ci])
		body = strings.TrimSpace(rest[ci+1:])
	} else {
		PrintError(errSyntax("expected ':' or '{' in with", raw, 5))
		return 1
	}

	// Parse "name = expr"
	eqIdx := strings.Index(binding, "=")
	if eqIdx <= 0 {
		PrintError(errSyntax("with binding must be 'name = value'", raw, 5))
		return 1
	}
	varName := strings.TrimSpace(binding[:eqIdx])
	valExpr := strings.TrimSpace(binding[eqIdx+1:])
	if !isIdent(varName) {
		PrintError(errSyntax("invalid variable name in with", raw, 5))
		return 1
	}

	// Save old value, bind new
	oldVal, hadOld := sh.vars[varName]
	sh.setVar(varName, sh.evalRHS(valExpr, raw))
	defer func() {
		if hadOld {
			sh.setVar(varName, oldVal)
		} else {
			sh.delVar(varName)
		}
	}()

	return sh.execBodyLines(body)
}

// ─── 9. goto / label ─────────────────────────────────────────────────────────
//
//   label top:
//   x++
//   goto top when $x < 10
//
// goto is implemented at the execBodyLines level by scanning for labels.
// Labels are declared with:  label <name>:
// Jumped to with:            goto <name>
//
// To prevent infinite loops without any other exit condition,
// a max-iteration guard is applied (100,000 jumps).

// execBodyLinesWithGoto runs body lines supporting goto/label.
func (sh *Shell) execBodyLinesWithGoto(body string) int {
	lines := bodyLines(body)
	maxJumps := 100000
	jumps := 0

	// Build label index
	labels := map[string]int{}
	for i, line := range lines {
		line = strings.TrimSpace(line)
		low := strings.ToLower(line)
		if strings.HasPrefix(low, "label ") {
			name := strings.TrimSuffix(strings.TrimSpace(line[6:]), ":")
			labels[name] = i
		}
	}

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			i++
			continue
		}
		low := strings.ToLower(line)

		// Skip label declarations
		if strings.HasPrefix(low, "label ") {
			i++
			continue
		}

		// goto
		if strings.HasPrefix(low, "goto ") {
			rest := strings.TrimSpace(line[5:])
			// Support: goto label when cond
			targetLabel := rest
			if whenIdx := findOutside(rest, " when "); whenIdx >= 0 {
				cond := strings.TrimSpace(rest[whenIdx+6:])
				targetLabel = strings.TrimSpace(rest[:whenIdx])
				if !sh.evalCond(cond) {
					i++
					continue
				}
			}
			idx, ok := labels[targetLabel]
			if !ok {
				PrintError(&ShellError{
					Code: "E002", Kind: "SyntaxError",
					Message: fmt.Sprintf("undefined label %q", targetLabel),
					Source:  line, Col: 5, Span: len(targetLabel),
					Hint: "declare with:  label " + targetLabel + ":",
				})
				return 1
			}
			jumps++
			if jumps > maxJumps {
				PrintError(&ShellError{
					Code: "E004", Kind: "RuntimeError",
					Message: "too many goto jumps (infinite loop?)",
					Source:  line, Col: -1,
					Hint: "add a break condition to your goto loop",
				})
				return 1
			}
			i = idx
			continue
		}

		// break / continue / pass
		if low == "break" {
			return codeBreak
		}
		if low == "continue" {
			return codeContinue
		}
		if low == "pass" {
			i++
			continue
		}

		// return
		if strings.HasPrefix(low, "return") {
			val := strings.TrimSpace(line[6:])
			if val != "" {
				sh.setVar("_return", sh.evalExpr(val))
			}
			return codeReturn
		}

		code := sh.execLine(line)
		if code == codeBreak || code == codeContinue || code == codeReturn {
			return code
		}
		if code != 0 {
			return code
		}
		i++
	}
	return 0
}

// ─── 10. throw / raise ────────────────────────────────────────────────────────
//
//   throw "something went wrong"
//   throw $errorMsg
//   raise "custom error"   (alias)
//
// throw stores the message in _error and returns a non-zero code.
// Inside try/catch, the catch block captures $e = the thrown message.

func (sh *Shell) evalThrow(raw string) int {
	// strip "throw " or "raise "
	msg := strings.TrimSpace(raw[6:])
	if strings.HasPrefix(strings.ToLower(raw), "raise ") {
		msg = strings.TrimSpace(raw[6:])
	}
	msg = sh.evalExpr(stripQuotes(msg))
	msg = sh.expandVars(msg)
	sh.setVar("_error", msg)
	sh.lastErrMsg = msg

	PrintError(&ShellError{
		Code:    "E004",
		Kind:    "RuntimeError",
		Message: msg,
		Source:  raw,
		Col:     -1,
		Hint:    "catch this with:  try { ... } catch e { echo $e }",
	})
	return 1
}

// ─── Wire everything into evalScript ─────────────────────────────────────────

// evalScript2 handles the 10 new constructs.
// Called from evalScript before the original dispatch.
func (sh *Shell) evalScript2(raw string) (bool, int) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false, 0
	}

	lp := len(raw)
	if lp > 25 {
		lp = 25
	}
	lower := strings.ToLower(raw[:lp])

	switch {

	// switch
	case strings.HasPrefix(lower, "switch "):
		return true, sh.evalSwitch(raw)

	// enum
	case strings.HasPrefix(lower, "enum "):
		return true, sh.evalEnum(raw)

	// struct declaration
	case strings.HasPrefix(lower, "struct "):
		return true, sh.evalStruct(raw)

	// defer
	case strings.HasPrefix(lower, "defer "):
		return true, sh.evalDefer(raw)

	// with
	case strings.HasPrefix(lower, "with "):
		return true, sh.evalWith(raw)

	// goto
	case strings.HasPrefix(lower, "goto "):
		// goto at top level just records; execBodyLinesWithGoto handles it
		// But if someone types it interactively, show a hint
		PrintError(&ShellError{
			Code: "E002", Kind: "SyntaxError",
			Message: "goto can only be used inside a function or block body",
			Source:  raw, Col: 0, Span: 4,
			Hint: "use goto inside a func { } block with matching label <name>:",
		})
		return true, 1

	// label at top level — no-op
	case strings.HasPrefix(lower, "label "):
		return true, 0

	// throw / raise
	case strings.HasPrefix(lower, "throw "), strings.HasPrefix(lower, "raise "):
		return true, sh.evalThrow(raw)
	}

	// when guard modifier:  cmd when cond
	if code, ok := sh.tryWhenGuard(raw); ok {
		return true, code
	}

	// pipe expression:  x = expr |> op |> op
	// or standalone:  "hello" |> upper
	if strings.Contains(raw, "|>") {
		// Check if it's an assignment
		eqIdx := strings.Index(raw, "=")
		if eqIdx > 0 {
			lhs := strings.TrimSpace(raw[:eqIdx])
			rhs := strings.TrimSpace(raw[eqIdx+1:])
			if isIdent(lhs) && strings.Contains(rhs, "|>") {
				if val, ok := sh.evalPipeExpr(rhs); ok {
					sh.setVar(lhs, val)
					return true, 0
				}
			}
		}
		// Standalone pipe expr
		if val, ok := sh.evalPipeExpr(raw); ok {
			fmt.Println("\n  " + val + "\n")
			return true, 0
		}
	}

	// Function call with return value capture:  result = funcName(args)
	if code, ok := sh.tryFuncCallCapture(raw); ok {
		return true, code
	}

	return false, 0
}

// tryFuncCallCapture wraps tryFuncCallAssign for use in evalScript2.
func (sh *Shell) tryFuncCallCapture(raw string) (int, bool) {
	eqIdx := strings.Index(raw, "=")
	if eqIdx <= 0 {
		return 0, false
	}
	if eqIdx+1 < len(raw) && raw[eqIdx+1] == '=' {
		return 0, false
	}
	if eqIdx > 0 && (raw[eqIdx-1] == '!' || raw[eqIdx-1] == '<' || raw[eqIdx-1] == '>') {
		return 0, false
	}

	lhs := strings.TrimSpace(raw[:eqIdx])
	if !isIdent(lhs) {
		return 0, false
	}

	rhs := strings.TrimSpace(raw[eqIdx+1:])
	// Must reference a known user function
	firstWord := ""
	for _, ch := range rhs {
		if ch == '(' || ch == ' ' || ch == '\t' {
			break
		}
		firstWord += string(ch)
	}
	if _, ok := sh.funcs[firstWord]; !ok {
		return 0, false
	}
	// Has parens → definitely a call expression
	if !strings.Contains(rhs, "(") {
		return 0, false
	}

	_, captured := sh.tryFuncCallAssign(raw)
	return 0, captured
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// errDivZeroInline returns a division by zero error for inline use.
func errDivZeroInline(src string) error {
	return &ShellError{
		Code: "E006", Kind: "DivisionByZero",
		Message: "division by zero", Source: src, Col: -1,
		Hint: "Check the denominator before dividing",
	}
}

var _ = math.Pi  // ensure math is used
var _ = os.Stdin // ensure os is used
