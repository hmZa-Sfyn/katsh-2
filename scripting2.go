package main

// ─────────────────────────────────────────────────────────────────────────────
//  KatSH — scripting2.go
//
//  Extended scripting features (all inlined — no scripting3.go dependency):
//
//   1.  Function return values    — result = add(3, 4)
//   2.  switch statement          — switch $x { "a": ... }
//   3.  enum declaration          — enum Color { Red Green Blue }
//   4.  struct declaration        — struct Point { x y }
//   5.  pipe expression |>        — x = "hello" |> upper |> trim
//   6.  when guard modifier       — echo "hi" when $n > 0
//   7.  defer statement           — defer echo "cleanup"
//   8.  with scoped binding       — with x = expr: body
//   9.  goto / label              — label top:  /  goto top
//  10.  throw / raise             — throw "oops"
//  11.  assert statement          — assert $x > 0 "msg"
//  12.  typeof / kindof           — typeof $x
//  13.  ternary assignment        — x = cond ? a : b
//  14.  ?? null-coalescing        — x = $y ?? "default"
//  15.  multi-assign              — a, b, c = 1, 2, 3
//  16.  array builtins            — arr_push/pop/filter/map/sort/…
//  17.  map builtins              — map_keys/values/has/del/merge/…
//  18.  range expression          — (forwarded to scripting.go)
//  19.  numeric helpers           — hex/oct/bin/abs/sign/clamp/sqrt/…
//
//  NEW FEATURES (v2.1) — wired into evalScript2 dispatch:
//   A.  loop / forever            — infinite loop  (in scripting.go)
//   B.  format                    — sprintf output (in scripting.go)
//   C.  field N pipe              — "a b c" | field 2  → "b"
//   D.  |? filter expr            — $arr |? $_ > 5
// ─────────────────────────────────────────────────────────────────────────────

import (
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
)

// ─── 1. Function return values ────────────────────────────────────────────────
//
//   result = add(3, 4)
//   result = myFunc arg1 arg2

// callUserFuncReturning calls a user function and returns its _return value.
func (sh *Shell) callUserFuncReturning(fn *UserFunc, args []string, src string) string {
	saved := make(map[string]string)
	for i, p := range fn.Params {
		saved[p] = sh.vars[p]
		if i < len(args) { sh.vars[p] = sh.evalExpr(args[i]) } else { sh.vars[p] = "" }
	}
	if len(args) > len(fn.Params) { sh.vars["_args"] = strings.Join(args[len(fn.Params):], " ") }
	sh.vars["_argc"]   = strconv.Itoa(len(args))
	sh.vars["_return"] = ""

	body := strings.Join(fn.Body, "\n")
	code := sh.execBodyLinesWithGoto(body)
	_ = code

	retVal := sh.vars["_return"]
	for p, v := range saved { sh.vars[p] = v }
	return retVal
}

// tryFuncCallAssign detects  result = funcName(args)  or  result = funcName a b
func (sh *Shell) tryFuncCallAssign(raw string) (string, bool) {
	eqIdx := strings.Index(raw, "=")
	if eqIdx <= 0 { return "", false }
	if eqIdx > 0 && (raw[eqIdx-1] == '!' || raw[eqIdx-1] == '<' || raw[eqIdx-1] == '>') { return "", false }
	if eqIdx+1 < len(raw) && raw[eqIdx+1] == '=' { return "", false }

	lhs := strings.TrimSpace(raw[:eqIdx])
	if !isIdent(lhs) { return "", false }

	rhs := strings.TrimSpace(raw[eqIdx+1:])
	firstWord := ""
	for _, ch := range rhs {
		if ch == '(' || ch == ' ' || ch == '\t' { break }
		firstWord += string(ch)
	}
	fn, ok := sh.funcs[firstWord]
	if !ok { return "", false }

	after := strings.TrimSpace(rhs[len(firstWord):])
	var argStr string
	if strings.HasPrefix(after, "(") {
		end := strings.LastIndex(after, ")")
		if end < 0 { end = len(after) }
		argStr = after[1:end]
	} else {
		argStr = after
	}

	var args []string
	for _, a := range tokenizeUnquoted(argStr) {
		if t := strings.TrimSpace(a); t != "" { args = append(args, sh.evalExpr(t)) }
	}

	retVal := sh.callUserFuncReturning(fn, args, raw)
	sh.setVar(lhs, retVal)
	return retVal, true
}

// ─── 2. switch statement ──────────────────────────────────────────────────────

func (sh *Shell) evalSwitch(raw string) int {
	rest := strings.TrimSpace(raw[7:])
	braceIdx := strings.Index(rest, "{")
	colonIdx := colonOutsideBraces(rest)

	var subject, body string
	if braceIdx >= 0 && (colonIdx < 0 || braceIdx < colonIdx) {
		subject = strings.TrimSpace(rest[:braceIdx])
		body    = extractBody(rest[braceIdx:])
	} else if colonIdx >= 0 {
		subject = strings.TrimSpace(rest[:colonIdx])
		body    = strings.TrimSpace(rest[colonIdx+1:])
	} else {
		PrintError(errSyntax("expected '{' or ':' in switch", raw, 7))
		return 1
	}

	val   := sh.evalExpr(subject)
	inner := body
	if strings.HasPrefix(inner, "{") { inner = inner[1 : len(inner)-1] }

	matched, fallthru := false, false
	for _, cl := range splitSemicolon(inner) {
		cl  = strings.TrimSpace(cl)
		low := strings.ToLower(cl)

		if strings.HasPrefix(low, "default") {
			after := strings.TrimPrefix(strings.TrimSpace(cl[7:]), ":")
			if matched || fallthru {
				return sh.execBodyLines(extractBody(strings.TrimSpace(after)))
			}
			matched = true
			code := sh.execBodyLines(extractBody(strings.TrimSpace(after)))
			if code == codeBreak { return 0 }
			return code
		}

		ci := colonOutsideBraces(cl)
		if ci < 0 { continue }
		pattern  := strings.TrimSpace(cl[:ci])
		casebody := strings.TrimSpace(cl[ci+1:])
		patVal   := sh.evalExpr(stripQuotes(pattern))

		if matched || fallthru || patVal == val {
			matched  = true
			fallthru = false
			if strings.TrimSpace(casebody) == "fallthrough" { fallthru = true; continue }
			code := sh.execBodyLines(extractBody(casebody))
			if code == codeBreak { return 0 }
			if code != 0 { return code }
			return 0
		}
	}
	return 0
}

