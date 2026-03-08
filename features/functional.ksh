#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  35_functional.ksh — Memoization, higher-order functions, composition
#
#  Topics covered:
#    memoize pattern (map as cache) · call counting
#    function composition (compose/pipe) · partial application
#    map/filter/reduce with user functions
#    curry pattern · once (run-only-once wrapper)
#    trampolining for deep recursion
# ─────────────────────────────────────────────────────────────────────

# ── Memoization ───────────────────────────────────────────────────────
println "=== memoization ==="
_memo_cache  = map {}
_call_counts = map {}

func memoize(fn_name, arg) {
    key = "${fn_name}:${arg}"
    cached = map_get _memo_cache $key
    if $cached != "": return $cached

    # Track call count
    cnt = map_get _call_counts $fn_name
    if $cnt == "": cnt = 0
    map_set _call_counts $fn_name ($cnt + 1)

    result = $fn_name($arg)
    map_set _memo_cache $key $result
    return $result
}

func slow_fib(n) {
    if $n <= 1: return $n
    a = memoize "slow_fib" ($n - 1)
    b = memoize "slow_fib" ($n - 2)
    return $a + $b
}

println "Fibonacci with memoization:"
for i in range(0..20) {
    f = memoize "slow_fib" $i
    printf "  fib(%2d) = %d\n" $i $f
}
fn_calls = map_get _call_counts "slow_fib"
println "  unique slow_fib() calls: $fn_calls  (vs $(20 * 19 / 2) without memo)"
println ""

# ── Function composition ──────────────────────────────────────────────
println "=== function composition ==="

# compose(f, g)(x) = f(g(x))
# In KatSH: pass function names, apply right-to-left
func compose_apply(fns, val) {
    result = $val
    rev = $fns | arr_reverse
    for fn in $rev {
        result = $fn($result)
    }
    return $result
}

func double(x)    { return $x * 2 }
func inc(x)       { return $x + 1 }
func square(x)    { return $x * $x }
func neg(x)       { return $x * -1 }

# pipeline: square → double → inc
pipeline = ["inc", "double", "square"]
for n in [2, 3, 4, 5] {
    result = compose_apply($pipeline, $n)
    println "  compose(inc∘double∘square)($n) = $result"
}
println ""

# ── Partial application ───────────────────────────────────────────────
println "=== partial application ==="

func add(a, b)  { return $a + $b }
func mul(a, b)  { return $a * $b }
func pow_fn(a, b) { return $a ** $b }

# Simulate partial: store first arg, return a "bound" function name
partial_args = map {}

func partial(fn_name, bound_arg, alias_name) {
    map_set partial_args $alias_name "$fn_name:$bound_arg"
}

func call_partial(alias_name, b) {
    spec    = map_get partial_args $alias_name
    parts   = $spec | split ":"
    fn_name = $parts[0]
    a       = $parts[1]
    return $fn_name($a, $b)
}

partial "add" 10 "add10"
partial "mul" 3  "triple"
partial "mul" 2  "double2"
partial "pow_fn" 2 "pow2"

for n in [1, 2, 3, 4, 5] {
    a  = call_partial "add10"  $n
    b  = call_partial "triple" $n
    c  = call_partial "double2" $n
    d  = call_partial "pow2"   $n
    println "  n=$n  add10=$a  triple=$b  double=$c  2^n=$d"
}
println ""

# ── arr_map / arr_filter / reduce with user functions ──────────────────
println "=== map / filter / reduce ==="

nums = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
println "nums: $nums"

# Map
doubled  = $nums | arr_map double
squared  = $nums | arr_map square
println "  doubled: $doubled"
println "  squared: $squared"

# Filter (built-in predicates)
func is_even(n) { return $n % 2 == 0 }
func is_odd(n)  { return $n % 2 != 0 }
func gt5(n)     { return $n > 5 }

evens = $nums | arr_filter is_even
odds  = $nums | arr_filter is_odd
big   = $nums | arr_filter gt5
println "  evens:  $evens"
println "  odds:   $odds"
println "  > 5:    $big"

# Reduce (manual — katsh doesn't have a reduce builtin, so we write it)
func reduce(arr, fn_name, init) {
    acc = $init
    for val in $arr {
        acc = $fn_name($acc, $val)
    }
    return $acc
}

func add_fn(a, b) { return $a + $b }
func max_fn(a, b) { return if $a > $b: $a; else: $b }
func min_fn(a, b) { return if $a < $b: $a; else: $b }

sum  = reduce($nums, "add_fn", 0)
mx   = reduce($nums, "max_fn", $nums[0])
mn   = reduce($nums, "min_fn", $nums[0])
prod = reduce([1,2,3,4,5], "mul", 1)

println "  sum:     $sum"
println "  max:     $mx"
println "  min:     $mn"
println "  product: $prod"
println ""

# ── once: run a function only once ───────────────────────────────────
println "=== once (run only once) ==="
_once_done = set {}

func once(fn_name) {
    set_has _once_done $fn_name && return
    set_add _once_done $fn_name
    $fn_name()
}

func init_db() {
    println "  [init_db] connecting to database..."
}
func init_cache() {
    println "  [init_cache] warming up cache..."
}

println "First pass:"
once "init_db"
once "init_cache"
println "Second pass (should be silent):"
once "init_db"
once "init_cache"
println "Done — both functions ran exactly once ✓"
println ""

# ── Pipeline as first-class value ────────────────────────────────────
println "=== pipelines as data ==="

# A pipeline is just a list of function names
text_pipeline   = ["trim", "lower", "reverse"]
number_pipeline = ["square", "double", "inc"]

func run_pipeline(pipeline, val) {
    result = $val
    for fn in $pipeline {
        result = $fn($result)
    }
    return $result
}

for s in ["  HELLO  ", "  WORLD  ", "  KATSH  "] {
    println "  '$s' → '$(run_pipeline $text_pipeline $s)'"
}
println ""

for n in [2, 3, 4, 5] {
    println "  $n → $(run_pipeline $number_pipeline $n)"
}
println ""

# ── Memoize + compose: precompute a lookup table ─────────────────────
println "=== precomputed lookup table ==="
trig_table = map {}

func precompute_squares(limit) {
    for i in range(0..$limit) {
        map_set trig_table "sq_$i" ($i * $i)
    }
}

precompute_squares 15

println "Square lookup table (0-15):"
for i in range(0..15) {
    v = map_get trig_table "sq_$i"
    printf "  %2d² = %3d\n" $i $v
}