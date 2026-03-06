package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
//  Advanced data types and manipulation
//
//  New types (stored as typed strings in sh.vars):
//
//  map    — key:value pairs       m = map { "a":1 "b":2 }
//  set    — unique values         s = set { 1 2 3 }
//  stack  — LIFO collection       s = stack {}
//  queue  — FIFO collection       q = queue {}
//  tuple  — immutable record      t = (1, "hello", true)
//  matrix — 2D numeric grid       M = matrix(3,3)
//
//  Encoding (all as strings in sh.vars):
//   map    →  __map__key1\x02val1\x03key2\x02val2\x03...
//   set    →  __set__item1\x03item2\x03...
//   stack  →  __stack__item1\x03item2\x03...   (top = last)
//   queue  →  __queue__item1\x03item2\x03...   (front = first)
//   tuple  →  __tuple__item1\x03item2\x03...
//   matrix →  __matrix__rows\x02cols\x02v00\x03v01\x03...
//
//  All operations are available as:
//    - pipe operators:   $m | map_get "key"
//    - builtin commands: map_get $m "key"
//    - scripting syntax: x = $m["key"]   (map only)
// ─────────────────────────────────────────────────────────────────────────────

const (
	dtSep  = "\x03" // item separator
	dtKV   = "\x02" // key-value separator within a map entry
	mapPfx  = "__map__"
	setPfx  = "__set__"
	stackPfx= "__stack__"
	queuePfx= "__queue__"
	tupPfx  = "__tuple__"
	matPfx  = "__matrix__"
)

// ─── Type constructors ────────────────────────────────────────────────────────

func encodeMap(m map[string]string) string {
	var parts []string
	for k, v := range m { parts = append(parts, k+dtKV+v) }
	sort.Strings(parts)
	return mapPfx + strings.Join(parts, dtSep)
}

func decodeMap(s string) map[string]string {
	m := map[string]string{}
	s = strings.TrimPrefix(s, mapPfx)
	if s == "" { return m }
	for _, part := range strings.Split(s, dtSep) {
		kv := strings.SplitN(part, dtKV, 2)
		if len(kv) == 2 { m[kv[0]] = kv[1] }
	}
	return m
}

func encodeSet(items []string) string {
	seen := map[string]bool{}
	var unique []string
	for _, it := range items {
		if !seen[it] { seen[it] = true; unique = append(unique, it) }
	}
	return setPfx + strings.Join(unique, dtSep)
}

func decodeSet(s string) []string {
	s = strings.TrimPrefix(s, setPfx)
	if s == "" { return nil }
	return strings.Split(s, dtSep)
}

func encodeStack(items []string) string { return stackPfx + strings.Join(items, dtSep) }
func decodeStack(s string) []string {
	s = strings.TrimPrefix(s, stackPfx)
	if s == "" { return nil }
	return strings.Split(s, dtSep)
}

func encodeQueue(items []string) string { return queuePfx + strings.Join(items, dtSep) }
func decodeQueue(s string) []string {
	s = strings.TrimPrefix(s, queuePfx)
	if s == "" { return nil }
	return strings.Split(s, dtSep)
}

func encodeTuple(items []string) string { return tupPfx + strings.Join(items, dtSep) }
func decodeTuple(s string) []string {
	s = strings.TrimPrefix(s, tupPfx)
	if s == "" { return nil }
	return strings.Split(s, dtSep)
}

// Matrix: encoded as  __matrix__rows\x02cols\x02v00\x03v01\x03v10\x03...
func encodeMatrix(rows, cols int, data []float64) string {
	parts := []string{strconv.Itoa(rows), strconv.Itoa(cols)}
	for _, v := range data { parts = append(parts, fmtNum(v)) }
	return matPfx + strings.Join(parts[:2], dtKV) + dtKV + strings.Join(parts[2:], dtSep)
}

func decodeMatrix(s string) (rows, cols int, data []float64) {
	s = strings.TrimPrefix(s, matPfx)
	kv := strings.SplitN(s, dtKV, 3)
	if len(kv) < 2 { return 0, 0, nil }
	rows, _ = strconv.Atoi(kv[0])
	cols, _ = strconv.Atoi(kv[1])
	if len(kv) > 2 {
		for _, tok := range strings.Split(kv[2], dtSep) {
			v, _ := strconv.ParseFloat(tok, 64)
			data = append(data, v)
		}
	}
	for len(data) < rows*cols { data = append(data, 0) }
	return
}