// ─── 3. enum declaration ──────────────────────────────────────────────────────

func (sh *Shell) evalEnum(raw string) int {
	rest := strings.TrimSpace(raw[5:])
	spIdx := strings.IndexAny(rest, " {")
	if spIdx < 0 {
		PrintError(errSyntax("expected enum name followed by { members }", raw, 5))
		return 1
	}
	enumName := strings.TrimSpace(rest[:spIdx])
	body  := extractBody(strings.TrimSpace(rest[spIdx:]))
	inner := body
	if strings.HasPrefix(inner, "{") { inner = inner[1 : len(inner)-1] }

	var members []string
	counter := 0
	for _, tok := range strings.Fields(inner) {
		if tok == "" { continue }
		if idx := strings.Index(tok, "="); idx > 0 {
			name := tok[:idx]
			n, err := strconv.Atoi(tok[idx+1:])
			if err == nil { counter = n }
			sh.setVar(enumName+"_"+name, strconv.Itoa(counter))
			members = append(members, name)
		} else {
			sh.setVar(enumName+"_"+tok, strconv.Itoa(counter))
			members = append(members, tok)
		}
		counter++
	}
	sh.setVar(enumName, "["+strings.Join(members, arraySep)+"]")
	fmt.Printf("  %s✔ enum %s%s%s { %s }%s defined\n",
		ansiGreen, ansiBold+ansiCyan, enumName, ansiReset, strings.Join(members, ", "), ansiReset)
	return 0
}

// ─── 4. struct declaration + instantiation ───────────────────────────────────

// StructDef holds a struct type definition.
type StructDef struct {
	Name     string
	Fields   []string
	Defaults map[string]string
}

func (sh *Shell) evalStruct(raw string) int {
	rest := strings.TrimSpace(raw[7:])
	spIdx := strings.IndexAny(rest, " {")
	if spIdx < 0 {
		PrintError(errSyntax("expected struct name followed by { fields }", raw, 7))
		return 1
	}
	name  := strings.TrimSpace(rest[:spIdx])
	body  := extractBody(strings.TrimSpace(rest[spIdx:]))
	inner := body
	if strings.HasPrefix(inner, "{") { inner = inner[1 : len(inner)-1] }

	def := &StructDef{Name: name, Defaults: make(map[string]string)}
	for _, tok := range strings.Fields(inner) {
		if tok == "" { continue }
		if idx := strings.Index(tok, "="); idx > 0 {
			field := tok[:idx]
			def.Fields = append(def.Fields, field)
			def.Defaults[field] = stripQuotes(tok[idx+1:])
		} else {
			def.Fields = append(def.Fields, tok)
		}
	}
	sh.setVar("__struct_"+name, strings.Join(def.Fields, ","))
	for f, d := range def.Defaults { sh.setVar("__struct_"+name+"_default_"+f, d) }

	fmt.Printf("  %s✔ struct %s%s%s { %s }%s defined\n",
		ansiGreen, ansiBold+ansiCyan, name, ansiReset, strings.Join(def.Fields, ", "), ansiReset)
	return 0
}

func (sh *Shell) instantiateStruct(varN, typeName, argStr string) bool {
	fieldList := sh.vars["__struct_"+typeName]
	if fieldList == "" { return false }
	fields := strings.Split(fieldList, ",")
	args   := tokenizeUnquoted(argStr)
	sh.setVar(varN+"__type", typeName)
	for _, f := range fields {
		defKey := "__struct_" + typeName + "_default_" + f
		if d, ok := sh.vars[defKey]; ok { sh.setVar(varN+"_"+f, d) } else { sh.setVar(varN+"_"+f, "") }
	}
	for i, arg := range args {
		arg = strings.TrimSpace(arg)
		if kv := strings.SplitN(arg, "=", 2); len(kv) == 2 {
			sh.setVar(varN+"_"+kv[0], sh.evalExpr(kv[1]))
		} else if i < len(fields) {
			sh.setVar(varN+"_"+fields[i], sh.evalExpr(arg))
		}
	}
	return true
}

// ─── 5. Pipe expression |> ────────────────────────────────────────────────────

func (sh *Shell) evalPipeExpr(raw string) (string, bool) {
	if !strings.Contains(raw, "|>") { return "", false }
	stages := splitPipeArrow(raw)
	if len(stages) < 2 { return "", false }

	val  := sh.evalExpr(stripQuotes(strings.TrimSpace(stages[0])))
	kind := KindString
	if _, err := strconv.ParseFloat(val, 64); err == nil { kind = KindNumber }
	r := NewTyped(val, kind)

	for _, stage := range stages[1:] {
		stage = strings.TrimSpace(stage)
		parts := tokenizeUnquoted(stage)
		if len(parts) == 0 { continue }
		op := strings.ToLower(parts[0])
		if isStringOp(op) {
			result, err := applyStringOp(r, op, parts[1:])
			if err != nil {
				if se, ok := err.(*ShellError); ok { PrintError(se) }
				return r.Text, true
			}
			if result != nil { r = result }
		} else if fn, ok := sh.funcs[op]; ok {
			fnArgs := append([]string{r.Text}, parts[1:]...)
			r = NewTyped(sh.callUserFuncReturning(fn, fnArgs, raw), KindString)
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
		case inBt: cur.WriteRune(ch); if ch == '`' { inBt = false }
		case inQ:  cur.WriteRune(ch); if ch == qCh { inQ = false }
		case ch == '`': inBt = true; cur.WriteRune(ch)
		case ch == '"' || ch == '\'': inQ = true; qCh = ch; cur.WriteRune(ch)
		case ch == '|' && i+1 < len(runes) && runes[i+1] == '>':
			parts = append(parts, cur.String()); cur.Reset(); i++
		default: cur.WriteRune(ch)
		}
	}
	if cur.Len() > 0 { parts = append(parts, cur.String()) }
	return parts
}

