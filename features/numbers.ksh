#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  04_numbers.ksh — Numbers, math operations, type conversion
#
#  Topics:
#    arithmetic · pipe ops · abs/ceil/floor/round/sqrt
#    hex/oct/bin · tonum/tostr · bc · random
# ─────────────────────────────────────────────────────────────────────

# ── Basic arithmetic expressions ──────────────────────────────────────
println "3 + 4      = $(3 + 4)"
println "10 - 3     = $(10 - 3)"
println "6 * 7      = $(6 * 7)"
println "22 / 7     = $(22 / 7)"
println "17 % 5     = $(17 % 5)"
println "2 ** 10    = $(2 ** 10)"
println "(3+4)*2    = $((3+4)*2)"
println ""

# ── Pipe-style numeric operations ─────────────────────────────────────
println "5  | add 3     = $(5  | add 3)"
println "10 | sub_n 4   = $(10 | sub_n 4)"
println "4  | mul 7     = $(4  | mul 7)"
println "10 | div 4     = $(10 | div 4)"
println "10 | mod 3     = $(10 | mod 3)"
println "3  | pow 4     = $(3  | pow 4)"
println ""

# ── Rounding functions ────────────────────────────────────────────────
for n in [3.2, 3.5, 3.7, -2.3, -2.7] {
    c = $n | ceil
    f = $n | floor
    r = $n | round
    println "$n  ceil=$c  floor=$f  round=$r"
}
println ""

# ── Other math ────────────────────────────────────────────────────────
println "sqrt(144)  = $(144 | sqrt)"
println "sqrt(2)    = $(2   | sqrt)"
println "abs(-42)   = $(-42 | abs)"
println "abs(42)    = $(42  | abs)"
println "negate(7)  = $(7   | negate)"
println ""

# ── Base conversion ───────────────────────────────────────────────────
for n in [0, 10, 15, 255, 256, 1024] {
    h = $n | hex
    o = $n | oct
    b = $n | bin
    println "$n  hex=$h  oct=$o  bin=$b"
}
println ""

# ── Type conversion ───────────────────────────────────────────────────
s  = "42"
n  = tonum $s
n2 = $n * 3
println "tonum '$s' → $n  (× 3 = $n2)"

n3 = 3.14159
s2 = tostr $n3
println "tostr $n3 → '$s2' (type: $(typeof $s2))"

words_str = "alpha beta gamma"
arr = toarray $words_str
println "toarray '$words_str' → $arr"
println ""

# ── bc — arbitrary precision ──────────────────────────────────────────
println "bc  22/7           = $(bc '22/7')"
println "bc  scale=15; 22/7 = $(bc 'scale=15; 22/7')"
println "bc  sqrt(2)        = $(bc 'sqrt(2)')"
println "bc  2^64           = $(bc '2^64')"
println "bc  1+2+3+4+5      = $(bc '1+2+3+4+5')"
println ""

# ── factor ────────────────────────────────────────────────────────────
for n in [12, 28, 97, 100, 360] {
    factor $n
}
println ""

# ── random ────────────────────────────────────────────────────────────
println "random (any):        $(random)"
println "random 100 (0-99):   $(random 100)"
println "random 50 100:       $(random 50 100)"
println ""

# ── Compound example: grade calculator ────────────────────────────────
scores = [88, 92, 75, 95, 83, 67, 91]
total  = $scores | arr_sum
count  = $scores | arr_len
avg    = $total / $count
min_s  = $scores | arr_min
max_s  = $scores | arr_max

println "Scores:  $scores"
println "Total:   $total"
println "Count:   $count"
println "Average: $avg"
println "Min:     $min_s"
println "Max:     $max_s"

grade = if $avg >= 90: "A"
grade = if $avg >= 80 and $grade == "": "B"; else: $grade
grade = if $avg >= 70 and $grade == "": "C"; else: $grade
grade = if $grade == "": "D"; else: $grade
println "Grade:   $grade"