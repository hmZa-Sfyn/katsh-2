package main

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ─────────────────────────────────────────────────────────────────────────────
//  String / Array / Number operations
//
//  Every operation works as BOTH:
//    a) a pipe operator:  "hello world" | split " "
//    b) a builtin command: split " " "hello world"
//
//  Type checking:
//    If a pipe receives the wrong kind (e.g. number piped to upper),
//    a descriptive E003 TypeError is printed with the actual vs expected type.
//
//  All operations are registered in:
//    applyPipe()   → called by the pipe engine
//    handleBuiltin() via handleStringOp() → called as commands
//
//  Supported operations (50):
//
//  String manipulation
//    upper         HELLO WORLD
//    lower         hello world
//    title         Hello World
//    trim          strip leading/trailing whitespace
//    ltrim         strip leading whitespace
//    rtrim         strip trailing whitespace
//    strip <chars> strip specific chars from both ends
//    len           length (chars)
//    reverse       reverse characters
//    repeat <N>    repeat string N times
//    replace <old> <new>   replace all occurrences
//    replace1 <old> <new>  replace first occurrence
//    sub <start> [end]     substring by index
//    pad <width> [char]    right-pad to width
//    lpad <width> [char]   left-pad to width
//    center <width> [char] center-pad to width
//
//  String tests (return "true" / "false")
//    startswith <prefix>
//    endswith <suffix>
//    contains <substring>
//    match <regex>         full regex match
//    isnum                 is numeric?
//    isalpha               all letters?
//    isalnum               letters or digits?
//    isspace               all whitespace?
//    isupper               all uppercase?
//    islower               all lowercase?
//
//  String splitting / joining
//    split [sep]           split on sep (default " ") → array result
//    lines                 split on newlines → array result
//    words                 split on whitespace → array result
//    chars                 split into individual characters → array result
//    join [sep]            join array items with sep (default " ")
//    concat <str>          append string to end
//    prepend <str>         prepend string to start
//
//  Array operations  (input must be array kind)
//    first                 first element
//    last                  last element
//    nth <N>               Nth element (0-based)
//    slice <start> [end]   sub-array
//    push <val>            append element → new array
//    pop                   remove last element → new array
//    flatten               flatten one level of nesting
//    arr_uniq              remove duplicate elements
//    arr_sort              sort elements
//    arr_reverse           reverse element order
//    arr_len               element count
//    arr_join [sep]        alias for join
//    arr_contains <val>    test membership → "true"/"false"
//    arr_map <expr>        apply expression to each element
//    arr_filter <expr>     keep elements where expr is true
//    arr_sum               numeric sum of elements
//    arr_min               minimum numeric element
//    arr_max               maximum numeric element
//    arr_avg               average of numeric elements
//
//  Number operations  (input must be number kind)
//    add <N>    +N
//    sub <N>    -N
//    mul <N>    *N
//    div <N>    /N
//    mod <N>    %N
//    pow <N>    ^N
//    abs        absolute value
//    ceil       round up
//    floor      round down
//    round [N]  round to N decimal places
//    sqrt       square root
//    negate     negate
//    hex        convert to hex string
//    oct        convert to octal string
//    bin        convert to binary string
// ─────────────────────────────────────────────────────────────────────────────

// ─── Type-checking helpers ───────────────────────────────────────────────────

const (
	KindString = "string"
	KindNumber = "number"
	KindArray  = "array"
	KindAny    = ""
)

// requireKind checks that r.ValueKind matches want.
// If want is "" (KindAny) anything passes.
// Returns a ShellError if the type is wrong.
func requireKind(r *Result, op, want string) *ShellError {
	if want == KindAny { return nil }
	got := r.ValueKind
	if got == "" { got = "text" }
	if got == want { return nil }
	return &ShellError{
		Code:    "E003",
		Kind:    "TypeError",
		Message: fmt.Sprintf("%q expects %s input, got %s", op, want, got),
		Source:  r.Text,
		Col:     -1,
		Hint:    fmt.Sprintf("pipe a %s value before %q — e.g. \"hello\" | %s", want, op, op),
		Fix:     fmt.Sprintf("\"your text here\" | %s", op),
	}
}

