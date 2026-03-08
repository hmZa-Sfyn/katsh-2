#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  01_variables.ksh — Variables, assignment, operators, interpolation
#
#  Topics:
#    basic assignment · compound operators · string interpolation
#    ${...} expansion · readonly · with scoping · typeof
# ─────────────────────────────────────────────────────────────────────

# ── Basic assignment ──────────────────────────────────────────────────
name  = "Alice"
age   = 30
score = 98.5
flag  = true

println "Name:  $name"
println "Age:   $age"
println "Score: $score"
println "Flag:  $flag"
println ""

# ── Compound operators ────────────────────────────────────────────────
x = 10
println "Start: $x"
x += 5    ;  println "  += 5  → $x"
x -= 3    ;  println "  -= 3  → $x"
x *= 4    ;  println "  *= 4  → $x"
x /= 6    ;  println "  /= 6  → $x"
x %= 5    ;  println "  %= 5  → $x"
x **= 3   ;  println "  **= 3 → $x"
x++       ;  println "  x++   → $x"
x--       ;  println "  x--   → $x"
++x       ;  println "  ++x   → $x"
println ""

# ── String interpolation ──────────────────────────────────────────────
city = "NYC"
println "Hello $name, you live in $city"
println "Name has ${#name} characters"
println "Uppercase: $(echo $name | upper)"

# ${VAR:-default} — use default if unset or empty
unset greeting
println "Greeting: ${greeting:-Good morning}"

greeting = "Hey"
println "Greeting: ${greeting:-Good morning}"

# ${VAR:+value} — use value only if VAR is set
println "Badge: ${score:+[${score} pts]}"
println ""

# ── Readonly variables ────────────────────────────────────────────────
VERSION = "1.0.0"
readonly VERSION
println "Version: $VERSION"

# Uncommenting the next line would cause E005 ReadonlyError:
# VERSION = "2.0.0"
println ""

# ── typeof ────────────────────────────────────────────────────────────
nums = [1, 2, 3]
m    = map { "key":"val" }

println "typeof name: $(typeof $name)"
println "typeof age:  $(typeof $age)"
println "typeof nums: $(typeof $nums)"
println "typeof m:    $(typeof $m)"
println ""

# ── with scoping ──────────────────────────────────────────────────────
with temp = "scoped" {
    println "Inside with: $temp"
}
# $temp is gone here — would be empty:
println "Outside with: '${temp:-<unset>}'"