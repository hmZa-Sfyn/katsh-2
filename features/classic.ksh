#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  32_classic_exercises.ksh — Timeless coding exercises in KatSH
#
#  Exercises:
#    FizzBuzz · Collatz sequence · Caesar cipher
#    ROT13 · reverse words · count vowels
#    Run-length encoding · Pascal's triangle
#    Number to words · roman numerals
#    99 bottles · 12 days of Christmas (compressed)
# ─────────────────────────────────────────────────────────────────────

# ── FizzBuzz ──────────────────────────────────────────────────────────
println "=== FizzBuzz (1-30) ==="
result = []
for i in range(1..30) {
    out = match {
        $i % 15 == 0: "FizzBuzz"
        $i % 3  == 0: "Fizz"
        $i % 5  == 0: "Buzz"
        *:            tostr $i
    }
    result[] = $out
}
println "$result"
println ""

# ── Collatz sequence ──────────────────────────────────────────────────
println "=== Collatz sequence ==="
func collatz(n) {
    seq = [$n]
    while $n != 1 {
        if $n % 2 == 0 { n = $n / 2 } else { n = $n * 3 + 1 }
        seq[] = $n
    }
    return $seq
}

for start in [6, 11, 27] {
    seq = collatz($start)
    len = $seq | arr_len
    println "  $start → ... → 1  ($len steps)"
}
println ""

# ── Caesar cipher ─────────────────────────────────────────────────────
println "=== Caesar cipher ==="
func caesar_char(ch, shift) {
    alphabet = "abcdefghijklmnopqrstuvwxyz"
    pos = $alphabet | sub 0 26    # full string
    # Find char position
    for i in range(0..25) {
        c = $alphabet | sub $i 1
        if $c == $ch {
            new_pos = ($i + $shift) % 26
            return $alphabet | sub $new_pos 1
        }
    }
    return $ch   # non-alpha: unchanged
}

func caesar(text, shift) {
    lower_text = $text | lower
    result = ""
    for ch in $($lower_text | chars) {
        result = $result | concat $(caesar_char $ch $shift)
    }
    return $result
}

msg       = "Hello KatSH World"
encoded   = caesar($msg, 13)
decoded   = caesar($encoded, 13)     # ROT13 is its own inverse
println "  Original:  $msg"
println "  ROT13:     $encoded"
println "  Decoded:   $decoded"
println ""

# ── ROT13 (letter-swapping shortcut) ─────────────────────────────────
println "=== ROT13 ==="
func rot13(s) {
    result = ""
    for ch in $($s | chars) {
        is_up = $ch | isupper
        lch   = $ch | lower
        rotated = caesar_char $lch 13
        result = $result | concat $(if $is_up: ($rotated | upper); else: $rotated)
    }
    return $result
}

for phrase in ["Attack at dawn", "The Quick Brown Fox", "KatSH is cool"] {
    enc = rot13($phrase)
    dec = rot13($enc)
    println "  '$phrase'"
    println "    → '$enc'"
    println "    → '$dec'"
}
println ""

# ── Reverse words ─────────────────────────────────────────────────────
println "=== reverse words ==="
func reverse_words(s) {
    return $s | words | arr_reverse | arr_join " "
}

for s in ["Hello World", "the quick brown fox", "one"] {
    println "  '$s' → '$(reverse_words $s)'"
}
println ""

# ── Count vowels ─────────────────────────────────────────────────────
println "=== count vowels ==="
func count_vowels(s) {
    vowels = 0
    for ch in $($s | lower | chars) {
        match $ch {
            "a" | "e" | "i" | "o" | "u": vowels++
        }
    }
    return $vowels
}

for word in ["hello", "rhythm", "aeiou", "KatSH", "extraordinary"] {
    v = count_vowels($word)
    println "  '$word': $v vowel(s)"
}
println ""

# ── Run-length encoding ───────────────────────────────────────────────
println "=== run-length encoding ==="
func rle_encode(s) {
    chars = $s | chars
    result = ""
    i = 0
    n = $chars | arr_len
    while $i < $n {
        ch    = $chars[$i]
        count = 1
        while $i + $count < $n and $chars[$i + $count] == $ch {
            count++
        }
        result = $result | concat "$count$ch"
        i = $i + $count
    }
    return $result
}

func rle_decode(s) {
    result = ""
    chars  = $s | chars
    i      = 0
    while $i < ($chars | arr_len) {
        # Collect digits
        num_str = ""
        while $i < ($chars | arr_len) and ($chars[$i] | isnum) {
            num_str = $num_str | concat $chars[$i]
            i++
        }
        ch     = $chars[$i]
        count  = tonum $num_str
        result = $result | concat ($ch | repeat $count)
        i++
    }
    return $result
}

for s in ["aabbbcccc", "AAABBCDDDD", "abc", "aaaaaaaa"] {
    enc = rle_encode($s)
    dec = rle_decode($enc)
    println "  '$s' → '$enc' → '$dec'  $(if $dec == $s: '✓'; else: '✗')"
}
println ""

# ── Pascal's triangle ────────────────────────────────────────────────
println "=== Pascal's triangle (10 rows) ==="
row = [1]
for r in range(0..9) {
    # Print centered
    spaces = " " | repeat ((10 - $r) * 2)
    line   = $spaces
    for n in $row {
        line = $line | concat "$(tostr $n | lpad 4)"
    }
    println "$line"

    # Compute next row
    new_row = [1]
    for i in range(0..($row | arr_len - 2)) {
        new_row[] = tonum($row[$i]) + tonum($row[$i + 1])
    }
    new_row[] = 1
    row = $new_row
}
println ""

# ── Number to words (1-20) ────────────────────────────────────────────
println "=== number to words ==="
func num_to_word(n) {
    words_list = ["", "one","two","three","four","five","six","seven",
                  "eight","nine","ten","eleven","twelve","thirteen",
                  "fourteen","fifteen","sixteen","seventeen","eighteen",
                  "nineteen","twenty"]
    if $n >= 1 and $n <= 20: return $words_list[$n]
    if $n == 0: return "zero"
    return "unknown"
}

for n in range(1..20) {
    println "  $n = $(num_to_word $n)"
}
println ""

# ── Roman numerals ────────────────────────────────────────────────────
println "=== roman numerals ==="
func to_roman(n) {
    values  = [1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1]
    symbols = ["M","CM","D","CD","C","XC","L","XL","X","IX","V","IV","I"]
    result  = ""
    for i in range(0..($values | arr_len)) {
        val = $values[$i]
        sym = $symbols[$i]
        while $n >= $val {
            result = $result | concat $sym
            n      = $n - $val
        }
    }
    return $result
}

for n in [1, 4, 9, 14, 40, 90, 399, 400, 1994, 2024, 3999] {
    println "  $n = $(to_roman $n)"
}