// resultText returns the text content of a result (works for both string and array).
func resultText(r *Result) string { return r.Text }

// ─── Pipe dispatch ────────────────────────────────────────────────────────────
// applyStringOp is called from applyPipe for all string/array/number ops.

func applyStringOp(r *Result, op string, args []string) (*Result, error) {
	switch op {

	// ── String manipulation ─────────────────────────────────────────────────

	case "upper":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		return NewTyped(strings.ToUpper(r.Text), KindString), nil

	case "lower":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		return NewTyped(strings.ToLower(r.Text), KindString), nil

	case "title":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		return NewTyped(toTitle(r.Text), KindString), nil

	case "trim":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		return NewTyped(strings.TrimSpace(r.Text), KindString), nil

	case "ltrim":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		return NewTyped(strings.TrimLeftFunc(r.Text, unicode.IsSpace), KindString), nil

	case "rtrim":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		return NewTyped(strings.TrimRightFunc(r.Text, unicode.IsSpace), KindString), nil

	case "strip":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		chars := " \t\n"
		if len(args) > 0 { chars = args[0] }
		return NewTyped(strings.Trim(r.Text, chars), KindString), nil

	case "len":
		if e := requireKind(r, op, KindString); e != nil {
			// Allow on array too
			if r.ValueKind == KindArray {
				items := splitArrayResult(r.Text)
				return NewTyped(strconv.Itoa(len(items)), KindNumber), nil
			}
			return nil, e
		}
		return NewTyped(strconv.Itoa(utf8.RuneCountInString(r.Text)), KindNumber), nil

	case "reverse":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		runes := []rune(r.Text)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 { runes[i], runes[j] = runes[j], runes[i] }
		return NewTyped(string(runes), KindString), nil

	case "repeat":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		n := 2
		if len(args) > 0 { fmt.Sscanf(args[0], "%d", &n) }
		return NewTyped(strings.Repeat(r.Text, n), KindString), nil

	case "replace":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		if len(args) < 2 { return nil, fmt.Errorf("replace: need <old> <new>") }
		return NewTyped(strings.ReplaceAll(r.Text, args[0], args[1]), KindString), nil

	case "replace1":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		if len(args) < 2 { return nil, fmt.Errorf("replace1: need <old> <new>") }
		return NewTyped(strings.Replace(r.Text, args[0], args[1], 1), KindString), nil

	case "sub":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		runes := []rune(r.Text)
		start, end := 0, len(runes)
		if len(args) > 0 { fmt.Sscanf(args[0], "%d", &start) }
		if len(args) > 1 { fmt.Sscanf(args[1], "%d", &end) }
		if start < 0 { start = localMax(0, len(runes)+start) }
		if end < 0   { end   = localMax(0, len(runes)+end) }
		if start > len(runes) { start = len(runes) }
		if end   > len(runes) { end   = len(runes) }
		if start > end { start, end = end, start }
		return NewTyped(string(runes[start:end]), KindString), nil

	case "pad":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		w, ch := 0, " "
		if len(args) > 0 { fmt.Sscanf(args[0], "%d", &w) }
		if len(args) > 1 { ch = args[1] }
		s := r.Text
		for utf8.RuneCountInString(s) < w { s += ch }
		return NewTyped(s, KindString), nil

	case "lpad":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		w, ch := 0, " "
		if len(args) > 0 { fmt.Sscanf(args[0], "%d", &w) }
		if len(args) > 1 { ch = args[1] }
		s := r.Text
		for utf8.RuneCountInString(s) < w { s = ch + s }
		return NewTyped(s, KindString), nil

	case "center":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		w, ch := 0, " "
		if len(args) > 0 { fmt.Sscanf(args[0], "%d", &w) }
		if len(args) > 1 { ch = args[1] }
		s := r.Text
		for utf8.RuneCountInString(s) < w {
			s = ch + s
			if utf8.RuneCountInString(s) < w { s = s + ch }
		}
		return NewTyped(s, KindString), nil

	// ── String tests ────────────────────────────────────────────────────────

	case "startswith":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		if len(args) == 0 { return nil, fmt.Errorf("startswith: need <prefix>") }
		return NewTyped(boolStr(strings.HasPrefix(r.Text, args[0])), KindString), nil

	case "endswith":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		if len(args) == 0 { return nil, fmt.Errorf("endswith: need <suffix>") }
		return NewTyped(boolStr(strings.HasSuffix(r.Text, args[0])), KindString), nil

	case "contains":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		if len(args) == 0 { return nil, fmt.Errorf("contains: need <substring>") }
		return NewTyped(boolStr(strings.Contains(r.Text, args[0])), KindString), nil

	case "match":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		if len(args) == 0 { return nil, fmt.Errorf("match: need <regex>") }
		rx, err := regexp.Compile(args[0])
		if err != nil { return nil, fmt.Errorf("match: invalid regex: %v", err) }
		return NewTyped(boolStr(rx.MatchString(r.Text)), KindString), nil

	case "isnum":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		_, err := strconv.ParseFloat(r.Text, 64)
		return NewTyped(boolStr(err == nil), KindString), nil

	case "isalpha":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		ok := len(r.Text) > 0
		for _, ch := range r.Text { if !unicode.IsLetter(ch) { ok = false; break } }
		return NewTyped(boolStr(ok), KindString), nil

	case "isalnum":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		ok := len(r.Text) > 0
		for _, ch := range r.Text { if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) { ok = false; break } }
		return NewTyped(boolStr(ok), KindString), nil

	case "isspace":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		ok := len(r.Text) > 0
		for _, ch := range r.Text { if !unicode.IsSpace(ch) { ok = false; break } }
		return NewTyped(boolStr(ok), KindString), nil

	case "isupper":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		ok := len(r.Text) > 0
		for _, ch := range r.Text { if unicode.IsLetter(ch) && !unicode.IsUpper(ch) { ok = false; break } }
		return NewTyped(boolStr(ok), KindString), nil

	case "islower":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		ok := len(r.Text) > 0
		for _, ch := range r.Text { if unicode.IsLetter(ch) && !unicode.IsLower(ch) { ok = false; break } }
		return NewTyped(boolStr(ok), KindString), nil

	// ── String split / join ─────────────────────────────────────────────────

	case "split":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		sep := " "
		if len(args) > 0 { sep = args[0] }
		parts := strings.Split(r.Text, sep)
		return makeArrayResult(parts), nil

	case "lines":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		parts := strings.Split(strings.TrimRight(r.Text, "\n"), "\n")
		return makeArrayResult(parts), nil

	case "words":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		return makeArrayResult(strings.Fields(r.Text)), nil

	case "chars":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		var cs []string
		for _, ch := range r.Text { cs = append(cs, string(ch)) }
		return makeArrayResult(cs), nil

	case "join":
		sep := " "
		if len(args) > 0 { sep = args[0] }
		// Accept array or raw text (newline-separated)
		items := toItems(r)
		return NewTyped(strings.Join(items, sep), KindString), nil

	case "concat":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		suffix := ""
		if len(args) > 0 { suffix = strings.Join(args, " ") }
		return NewTyped(r.Text+suffix, KindString), nil

	case "prepend":
		if e := requireKind(r, op, KindString); e != nil { return nil, e }
		prefix := ""
		if len(args) > 0 { prefix = strings.Join(args, " ") }
		return NewTyped(prefix+r.Text, KindString), nil

	// ── Array operations ────────────────────────────────────────────────────

	case "first":
		items := toItems(r)
		if len(items) == 0 { return NewTyped("", KindString), nil }
		return NewTyped(items[0], KindString), nil

	case "last":
		items := toItems(r)
		if len(items) == 0 { return NewTyped("", KindString), nil }
		return NewTyped(items[len(items)-1], KindString), nil

	case "nth":
		items := toItems(r)
		n := 0
		if len(args) > 0 { fmt.Sscanf(args[0], "%d", &n) }
		if n < 0 { n = len(items) + n }
		if n < 0 || n >= len(items) { return NewTyped("", KindString), nil }
		return NewTyped(items[n], KindString), nil

	case "slice":
		items := toItems(r)
		start, end := 0, len(items)
		if len(args) > 0 { fmt.Sscanf(args[0], "%d", &start) }
		if len(args) > 1 { fmt.Sscanf(args[1], "%d", &end) }
		if start < 0 { start = localMax(0, len(items)+start) }
		if end   < 0 { end   = localMax(0, len(items)+end) }
		if start > len(items) { start = len(items) }
		if end   > len(items) { end   = len(items) }
		return makeArrayResult(items[start:end]), nil

	case "push":
		items := toItems(r)
		val := ""
		if len(args) > 0 { val = strings.Join(args, " ") }
		return makeArrayResult(append(items, val)), nil

	case "pop":
		items := toItems(r)
		if len(items) == 0 { return makeArrayResult(nil), nil }
		return makeArrayResult(items[:len(items)-1]), nil

	case "arr_uniq":
		items := toItems(r)
		seen := map[string]bool{}
		var out []string
		for _, it := range items { if !seen[it] { seen[it]=true; out=append(out,it) } }
		return makeArrayResult(out), nil

	case "arr_sort":
		items := toItems(r)
		sorted := make([]string, len(items))
		copy(sorted, items)
		sort.Strings(sorted)
		return makeArrayResult(sorted), nil

	case "arr_reverse":
		items := toItems(r)
		rev := make([]string, len(items))
		for i, v := range items { rev[len(items)-1-i] = v }
		return makeArrayResult(rev), nil

	case "arr_len":
		items := toItems(r)
		return NewTyped(strconv.Itoa(len(items)), KindNumber), nil

	case "arr_join":
		sep := " "
		if len(args) > 0 { sep = args[0] }
		return NewTyped(strings.Join(toItems(r), sep), KindString), nil

	case "arr_contains":
		if len(args) == 0 { return nil, fmt.Errorf("arr_contains: need <value>") }
		items := toItems(r)
		for _, it := range items { if it == args[0] { return NewTyped("true", KindString), nil } }
		return NewTyped("false", KindString), nil

	case "arr_map":
		if len(args) == 0 { return nil, fmt.Errorf("arr_map: need <expr>  (use $it for element)") }
		expr := strings.Join(args, " ")
		items := toItems(r)
		out := make([]string, len(items))
		for i, it := range items {
			// Simple expression: replace $it with value
			expanded := strings.ReplaceAll(expr, "$it", it)
			out[i] = evalSimpleExpr(expanded)
		}
		return makeArrayResult(out), nil

	case "arr_filter":
		if len(args) == 0 { return nil, fmt.Errorf("arr_filter: need <expr>  (use $it for element)") }
		expr := strings.Join(args, " ")
		items := toItems(r)
		var out []string
		for _, it := range items {
			expanded := strings.ReplaceAll(expr, "$it", it)
			if evalSimpleCond(expanded) { out = append(out, it) }
		}
		return makeArrayResult(out), nil

	case "arr_sum":
		sum := 0.0
		for _, it := range toItems(r) { v, _ := strconv.ParseFloat(it, 64); sum += v }
		return NewTyped(fmtNum(sum), KindNumber), nil

	case "arr_min":
		items := toItems(r)
		if len(items) == 0 { return NewTyped("", KindNumber), nil }
		min, _ := strconv.ParseFloat(items[0], 64)
		for _, it := range items[1:] { v, _ := strconv.ParseFloat(it, 64); if v < min { min = v } }
		return NewTyped(fmtNum(min), KindNumber), nil

	case "arr_max":
		items := toItems(r)
		if len(items) == 0 { return NewTyped("", KindNumber), nil }
		max, _ := strconv.ParseFloat(items[0], 64)
		for _, it := range items[1:] { v, _ := strconv.ParseFloat(it, 64); if v > max { max = v } }
		return NewTyped(fmtNum(max), KindNumber), nil

	case "arr_avg":
		items := toItems(r)
		if len(items) == 0 { return NewTyped("", KindNumber), nil }
		sum := 0.0
		for _, it := range items { v, _ := strconv.ParseFloat(it, 64); sum += v }
		return NewTyped(fmtNum(sum/float64(len(items))), KindNumber), nil

	case "flatten":
		// Flatten items that themselves contain newlines
		var out []string
		for _, it := range toItems(r) {
			for _, sub := range strings.Split(it, "\n") {
				if s := strings.TrimSpace(sub); s != "" { out = append(out, s) }
			}
		}
		return makeArrayResult(out), nil

	// ── Number operations ───────────────────────────────────────────────────

	case "add":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		n := parseF(r.Text); delta := 0.0
		if len(args) > 0 { delta, _ = strconv.ParseFloat(args[0], 64) }
		return NewTyped(fmtNum(n+delta), KindNumber), nil

	case "sub_n":  // "sub" already taken by substring above, use sub_n for numeric subtract
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		n := parseF(r.Text); delta := 0.0
		if len(args) > 0 { delta, _ = strconv.ParseFloat(args[0], 64) }
		return NewTyped(fmtNum(n-delta), KindNumber), nil

	case "mul":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		n := parseF(r.Text); factor := 1.0
		if len(args) > 0 { factor, _ = strconv.ParseFloat(args[0], 64) }
		return NewTyped(fmtNum(n*factor), KindNumber), nil

	case "div":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		n := parseF(r.Text); divisor := 1.0
		if len(args) > 0 { divisor, _ = strconv.ParseFloat(args[0], 64) }
		if divisor == 0 { return nil, errDivZero(r.Text) }
		return NewTyped(fmtNum(n/divisor), KindNumber), nil

	case "mod":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		n := parseF(r.Text); m := 1.0
		if len(args) > 0 { m, _ = strconv.ParseFloat(args[0], 64) }
		if m == 0 { return nil, errDivZero(r.Text) }
		return NewTyped(fmtNum(math.Mod(n, m)), KindNumber), nil

	case "pow":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		n := parseF(r.Text); exp := 2.0
		if len(args) > 0 { exp, _ = strconv.ParseFloat(args[0], 64) }
		return NewTyped(fmtNum(math.Pow(n, exp)), KindNumber), nil

	case "abs":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		return NewTyped(fmtNum(math.Abs(parseF(r.Text))), KindNumber), nil

	case "ceil":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		return NewTyped(fmtNum(math.Ceil(parseF(r.Text))), KindNumber), nil

	case "floor":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		return NewTyped(fmtNum(math.Floor(parseF(r.Text))), KindNumber), nil

	case "round":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		places := 0
		if len(args) > 0 { fmt.Sscanf(args[0], "%d", &places) }
		factor := math.Pow(10, float64(places))
		return NewTyped(fmtNum(math.Round(parseF(r.Text)*factor)/factor), KindNumber), nil

	case "sqrt":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		n := parseF(r.Text)
		if n < 0 { return nil, fmt.Errorf("sqrt: cannot take square root of negative number") }
		return NewTyped(fmtNum(math.Sqrt(n)), KindNumber), nil

	case "negate":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		return NewTyped(fmtNum(-parseF(r.Text)), KindNumber), nil

	case "hex":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		n := int64(parseF(r.Text))
		return NewTyped(fmt.Sprintf("0x%x", n), KindString), nil

	case "oct":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		n := int64(parseF(r.Text))
		return NewTyped(fmt.Sprintf("0o%o", n), KindString), nil

	case "bin":
		if e := requireKind(r, op, KindNumber); e != nil { return nil, e }
		n := int64(parseF(r.Text))
		return NewTyped(fmt.Sprintf("0b%b", n), KindString), nil

	// ── Display helpers ─────────────────────────────────────────────────────

	case "echo", "print", "println":
		// Print the value — works on any kind
		out := r.Text
		if r.ValueKind == KindArray {
			items := splitArrayResult(r.Text)
			out = strings.Join(items, "\n")
		}
		return NewText(out), nil

	case "type":
		kind := r.ValueKind
		if kind == "" { kind = "text" }
		return NewTyped(kind, KindString), nil

	case "tonum":
		f, err := strconv.ParseFloat(strings.TrimSpace(r.Text), 64)
		if err != nil {
			return nil, &ShellError{
				Code: "E003", Kind: "TypeError",
				Message: fmt.Sprintf("cannot convert %q to number", r.Text),
				Source:  r.Text, Col: -1,
				Hint: "make sure the string contains only digits",
			}
		}
		return NewTyped(fmtNum(f), KindNumber), nil

	case "tostr":
		return NewTyped(r.Text, KindString), nil

	case "toarray":
		// Convert text to array (split on newlines)
		lines := strings.Split(strings.TrimRight(r.Text, "\n"), "\n")
		return makeArrayResult(lines), nil
	}

	return nil, nil // not a string op
}