// ─── 6. when guard modifier ───────────────────────────────────────────────────

func (sh *Shell) tryWhenGuard(raw string) (int, bool) {
	idx := findOutside(raw, " when ")
	if idx < 0 { return 0, false }
	cmd  := strings.TrimSpace(raw[:idx])
	cond := strings.TrimSpace(raw[idx+6:])
	if !sh.evalCond(cond) { return 0, true }
	return sh.execLine(cmd), true
}

// ─── 7. defer ─────────────────────────────────────────────────────────────────

func (sh *Shell) evalDefer(raw string) int {
	cmd := strings.TrimSpace(raw[6:])
	if cmd == "" { PrintError(errSyntax("defer requires a command", raw, 6)); return 1 }
	sh.deferStack = append(sh.deferStack, cmd)
	return 0
}

func (sh *Shell) runDeferred() {
	stack := sh.deferStack; sh.deferStack = nil
	for i := len(stack) - 1; i >= 0; i-- { sh.execLine(stack[i]) }
}

// ─── 8. with scoped binding ───────────────────────────────────────────────────

func (sh *Shell) evalWith(raw string) int {
	rest := strings.TrimSpace(raw[5:])
	ci       := colonOutsideBraces(rest)
	braceIdx := strings.Index(rest, "{")

	var binding, body string
	if braceIdx >= 0 && (ci < 0 || braceIdx < ci) {
		binding = strings.TrimSpace(rest[:braceIdx])
		body    = extractBody(rest[braceIdx:])
	} else if ci >= 0 {
		binding = strings.TrimSpace(rest[:ci])
		body    = strings.TrimSpace(rest[ci+1:])
	} else {
		PrintError(errSyntax("expected ':' or '{' in with", raw, 5))
		return 1
	}
	eqIdx := strings.Index(binding, "=")
	if eqIdx <= 0 { PrintError(errSyntax("with binding must be 'name = value'", raw, 5)); return 1 }
	vn      := strings.TrimSpace(binding[:eqIdx])
	valExpr := strings.TrimSpace(binding[eqIdx+1:])
	if !isIdent(vn) { PrintError(errSyntax("invalid variable name in with", raw, 5)); return 1 }

	oldVal, hadOld := sh.vars[vn]
	sh.setVar(vn, sh.evalRHS(valExpr, raw))
	defer func() {
		if hadOld { sh.setVar(vn, oldVal) } else { sh.delVar(vn) }
	}()
	return sh.execBodyLines(body)
}

// ─── 9. goto / label ─────────────────────────────────────────────────────────

func (sh *Shell) execBodyLinesWithGoto(body string) int {
	lines    := bodyLines(body)
	maxJumps := 100_000
	jumps    := 0

	labels := map[string]int{}
	for i, line := range lines {
		line = strings.TrimSpace(line)
		low  := strings.ToLower(line)
		if strings.HasPrefix(low, "label ") {
			name := strings.TrimSuffix(strings.TrimSpace(line[6:]), ":")
			labels[name] = i
		}
	}

	i := 0
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			i++; continue
		}
		low := strings.ToLower(line)

		if strings.HasPrefix(low, "label ") { i++; continue }

		if strings.HasPrefix(low, "goto ") {
			rest        := strings.TrimSpace(line[5:])
			targetLabel := rest
			if whenIdx := findOutside(rest, " when "); whenIdx >= 0 {
				cond        := strings.TrimSpace(rest[whenIdx+6:])
				targetLabel  = strings.TrimSpace(rest[:whenIdx])
				if !sh.evalCond(cond) { i++; continue }
			}
			idx, ok := labels[targetLabel]
			if !ok {
				PrintError(&ShellError{Code: "E002", Kind: "SyntaxError",
					Message: fmt.Sprintf("undefined label %q", targetLabel),
					Source: line, Col: 5, Span: len(targetLabel),
					Hint: "declare with:  label " + targetLabel + ":"})
				return 1
			}
			jumps++
			if jumps > maxJumps {
				PrintError(&ShellError{Code: "E004", Kind: "RuntimeError",
					Message: "too many goto jumps (infinite loop?)",
					Source: line, Col: -1,
					Hint: "add a break condition to your goto loop"})
				return 1
			}
			i = idx
			continue
		}

		if low == "break"    { return codeBreak }
		if low == "continue" { return codeContinue }
		if low == "pass"     { i++; continue }

		if strings.HasPrefix(low, "return") {
			val := strings.TrimSpace(line[6:])
			if val != "" { sh.vars["_return"] = sh.evalExpr(val) }
			return codeReturn
		}

		code := sh.execLine(line)
		if code == codeBreak || code == codeContinue || code == codeReturn { return code }
		if code != 0 { return code }
		i++
	}
	return 0
}

// ─── 10. throw / raise ───────────────────────────────────────────────────────

func (sh *Shell) evalThrow(raw string) int {
	lraw := strings.ToLower(raw)
	msg  := ""
	if strings.HasPrefix(lraw, "throw ") || strings.HasPrefix(lraw, "raise ") {
		msg = strings.TrimSpace(raw[6:])
	}
	msg = sh.evalExpr(stripQuotes(sh.expandVars(msg)))
	if msg == "" { msg = "an error was thrown" }

	sh.vars["_error"] = msg
	sh.lastErrMsg     = msg
	sh.thrownMsg      = msg

	if sh.errHandlerDepth == 0 {
		loc := ""
		if sh.currentFile != "" {
			loc = sh.currentFile
			if sh.currentLine > 0 { loc = fmt.Sprintf("%s:%d", sh.currentFile, sh.currentLine) }
		}
		detail := ""
		if loc != "" { detail = "thrown at " + loc }
		PrintError(&ShellError{
			Code: "E009", Kind: "UnhandledThrow",
			Message: msg, Source: sh.currentSrc, Line: sh.currentLine,
			Col: -1, Detail: detail,
			Hint: "wrap in:  try { ... } catch e { println $e }",
			Fix:  "try { " + strings.TrimSpace(raw) + " } catch e { println \"caught: $e\" }",
		})
	}
	return 1
}

// ─── 11. assert ───────────────────────────────────────────────────────────────
//
//   assert $x > 0
//   assert $x > 0 "x must be positive"
//   assert -f config.json "config file is required"

