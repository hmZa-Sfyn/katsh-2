#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  27_types.ksh — Type system, coercion, typeof, edge cases
#
#  Topics covered:
#    typeof on all types · implicit vs explicit coercion
#    tonum / tostr / toarray · isnum/isalpha/isalnum edge cases
#    arithmetic with string numbers · strict comparison
#    map/set/stack/queue/tuple/matrix typeof
#    defensive type checking in functions · coercion in conditions
# ─────────────────────────────────────────────────────────────────────

# ── typeof on all value kinds ─────────────────────────────────────────
println "=== typeof ==="
s   = "hello"
n   = 42
f   = 3.14
b   = true
arr = [1, 2, 3]
m   = map { a=1 b=2 }
st  = set { 1 2 3 }
stk = stack {}
q   = queue {}
t   = (1, 2, 3)
mx  = matrix(2, 2)
stack_push stk "x"
enqueue q "y"

println "  'hello'      → $(typeof $s)"
println "  42           → $(typeof $n)"
println "  3.14         → $(typeof $f)"
println "  true         → $(typeof $b)"
println "  [1,2,3]      → $(typeof $arr)"
println "  map{...}     → $(typeof $m)"
println "  set{...}     → $(typeof $st)"
println "  stack{}      → $(typeof $stk)"
println "  queue{}      → $(typeof $q)"
println "  (1,2,3)      → $(typeof $t)"
println "  matrix(2,2)  → $(typeof $mx)"
println ""

# ── tonum: string → number ────────────────────────────────────────────
println "=== tonum ==="
for s in ["42", "3.14", "0", "-7", "1e3", "  99  ", "0xFF", "abc"] {
    is_n = $s | trim | isnum
    if $is_n {
        n = tonum ($s | trim)
        println "  tonum('$s') = $n  (type: $(typeof $n))"
    } else {
        println "  tonum('$s') → not numeric"
    }
}
println ""

# ── tostr: number → string ────────────────────────────────────────────
println "=== tostr ==="
for n in [0, 42, -7, 3.14159, 1000000] {
    s   = tostr $n
    len = $s | len
    println "  tostr($n) = '$s'  len=$len  type=$(typeof $s)"
}
println ""

# ── toarray: string → array ───────────────────────────────────────────
println "=== toarray ==="
for s in ["hello world", "a b c d e", "  spaces  here  ", "nosplit"] {
    arr = toarray $s
    len = $arr | arr_len
    println "  toarray('$s') → $arr  (len=$len)"
}
println ""

# ── isnum edge cases ──────────────────────────────────────────────────
println "=== isnum edge cases ==="
for val in ["0", "1", "-1", "3.14", "3.14.15", "1e5", "0x1F",
            " 42 ", "42abc", "", "∞", "NaN", ".5", "-.5"] {
    trimmed = $val | trim
    n = $trimmed | isnum
    println "  isnum('$val') = $n"
}
println ""

# ── Implicit coercion in arithmetic ──────────────────────────────────
println "=== implicit coercion ==="
# When a string looks like a number, arithmetic works on it
s_num = "10"
result = $s_num + 5
println "  '10' + 5 = $result  (type: $(typeof $result))"

s_float = "2.5"
result2 = $s_float * 4
println "  '2.5' * 4 = $result2"

# String that's NOT a number — arithmetic falls back to string concat
s_word = "hello"
try {
    result3 = $s_word + 5
    println "  'hello' + 5 = $result3  (type: $(typeof $result3))"
} catch e {
    println "  'hello' + 5 → $e"
}
println ""

# ── Strict equality vs coerced equality ──────────────────────────────
println "=== equality comparisons ==="
# In KatSH, == compares string representations
println "  42 == '42'   : $(if 42 == '42': 'true'; else: 'false')"
println "  42 == 42     : $(if 42 == 42: 'true'; else: 'false')"
println "  '42' == '42' : $(if '42' == '42': 'true'; else: 'false')"
println "  true == 'true': $(if true == 'true': 'true'; else: 'false')"
println "  0 == false   : $(if 0 == false: 'true'; else: 'false')"
println "  '' == false  : $(if '' == false: 'true'; else: 'false')"
println ""

# ── Defensive type checking in functions ──────────────────────────────
println "=== defensive typed functions ==="
func typed_add(a, b) {
    typeof($a) == "number" || tonum($a) != "" || throw "typed_add: a='$a' is not numeric"
    typeof($b) == "number" || tonum($b) != "" || throw "typed_add: b='$b' is not numeric"
    return tonum($a) + tonum($b)
}

func expect_map(val, name) {
    typeof($val) == "map" || throw "TypeError: $name must be a map, got $(typeof $val)"
}

func expect_array(val, name) {
    # Arrays are encoded as strings in KatSH — check via arr_len
    try {
        arr_len $val
    } catch e {
        throw "TypeError: $name must be an array"
    }
}

for pair in [["3","7"], ["2.5","1.5"], ["hello","5"], ["10","x"]] {
    a = $pair[0]
    b = $pair[1]
    try {
        r = typed_add($a, $b)
        println "  typed_add($a, $b) = $r"
    } catch e {
        println "  typed_add($a, $b) → $e"
    }
}
println ""

# ── Type-safe container pattern ───────────────────────────────────────
println "=== type-safe value wrapper ==="
func make_typed(type_name, val) {
    w = map {}
    map_set w "type"  $type_name
    map_set w "value" $val
    return $w
}

func unwrap_typed(w, expected_type) {
    actual = map_get w "type"
    $actual == $expected_type || throw "TypeError: expected $expected_type, got $actual"
    return map_get w "value"
}

# Create typed values
age_val    = make_typed "age"    "25"
score_val  = make_typed "score"  "88"
name_val   = make_typed "name"   "Alice"

# Unwrap correctly
age   = unwrap_typed($age_val,   "age")
score = unwrap_typed($score_val, "score")
println "  age=$age  score=$score"

# Unwrap with wrong type
try {
    wrong = unwrap_typed($name_val, "age")
} catch e {
    println "  $e"
}
println ""

# ── Coercion in conditionals ──────────────────────────────────────────
println "=== truthiness in conditions ==="
func truthy(val) {
    if $val: return "truthy"; else: return "falsy"
}
for v in ["", "0", "false", "no", "hello", "1", "true", "yes"] {
    println "  '$v' is $(truthy $v)"
}