// ─── Register string ops into the main applyPipe dispatcher ──────────────────

// isStringOp returns true if op is handled by applyStringOp.
func isStringOp(op string) bool {
	switch op {
	case "upper","lower","title","trim","ltrim","rtrim","strip",
		"len","reverse","repeat","replace","replace1","sub","pad","lpad","center",
		"startswith","endswith","contains","match",
		"isnum","isalpha","isalnum","isspace","isupper","islower",
		"split","lines","words","chars","join","concat","prepend",
		"first","last","nth","slice","push","pop","flatten",
		"arr_uniq","arr_sort","arr_reverse","arr_len","arr_join",
		"arr_contains","arr_map","arr_filter",
		"arr_sum","arr_min","arr_max","arr_avg",
		"add","sub_n","mul","div","mod","pow",
		"abs","ceil","floor","round","sqrt","negate","hex","oct","bin",
		"echo","print","println","type","tonum","tostr","toarray":
		return true
	}
	return false
}

// ─── Standalone command handler ───────────────────────────────────────────────
// handleStringOp handles string ops as standalone commands:
//   upper "hello"   →  HELLO
//   len "hello"     →  5
//   split " " "hello world"  →  ["hello", "world"]

func handleStringOp(sh *Shell, command string, args []string) (*Result, bool, error) {
	if !isStringOp(command) { return nil, false, nil }

	// Build input Result from args:
	//   last arg = input value, earlier args = op args
	// If only 1 arg and no input context: treat it as both
	if len(args) == 0 {
		return NewText(fmt.Sprintf("usage: %s <value> [args...]", command)), true, nil
	}

	// Last arg is the string/number to operate on
	input := args[len(args)-1]
	opArgs := args[:len(args)-1]

	// If command itself has flags that come first, swap:
	// e.g. split " " "hello world" → opArgs=[" "] input="hello world"
	// Already correct since last arg is always the input value.

	// Detect kind
	kind := KindString
	if _, err := strconv.ParseFloat(input, 64); err == nil {
		kind = KindNumber
	}

	r := NewTyped(input, kind)
	result, err := applyStringOp(r, command, opArgs)
	if err != nil {
		return nil, true, err
	}
	if result == nil {
		return NewText(""), true, nil
	}

	// Pretty-print array results as a table
	if result.ValueKind == KindArray {
		items := splitArrayResult(result.Text)
		cols := []string{"index", "value"}
		rows := make([]Row, len(items))
		for i, it := range items { rows[i] = Row{"index": strconv.Itoa(i), "value": it} }
		return NewTable(cols, rows), true, nil
	}

	return result, true, nil
}