// isDataType returns the type tag or "" if s is not an advanced data type.
func isDataType(s string) string {
	switch {
	case strings.HasPrefix(s, mapPfx):   return "map"
	case strings.HasPrefix(s, setPfx):   return "set"
	case strings.HasPrefix(s, stackPfx): return "stack"
	case strings.HasPrefix(s, queuePfx): return "queue"
	case strings.HasPrefix(s, tupPfx):   return "tuple"
	case strings.HasPrefix(s, matPfx):   return "matrix"
	}
	return ""
}

// ─── Scripting syntax integration ────────────────────────────────────────────
// evalScript recognises:
//   m = map { "a":1  "b":2 }
//   s = set { 1 2 3 }
//   st = stack {}
//   q  = queue {}
//   t  = (1, "hello", true)
//   M  = matrix(3, 3)
//   x  = $m["key"]

func (sh *Shell) evalDataTypeLiteral(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	lower := strings.ToLower(raw)

	// map { ... }
	if strings.HasPrefix(lower, "map") {
		rest := strings.TrimSpace(raw[3:])
		body := extractBody(rest)
		inner := body
		if strings.HasPrefix(inner, "{") { inner = inner[1:len(inner)-1] }
		m := map[string]string{}
		// parse "key":value or key=value pairs
		for _, tok := range tokenizeKVPairs(inner) {
			if idx := strings.Index(tok, ":"); idx > 0 {
				k := strings.TrimSpace(stripQuotes(tok[:idx]))
				v := strings.TrimSpace(sh.evalExpr(stripQuotes(tok[idx+1:])))
				m[k] = v
			} else if idx := strings.Index(tok, "="); idx > 0 {
				k := strings.TrimSpace(tok[:idx])
				v := strings.TrimSpace(sh.evalExpr(tok[idx+1:]))
				m[k] = v
			}
		}
		return encodeMap(m), true
	}

	// set { ... }
	if strings.HasPrefix(lower, "set") {
		rest := strings.TrimSpace(raw[3:])
		body := extractBody(rest)
		inner := body
		if strings.HasPrefix(inner, "{") { inner = inner[1:len(inner)-1] }
		var items []string
		for _, tok := range strings.Fields(inner) {
			items = append(items, sh.evalExpr(stripQuotes(tok)))
		}
		return encodeSet(items), true
	}

	// stack { ... }  or  stack()
	if strings.HasPrefix(lower, "stack") {
		rest := strings.TrimSpace(raw[5:])
		var items []string
		if strings.HasPrefix(rest, "{") {
			inner := rest[1:]
			if idx := strings.LastIndex(inner, "}"); idx >= 0 { inner = inner[:idx] }
			for _, tok := range strings.Fields(inner) { items = append(items, sh.evalExpr(stripQuotes(tok))) }
		}
		return encodeStack(items), true
	}

	// queue { ... }
	if strings.HasPrefix(lower, "queue") {
		rest := strings.TrimSpace(raw[5:])
		var items []string
		if strings.HasPrefix(rest, "{") {
			inner := rest[1:]
			if idx := strings.LastIndex(inner, "}"); idx >= 0 { inner = inner[:idx] }
			for _, tok := range strings.Fields(inner) { items = append(items, sh.evalExpr(stripQuotes(tok))) }
		}
		return encodeQueue(items), true
	}

	// tuple: (1, "hello", true)
	if strings.HasPrefix(raw, "(") && strings.HasSuffix(raw, ")") {
		inner := raw[1 : len(raw)-1]
		var items []string
		for _, tok := range strings.Split(inner, ",") {
			items = append(items, sh.evalExpr(strings.TrimSpace(stripQuotes(strings.TrimSpace(tok)))))
		}
		return encodeTuple(items), true
	}

	// matrix(rows, cols)  or  matrix(rows, cols, fillVal)
	if strings.HasPrefix(lower, "matrix") {
		rest := strings.TrimSpace(raw[6:])
		if strings.HasPrefix(rest, "(") {
			end := strings.Index(rest, ")")
			if end < 0 { end = len(rest) }
			inner := rest[1:end]
			parts := strings.Split(inner, ",")
			rows, cols, fill := 0, 0, 0.0
			if len(parts) >= 1 { fmt.Sscanf(sh.evalExpr(strings.TrimSpace(parts[0])), "%d", &rows) }
			if len(parts) >= 2 { fmt.Sscanf(sh.evalExpr(strings.TrimSpace(parts[1])), "%d", &cols) }
			if len(parts) >= 3 { fill, _ = strconv.ParseFloat(sh.evalExpr(strings.TrimSpace(parts[2])), 64) }
			data := make([]float64, rows*cols)
			for i := range data { data[i] = fill }
			return encodeMatrix(rows, cols, data), true
		}
	}

	return "", false
}

