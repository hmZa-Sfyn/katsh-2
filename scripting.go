package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"
)

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

// evalScript — scripting entry point (called before builtins in execLine)
func (sh *Shell) evalScript(raw string) (bool, int) {
	raw = strings.TrimSpace(raw)
	if raw == "" { return true, 0 }

	// Comments
	if (strings.HasPrefix(raw, "#") && !strings.HasPrefix(raw, "#=")) ||
		strings.HasPrefix(raw, "//") || strings.HasPrefix(raw, "///") {
		return true, 0
	}

	// ── New features (scripting2.go) ──────────────────────────────────────
	if handled, code := sh.evalScript2(raw); handled {
		return true, code
	}

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
	case strings.HasPrefix(lower, "do ") || lower == "do{" || strings.HasPrefix(lower, "do{"):
		return true, sh.evalDo(raw)
	case strings.HasPrefix(lower, "repeat "):
		return true, sh.evalRepeat(raw)
	case strings.HasPrefix(lower, "match "):
		return true, sh.evalMatch(raw)
	case strings.HasPrefix(lower, "try ") || strings.HasPrefix(lower, "try{"):
		return true, sh.evalTry(raw)
	case strings.HasPrefix(lower, "func "):
		return true, sh.evalFuncDef(raw)
	case strings.HasPrefix(lower, "print ") || strings.HasPrefix(lower, "println "):
		prefix := 6
		if strings.HasPrefix(lower, "println ") { prefix = 8 }
		text := sh.expandBackticks(sh.expandVars(raw[prefix:]))
		text = expandStringExpr(sh, text)
		fmt.Println("  " + text)
		return true, 0
	case lower == "pass":
		return true, 0
	case strings.HasPrefix(lower, "local "):
		inner := strings.TrimSpace(raw[6:])
		if code, ok := sh.tryVarAssign(inner); ok { return true, code }
	}

	// && / ||
	if code, ok := sh.tryAndOr(raw); ok { return true, code }

	// Variable assignment
	if code, ok := sh.tryVarAssign(raw); ok { return true, code }

	// Increment/decrement
	if code, ok := sh.tryIncrDecr(raw); ok { return true, code }

	// User function call
	parts := tokenize(raw)
	if len(parts) > 0 {
		if fn, ok := sh.funcs[parts[0]]; ok {
			return true, sh.callUserFunc(fn, parts[1:], raw)
		}
	}
	return false, 0
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
	// Compound operators **= += -= *= /= %=
	for _, op := range []string{"**=", "+=", "-=", "*=", "/=", "%="} {
		if idx := strings.Index(raw, op); idx > 0 {
			name := strings.TrimSpace(raw[:idx])
			if !isIdent(name) { continue }
			rhs := strings.TrimSpace(raw[idx+len(op):])
			curF, _ := strconv.ParseFloat(sh.getVar(name), 64)
			rhsF, err := strconv.ParseFloat(sh.evalExpr(rhs), 64)
			if err != nil {
				_, col, src := FindErrorPos(raw, rhs)
				PrintError(&ShellError{Code:"E003",Kind:"TypeError",
					Message:fmt.Sprintf("cannot apply %s: %q is not a number",op,rhs),
					Source:src,Col:col,Hint:"Use a numeric value on the right-hand side"})
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
				idxStr := strings.TrimSpace(raw[lbIdx+1:rbIdx])
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
	if idx > 0 && (raw[idx-1]=='!' || raw[idx-1]=='<' || raw[idx-1]=='>') { return 0, false }
	if idx+1 < len(raw) && raw[idx+1] == '=' { return 0, false }
	name := strings.TrimSpace(raw[:idx])
	if !isIdent(name) { return 0, false }
	sh.setVar(name, sh.evalRHS(strings.TrimSpace(raw[idx+1:]), raw))
	return 0, true
}

func (sh *Shell) tryIncrDecr(raw string) (int, bool) {
	raw = strings.TrimSpace(raw)
	if strings.HasSuffix(raw, "++") { name := raw[:len(raw)-2]; if isIdent(strings.TrimSpace(name)) { sh.incrVar(strings.TrimSpace(name),1); return 0,true } }
	if strings.HasSuffix(raw, "--") { name := raw[:len(raw)-2]; if isIdent(strings.TrimSpace(name)) { sh.incrVar(strings.TrimSpace(name),-1); return 0,true } }
	if strings.HasPrefix(raw, "++") { name := raw[2:]; if isIdent(strings.TrimSpace(name)) { sh.incrVar(strings.TrimSpace(name),1); return 0,true } }
	if strings.HasPrefix(raw, "--") { name := raw[2:]; if isIdent(strings.TrimSpace(name)) { sh.incrVar(strings.TrimSpace(name),-1); return 0,true } }
	return 0, false
}
func (sh *Shell) incrVar(name string, d float64) {
	v, _ := strconv.ParseFloat(sh.getVar(name), 64)
	sh.setVar(name, fmtNum(v+d))
}

// evalRHS evaluates the right side of an assignment.
func (sh *Shell) evalRHS(rhs, src string) string {
	rhs = strings.TrimSpace(rhs)
	if rhs == "" { return "" }
	if strings.HasPrefix(strings.ToLower(rhs), "if ") { return sh.evalInlineIf(rhs, src) }
	if strings.HasPrefix(rhs,"[") && strings.HasSuffix(rhs,"]") { return sh.makeArray(sh.parseArrayLiteral(rhs)) }
	if strings.HasPrefix(rhs,"`") && strings.HasSuffix(rhs,"`") { return sh.runSubshell(rhs[1:len(rhs)-1]) }
	if strings.HasPrefix(rhs,`"`) && strings.HasSuffix(rhs,`"`) { return sh.interpolate(rhs[1:len(rhs)-1]) }
	if strings.HasPrefix(rhs,"'") && strings.HasSuffix(rhs,"'") { return rhs[1:len(rhs)-1] }
	return sh.evalExpr(rhs)
}

// interpolate expands $VAR, ${VAR}, ${VAR:-default}, ${#VAR}, and backticks.
func (sh *Shell) interpolate(s string) string {
	s = sh.expandBackticks(s)
	return os.Expand(s, func(key string) string {
		if strings.HasPrefix(key,"#") { return strconv.Itoa(len(sh.getVar(key[1:]))) }
		if strings.Contains(key,":-") {
			p := strings.SplitN(key,":-",2)
			if v := sh.getVar(p[0]); v != "" { return v }
			return p[1]
		}
		if strings.Contains(key,":+") {
			p := strings.SplitN(key,":+",2)
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
	inner := strings.TrimSpace(s[1:len(s)-1])
	if inner == "" { return nil }
	var out []string
	for _, item := range strings.Split(inner, ",") {
		out = append(out, sh.evalExpr(stripQuotes(strings.TrimSpace(item))))
	}
	return out
}
func (sh *Shell) arrayItems(name string) []string {
	raw := sh.vars[name]
	if !strings.HasPrefix(raw,"[") { return strings.Fields(raw) }
	content := raw[1:len(raw)-1]
	if content == "" { return nil }
	return strings.Split(content, arraySep)
}
func (sh *Shell) arrayGet(name, idx string) string {
	items := sh.arrayItems(name)
	if idx == "len" || idx == "#" { return strconv.Itoa(len(items)) }
	i, err := strconv.Atoi(idx)
	if err != nil { return "" }
	if i < 0 { i = len(items)+i }
	if i < 0 || i >= len(items) { return "" }
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
	items := sh.arrayItems(name)
	sh.vars[name] = sh.makeArray(append(items, val))
}

// ─── Expression evaluator ─────────────────────────────────────────────────────

func (sh *Shell) evalExpr(expr string) string {
	expr = strings.TrimSpace(expr)
	if expr == "" { return "" }

	// ${VAR...}
	if strings.HasPrefix(expr,"${") && strings.HasSuffix(expr,"}") {
		return sh.interpolate(expr)
	}
	// $VAR or $arr[N]
	if strings.HasPrefix(expr,"$") {
		name := expr[1:]
		if lbIdx := strings.Index(name,"["); lbIdx >= 0 {
			arrName := name[:lbIdx]
			idxStr := strings.TrimSuffix(name[lbIdx+1:],"]")
			return sh.arrayGet(arrName, sh.evalExpr(idxStr))
		}
		return sh.getVar(name)
	}
	// arr[N] or arr.len
	if lbIdx := strings.Index(expr,"["); lbIdx > 0 && strings.HasSuffix(expr,"]") {
		arrName := expr[:lbIdx]
		if isIdent(arrName) { return sh.arrayGet(arrName, sh.evalExpr(expr[lbIdx+1:len(expr)-1])) }
	}
	if strings.HasSuffix(expr,".len") {
		arrName := expr[:len(expr)-4]
		if isIdent(arrName) { return sh.arrayGet(arrName,"len") }
	}
	// bare var
	if isIdent(expr) {
		if v, ok := sh.vars[expr]; ok { return v }
	}
	// backtick
	if strings.HasPrefix(expr,"`") && strings.HasSuffix(expr,"`") {
		return sh.runSubshell(expr[1:len(expr)-1])
	}
	// string concat
	if strings.ContainsAny(expr,`"'`) || strings.Contains(expr,"` ") {
		return expandStringExpr(sh, expr)
	}
	// **
	if idx := strings.LastIndex(expr," ** "); idx >= 0 {
		base, _ := strconv.ParseFloat(sh.evalExpr(expr[:idx]),64)
		exp, _  := strconv.ParseFloat(sh.evalExpr(expr[idx+4:]),64)
		return fmtNum(math.Pow(base,exp))
	}
	// arithmetic
	if r, ok := tryArith(sh, expr); ok { return r }
	return expr
}

func tryArith(sh *Shell, expr string) (string, bool) {
	for _, op := range []string{"+","-","*","/","%"} {
		idx := strings.LastIndex(expr," "+op+" ")
		if idx < 0 { continue }
		lv, _ := strconv.ParseFloat(sh.evalExpr(expr[:idx]),64)
		rv, _ := strconv.ParseFloat(sh.evalExpr(expr[idx+3:]),64)
		var r float64
		switch op {
		case "+": r = lv+rv
		case "-": r = lv-rv
		case "*": r = lv*rv
		case "/": if rv==0{return "0",true}; r=lv/rv
		case "%": if rv==0{return "0",true}; r=math.Mod(lv,rv)
		}
		return fmtNum(r), true
	}
	return "", false
}

func expandStringExpr(sh *Shell, expr string) string {
	var sb strings.Builder
	for _, p := range splitOnDot(expr) {
		sb.WriteString(evalStringPart(sh, strings.TrimSpace(p)))
	}
	return sb.String()
}
func evalStringPart(sh *Shell, p string) string {
	if xIdx := strings.LastIndex(p,`"x`); xIdx > 0 {
		n := 0; fmt.Sscanf(p[xIdx+2:],"%d",&n)
		return strings.Repeat(stripQuotes(p[:xIdx+1]),n)
	}
	if strings.HasPrefix(p,"`") && strings.HasSuffix(p,"`") { return strings.TrimSpace(sh.runSubshell(p[1:len(p)-1])) }
	if strings.HasPrefix(p,"$") { return sh.getVar(p[1:]) }
	if strings.HasPrefix(p,`"`) || strings.HasPrefix(p,"'") { return sh.interpolate(stripQuotes(p)) }
	return sh.expandVars(p)
}
func splitOnDot(s string) []string {
	var parts []string; var cur strings.Builder
	inQ,inBt := false,false; qCh := rune(0)
	for _, ch := range s {
		switch {
		case inBt: cur.WriteRune(ch); if ch=='`'{inBt=false}
		case inQ:  cur.WriteRune(ch); if ch==qCh{inQ=false}
		case ch=='`': inBt=true; cur.WriteRune(ch)
		case ch=='"'||ch=='\'': inQ=true; qCh=ch; cur.WriteRune(ch)
		case ch=='.': parts=append(parts,cur.String()); cur.Reset()
		default: cur.WriteRune(ch)
		}
	}
	if cur.Len()>0 { parts=append(parts,cur.String()) }
	return parts
}
func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s)>=2 && ((s[0]=='"'&&s[len(s)-1]=='"')||(s[0]=='\''&&s[len(s)-1]=='\'')) { return s[1:len(s)-1] }
	return s
}

// ─── Condition evaluator ──────────────────────────────────────────────────────

func (sh *Shell) evalCond(cond string) bool {
	cond = strings.TrimSpace(cond)
	lower := strings.ToLower(cond)

	// not / !
	if strings.HasPrefix(lower,"not ") { return !sh.evalCond(cond[4:]) }
	if strings.HasPrefix(cond,"!") && !strings.HasPrefix(cond,"!=") { return !sh.evalCond(cond[1:]) }

	// and / or / && / ||
	if idx := findOutside(cond," and "); idx>=0 { return sh.evalCond(cond[:idx])&&sh.evalCond(cond[idx+5:]) }
	if idx := findOutside(cond," or ");  idx>=0 { return sh.evalCond(cond[:idx])||sh.evalCond(cond[idx+4:]) }
	if idx := findOutside(cond," && ");  idx>=0 { return sh.evalCond(cond[:idx])&&sh.evalCond(cond[idx+4:]) }
	if idx := findOutside(cond," || ");  idx>=0 { return sh.evalCond(cond[:idx])||sh.evalCond(cond[idx+4:]) }

	// Comparison operators
	for _, op := range []string{"!=",">=","<=","==","!~","~=",">","<","~"} {
		idx := strings.Index(cond,op)
		if idx <= 0 { continue }
		lv := strings.TrimSpace(sh.evalExpr(cond[:idx]))
		rv := strings.TrimSpace(stripQuotes(sh.evalExpr(cond[idx+len(op):])))
		lf,lNum := parseNum(lv); rf,rNum := parseNum(rv)
		switch op {
		case "==": return lv==rv
		case "!=": return lv!=rv
		case "~","~=": return strings.Contains(lv,rv)
		case "!~": return !strings.Contains(lv,rv)
		case ">":  if lNum&&rNum{return lf>rf}; return lv>rv
		case "<":  if lNum&&rNum{return lf<rf}; return lv<rv
		case ">=": if lNum&&rNum{return lf>=rf}; return lv>=rv
		case "<=": if lNum&&rNum{return lf<=rf}; return lv<=rv
		}
	}

	// Test flags: -z -n -f -d -e -r
	if strings.HasPrefix(cond,"-") {
		parts := strings.Fields(cond)
		if len(parts)==2 { return evalTestFlag(parts[0],sh.evalExpr(parts[1])) }
	}

	v := strings.ToLower(strings.TrimSpace(sh.evalExpr(cond)))
	switch v { case "true","1","yes": return true; case "false","0","no","": return false }
	return v != ""
}

func evalTestFlag(flag, val string) bool {
	switch flag {
	case "-z": return val==""
	case "-n": return val!=""
	case "-f": info,err:=os.Stat(val); return err==nil&&!info.IsDir()
	case "-d": info,err:=os.Stat(val); return err==nil&&info.IsDir()
	case "-e": _,err:=os.Stat(val); return err==nil
	case "-r": f,err:=os.Open(val); if err!=nil{return false}; f.Close(); return true
	}
	return false
}

// ─── if / elif / else ─────────────────────────────────────────────────────────

func (sh *Shell) evalIf(raw string) int {
	rest := strings.TrimSpace(raw[3:])
	type branch struct{ cond,body string }
	var branches []branch; var elsebody string

	for _, cl := range splitSemicolon(rest) {
		cl = strings.TrimSpace(cl); low := strings.ToLower(cl)
		switch {
		case strings.HasPrefix(low,"elif ") || strings.HasPrefix(low,"else if "):
			off := 5; if strings.HasPrefix(low,"else if "){off=8}
			cond,body := splitColon(cl[off:])
			branches = append(branches,branch{cond,extractBody(body)})
		case strings.HasPrefix(low,"else:") || strings.HasPrefix(low,"else ") || low=="else" || strings.HasPrefix(low,"else{"):
			after := strings.TrimSpace(cl[4:]); after = strings.TrimPrefix(after,":")
			elsebody = extractBody(strings.TrimSpace(after))
		default:
			cond,body := splitColon(cl)
			branches = append(branches,branch{cond,extractBody(body)})
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
	var cond,thenVal,elseVal string
	for _, p := range splitSemicolon(rest) {
		p = strings.TrimSpace(p); low := strings.ToLower(p)
		if strings.HasPrefix(low,"else:") || strings.HasPrefix(low,"else ") {
			elseVal = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(p[4:]),":"))
		} else {
			c,body := splitColon(p); if cond==""{cond=c;thenVal=body}
		}
	}
	if sh.evalCond(cond) { return sh.evalRHS(thenVal,src) }
	return sh.evalRHS(elseVal,src)
}

// ─── unless ───────────────────────────────────────────────────────────────────

func (sh *Shell) evalUnless(raw string) int {
	rest := strings.TrimSpace(raw[7:])
	cond,body := splitColon(rest)
	elsebody := ""
	if idx := strings.Index(body,"; else"); idx>=0 {
		parts := splitSemicolon(body); body=strings.TrimSpace(parts[0])
		if len(parts)>1 { after := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(parts[1]),"else:"),"else "); elsebody=extractBody(strings.TrimSpace(after)) }
	}
	if !sh.evalCond(cond) { return sh.execBodyLines(extractBody(body)) }
	if elsebody != "" { return sh.execBodyLines(elsebody) }
	return 0
}

// ─── match / case ─────────────────────────────────────────────────────────────

func (sh *Shell) evalMatch(raw string) int {
	rest := strings.TrimSpace(raw[6:])
	braceIdx := strings.Index(rest,"{")
	if braceIdx < 0 {
		colonIdx := colonOutsideBraces(rest)
		if colonIdx < 0 { PrintError(errSyntax("expected '{' in match",raw,6)); return 1 }
		return sh.runMatch(strings.TrimSpace(rest[:colonIdx]),rest[colonIdx+1:],raw)
	}
	return sh.runMatch(strings.TrimSpace(rest[:braceIdx]),rest[braceIdx:],raw)
}

func (sh *Shell) runMatch(subject, body, raw string) int {
	val := sh.evalExpr(subject)
	inner := extractBody(body)
	if strings.HasPrefix(inner,"{") { inner=inner[1:len(inner)-1] }

	for _, cl := range splitSemicolon(inner) {
		cl = strings.TrimSpace(cl); low := strings.ToLower(cl)
		if strings.HasPrefix(low,"default") {
			after := strings.TrimPrefix(strings.TrimSpace(cl[7:]),":"); return sh.execBodyLines(extractBody(strings.TrimSpace(after)))
		}
		if !strings.HasPrefix(low,"case ") { continue }
		caseExpr := cl[5:]; ci := colonOutsideBraces(caseExpr); if ci<0{continue}
		pattern := strings.TrimSpace(caseExpr[:ci]); casebody := strings.TrimSpace(caseExpr[ci+1:])
		if sh.matchPattern(val,pattern) { return sh.execBodyLines(extractBody(casebody)) }
	}
	return 0
}

func (sh *Shell) matchPattern(val, pattern string) bool {
	pattern = strings.TrimSpace(pattern)
	if strings.Contains(pattern,"|") {
		for _, alt := range strings.Split(pattern,"|") { if sh.matchPattern(val,strings.TrimSpace(alt)){return true} }
		return false
	}
	if pattern=="*"||pattern=="_" { return true }
	for _, op := range []string{">=","<=",">","<","!="} {
		if strings.HasPrefix(pattern,op) { return sh.evalCond(val+" "+op+" "+strings.TrimSpace(pattern[len(op):])) }
	}
	p := stripQuotes(pattern)
	if strings.Contains(p,"*") { return globMatch(p,val) }
	return val==p || val==pattern
}

func globMatch(pattern,s string) bool {
	if pattern=="*" { return true }
	parts := strings.Split(pattern,"*"); pos := 0
	for i,p := range parts {
		if p=="" { continue }
		idx := strings.Index(s[pos:],p); if idx<0{return false}
		if i==0&&idx!=0{return false}
		pos += idx+len(p)
	}
	return true
}

// ─── for loop ─────────────────────────────────────────────────────────────────

func (sh *Shell) evalFor(raw string) int {
	rest := strings.TrimSpace(raw[4:])
	inIdx := strings.Index(strings.ToLower(rest)," in ")
	if inIdx < 0 { PrintError(errSyntax("expected 'for <var> in <iterable>: <body>'",raw,0)); return 1 }
	varName := strings.TrimSpace(rest[:inIdx])
	after := strings.TrimSpace(rest[inIdx+4:])
	colonIdx := colonOutsideBraces(after)

	var iterExpr, body string
	if colonIdx >= 0 {
		iterExpr = strings.TrimSpace(after[:colonIdx])
		body = extractBody(strings.TrimSpace(after[colonIdx+1:]))
	} else if strings.HasPrefix(strings.TrimSpace(after),"[") {
		rb := strings.Index(after,"]"); if rb<0{PrintError(errSyntax("unclosed '[' in for loop",raw,0));return 1}
		iterExpr = strings.TrimSpace(after[:rb+1]); body = extractBody(strings.TrimSpace(after[rb+1:]))
	} else {
		PrintError(errSyntax("expected ':' or '{' after iterable in for",raw,0)); return 1
	}

	for _, item := range sh.evalIterable(iterExpr,raw) {
		sh.setVar(varName,item)
		code := sh.execBodyLines(body)
		if code==codeBreak { break }; if code==codeContinue { continue }; if code!=0 { return code }
	}
	sh.delVar(varName); return 0
}

func (sh *Shell) evalIterable(expr, src string) []string {
	expr = strings.TrimSpace(expr)
	if strings.HasPrefix(strings.ToLower(expr),"range") {
		inner := strings.Trim(expr[5:],"()")
		if strings.Contains(inner,"..") {
			p := strings.SplitN(inner,"..",2); a,b:=0,0
			fmt.Sscanf(sh.evalExpr(strings.TrimSpace(p[0])),"%d",&a)
			fmt.Sscanf(sh.evalExpr(strings.TrimSpace(p[1])),"%d",&b)
			return makeRange(a,b)
		}
		p := strings.SplitN(inner,",",2); a,b:=0,0
		fmt.Sscanf(sh.evalExpr(strings.TrimSpace(p[0])),"%d",&a)
		if len(p)>1 { fmt.Sscanf(sh.evalExpr(strings.TrimSpace(p[1])),"%d",&b) }
		return makeRange(a,b)
	}
	if strings.HasPrefix(expr,"[") && strings.HasSuffix(expr,"]") { return sh.parseArrayLiteral(expr) }
	if strings.HasPrefix(expr,"`") && strings.HasSuffix(expr,"`") {
		var lines []string
		for _, l := range strings.Split(sh.runSubshell(expr[1:len(expr)-1]),"\n") { if t:=strings.TrimSpace(l);t!=""{lines=append(lines,t)} }
		return lines
	}
	if strings.HasPrefix(expr,"$") {
		val := sh.getVar(expr[1:])
		if strings.HasPrefix(val,"[") { return sh.arrayItems(expr[1:]) }
		return strings.Fields(val)
	}
	return strings.Fields(expr)
}

func makeRange(a,b int) []string {
	if a>b { out:=make([]string,a-b); for i:=range out{out[i]=strconv.Itoa(a-i)}; return out }
	out:=make([]string,b-a); for i:=range out{out[i]=strconv.Itoa(a+i)}; return out
}

// ─── while ────────────────────────────────────────────────────────────────────

func (sh *Shell) evalWhile(raw string) int {
	rest := strings.TrimSpace(raw[6:])
	ci := colonOutsideBraces(rest)
	if ci<0 { PrintError(errSyntax("expected ':' in while",raw,6)); return 1 }
	cond := strings.TrimSpace(rest[:ci]); body := extractBody(strings.TrimSpace(rest[ci+1:]))
	for i:=0; i<1000000; i++ {
		if !sh.evalCond(cond) { break }
		code := sh.execBodyLines(body)
		if code==codeBreak{break}; if code==codeContinue{continue}; if code!=0{return code}
	}
	return 0
}

// ─── do { } while / until ─────────────────────────────────────────────────────

func (sh *Shell) evalDo(raw string) int {
	rest := strings.TrimSpace(raw[3:]) // strip "do "
	if strings.HasPrefix(rest,"{") { rest = strings.TrimSpace(rest) }
	body := extractBody(rest)
	remaining := strings.TrimSpace(rest[len(body):])
	low := strings.ToLower(remaining)
	isUntil := strings.HasPrefix(low,"until ")
	isWhile := strings.HasPrefix(low,"while ")
	if !isUntil && !isWhile { return sh.execBodyLines(body) }
	cond := strings.TrimSpace(remaining[6:])
	for i:=0; i<1000000; i++ {
		code := sh.execBodyLines(body)
		if code==codeBreak{break}; if code!=0&&code!=codeContinue{return code}
		c := sh.evalCond(cond)
		if isWhile && !c { break }
		if isUntil && c  { break }
	}
	return 0
}

// ─── repeat N: body ───────────────────────────────────────────────────────────

func (sh *Shell) evalRepeat(raw string) int {
	rest := strings.TrimSpace(raw[7:])
	ci := colonOutsideBraces(rest)
	if ci<0 { PrintError(errSyntax("expected ':' in repeat",raw,7)); return 1 }
	n:=0; fmt.Sscanf(sh.evalExpr(strings.TrimSpace(rest[:ci])),"%d",&n)
	body := extractBody(strings.TrimSpace(rest[ci+1:]))
	for i:=0; i<n; i++ {
		sh.setVar("_i",strconv.Itoa(i))
		code := sh.execBodyLines(body)
		if code==codeBreak{break}; if code==codeContinue{continue}; if code!=0{return code}
	}
	sh.delVar("_i"); return 0
}

// ─── try / catch / finally ───────────────────────────────────────────────────

func (sh *Shell) evalTry(raw string) int {
	rest := strings.TrimSpace(raw[4:]) // strip "try "
	if !strings.HasPrefix(rest,"{") { rest="{"+rest+"}" }
	body := extractBody(rest)
	remaining := strings.TrimSpace(rest[len(body):])
	catchBody,finallyBody,catchVar := "","",""

	low := strings.ToLower(remaining)
	if strings.HasPrefix(low,"catch") {
		after := strings.TrimSpace(remaining[5:])
		if after!=""&&after[0]!='{' {
			p:=strings.Fields(after)
			if isIdent(p[0]) { catchVar=p[0]; if len(p)>1{after=strings.Join(p[1:]," ")} }
		}
		catchBody = extractBody(after)
		remaining = strings.TrimSpace(after[len(catchBody):])
	}
	low = strings.ToLower(remaining)
	if strings.HasPrefix(low,"finally") {
		finallyBody = extractBody(strings.TrimSpace(remaining[7:]))
	}

	code := sh.execBodyLines(body)
	if code!=0 && catchBody!="" {
		if catchVar!="" { sh.setVar(catchVar,fmt.Sprintf("exit code %d",code)) }
		code = sh.execBodyLines(catchBody)
	}
	if finallyBody!="" { sh.execBodyLines(finallyBody) }
	return code
}

// ─── func def ────────────────────────────────────────────────────────────────

func (sh *Shell) evalFuncDef(raw string) int {
	rest := strings.TrimSpace(raw[5:])
	parenIdx := strings.IndexAny(rest,"( {")
	if parenIdx < 0 { PrintError(errSyntax("expected '(' after func name",raw,5)); return 1 }
	name := strings.TrimSpace(rest[:parenIdx])
	exported := strings.HasSuffix(name,"!"); if exported{name=name[:len(name)-1]}

	var params []string
	afterParams := rest
	if rest[parenIdx]=='(' {
		ci := strings.Index(rest,")"); if ci<0{PrintError(errSyntax("unclosed '(' in func",raw,parenIdx));return 1}
		for _, p := range strings.Split(rest[parenIdx+1:ci],",") {
			p = strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(p),"[]"),"...")
			if p!="" { params=append(params,p) }
		}
		afterParams = strings.TrimSpace(rest[ci+1:])
	}
	body := extractBody(afterParams)
	sh.funcs[name] = &UserFunc{Name:name,Params:params,Body:bodyLines(body),Exported:exported}
	fmt.Printf("  %s✔ func %s%s%s(%s)%s defined\n",ansiGreen,ansiBold+ansiCyan,name,ansiReset,strings.Join(params,", "),ansiReset)
	return 0
}

