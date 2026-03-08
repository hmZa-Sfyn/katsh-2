#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  06_loops.ksh — for, while, do-while, do-until, repeat
#
#  Topics:
#    for over list · for over range · for over array var
#    for over command output · while · do-while · do-until
#    repeat · break · continue · break/continue with when
# ─────────────────────────────────────────────────────────────────────

# ── for over literal list ─────────────────────────────────────────────
println "=== for over literal list ==="
for color in ["red", "green", "blue", "yellow"] {
    upper_color = $color | upper
    println "  $upper_color"
}
println ""

# ── for over numeric range ────────────────────────────────────────────
println "=== range(1..5) ==="
for i in range(1..5) {
    println "  i = $i"
}
println ""

# ── range with step — using modulo trick ─────────────────────────────
println "=== even numbers 0..20 ==="
evens = []
for i in range(0..20) {
    evens[] = $i when $i % 2 == 0
}
println "  $evens"
println ""

# ── for over array variable ───────────────────────────────────────────
println "=== iterating array variable ==="
fruits = ["apple", "banana", "cherry", "date"]
for fruit in $fruits {
    println "  $fruit has $(echo $fruit | len) letters"
}
println ""

# ── for over command output ───────────────────────────────────────────
println "=== for over $() command output ==="
for word in $(echo "the quick brown fox jumps" | words) {
    println "  '$word'"
}
println ""

# ── break ─────────────────────────────────────────────────────────────
println "=== break at first multiple of 7 ==="
for i in range(1..100) {
    if $i % 7 == 0 {
        println "  Found: $i — breaking"
        break
    }
}
println ""

# ── continue ──────────────────────────────────────────────────────────
println "=== continue — skip multiples of 3 ==="
output = []
for i in range(1..15) {
    continue when $i % 3 == 0
    output[] = $i
}
println "  $output"
println ""

# ── break / continue with when (clean postfix style) ─────────────────
println "=== collect until sum > 50 ==="
nums  = [5, 12, 8, 20, 3, 15, 9, 7, 18]
total = 0
kept  = []
for n in $nums {
    break when $total + $n > 50
    total += $n
    kept[] = $n
}
println "  kept=$kept  total=$total"
println ""

# ── while loop ────────────────────────────────────────────────────────
println "=== while — countdown ==="
n = 5
while $n > 0 {
    println "  $n..."
    n--
}
println "  Liftoff!"
println ""

# ── while with break ──────────────────────────────────────────────────
println "=== while with break — find first prime > 20 ==="
func is_prime(n) {
    if $n < 2: return false
    i = 2
    while $i * $i <= $n {
        if $n % $i == 0: return false
        i++
    }
    return true
}

candidate = 21
while true {
    p = is_prime($candidate)
    if $p {
        println "  First prime > 20: $candidate"
        break
    }
    candidate++
}
println ""

# ── do-while — body runs at least once ────────────────────────────────
println "=== do-while — at least one iteration ==="
count = 0
do {
    count++
    println "  iteration $count"
} while $count < 3
println ""

# ── do-until — stop when condition becomes true ───────────────────────
println "=== do-until — double until > 100 ==="
val = 1
do {
    val *= 2
    println "  val = $val"
} until $val > 100
println ""

# ── repeat N — simplest loop ─────────────────────────────────────────
println "=== repeat 5 — uses \$_i counter ==="
repeat 5 {
    stars = "*" | repeat ($_i + 1)
    println "  $_i: $stars"
}
println ""

# ── Nested loops ─────────────────────────────────────────────────────
println "=== nested loops — multiplication table (1-4) ==="
for row in range(1..4) {
    line = ""
    for col in range(1..4) {
        product = $row * $col
        padded  = tostr $product | lpad 4
        line    = $line | concat $padded
    }
    println "$line"
}