// tokenizeKVPairs splits a map body into individual k:v tokens.
func tokenizeKVPairs(s string) []string {
	var parts []string
	var cur strings.Builder
	inQ := false
	qCh := rune(0)
	for _, ch := range s {
		switch {
		case inQ:
			cur.WriteRune(ch)
			if ch == qCh { inQ = false }
		case ch == '"' || ch == '\'':
			inQ = true; qCh = ch; cur.WriteRune(ch)
		case ch == ' ' || ch == '\t' || ch == '\n':
			if t := strings.TrimSpace(cur.String()); t != "" {
				parts = append(parts, t)
				cur.Reset()
			}
		default:
			cur.WriteRune(ch)
		}
	}
	if t := strings.TrimSpace(cur.String()); t != "" { parts = append(parts, t) }
	return parts
}

// ─── Map operations ───────────────────────────────────────────────────────────

func mapGet(encoded, key string) string {
	m := decodeMap(encoded)
	return m[key]
}

func mapSet(encoded, key, val string) string {
	m := decodeMap(encoded)
	m[key] = val
	return encodeMap(m)
}

func mapDel(encoded, key string) string {
	m := decodeMap(encoded)
	delete(m, key)
	return encodeMap(m)
}

func mapHas(encoded, key string) string {
	m := decodeMap(encoded)
	_, ok := m[key]
	return boolStr(ok)
}

func mapKeys(encoded string) *Result {
	m := decodeMap(encoded)
	keys := make([]string, 0, len(m))
	for k := range m { keys = append(keys, k) }
	sort.Strings(keys)
	return makeArrayResult(keys)
}

func mapValues(encoded string) *Result {
	m := decodeMap(encoded)
	keys := make([]string, 0, len(m))
	for k := range m { keys = append(keys, k) }
	sort.Strings(keys)
	vals := make([]string, len(keys))
	for i, k := range keys { vals[i] = m[k] }
	return makeArrayResult(vals)
}

func mapLen(encoded string) string {
	return strconv.Itoa(len(decodeMap(encoded)))
}

func mapToTable(encoded string) *Result {
	m := decodeMap(encoded)
	keys := make([]string, 0, len(m))
	for k := range m { keys = append(keys, k) }
	sort.Strings(keys)
	cols := []string{"key", "value"}
	rows := make([]Row, len(keys))
	for i, k := range keys { rows[i] = Row{"key": k, "value": m[k]} }
	return NewTable(cols, rows)
}

func mapMerge(a, b string) string {
	ma := decodeMap(a)
	mb := decodeMap(b)
	for k, v := range mb { ma[k] = v }
	return encodeMap(ma)
}

// ─── Set operations ───────────────────────────────────────────────────────────

func setAdd(encoded, val string) string {
	items := decodeSet(encoded)
	for _, it := range items { if it == val { return encoded } }
	return encodeSet(append(items, val))
}

func setRemove(encoded, val string) string {
	items := decodeSet(encoded)
	var out []string
	for _, it := range items { if it != val { out = append(out, it) } }
	return encodeSet(out)
}

func setHas(encoded, val string) string {
	for _, it := range decodeSet(encoded) { if it == val { return "true" } }
	return "false"
}

func setUnion(a, b string) string {
	seen := map[string]bool{}
	var items []string
	for _, it := range decodeSet(a) { if !seen[it] { seen[it]=true; items=append(items,it) } }
	for _, it := range decodeSet(b) { if !seen[it] { seen[it]=true; items=append(items,it) } }
	return encodeSet(items)
}