func (sh *Shell) callUserFunc(fn *UserFunc, args []string, src string) int {
	saved := make(map[string]string)
	for i,p := range fn.Params {
		saved[p] = sh.vars[p]
		if i<len(args){sh.vars[p]=sh.evalExpr(args[i])}else{sh.vars[p]=""}
	}
	if len(args)>len(fn.Params) { sh.vars["_args"]=strings.Join(args[len(fn.Params):]," ") }
	sh.vars["_argc"] = strconv.Itoa(len(args))
	sh.vars["_return"] = ""
	outerDefer := sh.deferStack
	sh.deferStack = nil
	body := strings.Join(fn.Body, "\n")
	code := sh.execBodyLinesWithGoto(body)
	if code == codeReturn { code = 0 }
	for i := len(sh.deferStack)-1; i >= 0; i-- { sh.execLine(sh.deferStack[i]) }
	sh.deferStack = outerDefer
	for p,v := range saved { sh.vars[p]=v }
	return code
}

// ─── Subshell / backtick expansion ──────────────────────────────────────────

func (sh *Shell) runSubshell(cmd string) string {
	cmd=strings.TrimSpace(cmd); if cmd==""{return ""}
	old:=sh.captureMode; sh.captureMode=true; sh.captureOut.Reset()
	sh.execLine(cmd)
	sh.captureMode=false; out:=strings.TrimRight(sh.captureOut.String(),"\n ")
	sh.captureOut.Reset(); sh.captureMode=old
	return out
}