// ─── Array result encoding ────────────────────────────────────────────────────
// Arrays flowing through the pipe are encoded as a special Result:
//   ValueKind = "array"
//   Text      = items joined by "\x01" (unit separator)

const pipeArraySep = "\x01"

func makeArrayResult(items []string) *Result {
	return &Result{
		Text:      strings.Join(items, pipeArraySep),
		IsTable:   false,
		ValueKind: KindArray,
	}
}

func splitArrayResult(s string) []string {
	if s == "" { return nil }
	return strings.Split(s, pipeArraySep)
}

// toItems converts any Result to a string slice.
func toItems(r *Result) []string {
	if r.ValueKind == KindArray {
		return splitArrayResult(r.Text)
	}
	if r.IsTable {
		var items []string
		for _, row := range r.Rows {
			for _, col := range r.Cols {
				items = append(items, row[col])
			}
		}
		return items
	}
	// Plain text: split on newlines
	lines := strings.Split(strings.TrimRight(r.Text, "\n"), "\n")
	var out []string
	for _, l := range lines { if t := strings.TrimSpace(l); t != "" { out = append(out, t) } }
	return out
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func boolStr(b bool) string {
	if b { return "true" }
	return "false"
}

func parseF(s string) float64 {
	f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return f
}

func toTitle(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) == 0 { continue }
		r, size := utf8.DecodeRuneInString(w)
		words[i] = string(unicode.ToUpper(r)) + strings.ToLower(w[size:])
	}
	return strings.Join(words, " ")
}

