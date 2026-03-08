#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  24_logic_operators.ksh — && || short-circuit logic
#
#  Topics covered:
#    && (and-then) · || (or-else) · chaining multiple operators
#    short-circuit evaluation · using && for guard clauses
#    || for fallbacks · combining with if/unless/when
#    real patterns: flag checking, safe defaults, command chains
# ─────────────────────────────────────────────────────────────────────

# ── && runs right side only if left succeeds ─────────────────────────
println "=== && (and-then) ==="

x = 10
x > 5 && println "  x is greater than 5"
x > 20 && println "  this will NOT print (x <= 20)"
println ""

# Chain of actions — stops at first failure
println "Build-like chain:"
ok = true
$ok && println "  step 1: compile — OK"
$ok && println "  step 2: test    — OK"
$ok && println "  step 3: deploy  — OK"
println ""

# ── || runs right side only if left fails ────────────────────────────
println "=== || (or-else) ==="

y = 0
y > 0 || println "  y is not positive (fallback fired)"
y > 0 || y = 1
println "  y after fallback: $y"
println ""

# ── Short-circuit as guard clause ────────────────────────────────────
println "=== && as guard clause ==="
func process(val) {
    $val != "" || throw "process: value cannot be empty"
    $val | isnum || throw "process: value must be numeric — got '$val'"
    n = tonum $val
    n > 0 || throw "process: value must be positive — got $n"
    println "  processing $n → $(n * n)"
}

for input in ["9", "", "abc", "-3", "7"] {
    try {
        process $input
    } catch e {
        println "  error: $e"
    }
}
println ""

# ── || for default values ─────────────────────────────────────────────
println "=== || for defaults ==="
func get_config(key, default_val) {
    # Simulate: some keys exist in env
    known = map { "HOST":"localhost" "PORT":"8080" }
    val   = map_get known $key
    $val != "" || val = $default_val
    return $val
}

host    = get_config "HOST"    "127.0.0.1"
port    = get_config "PORT"    "3000"
timeout = get_config "TIMEOUT" "30"
region  = get_config "REGION"  "us-east-1"

println "  HOST=$host  PORT=$port  TIMEOUT=$timeout  REGION=$region"
println ""

# ── Combining && and || ───────────────────────────────────────────────
println "=== combined && || ==="

func check_user(name, age) {
    $name != "" && $age > 0 && println "  valid user: $name (age $age)"
    $name == "" || $age > 0 || println "  invalid: empty name or bad age"
}

check_user "Alice" 30
check_user "Bob"   25
check_user ""      30
check_user "Dave"  -1
println ""

# ── && || in file operations ─────────────────────────────────────────
println "=== && || in file operations ==="
tmp = "/tmp/katsh_logic_test_$$"

# Create file, then act on it
bash! touch "$tmp" && echo "  created: $tmp"
bash! test -f "$tmp" && echo "  file exists: yes" || echo "  file exists: no"
bash! rm -f "$tmp" && echo "  removed: $tmp"
bash! test -f "$tmp" && echo "  still exists" || echo "  confirmed gone"
println ""

# ── Building validated pipelines with && ─────────────────────────────
println "=== validated processing chain ==="
func parse_int(s)   { $s | isnum || throw "not a number: $s";  return tonum $s }
func positive(n)    { $n > 0    || throw "not positive: $n";   return $n }
func under_100(n)   { $n < 100  || throw "too large: $n";      return $n }

func validated_score(s) {
    n = parse_int($s)
    positive $n
    under_100 $n
    return $n
}

for raw in ["85", "0", "abc", "150", "42", "-5"] {
    try {
        score = validated_score($raw)
        println "  '$raw' → valid score: $score"
    } catch e {
        println "  '$raw' → $e"
    }
}
println ""

# ── Short-circuit in loops ────────────────────────────────────────────
println "=== short-circuit in loop logic ==="
items = ["apple", "", "banana", "", "cherry", "date", ""]
good  = []
blank_count = 0

for item in $items {
    $item == "" && { blank_count++ ; continue }
    good[] = $item
}
println "  non-empty: $good"
println "  blanks skipped: $blank_count"
println ""

# ── Operator precedence demo ──────────────────────────────────────────
println "=== operator chaining ==="
a = true
b = false
c = true

# Both && and || are left-to-right
$a && $b && println "  a && b: yes"
$a && $b || println "  a && b failed, || fires"
$a || $b && println "  a || b: a was true, chain ran"
$b || $c && println "  b || c: c was true, chain ran"
$b && $a || $c && println "  (b && a) || c: c saved it"