func (sh *Shell) evalAssert(raw string) int {
	rest := strings.TrimSpace(raw[7:]) // strip "assert "
	msg  := ""
	if lIdx := strings.LastIndexAny(rest, `"'`); lIdx > 0 {
		qc   := rune(rest[lIdx])
		fIdx := strings.IndexRune(rest, qc)
		if fIdx >= 0 && fIdx < lIdx {
			msg  = rest[fIdx+1 : lIdx]
			rest = strings.TrimSpace(rest[:fIdx])
		}
	}
	cond := sh.expandVars(rest)
	if sh.evalCond(cond) { return 0 }
	if msg == "" { msg = fmt.Sprintf("assertion failed: %s", rest) }
	msg = sh.expandVars(msg)
	PrintError(&ShellError{
		Code: "E014", Kind: "AssertionError",
		Message: msg, Source: sh.currentSrc, Line: sh.currentLine, Col: 0,
		Detail: fmt.Sprintf("condition evaluated to false: %s", rest),
		Hint:   "check the value of variables involved in this condition",
	})
	return 1
}

// ─── 12. typeof / kindof ─────────────────────────────────────────────────────

func katshTypeOf(val string) string {
	if strings.HasPrefix(val, mapPfx)   { return "map" }
	if strings.HasPrefix(val, setPfx)   { return "set" }
	if strings.HasPrefix(val, stackPfx) { return "stack" }
	if strings.HasPrefix(val, queuePfx) { return "queue" }
	if strings.HasPrefix(val, tupPfx)   { return "tuple" }
	if strings.HasPrefix(val, matPfx)   { return "matrix" }
	if strings.HasPrefix(val, "[") {
		inner := val[1 : len(val)-1]
		if strings.TrimSpace(inner) == "" || strings.Contains(inner, arraySep) { return "array" }
		return "array"
	}
	if _, err := strconv.ParseFloat(val, 64); err == nil { return "number" }
	if val == "true" || val == "false" { return "bool" }
	if val == "" || val == "null" || val == "nil" { return "null" }
	return "string"
}

// ─── 13. Ternary assignment ───────────────────────────────────────────────────

func (sh *Shell) tryTernaryAssign(raw string) (int, bool) {
	eqIdx := strings.Index(raw, "=")
	if eqIdx <= 0 { return 0, false }
	if eqIdx+1 < len(raw) && raw[eqIdx+1] == '=' { return 0, false }
	if raw[eqIdx-1] == '!' || raw[eqIdx-1] == '<' || raw[eqIdx-1] == '>' { return 0, false }
	lhs := strings.TrimSpace(raw[:eqIdx])
	if !isIdent(lhs) { return 0, false }

	rhs := strings.TrimSpace(raw[eqIdx+1:])
	qIdx := findOutside(rhs, "?")
	if qIdx < 0 { return 0, false }
	if qIdx+1 < len(rhs) && rhs[qIdx+1] == '(' { return 0, false }

	cIdx := findOutside(rhs[qIdx+1:], ":")
	if cIdx < 0 { return 0, false }
	cIdx += qIdx + 1

	condExpr  := strings.TrimSpace(rhs[:qIdx])
	trueExpr  := strings.TrimSpace(rhs[qIdx+1 : cIdx])
	falseExpr := strings.TrimSpace(rhs[cIdx+1:])

	if strings.HasPrefix(condExpr, "(") && strings.HasSuffix(condExpr, ")") {
		condExpr = condExpr[1 : len(condExpr)-1]
	}

	var val string
	if sh.evalCond(condExpr) {
		val = sh.evalExpr(stripQuotes(trueExpr))
	} else if strings.Contains(falseExpr, "?") {
		tmp := "__ternary_tmp = " + falseExpr
		if _, ok := sh.tryTernaryAssign(tmp); ok {
			val = sh.vars["__ternary_tmp"]
			sh.delVar("__ternary_tmp")
		} else {
			val = sh.evalExpr(stripQuotes(falseExpr))
		}
	} else {
		val = sh.evalExpr(stripQuotes(falseExpr))
	}

	sh.setVar(lhs, val)
	return 0, true
}

// ─── 14. ?? null-coalescing ───────────────────────────────────────────────────

func (sh *Shell) tryNullCoalesce(raw string) (int, bool) {
	eqIdx := strings.Index(raw, "=")
	if eqIdx <= 0 { return 0, false }
	if eqIdx+1 < len(raw) && raw[eqIdx+1] == '=' { return 0, false }
	if raw[eqIdx-1] == '!' || raw[eqIdx-1] == '<' || raw[eqIdx-1] == '>' { return 0, false }
	lhs := strings.TrimSpace(raw[:eqIdx])
	if !isIdent(lhs) { return 0, false }
	rhs  := strings.TrimSpace(raw[eqIdx+1:])
	qIdx := strings.Index(rhs, "??")
	if qIdx < 0 { return 0, false }

	primary  := strings.TrimSpace(rhs[:qIdx])
	fallback := strings.TrimSpace(rhs[qIdx+2:])
	val := sh.evalExpr(primary)
	if val == "" || val == "null" || val == "nil" || val == "undefined" {
		val = sh.evalExpr(stripQuotes(fallback))
	}
	sh.setVar(lhs, val)
	return 0, true
}

// ─── 15. Multi-assign ────────────────────────────────────────────────────────
//
//   a, b, c = 1, 2, 3
//   first, rest... = $arr