// evalSimpleExpr evaluates a very simple expression for arr_map.
// Supports: bare string, basic arithmetic tokens.
func evalSimpleExpr(expr string) string {
	expr = strings.TrimSpace(expr)
	for _, op := range []string{" + "," - "," * "," / "} {
		if idx := strings.LastIndex(expr, op); idx >= 0 {
			l, _ := strconv.ParseFloat(strings.TrimSpace(expr[:idx]), 64)
			r, _ := strconv.ParseFloat(strings.TrimSpace(expr[idx+3:]), 64)
			switch strings.TrimSpace(op) {
			case "+": return fmtNum(l+r)
			case "-": return fmtNum(l-r)
			case "*": return fmtNum(l*r)
			case "/": if r!=0{return fmtNum(l/r)}
			}
		}
	}
	return expr
}

// evalSimpleCond evaluates a simple boolean condition for arr_filter.
func evalSimpleCond(expr string) bool {
	expr = strings.TrimSpace(expr)
	for _, op := range []string{">=","<=","!=","==",">","<","~"} {
		if idx := strings.Index(expr, op); idx > 0 {
			lv := strings.TrimSpace(expr[:idx])
			rv := strings.TrimSpace(expr[idx+len(op):])
			lf, lNumErr := strconv.ParseFloat(lv, 64)
			rf, rNumErr := strconv.ParseFloat(rv, 64)
			switch op {
			case "==": return lv==rv
			case "!=": return lv!=rv
			case "~":  return strings.Contains(lv, rv)
			case ">":  if lNumErr==nil&&rNumErr==nil{return lf>rf}; return lv>rv
			case "<":  if lNumErr==nil&&rNumErr==nil{return lf<rf}; return lv<rv
			case ">=": if lNumErr==nil&&rNumErr==nil{return lf>=rf}; return lv>=rv
			case "<=": if lNumErr==nil&&rNumErr==nil{return lf<=rf}; return lv<=rv
			}
		}
	}
	switch strings.ToLower(expr) {
	case "true","1","yes": return true
	case "false","0","no","": return false
	}
	return expr != ""
}