func setIntersect(a, b string) string {
	inB := map[string]bool{}
	for _, it := range decodeSet(b) { inB[it] = true }
	var items []string
	for _, it := range decodeSet(a) { if inB[it] { items = append(items, it) } }
	return encodeSet(items)
}

func setDiff(a, b string) string {
	inB := map[string]bool{}
	for _, it := range decodeSet(b) { inB[it] = true }
	var items []string
	for _, it := range decodeSet(a) { if !inB[it] { items = append(items, it) } }
	return encodeSet(items)
}

func setToTable(encoded string) *Result {
	items := decodeSet(encoded)
	cols := []string{"index", "value"}
	rows := make([]Row, len(items))
	for i, it := range items { rows[i] = Row{"index": strconv.Itoa(i), "value": it} }
	return NewTable(cols, rows)
}

// ─── Stack operations ─────────────────────────────────────────────────────────

func stackPush(encoded, val string) string {
	return encodeStack(append(decodeStack(encoded), val))
}

func stackPop(encoded string) (string, string) { // returns (newStack, poppedVal)
	items := decodeStack(encoded)
	if len(items) == 0 { return encoded, "" }
	return encodeStack(items[:len(items)-1]), items[len(items)-1]
}

func stackPeek(encoded string) string {
	items := decodeStack(encoded)
	if len(items) == 0 { return "" }
	return items[len(items)-1]
}

func stackLen(encoded string) string { return strconv.Itoa(len(decodeStack(encoded))) }

func stackToTable(encoded string) *Result {
	items := decodeStack(encoded)
	cols := []string{"depth", "value"}
	rows := make([]Row, len(items))
	for i, it := range items {
		rows[len(items)-1-i] = Row{"depth": strconv.Itoa(i), "value": it}
	}
	return NewTable(cols, rows)
}

// ─── Queue operations ─────────────────────────────────────────────────────────

func queueEnqueue(encoded, val string) string {
	return encodeQueue(append(decodeQueue(encoded), val))
}

func queueDequeue(encoded string) (string, string) { // returns (newQueue, dequeuedVal)
	items := decodeQueue(encoded)
	if len(items) == 0 { return encoded, "" }
	return encodeQueue(items[1:]), items[0]
}

func queuePeek(encoded string) string {
	items := decodeQueue(encoded)
	if len(items) == 0 { return "" }
	return items[0]
}

func queueLen(encoded string) string { return strconv.Itoa(len(decodeQueue(encoded))) }

func queueToTable(encoded string) *Result {
	items := decodeQueue(encoded)
	cols := []string{"position", "value"}
	rows := make([]Row, len(items))
	for i, it := range items { rows[i] = Row{"position": strconv.Itoa(i), "value": it} }
	return NewTable(cols, rows)
}

// ─── Tuple operations ─────────────────────────────────────────────────────────

func tupleGet(encoded string, idx int) string {
	items := decodeTuple(encoded)
	if idx < 0 { idx = len(items) + idx }
	if idx < 0 || idx >= len(items) { return "" }
	return items[idx]
}

func tupleLen(encoded string) string { return strconv.Itoa(len(decodeTuple(encoded))) }

func tupleToTable(encoded string) *Result {
	items := decodeTuple(encoded)
	cols := []string{"index", "value"}
	rows := make([]Row, len(items))
	for i, it := range items { rows[i] = Row{"index": strconv.Itoa(i), "value": it} }
	return NewTable(cols, rows)
}

// ─── Matrix operations ────────────────────────────────────────────────────────

func matrixGet(encoded string, r, c int) string {
	rows, cols, data := decodeMatrix(encoded)
	if r < 0 { r = rows + r }
	if c < 0 { c = cols + c }
	if r < 0 || r >= rows || c < 0 || c >= cols { return "" }
	return fmtNum(data[r*cols+c])
}

func matrixSet(encoded string, r, c int, val float64) string {
	rows, cols, data := decodeMatrix(encoded)
	if r < 0 { r = rows + r }
	if c < 0 { c = cols + c }
	if r >= 0 && r < rows && c >= 0 && c < cols { data[r*cols+c] = val }
	return encodeMatrix(rows, cols, data)
}

func matrixAdd(a, b string) string {
	rows, cols, da := decodeMatrix(a)
	_, _, db := decodeMatrix(b)
	out := make([]float64, len(da))
	for i := range da {
		v := da[i]
		if i < len(db) { v += db[i] }
		out[i] = v
	}
	return encodeMatrix(rows, cols, out)
}

