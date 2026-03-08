#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  07_functions.ksh — Functions, return values, recursion, defer
#
#  Topics:
#    basic definition · return · multiple params · $_return
#    variadic ($_args) · local · defer · recursion · higher-order
# ─────────────────────────────────────────────────────────────────────

# ── Basic function ────────────────────────────────────────────────────
func greet(name) {
    println "Hello, $name!"
}
greet "Alice"
greet "Bob"
println ""

# ── Return values ─────────────────────────────────────────────────────
func add(a, b) {
    return $a + $b
}

result = add(3, 7)
println "add(3, 7) = $result"

# Return value also stored in $_return
add 10 20
println "_return   = $_return"
println ""

# ── Multiple parameters ───────────────────────────────────────────────
func clamp(val, lo, hi) {
    if $val < $lo: return $lo
    if $val > $hi: return $hi
    return $val
}

for v in [-10, 0, 50, 75, 150] {
    c = clamp($v, 0, 100)
    println "clamp($v, 0, 100) = $c"
}
println ""

# ── local variables ───────────────────────────────────────────────────
func build_tag(text, tag) {
    local open  = "<$tag>"
    local close = "</$tag>"
    return "$open$text$close"
}
println "$(build_tag 'Hello World' 'h1')"
println "$(build_tag 'subtitle' 'h2')"
println ""

# ── Variadic: extra args beyond named params go to $_args ─────────────
func log(level) {
    println "[$level] $_args"
}
log "INFO" "server started on port" 8080
log "WARN" "memory usage above 80%"
log "ERROR" "failed to connect to" "database" "— retrying"
println ""

# ── Default-like behavior using ${:-} ─────────────────────────────────
func connect(host, port) {
    port = ${port:-443}
    println "Connecting to $host:$port"
}
connect "api.example.com" 8080
connect "api.example.com"
println ""

# ── Recursion: factorial ──────────────────────────────────────────────
func factorial(n) {
    if $n <= 1: return 1
    prev = factorial($n - 1)
    return $n * $prev
}

println "Factorials:"
for n in range(0..10) {
    f = factorial($n)
    println "  $n! = $f"
}
println ""

# ── Recursion: Fibonacci ──────────────────────────────────────────────
func fib(n) {
    if $n <= 1: return $n
    a = fib($n - 1)
    b = fib($n - 2)
    return $a + $b
}

println "Fibonacci sequence:"
seq = []
for i in range(0..12) {
    seq[] = fib($i)
}
println "  $seq"
println ""

# ── Recursion: binary search ──────────────────────────────────────────
sorted_arr = [2, 5, 8, 12, 16, 23, 38, 42, 55, 72, 91]

func binary_search(arr, target, lo, hi) {
    if $lo > $hi: return -1
    mid = ($lo + $hi) / 2
    mid_val = $arr[$mid]
    if $mid_val == $target: return $mid
    if $mid_val < $target: return binary_search($arr, $target, $mid + 1, $hi)
    return binary_search($arr, $target, $lo, $mid - 1)
}

for t in [12, 55, 100, 2, 91] {
    last_idx = $sorted_arr.len - 1
    idx = binary_search($sorted_arr, $t, 0, $last_idx)
    if $idx >= 0 {
        println "Found $t at index $idx"
    } else {
        println "$t not found"
    }
}
println ""

# ── defer — runs when function exits, LIFO order ─────────────────────
func with_logging(name) {
    defer println "  [defer 3] function '$name' exited"
    defer println "  [defer 2] cleanup complete"
    defer println "  [defer 1] starting cleanup"
    println "  [body] function '$name' running..."
    println "  [body] doing work..."
}
println "with_logging demonstration:"
with_logging "process_data"
println ""

# ── Higher-order: pass op name to arr_map ────────────────────────────
func apply_to_all(arr, op) {
    return $arr | arr_map $op
}

words = ["hello", "world", "katsh", "scripting"]
println "Original:    $words"
println "upper:       $(apply_to_all $words upper)"
println "reverse:     $(apply_to_all $words reverse)"
println "len (chars): $(apply_to_all $words len)"
println ""

# ── Functions calling functions ───────────────────────────────────────
func slugify(s) {
    return $s | lower | trim | replace " " "-"
}

func make_url(base, title) {
    slug = slugify($title)
    return "$base/$slug"
}

titles = ["Hello World", "My First Post", "KatSH is Cool"]
for title in $titles {
    url = make_url "https://blog.example.com" $title
    println "$url"
}