func (sh *Shell) expandBackticks(s string) string {
	for {
		start:=strings.Index(s,"`"); if start<0{break}
		end:=strings.Index(s[start+1:],"`"); if end<0{break}
		end+=start+1
		s=s[:start]+sh.runSubshell(s[start+1:end])+s[end+1:]
	}
	return s
}

// ─── Body execution helpers ───────────────────────────────────────────────────

const (codeBreak=-1; codeContinue=-2; codeReturn=-3)

func (sh *Shell) execBodyLines(body string) int {
	for _, line := range bodyLines(body) {
		line=strings.TrimSpace(line); if line==""||strings.HasPrefix(line,"#")||strings.HasPrefix(line,"//"){continue}
		low:=strings.ToLower(line)
		if low=="break"{return codeBreak}; if low=="continue"{return codeContinue}; if low=="pass"{continue}
		if strings.HasPrefix(low,"return") { val:=strings.TrimSpace(line[6:]); if val!=""{sh.setVar("_return",sh.evalExpr(val))}; return codeReturn }
		code:=sh.execLine(line)
		if code==codeBreak||code==codeContinue||code==codeReturn{return code}
		if code!=0{return code}
	}
	return 0
}

func bodyLines(body string) []string {
	body=strings.TrimSpace(body)
	if strings.HasPrefix(body,"{")&&strings.HasSuffix(body,"}") { body=body[1:len(body)-1] }
	var lines []string
	for _, l := range strings.Split(body,";") {
		for _, s := range strings.Split(l,"\n") { if t:=strings.TrimSpace(s);t!=""{lines=append(lines,t)} }
	}
	return lines
}
func extractBody(s string) string {
	s=strings.TrimSpace(s); if !strings.HasPrefix(s,"{"){return s}
	depth:=0
	for i,ch := range s {
		if ch=='{'{depth++}; if ch=='}'{depth--;if depth==0{return s[:i+1]}}
	}
	return s
}
func splitColon(s string) (string,string) {
	idx:=colonOutsideBraces(s); if idx<0{return s,""}
	return strings.TrimSpace(s[:idx]),strings.TrimSpace(s[idx+1:])
}
func colonOutsideBraces(s string) int {
	depth,inQ:=0,false; qCh:=rune(0)
	for i,ch := range s {
		if inQ{if ch==qCh{inQ=false};continue}
		if ch=='"'||ch=='\''{inQ=true;qCh=ch;continue}
		if ch=='{'||ch=='('||ch=='['{depth++;continue}
		if ch=='}'||ch==')'||ch==']'{depth--;continue}
		if ch==':'&&depth==0{return i}
	}
	return -1
}
func splitSemicolon(s string) []string {
	var parts []string; var cur strings.Builder; depth,inQ:=0,false; qCh:=rune(0)
	for _,ch:=range s {
		if inQ{cur.WriteRune(ch);if ch==qCh{inQ=false};continue}
		if ch=='"'||ch=='\''{inQ=true;qCh=ch;cur.WriteRune(ch);continue}
		if ch=='{'||ch=='('{depth++;cur.WriteRune(ch);continue}
		if ch=='}'||ch==')'{depth--;cur.WriteRune(ch);continue}
		if ch==';'&&depth==0{if t:=strings.TrimSpace(cur.String());t!=""{parts=append(parts,t)};cur.Reset()}else{cur.WriteRune(ch)}
	}
	if t:=strings.TrimSpace(cur.String());t!=""{parts=append(parts,t)}
	return parts
}

// ─── Variable helpers ─────────────────────────────────────────────────────────

func (sh *Shell) getVar(name string) string {
	if v,ok:=sh.vars[name];ok{return v}
	return os.Getenv(name)
}
func (sh *Shell) setVar(name,val string) { sh.vars[name]=val }
func (sh *Shell) delVar(name string) { delete(sh.vars,name) }

// ─── Misc helpers ─────────────────────────────────────────────────────────────

func isIdent(s string) bool {
	if s==""{ return false }
	for i,ch:=range s {
		if i==0&&!unicode.IsLetter(ch)&&ch!='_'{return false}
		if i>0&&!unicode.IsLetter(ch)&&!unicode.IsDigit(ch)&&ch!='_'{return false}
	}
	return true
}
func parseNum(s string) (float64,bool) { f,err:=strconv.ParseFloat(s,64); return f,err==nil }
func fmtNum(f float64) string {
	if f==math.Trunc(f){return strconv.FormatInt(int64(f),10)}
	return strconv.FormatFloat(f,'f',-1,64)
}
