#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  34_modules.ksh — Source/import and multi-file patterns
#
#  Topics covered:
#    source . for local libraries · import from URL
#    exporting functions with func! · namespace conventions
#    guard against double-sourcing (__MODULE__ guard)
#    building a local "standard library" of helpers
#    using library functions across scripts
# ─────────────────────────────────────────────────────────────────────

# ── Module guard pattern ──────────────────────────────────────────────
# Each library sets a flag so it can be sourced safely multiple times
# without re-running its body.

# ────────────────────────────────────────────────────────────────────
# Simulate: lib/math.ksh
# (In real use: source ./lib/math.ksh)
# ────────────────────────────────────────────────────────────────────
println "--- loading lib/math.ksh ---"
__MATH_LOADED = ${__MATH_LOADED:-false}
unless $__MATH_LOADED {
    __MATH_LOADED = true

    func! math_gcd(a, b) {
        while $b != 0 { tmp = $b; b = $a % $b; a = $tmp }
        return $a
    }
    func! math_lcm(a, b) {
        return $a * $b / math_gcd($a, $b)
    }
    func! math_is_prime(n) {
        if $n < 2: return false
        i = 2
        while $i * $i <= $n {
            if $n % $i == 0: return false
            i++
        }
        return true
    }
    func! math_clamp(v, lo, hi) {
        if $v < $lo: return $lo
        if $v > $hi: return $hi
        return $v
    }
    func! math_lerp(a, b, t) {
        return $a + ($b - $a) * $t
    }
    println "  math functions: gcd lcm is_prime clamp lerp"
}

# ────────────────────────────────────────────────────────────────────
# Simulate: lib/strings.ksh
# ────────────────────────────────────────────────────────────────────
println "--- loading lib/strings.ksh ---"
__STRINGS_LOADED = ${__STRINGS_LOADED:-false}
unless $__STRINGS_LOADED {
    __STRINGS_LOADED = true

    func! str_slugify(s) {
        return $s | lower | trim | replace " " "-"
    }
    func! str_truncate(s, max_len) {
        if ($s | len) <= $max_len: return $s
        return ($s | sub 0 ($max_len - 3)) | concat "..."
    }
    func! str_pad_center(s, w) {
        return $s | center $w
    }
    func! str_starts_with_any(s, prefixes) {
        for p in $prefixes {
            if $s | startswith $p: return true
        }
        return false
    }
    func! str_count(haystack, needle) {
        parts   = $haystack | split $needle
        return ($parts | arr_len) - 1
    }
    func! str_wrap(text, width) {
        words = $text | words
        line  = ""
        lines = []
        for word in $words {
            candidate = if $line == "": $word; else: "$line $word"
            if ($candidate | len) <= $width {
                line = $candidate
            } else {
                if $line != "": lines[] = $line
                line = $word
            }
        }
        if $line != "": lines[] = $line
        return $lines | arr_join "\n"
    }
    println "  string functions: slugify truncate pad_center starts_with_any count wrap"
}

# ────────────────────────────────────────────────────────────────────
# Simulate: lib/validation.ksh
# ────────────────────────────────────────────────────────────────────
println "--- loading lib/validation.ksh ---"
__VALIDATION_LOADED = ${__VALIDATION_LOADED:-false}
unless $__VALIDATION_LOADED {
    __VALIDATION_LOADED = true

    func! valid_email(e) {
        return ($e | contains "@") and ($e | contains ".")
    }
    func! valid_url(u) {
        return ($u | startswith "http://") or ($u | startswith "https://")
    }
    func! valid_range(n, lo, hi) {
        $n | isnum || return false
        v = tonum $n
        return $v >= $lo and $v <= $hi
    }
    func! valid_nonempty(s) {
        return ($s | trim | len) > 0
    }
    func! valid_maxlen(s, max) {
        return ($s | len) <= $max
    }
    func! validate_all(checks) {
        # checks = array of [result, message] pairs
        for check in $checks {
            parts = $check | split ":"
            ok    = $parts[0]
            msg   = $parts[1]
            $ok == "true" || throw $msg
        }
        return true
    }
    println "  validation functions: email url range nonempty maxlen validate_all"
}

# ────────────────────────────────────────────────────────────────────
# Main script — uses all three libraries
# ────────────────────────────────────────────────────────────────────
println ""
println "=== using lib/math.ksh ==="
println "gcd(48, 18) = $(math_gcd 48 18)"
println "lcm(12, 8)  = $(math_lcm 12 8)"
println "clamp(150, 0, 100) = $(math_clamp 150 0 100)"
println "lerp(0, 100, 0.25) = $(math_lerp 0 100 0.25)"

primes = []
for n in range(2..30) {
    if $(math_is_prime $n): primes[] = $n
}
println "primes 2-30: $primes"
println ""

println "=== using lib/strings.ksh ==="
titles = ["Hello World", "My Awesome Blog Post", "KatSH Scripting Guide"]
for t in $titles {
    slug     = str_slugify($t)
    short    = str_truncate($t, 20)
    centered = str_pad_center($t, 30)
    println "  '$t'"
    println "    slug='$slug'  truncated='$short'"
    println "    centered='$centered'"
}
println ""

long_text = "The quick brown fox jumps over the lazy dog and then ran away into the forest"
wrapped   = str_wrap($long_text, 30)
println "Wrapped at 30 chars:"
for line in $($wrapped | lines) {
    println "  |$line"
}
println ""

println "=== using lib/validation.ksh ==="
users = [
    "alice:alice@example.com:25",
    "bob:notanemail:30",
    "carol:carol@test.io:999",
    "dave::28",
    "eve:eve@company.com:35",
]
for row in $users {
    parts = $row | split ":"
    name  = $parts[0]
    email = $parts[1]
    age   = $parts[2]

    try {
        validate_all [
            "$(valid_nonempty $name):name is required",
            "$(valid_email $email):invalid email '$email'",
            "$(valid_range $age 18 120):age '$age' must be 18-120",
        ]
        println "  ✓ $name ($email, age=$age)"
    } catch e {
        println "  ✗ $name — $e"
    }
}
println ""

# ── Prevent double-source ─────────────────────────────────────────────
println "=== double-source guard ==="
# Source the same library again — the guard prevents re-running
unless $__MATH_LOADED {
    println "  math.ksh loading again (should not see this)"
}
unless $__STRINGS_LOADED {
    println "  strings.ksh loading again (should not see this)"
}
println "All libraries already loaded — skipped re-initialization ✓"