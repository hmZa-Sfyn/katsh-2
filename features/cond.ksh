#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  05_conditionals.ksh — if / elif / else, unless, when
#
#  Topics:
#    inline if · block if · multi-branch · unless
#    when (postfix guard) · and/or/not · file tests
# ─────────────────────────────────────────────────────────────────────

# ── Inline if ─────────────────────────────────────────────────────────
x = 42
if $x > 10: println "x is greater than 10"
if $x == 42: println "x is exactly 42"
if $x > 100: println "huge"; else: println "not huge"

println ""

# ── Block if ──────────────────────────────────────────────────────────
if $x > 40 {
    println "x ($x) is in the 40s or above"
    doubled = $x * 2
    println "  doubled = $doubled"
}
println ""

# ── Multi-branch if / elif / else ─────────────────────────────────────
func classify_score(s) {
    if $s >= 90:
        return "A — Excellent"
    elif $s >= 80:
        return "B — Good"
    elif $s >= 70:
        return "C — Average"
    elif $s >= 60:
        return "D — Below Average"
    else:
        return "F — Failing"
}

for score in [95, 85, 72, 65, 42] {
    label = classify_score($score)
    println "  $score → $label"
}
println ""

# ── Comparison operators ──────────────────────────────────────────────
a = "hello"
b = "world"
println "a == 'hello':  $(if $a == 'hello': 'yes'; else: 'no')"
println "a != b:        $(if $a != $b: 'yes'; else: 'no')"
println "10 > 5:        $(if 10 > 5: 'yes'; else: 'no')"
println "10 >= 10:      $(if 10 >= 10: 'yes'; else: 'no')"
println "5 <= 5:        $(if 5 <= 5: 'yes'; else: 'no')"
println ""

# ── Regex match with ~ and !~ ─────────────────────────────────────────
email = "user@example.com"
if $email ~ "@": println "'$email' contains @"
if $email !~ "gmail": println "'$email' is not a Gmail address"
println ""

# ── and / or / not ────────────────────────────────────────────────────
age   = 25
score2 = 88

if $age >= 18 and $score2 >= 80 {
    println "Adult with good score"
}

if $age < 18 or $score2 < 60 {
    println "Young or low score"
} else {
    println "Adult with passing score"
}

ready = false
if not $ready: println "System not ready"
println ""

# ── unless — opposite of if ───────────────────────────────────────────
connected = false
unless $connected: println "Not connected to network"
unless $connected {
    println "Attempting to reconnect..."
    connected = true
    println "Connected: $connected"
}
println ""

# ── when — postfix guard (runs statement only if condition is true) ────
n = 7
println "n = $n"
println "  positive" when $n > 0
println "  even"     when $n % 2 == 0
println "  odd"      when $n % 2 != 0
println "  large"    when $n > 100
println ""

# Useful in loops:
results = []
for i in range(1..20) {
    results[] = $i when $i % 3 == 0   # append only if divisible by 3
}
println "Multiples of 3 (1-20): $results"
println ""

# ── File existence tests ──────────────────────────────────────────────
# Note: -f (file), -d (directory), -e (exists) work in conditions
# Here we test on the script's own directory

script_dir = $__script_dir

if -d $script_dir {
    println "Script dir exists: $script_dir"
}

test_file = "$script_dir/01_variables.ksh"
if -f $test_file {
    println "Found: 01_variables.ksh"
} else {
    println "File not found: $test_file"
}

# Inline ternary-style using if expression
status = if -d $script_dir: "present"; else: "missing"
println "examples/ directory: $status"