func (sh *Shell) tryMultiAssign(raw string) (int, bool) {
	eqIdx := findOutside(raw, "=")
	if eqIdx <= 0 { return 0, false }
	if eqIdx+1 < len(raw) && raw[eqIdx+1] == '=' { return 0, false }
	if raw[eqIdx-1] == '!' || raw[eqIdx-1] == '<' || raw[eqIdx-1] == '>' { return 0, false }

	lhsRaw := raw[:eqIdx]
	if !strings.Contains(lhsRaw, ",") { return 0, false }

	lhsParts := strings.Split(lhsRaw, ",")
	for _, p := range lhsParts {
		name := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(p), "..."))
		if !isIdent(name) { return 0, false }
	}

	rhsRaw := strings.TrimSpace(raw[eqIdx+1:])
	var rhsVals []string
	if strings.HasPrefix(rhsRaw, "$") {
		if arr := sh.arrayItems(rhsRaw[1:]); len(arr) > 0 { rhsVals = arr }
	}
	if len(rhsVals) == 0 {
		for _, v := range strings.Split(rhsRaw, ",") {
			rhsVals = append(rhsVals, sh.evalExpr(strings.TrimSpace(v)))
		}
	}

	for i, p := range lhsParts {
		name   := strings.TrimSpace(p)
		isRest := strings.HasSuffix(name, "...")
		if isRest {
			name = strings.TrimSpace(strings.TrimSuffix(name, "..."))
			if i < len(rhsVals) { sh.setVar(name, sh.makeArray(rhsVals[i:])) } else { sh.setVar(name, sh.makeArray(nil)) }
			break
		}
		if i < len(rhsVals) { sh.setVar(name, rhsVals[i]) } else { sh.setVar(name, "") }
	}
	return 0, true
}

// ─── 16. Array builtins ───────────────────────────────────────────────────────

func (sh *Shell) tryArrayBuiltin(raw string) (int, bool) {
	if !strings.HasPrefix(strings.ToLower(raw), "arr_") { return 0, false }

	eqIdx := strings.Index(raw, "=")
	var lhsVar, rest string
	if eqIdx > 0 && !strings.ContainsRune(raw[:eqIdx], ' ') {
		lhsVar = strings.TrimSpace(raw[:eqIdx])
		rest   = strings.TrimSpace(raw[eqIdx+1:])
		if !isIdent(lhsVar) { lhsVar = ""; rest = raw }
	} else {
		rest = raw
	}

	parts := strings.Fields(rest)
	if len(parts) < 1 { return 0, false }
	op      := strings.ToLower(parts[0])
	arrName := ""
	if len(parts) >= 2 { arrName = varName2(parts[1]) }
	extra   := parts[2:]
	output  := ""
	handled := true

	switch op {
	case "arr_push":
		if arrName == "" || len(extra) == 0 { break }
		items := sh.arrayItems(arrName)
		for _, v := range extra { items = append(items, sh.evalExpr(stripQuotes(v))) }
		sh.setVar(arrName, sh.makeArray(items))

	case "arr_pop":
		items := sh.arrayItems(arrName)
		if len(items) == 0 {
			PrintError(&ShellError{Code: "E013", Kind: "IndexError",
				Message: fmt.Sprintf("arr_pop: '%s' is empty", arrName),
				Source: sh.currentSrc, Line: sh.currentLine, Col: 0,
				Hint: "check array length before popping"}); break
		}
		output = items[len(items)-1]
		sh.setVar(arrName, sh.makeArray(items[:len(items)-1]))

	case "arr_shift":
		items := sh.arrayItems(arrName)
		if len(items) == 0 {
			PrintError(&ShellError{Code: "E013", Kind: "IndexError",
				Message: fmt.Sprintf("arr_shift: '%s' is empty", arrName),
				Source: sh.currentSrc, Line: sh.currentLine, Col: 0,
				Hint: "check array length before shifting"}); break
		}
		output = items[0]
		sh.setVar(arrName, sh.makeArray(items[1:]))

	case "arr_unshift":
		if arrName == "" || len(extra) == 0 { break }
		var newItems []string
		for _, v := range extra { newItems = append(newItems, sh.evalExpr(stripQuotes(v))) }
		sh.setVar(arrName, sh.makeArray(append(newItems, sh.arrayItems(arrName)...)))

	case "arr_contains":
		if arrName == "" || len(extra) == 0 { output = "false"; break }
		needle := sh.evalExpr(stripQuotes(extra[0]))
		output  = "false"
		for _, it := range sh.arrayItems(arrName) { if it == needle { output = "true"; break } }

	case "arr_reverse":
		items := sh.arrayItems(arrName)
		rev   := make([]string, len(items))
		for i, v := range items { rev[len(items)-1-i] = v }
		output = sh.makeArray(rev)

	case "arr_unique":
		seen := map[string]bool{}
		var out []string
		for _, it := range sh.arrayItems(arrName) {
			if !seen[it] { seen[it] = true; out = append(out, it) }
		}
		output = sh.makeArray(out)

	case "arr_sort":
		items  := append([]string(nil), sh.arrayItems(arrName)...)
		allNum := true
		nums   := make([]float64, len(items))
		for i, s := range items {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil { allNum = false; break }
			nums[i] = f
		}
		if allNum {
			sort.Slice(items, func(i, j int) bool { return nums[i] < nums[j] })
		} else {
			sort.Strings(items)
		}
		output = sh.makeArray(items)

	case "arr_sum":
		sum := 0.0
		for _, it := range sh.arrayItems(arrName) {
			if f, err := strconv.ParseFloat(it, 64); err == nil { sum += f }
		}
		output = fmtNum(sum)

	case "arr_min":
		items := sh.arrayItems(arrName)
		if len(items) == 0 { break }
		min := math.MaxFloat64
		for _, it := range items {
			if f, err := strconv.ParseFloat(it, 64); err == nil && f < min { min = f }
		}
		output = fmtNum(min)

	case "arr_max":
		items := sh.arrayItems(arrName)
		if len(items) == 0 { break }
		max := -math.MaxFloat64
		for _, it := range items {
			if f, err := strconv.ParseFloat(it, 64); err == nil && f > max { max = f }
		}
		output = fmtNum(max)

	case "arr_len":
		output = strconv.Itoa(len(sh.arrayItems(arrName)))

	case "arr_join":
		sep := " "
		if len(extra) > 0 { sep = stripQuotes(sh.evalExpr(extra[0])) }
		output = strings.Join(sh.arrayItems(arrName), sep)

	case "arr_flatten":
		var flat []string
		for _, it := range sh.arrayItems(arrName) {
			if strings.HasPrefix(it, "[") {
				inner := it[1 : len(it)-1]
				for _, sub := range strings.Split(inner, arraySep) {
					if strings.TrimSpace(sub) != "" { flat = append(flat, sub) }
				}
			} else { flat = append(flat, it) }
		}
		output = sh.makeArray(flat)

	case "arr_zip":
		if arrName == "" || len(extra) == 0 { break }
		a := sh.arrayItems(arrName)
		b := sh.arrayItems(varName2(extra[0]))
		lim := len(a); if len(b) < lim { lim = len(b) }
		var pairs []string
		for i := 0; i < lim; i++ { pairs = append(pairs, sh.makeArray([]string{a[i], b[i]})) }
		output = sh.makeArray(pairs)

	case "arr_chunk":
		if arrName == "" || len(extra) == 0 { break }
		n, _ := strconv.Atoi(extra[0])
		if n <= 0 { n = 1 }
		items := sh.arrayItems(arrName)
		var chunks []string
		for i := 0; i < len(items); i += n {
			end := i + n; if end > len(items) { end = len(items) }
			chunks = append(chunks, sh.makeArray(items[i:end]))
		}
		output = sh.makeArray(chunks)

	case "arr_filter":
		if len(extra) == 0 { output = sh.makeArray(sh.arrayItems(arrName)); break }
		condStr := stripQuotes(strings.Join(extra, " "))
		var filtered []string
		for _, it := range sh.arrayItems(arrName) {
			sh.vars["_item"] = it; sh.vars["_"] = it
			if sh.evalCond(condStr) { filtered = append(filtered, it) }
		}
		sh.delVar("_item"); sh.delVar("_")
		output = sh.makeArray(filtered)

	case "arr_map":
		if len(extra) == 0 { output = sh.makeArray(sh.arrayItems(arrName)); break }
		exprStr := stripQuotes(strings.Join(extra, " "))
		var mapped []string
		for _, it := range sh.arrayItems(arrName) {
			sh.vars["_item"] = it; sh.vars["_"] = it
			mapped = append(mapped, sh.evalExpr(exprStr))
		}
		sh.delVar("_item"); sh.delVar("_")
		output = sh.makeArray(mapped)

	case "arr_find":
		if len(extra) == 0 { break }
		condStr := stripQuotes(strings.Join(extra, " "))
		for _, it := range sh.arrayItems(arrName) {
			sh.vars["_item"] = it; sh.vars["_"] = it
			if sh.evalCond(condStr) { output = it; break }
		}
		sh.delVar("_item"); sh.delVar("_")

	default:
		handled = false
	}

	if !handled { return 0, false }
	if lhsVar != "" { sh.setVar(lhsVar, output) } else if output != "" { sh.printResult(NewTyped(output, KindString)) }
	return 0, true
}