func matrixMul(a, b string) string {
	rA, cA, da := decodeMatrix(a)
	rB, cB, db := decodeMatrix(b)
	if cA != rB { return a } // dimension mismatch
	out := make([]float64, rA*cB)
	for i := 0; i < rA; i++ {
		for j := 0; j < cB; j++ {
			sum := 0.0
			for k := 0; k < cA; k++ { sum += da[i*cA+k] * db[k*cB+j] }
			out[i*cB+j] = sum
		}
	}
	return encodeMatrix(rA, cB, out)
}

func matrixTranspose(encoded string) string {
	rows, cols, data := decodeMatrix(encoded)
	out := make([]float64, rows*cols)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ { out[c*rows+r] = data[r*cols+c] }
	}
	return encodeMatrix(cols, rows, out)
}

func matrixToTable(encoded string) *Result {
	rows, cols, data := decodeMatrix(encoded)
	colNames := []string{"row"}
	for c := 0; c < cols; c++ { colNames = append(colNames, strconv.Itoa(c)) }
	var tableRows []Row
	for r := 0; r < rows; r++ {
		row := Row{"row": strconv.Itoa(r)}
		for c := 0; c < cols; c++ { row[strconv.Itoa(c)] = fmtNum(data[r*cols+c]) }
		tableRows = append(tableRows, row)
	}
	return NewTable(colNames, tableRows)
}

func matrixDet(encoded string) string {
	rows, cols, data := decodeMatrix(encoded)
	if rows != cols { return "NaN" }
	return fmtNum(det(data, rows))
}

func det(m []float64, n int) float64 {
	if n == 1 { return m[0] }
	if n == 2 { return m[0]*m[3] - m[1]*m[2] }
	result := 0.0
	sub := make([]float64, (n-1)*(n-1))
	for c := 0; c < n; c++ {
		si := 0
		for r := 1; r < n; r++ {
			for cc := 0; cc < n; cc++ {
				if cc == c { continue }
				sub[si] = m[r*n+cc]; si++
			}
		}
		sign := math.Pow(-1, float64(c))
		result += sign * m[c] * det(sub, n-1)
	}
	return result
}

// ─── Builtin command handler ──────────────────────────────────────────────────
// All data type commands routed through handleDataType.