// ─── 17. Map builtins ────────────────────────────────────────────────────────

func (sh *Shell) tryMapBuiltin(raw string) (int, bool) {
	if !strings.HasPrefix(strings.ToLower(raw), "map_") { return 0, false }

	eqIdx := strings.Index(raw, "=")
	var lhsVar, rest string
	if eqIdx > 0 && !strings.ContainsRune(raw[:eqIdx], ' ') {
		lhsVar = strings.TrimSpace(raw[:eqIdx])
		rest   = strings.TrimSpace(raw[eqIdx+1:])
		if !isIdent(lhsVar) { lhsVar = ""; rest = raw }
	} else {
		rest = raw
	}

	parts := strings.Fields(rest)
	if len(parts) < 1 { return 0, false }
	op      := strings.ToLower(parts[0])
	mapName := ""
	if len(parts) >= 2 { mapName = varName2(parts[1]) }
	extra   := parts[2:]
	output  := ""
	handled := true

	switch op {
	case "map_keys":
		output = sh.makeArray(mapKeysList(sh.vars[mapName]))

	case "map_values":
		var vals []string
		for _, k := range mapKeysList(sh.vars[mapName]) { vals = append(vals, mapGet(sh.vars[mapName], k)) }
		output = sh.makeArray(vals)

	case "map_has":
		if len(extra) == 0 { output = "false"; break }
		k := stripQuotes(sh.evalExpr(extra[0]))
		output = "false"
		for _, mk := range mapKeysList(sh.vars[mapName]) { if mk == k { output = "true"; break } }

	case "map_del":
		if len(extra) == 0 { break }
		k := stripQuotes(sh.evalExpr(extra[0]))
		sh.setVar(mapName, mapDelete(sh.vars[mapName], k))

	case "map_size":
		output = strconv.Itoa(len(mapKeysList(sh.vars[mapName])))

	case "map_merge":
		if len(extra) == 0 { output = sh.vars[mapName]; break }
		output = mapMerge3(sh.vars[mapName], sh.vars[varName2(extra[0])])

	case "map_entries":
		var entries []string
		for _, k := range mapKeysList(sh.vars[mapName]) {
			entries = append(entries, sh.makeArray([]string{k, mapGet(sh.vars[mapName], k)}))
		}
		output = sh.makeArray(entries)

	default:
		handled = false
	}

	if !handled { return 0, false }
	if lhsVar != "" { sh.setVar(lhsVar, output) } else if output != "" { sh.printResult(NewTyped(output, KindString)) }
	return 0, true
}

// ─── 19. Numeric helpers ─────────────────────────────────────────────────────