func handleDataType(sh *Shell, command string, args []string) (*Result, bool, error) {
	switch command {

	// ── map commands ─────────────────────────────────────────────────────
	case "map_new":
		return NewTyped(encodeMap(map[string]string{}), KindString), true, nil

	case "map_set":
		// map_set <varname> <key> <value>
		if len(args) < 3 { return nil, true, fmt.Errorf("map_set: usage: map_set <var> <key> <value>") }
		v := sh.vars[args[0]]
		v = mapSet(v, args[1], sh.evalExpr(args[2]))
		sh.setVar(args[0], v)
		return NewText(""), true, nil

	case "map_get":
		// map_get <varname|encoded> <key>
		if len(args) < 2 { return nil, true, fmt.Errorf("map_get: usage: map_get <var> <key>") }
		encoded := resolveVar(sh, args[0])
		return NewTyped(mapGet(encoded, args[1]), KindString), true, nil

	case "map_del":
		if len(args) < 2 { return nil, true, fmt.Errorf("map_del: usage: map_del <var> <key>") }
		v := sh.vars[args[0]]
		sh.setVar(args[0], mapDel(v, args[1]))
		return NewText(""), true, nil

	case "map_has":
		if len(args) < 2 { return nil, true, fmt.Errorf("map_has: usage: map_has <var> <key>") }
		encoded := resolveVar(sh, args[0])
		return NewTyped(mapHas(encoded, args[1]), KindString), true, nil

	case "map_keys":
		if len(args) < 1 { return nil, true, fmt.Errorf("map_keys: usage: map_keys <var>") }
		return mapKeys(resolveVar(sh, args[0])), true, nil

	case "map_values":
		if len(args) < 1 { return nil, true, fmt.Errorf("map_values: usage: map_values <var>") }
		return mapValues(resolveVar(sh, args[0])), true, nil

	case "map_len":
		if len(args) < 1 { return nil, true, fmt.Errorf("map_len: usage: map_len <var>") }
		return NewTyped(mapLen(resolveVar(sh, args[0])), KindNumber), true, nil

	case "map_show":
		if len(args) < 1 { return nil, true, fmt.Errorf("map_show: usage: map_show <var>") }
		return mapToTable(resolveVar(sh, args[0])), true, nil

	case "map_merge":
		if len(args) < 2 { return nil, true, fmt.Errorf("map_merge: usage: map_merge <var1> <var2>") }
		result := mapMerge(resolveVar(sh, args[0]), resolveVar(sh, args[1]))
		if len(args) >= 3 { sh.setVar(args[2], result) } else { sh.setVar(args[0], result) }
		return NewText(""), true, nil

	// ── set commands ─────────────────────────────────────────────────────
	case "set_new":
		return NewTyped(encodeSet(nil), KindString), true, nil

	case "set_add":
		if len(args) < 2 { return nil, true, fmt.Errorf("set_add: usage: set_add <var> <value>") }
		sh.setVar(args[0], setAdd(sh.vars[args[0]], sh.evalExpr(args[1])))
		return NewText(""), true, nil

	case "set_remove":
		if len(args) < 2 { return nil, true, fmt.Errorf("set_remove: usage: set_remove <var> <value>") }
		sh.setVar(args[0], setRemove(sh.vars[args[0]], sh.evalExpr(args[1])))
		return NewText(""), true, nil

	case "set_has":
		if len(args) < 2 { return nil, true, fmt.Errorf("set_has: usage: set_has <var> <value>") }
		return NewTyped(setHas(resolveVar(sh, args[0]), sh.evalExpr(args[1])), KindString), true, nil

	case "set_union":
		if len(args) < 2 { return nil, true, fmt.Errorf("set_union: usage: set_union <var1> <var2>") }
		result := setUnion(resolveVar(sh, args[0]), resolveVar(sh, args[1]))
		if len(args) >= 3 { sh.setVar(args[2], result) }
		return NewTyped(result, KindString), true, nil

	case "set_intersect":
		if len(args) < 2 { return nil, true, fmt.Errorf("set_intersect: usage: set_intersect <var1> <var2>") }
		result := setIntersect(resolveVar(sh, args[0]), resolveVar(sh, args[1]))
		if len(args) >= 3 { sh.setVar(args[2], result) }
		return NewTyped(result, KindString), true, nil

	case "set_diff":
		if len(args) < 2 { return nil, true, fmt.Errorf("set_diff: usage: set_diff <var1> <var2>") }
		result := setDiff(resolveVar(sh, args[0]), resolveVar(sh, args[1]))
		if len(args) >= 3 { sh.setVar(args[2], result) }
		return NewTyped(result, KindString), true, nil

	case "set_show":
		if len(args) < 1 { return nil, true, fmt.Errorf("set_show: usage: set_show <var>") }
		return setToTable(resolveVar(sh, args[0])), true, nil

	case "set_len":
		if len(args) < 1 { return nil, true, fmt.Errorf("set_len: usage: set_len <var>") }
		return NewTyped(strconv.Itoa(len(decodeSet(resolveVar(sh, args[0])))), KindNumber), true, nil

	// ── stack commands ────────────────────────────────────────────────────
	case "stack_new":
		return NewTyped(encodeStack(nil), KindString), true, nil

	case "stack_push":
		if len(args) < 2 { return nil, true, fmt.Errorf("stack_push: usage: stack_push <var> <value>") }
		sh.setVar(args[0], stackPush(sh.vars[args[0]], sh.evalExpr(args[1])))
		return NewText(""), true, nil

	case "stack_pop":
		if len(args) < 1 { return nil, true, fmt.Errorf("stack_pop: usage: stack_pop <var> [result_var]") }
		newEnc, val := stackPop(sh.vars[args[0]])
		sh.setVar(args[0], newEnc)
		if len(args) >= 2 { sh.setVar(args[1], val) }
		return NewTyped(val, KindString), true, nil

	case "stack_peek":
		if len(args) < 1 { return nil, true, fmt.Errorf("stack_peek: usage: stack_peek <var>") }
		return NewTyped(stackPeek(sh.vars[args[0]]), KindString), true, nil

	case "stack_len":
		if len(args) < 1 { return nil, true, fmt.Errorf("stack_len: usage: stack_len <var>") }
		return NewTyped(stackLen(sh.vars[args[0]]), KindNumber), true, nil

	case "stack_show":
		if len(args) < 1 { return nil, true, fmt.Errorf("stack_show: usage: stack_show <var>") }
		return stackToTable(sh.vars[args[0]]), true, nil

	// ── queue commands ────────────────────────────────────────────────────
	case "queue_new":
		return NewTyped(encodeQueue(nil), KindString), true, nil

	case "enqueue":
		if len(args) < 2 { return nil, true, fmt.Errorf("enqueue: usage: enqueue <var> <value>") }
		sh.setVar(args[0], queueEnqueue(sh.vars[args[0]], sh.evalExpr(args[1])))
		return NewText(""), true, nil

	case "dequeue":
		if len(args) < 1 { return nil, true, fmt.Errorf("dequeue: usage: dequeue <var> [result_var]") }
		newEnc, val := queueDequeue(sh.vars[args[0]])
		sh.setVar(args[0], newEnc)
		if len(args) >= 2 { sh.setVar(args[1], val) }
		return NewTyped(val, KindString), true, nil

	case "queue_peek":
		if len(args) < 1 { return nil, true, fmt.Errorf("queue_peek: usage: queue_peek <var>") }
		return NewTyped(queuePeek(sh.vars[args[0]]), KindString), true, nil

	case "queue_len":
		if len(args) < 1 { return nil, true, fmt.Errorf("queue_len: usage: queue_len <var>") }
		return NewTyped(queueLen(sh.vars[args[0]]), KindNumber), true, nil

	case "queue_show":
		if len(args) < 1 { return nil, true, fmt.Errorf("queue_show: usage: queue_show <var>") }
		return queueToTable(sh.vars[args[0]]), true, nil

	// ── tuple commands ────────────────────────────────────────────────────
	case "tuple_get":
		if len(args) < 2 { return nil, true, fmt.Errorf("tuple_get: usage: tuple_get <var> <index>") }
		encoded := resolveVar(sh, args[0])
		n := 0; fmt.Sscanf(sh.evalExpr(args[1]), "%d", &n)
		return NewTyped(tupleGet(encoded, n), KindString), true, nil

	case "tuple_len":
		if len(args) < 1 { return nil, true, fmt.Errorf("tuple_len: usage: tuple_len <var>") }
		return NewTyped(tupleLen(resolveVar(sh, args[0])), KindNumber), true, nil

	case "tuple_show":
		if len(args) < 1 { return nil, true, fmt.Errorf("tuple_show: usage: tuple_show <var>") }
		return tupleToTable(resolveVar(sh, args[0])), true, nil

	// ── matrix commands ───────────────────────────────────────────────────
	case "matrix_new":
		r, c, fill := 1, 1, 0.0
		if len(args) >= 1 { fmt.Sscanf(sh.evalExpr(args[0]), "%d", &r) }
		if len(args) >= 2 { fmt.Sscanf(sh.evalExpr(args[1]), "%d", &c) }
		if len(args) >= 3 { fill, _ = strconv.ParseFloat(sh.evalExpr(args[2]), 64) }
		data := make([]float64, r*c)
		for i := range data { data[i] = fill }
		return NewTyped(encodeMatrix(r, c, data), KindString), true, nil

	case "matrix_get":
		if len(args) < 3 { return nil, true, fmt.Errorf("matrix_get: usage: matrix_get <var> <row> <col>") }
		encoded := resolveVar(sh, args[0])
		r, c := 0, 0
		fmt.Sscanf(sh.evalExpr(args[1]), "%d", &r)
		fmt.Sscanf(sh.evalExpr(args[2]), "%d", &c)
		return NewTyped(matrixGet(encoded, r, c), KindNumber), true, nil

	case "matrix_set":
		if len(args) < 4 { return nil, true, fmt.Errorf("matrix_set: usage: matrix_set <var> <row> <col> <value>") }
		r, c := 0, 0
		fmt.Sscanf(sh.evalExpr(args[1]), "%d", &r)
		fmt.Sscanf(sh.evalExpr(args[2]), "%d", &c)
		v, _ := strconv.ParseFloat(sh.evalExpr(args[3]), 64)
		sh.setVar(args[0], matrixSet(sh.vars[args[0]], r, c, v))
		return NewText(""), true, nil

	case "matrix_add":
		if len(args) < 2 { return nil, true, fmt.Errorf("matrix_add: usage: matrix_add <var1> <var2> [result_var]") }
		result := matrixAdd(resolveVar(sh, args[0]), resolveVar(sh, args[1]))
		if len(args) >= 3 { sh.setVar(args[2], result) } else { sh.setVar(args[0], result) }
		return NewText(""), true, nil

	case "matrix_mul":
		if len(args) < 2 { return nil, true, fmt.Errorf("matrix_mul: usage: matrix_mul <var1> <var2> [result_var]") }
		result := matrixMul(resolveVar(sh, args[0]), resolveVar(sh, args[1]))
		if len(args) >= 3 { sh.setVar(args[2], result) } else { sh.setVar(args[0], result) }
		return NewText(""), true, nil

	case "matrix_transpose":
		if len(args) < 1 { return nil, true, fmt.Errorf("matrix_transpose: usage: matrix_transpose <var> [result_var]") }
		result := matrixTranspose(sh.vars[args[0]])
		if len(args) >= 2 { sh.setVar(args[1], result) } else { sh.setVar(args[0], result) }
		return NewText(""), true, nil

	case "matrix_det":
		if len(args) < 1 { return nil, true, fmt.Errorf("matrix_det: usage: matrix_det <var>") }
		return NewTyped(matrixDet(sh.vars[args[0]]), KindNumber), true, nil

	case "matrix_show":
		if len(args) < 1 { return nil, true, fmt.Errorf("matrix_show: usage: matrix_show <var>") }
		return matrixToTable(sh.vars[args[0]]), true, nil

	case "matrix_identity":
		n := 2
		if len(args) >= 1 { fmt.Sscanf(sh.evalExpr(args[0]), "%d", &n) }
		data := make([]float64, n*n)
		for i := 0; i < n; i++ { data[i*n+i] = 1 }
		return NewTyped(encodeMatrix(n, n, data), KindString), true, nil

	// ── type inspection ───────────────────────────────────────────────────
	case "typeof":
		if len(args) < 1 { return nil, true, fmt.Errorf("typeof: usage: typeof <var>") }
		val := resolveVar(sh, args[0])
		t := isDataType(val)
		if t == "" {
			// Fall back to scalar kind
			if _, err := strconv.ParseFloat(val, 64); err == nil { t = "number" } else { t = "string" }
		}
		return NewTyped(t, KindString), true, nil

	case "dt_show":
		// Generic show for any data type
		if len(args) < 1 { return nil, true, fmt.Errorf("dt_show: usage: dt_show <var>") }
		val := resolveVar(sh, args[0])
		switch isDataType(val) {
		case "map":    return mapToTable(val), true, nil
		case "set":    return setToTable(val), true, nil
		case "stack":  return stackToTable(val), true, nil
		case "queue":  return queueToTable(val), true, nil
		case "tuple":  return tupleToTable(val), true, nil
		case "matrix": return matrixToTable(val), true, nil
		}
		return NewText(val), true, nil
	}

	return nil, false, nil
}

// resolveVar returns the value of args[0]: if it starts with $ it's a var
// reference, otherwise it's treated as a literal encoded value.
func resolveVar(sh *Shell, s string) string {
	if strings.HasPrefix(s, "$") { return sh.getVar(s[1:]) }
	if v, ok := sh.vars[s]; ok { return v }
	return s
}

// isDataTypeCommand returns true for all data type command names.
func isDataTypeCommand(cmd string) bool {
	switch cmd {
	case "map_new","map_set","map_get","map_del","map_has","map_keys",
		"map_values","map_len","map_show","map_merge",
		"set_new","set_add","set_remove","set_has","set_union",
		"set_intersect","set_diff","set_show","set_len",
		"stack_new","stack_push","stack_pop","stack_peek","stack_len","stack_show",
		"queue_new","enqueue","dequeue","queue_peek","queue_len","queue_show",
		"tuple_get","tuple_len","tuple_show",
		"matrix_new","matrix_get","matrix_set","matrix_add","matrix_mul",
		"matrix_transpose","matrix_det","matrix_show","matrix_identity",
		"typeof","dt_show":
		return true
	}
	return false
}