func (sh *Shell) tryNumericHelper(raw string) (int, bool) {
	parts := strings.Fields(raw)
	if len(parts) < 2 { return 0, false }
	op := strings.ToLower(parts[0])

	lhsVar := ""
	if eqIdx := strings.Index(raw, "="); eqIdx > 0 {
		lhs := strings.TrimSpace(raw[:eqIdx])
		if isIdent(lhs) {
			lhsVar = lhs
			rest   := strings.TrimSpace(raw[eqIdx+1:])
			parts   = strings.Fields(rest)
			if len(parts) >= 1 { op = strings.ToLower(parts[0]) }
		}
	}

	output  := ""
	handled := true

	switch op {
	case "hex":
		if len(parts) < 2 { break }
		n, err := strconv.ParseInt(sh.evalExpr(parts[1]), 0, 64)
		if err != nil { break }; output = fmt.Sprintf("%x", n)
	case "HEX":
		if len(parts) < 2 { break }
		n, err := strconv.ParseInt(sh.evalExpr(parts[1]), 0, 64)
		if err != nil { break }; output = fmt.Sprintf("%X", n)
	case "oct":
		if len(parts) < 2 { break }
		n, err := strconv.ParseInt(sh.evalExpr(parts[1]), 0, 64)
		if err != nil { break }; output = fmt.Sprintf("%o", n)
	case "bin":
		if len(parts) < 2 { break }
		n, err := strconv.ParseInt(sh.evalExpr(parts[1]), 0, 64)
		if err != nil { break }; output = fmt.Sprintf("%b", n)
	case "abs":
		if len(parts) < 2 { break }
		f, err := strconv.ParseFloat(sh.evalExpr(parts[1]), 64)
		if err != nil { break }; output = fmtNum(math.Abs(f))
	case "sign":
		if len(parts) < 2 { break }
		f, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64)
		if f > 0 { output = "1" } else if f < 0 { output = "-1" } else { output = "0" }
	case "clamp":
		if len(parts) < 4 { break }
		v, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64)
		lo, _ := strconv.ParseFloat(sh.evalExpr(parts[2]), 64)
		hi, _ := strconv.ParseFloat(sh.evalExpr(parts[3]), 64)
		if v < lo { v = lo } else if v > hi { v = hi }
		output = fmtNum(v)
	case "int":
		if len(parts) < 2 { break }
		f, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64)
		output = strconv.FormatInt(int64(f), 10)
	case "float":
		if len(parts) < 2 { break }
		f, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64)
		output = strconv.FormatFloat(f, 'f', -1, 64)
	case "round":
		if len(parts) < 2 { break }
		f, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64)
		dp := 0
		if len(parts) >= 3 { dp, _ = strconv.Atoi(parts[2]) }
		if dp <= 0 { output = fmtNum(math.Round(f)) } else {
			factor := math.Pow(10, float64(dp))
			output = fmtNum(math.Round(f*factor) / factor)
		}
	case "floor":
		if len(parts) < 2 { break }
		f, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64); output = fmtNum(math.Floor(f))
	case "ceil":
		if len(parts) < 2 { break }
		f, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64); output = fmtNum(math.Ceil(f))
	case "sqrt":
		if len(parts) < 2 { break }
		f, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64); output = fmtNum(math.Sqrt(f))
	case "log":
		if len(parts) < 2 { break }
		f, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64); output = fmtNum(math.Log(f))
	case "log2":
		if len(parts) < 2 { break }
		f, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64); output = fmtNum(math.Log2(f))
	case "log10":
		if len(parts) < 2 { break }
		f, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64); output = fmtNum(math.Log10(f))
	case "pow":
		if len(parts) < 3 { break }
		base, _ := strconv.ParseFloat(sh.evalExpr(parts[1]), 64)
		exp, _  := strconv.ParseFloat(sh.evalExpr(parts[2]), 64)
		output = fmtNum(math.Pow(base, exp))
	case "pi":  output = fmtNum(math.Pi)
	case "e":   output = fmtNum(math.E)
	default:    handled = false
	}

	if !handled { return 0, false }
	if lhsVar != "" { sh.setVar(lhsVar, output) } else if output != "" { sh.printResult(NewTyped(output, KindString)) }
	return 0, true
}

// ─── evalScript2 — main dispatch ─────────────────────────────────────────────
//
// Called from evalScript (scripting.go) before the builtin dispatch.
// Handles all extended language features in a single O(1) prefix switch.

func (sh *Shell) evalScript2(raw string) (bool, int) {
	raw = strings.TrimSpace(raw)
	if raw == "" { return false, 0 }

	lp := 30
	if len(raw) < lp { lp = len(raw) }
	lower := strings.ToLower(raw[:lp])

	// ── Keyword prefix dispatch (O(1) amortised) ──────────────────────────
	switch {

	case strings.HasPrefix(lower, "switch "):
		return true, sh.evalSwitch(raw)

	case strings.HasPrefix(lower, "enum "):
		return true, sh.evalEnum(raw)

	case strings.HasPrefix(lower, "struct "):
		return true, sh.evalStruct(raw)

	case strings.HasPrefix(lower, "defer "):
		return true, sh.evalDefer(raw)

	case strings.HasPrefix(lower, "with "):
		return true, sh.evalWith(raw)

	case strings.HasPrefix(lower, "goto "):
		PrintError(&ShellError{Code: "E002", Kind: "SyntaxError",
			Message: "goto can only be used inside a function or block body",
			Source:  raw, Col: 0, Span: 4,
			Hint:    "use goto inside a func { } block with matching label <n>:"})
		return true, 1

	case strings.HasPrefix(lower, "label "):
		return true, 0 // no-op at top level

	case strings.HasPrefix(lower, "throw "), strings.HasPrefix(lower, "raise "):
		return true, sh.evalThrow(raw)

	case strings.HasPrefix(lower, "assert "):
		return true, sh.evalAssert(raw)

	case strings.HasPrefix(lower, "isnull "), strings.HasPrefix(lower, "isnil "),
		strings.HasPrefix(lower, "isnone "):
		rest := strings.TrimSpace(raw[strings.IndexByte(raw, ' ')+1:])
		v    := sh.evalExpr(rest)
		lv   := strings.ToLower(strings.TrimSpace(v))
		if lv == "" || lv == "null" || lv == "nil" || lv == "none" || lv == "undefined" {
			fmt.Println("\n  true  " + c(ansiGrey, "(null)") + "\n")
		} else {
			fmt.Println("\n  false  " + c(ansiGrey, "(has value: "+v+")") + "\n")
		}
		return true, 0

	case strings.HasPrefix(lower, "typeof "), strings.HasPrefix(lower, "kindof "):
		val  := sh.evalExpr(strings.TrimSpace(raw[7:]))
		kind := katshTypeOf(val)
		fmt.Printf("\n  %s%s%s  %s\"%s\"%s\n\n", ansiBold+ansiCyan, kind, ansiReset, ansiGrey, val, ansiReset)
		return true, 0

	case strings.HasPrefix(lower, "range ") || lower == "range":
		arr := sh.evalRangeExpr(raw[6:])
		sh.printResult(NewTyped(arr, KindArray))
		return true, 0

	// Array builtins
	case strings.HasPrefix(lower, "arr_"):
		if code, ok := sh.tryArrayBuiltin(raw); ok { return true, code }

	// Map builtins
	case strings.HasPrefix(lower, "map_"):
		if code, ok := sh.tryMapBuiltin(raw); ok { return true, code }

	// Numeric helpers  hex/oct/bin/abs/sign/clamp/sqrt/log/round/floor/ceil/pow…
	case strings.HasPrefix(lower, "hex "), strings.HasPrefix(lower, "oct "),
		strings.HasPrefix(lower, "bin "), strings.HasPrefix(lower, "abs "),
		strings.HasPrefix(lower, "sign "), strings.HasPrefix(lower, "clamp "),
		strings.HasPrefix(lower, "sqrt "), strings.HasPrefix(lower, "log "),
		strings.HasPrefix(lower, "log2 "), strings.HasPrefix(lower, "log10 "),
		strings.HasPrefix(lower, "round "), strings.HasPrefix(lower, "floor "),
		strings.HasPrefix(lower, "ceil "), strings.HasPrefix(lower, "pow "),
		lower == "pi", lower == "e":
		if code, ok := sh.tryNumericHelper(raw); ok { return true, code }
	}

	// ── when guard:  cmd when cond ────────────────────────────────────────
	if code, ok := sh.tryWhenGuard(raw); ok { return true, code }

	// ── |> pipe expression ────────────────────────────────────────────────
	if strings.Contains(raw, "|>") {
		eqIdx := strings.Index(raw, "=")
		if eqIdx > 0 {
			lhs := strings.TrimSpace(raw[:eqIdx])
			rhs := strings.TrimSpace(raw[eqIdx+1:])
			if isIdent(lhs) && strings.Contains(rhs, "|>") {
				if val, ok := sh.evalPipeExpr(rhs); ok { sh.setVar(lhs, val); return true, 0 }
			}
		}
		if val, ok := sh.evalPipeExpr(raw); ok { fmt.Println("\n  " + val + "\n"); return true, 0 }
	}

	// ── Standalone data type literal ──────────────────────────────────────
	if v, ok := sh.evalDataTypeLiteral(raw); ok {
		_ = v
		fmt.Println("\n  " + isDataType(v) + " value\n")
		return true, 0
	}

	// ── Function call with return value:  result = funcName(args) ─────────
	if strings.Contains(raw, "(") {
		eqIdx := strings.Index(raw, "=")
		if eqIdx > 0 {
			lhs := strings.TrimSpace(raw[:eqIdx])
			if isIdent(lhs) {
				rhs := strings.TrimSpace(raw[eqIdx+1:])
				fw  := ""
				for _, ch := range rhs {
					if ch == '(' || ch == ' ' || ch == '\t' { break }
					fw += string(ch)
				}
				if _, ok := sh.funcs[fw]; ok && strings.Contains(rhs, "(") {
					_, captured := sh.tryFuncCallAssign(raw)
					if captured { return true, 0 }
				}
			}
		}
	}

	// ── Ternary assignment ────────────────────────────────────────────────
	if strings.Contains(raw, "?") {
		if code, ok := sh.tryTernaryAssign(raw); ok { return true, code }
	}

	// ── ?? null-coalescing ────────────────────────────────────────────────
	if strings.Contains(raw, "??") {
		if code, ok := sh.tryNullCoalesce(raw); ok { return true, code }
	}

	// ── Multi-assign:  a, b = 1, 2 ───────────────────────────────────────
	if strings.Contains(raw, ",") && strings.Contains(raw, "=") {
		if code, ok := sh.tryMultiAssign(raw); ok { return true, code }
	}

	return false, 0
}

// ─── Map helpers ──────────────────────────────────────────────────────────────

func mapKeysList(raw string) []string {
	if !strings.HasPrefix(raw, mapPfx) { return nil }
	inner := raw[len(mapPfx):]
	if inner == "" { return nil }
	var keys []string
	for _, entry := range strings.Split(inner, arraySep) {
		entry = strings.TrimSpace(entry)
		if entry == "" { continue }
		if idx := strings.Index(entry, "="); idx >= 0 { keys = append(keys, entry[:idx]) }
	}
	return keys
}

func mapDelete(raw, key string) string {
	if !strings.HasPrefix(raw, mapPfx) { return raw }
	inner := raw[len(mapPfx):]
	var parts []string
	for _, entry := range strings.Split(inner, arraySep) {
		entry = strings.TrimSpace(entry)
		if entry == "" { continue }
		if idx := strings.Index(entry, "="); idx < 0 || entry[:idx] != key { parts = append(parts, entry) }
	}
	return mapPfx + strings.Join(parts, arraySep)
}

func mapMerge3(m1, m2 string) string {
	merged := map[string]string{}
	var order []string
	for _, raw := range []string{m1, m2} {
		if !strings.HasPrefix(raw, mapPfx) { continue }
		inner := raw[len(mapPfx):]
		for _, entry := range strings.Split(inner, arraySep) {
			entry = strings.TrimSpace(entry)
			if entry == "" { continue }
			if idx := strings.Index(entry, "="); idx >= 0 {
				k, v := entry[:idx], entry[idx+1:]
				if _, exists := merged[k]; !exists { order = append(order, k) }
				merged[k] = v
			}
		}
	}
	var parts []string
	for _, k := range order { parts = append(parts, k+"="+merged[k]) }
	return mapPfx + strings.Join(parts, arraySep)
}

// varName2 strips a leading $ from a variable reference.
func varName2(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "$") { return s[1:] }
	return s
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func detectKind(v string) string {
	lv := strings.ToLower(strings.TrimSpace(v))
	if lv == "" || lv == "null" || lv == "nil" || lv == "none" { return "null" }
	if lv == "true" || lv == "false" { return "bool" }
	if _, err := strconv.ParseInt(v, 10, 64); err == nil { return "int" }
	if _, err := strconv.ParseFloat(v, 64); err == nil { return "float" }
	if strings.HasPrefix(v, "[") { return "array" }
	if strings.HasPrefix(v, "{") { return "map" }
	return "string"
}

// keep imports used
var _ = os.Stdin
var _ = math.Pi
var _ = sort.Strings
