#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  28_unit_testing.ksh — Build and use a unit testing framework
#
#  Topics covered:
#    assert helpers · test suites · pass/fail tracking
#    expect_equal / expect_true / expect_error / expect_contains
#    test isolation · summary report · TAP-style output
#    testing string ops, number ops, functions
# ─────────────────────────────────────────────────────────────────────

# ── Test framework globals ────────────────────────────────────────────
_tests_run    = 0
_tests_passed = 0
_tests_failed = 0
_current_suite = ""

# ── Framework functions ───────────────────────────────────────────────
func suite(name) {
    _current_suite = $name
    println ""
    println "  ▶ $name"
}

func assert_eq(description, got, expected) {
    _tests_run++
    if $got == $expected {
        _tests_passed++
        println "    ✅  $description"
    } else {
        _tests_failed++
        println "    ❌  $description"
        println "        expected: '$expected'"
        println "        got:      '$got'"
    }
}

func assert_true(description, val) {
    assert_eq $description $val "true"
}

func assert_false(description, val) {
    assert_eq $description $val "false"
}

func assert_contains(description, haystack, needle) {
    _tests_run++
    result = $haystack | contains $needle
    if $result == "true" {
        _tests_passed++
        println "    ✅  $description"
    } else {
        _tests_failed++
        println "    ❌  $description"
        println "        '$haystack' does not contain '$needle'"
    }
}

func assert_throws(description, fn_name, fn_arg) {
    _tests_run++
    threw = false
    try {
        $fn_name $fn_arg
    } catch e {
        threw = true
    }
    if $threw {
        _tests_passed++
        println "    ✅  $description"
    } else {
        _tests_failed++
        println "    ❌  $description (expected throw but none occurred)"
    }
}

func test_summary() {
    println ""
    println "  ─────────────────────────────"
    total  = $_tests_run
    passed = $_tests_passed
    failed = $_tests_failed
    pct    = if $total > 0: ($passed * 100 / $total); else: 0

    println "  Results: $passed / $total passed ($pct%)"
    if $failed > 0 {
        println "  FAILED: $failed test(s)"
    } else {
        println "  All tests passed ✓"
    }
    println ""
}

# ══════════════════════════════════════════════════════════════════════
#  Functions under test
# ══════════════════════════════════════════════════════════════════════

func slugify(s) {
    return $s | lower | trim | replace " " "-" | replace "--" "-"
}

func clamp(val, lo, hi) {
    if $val < $lo: return $lo
    if $val > $hi: return $hi
    return $val
}

func palindrome(s) {
    clean = $s | lower | replace " " ""
    return $clean == ($clean | reverse)
}

func safe_div(a, b) {
    $b != 0 || throw "division by zero"
    return $a / $b
}

func word_count(text) {
    return $text | trim | words | arr_len
}

func title_case(s) {
    return $s | title
}

func is_valid_email(e) {
    return ($e | contains "@") and ($e | contains ".")
}

func fibonacci(n) {
    if $n <= 1: return $n
    a = 0
    b = 1
    for i in range(2..$n) {
        tmp = $b
        b   = $a + $b
        a   = $tmp
    }
    return $b
}

# ══════════════════════════════════════════════════════════════════════
#  Test suites
# ══════════════════════════════════════════════════════════════════════
println "═══════════════════════════════════"
println "  KatSH Unit Tests"
println "═══════════════════════════════════"

# ── slugify ───────────────────────────────────────────────────────────
suite "slugify()"
assert_eq "lowercase"                  $(slugify "Hello World")    "hello-world"
assert_eq "trims whitespace"           $(slugify "  hello  ")      "hello"
assert_eq "single word"                $(slugify "katsh")          "katsh"
assert_eq "already slug"               $(slugify "my-post-title")  "my-post-title"
assert_eq "multiple spaces"            $(slugify "a  b  c")        "a-b-c"

# ── clamp ─────────────────────────────────────────────────────────────
suite "clamp(val, lo, hi)"
assert_eq "within range"    $(clamp 50 0 100)   "50"
assert_eq "below min"       $(clamp -5 0 100)   "0"
assert_eq "above max"       $(clamp 200 0 100)  "100"
assert_eq "at min"          $(clamp 0 0 100)    "0"
assert_eq "at max"          $(clamp 100 0 100)  "100"
assert_eq "negative range"  $(clamp -3 -10 -1)  "-3"

# ── palindrome ────────────────────────────────────────────────────────
suite "palindrome()"
assert_true  "racecar"              $(palindrome "racecar")
assert_true  "level"                $(palindrome "level")
assert_true  "A man a plan a canal" $(palindrome "amanaplanacanal")
assert_false "hello"                $(palindrome "hello")
assert_false "world"                $(palindrome "world")
assert_true  "single char"         $(palindrome "x")

# ── safe_div ──────────────────────────────────────────────────────────
suite "safe_div(a, b)"
assert_eq    "10 / 2 = 5"          $(safe_div 10 2)   "5"
assert_eq    "9 / 3 = 3"           $(safe_div 9 3)    "3"
assert_eq    "7 / 2 = 3.5"         $(safe_div 7 2)    "3.5"
assert_throws "div by zero throws"  safe_div           "0"

# ── word_count ────────────────────────────────────────────────────────
suite "word_count()"
assert_eq "three words"           $(word_count "hello world foo")   "3"
assert_eq "single word"           $(word_count "hello")             "1"
assert_eq "extra whitespace"      $(word_count "  a  b  c  ")       "3"
assert_eq "empty string"          $(word_count "")                  "0"

# ── is_valid_email ────────────────────────────────────────────────────
suite "is_valid_email()"
assert_true  "user@example.com"   $(is_valid_email "user@example.com")
assert_true  "a@b.io"             $(is_valid_email "a@b.io")
assert_false "no at sign"         $(is_valid_email "noatsign.com")
assert_false "no dot"             $(is_valid_email "user@nodot")
assert_false "empty"              $(is_valid_email "")

# ── fibonacci ─────────────────────────────────────────────────────────
suite "fibonacci(n)"
assert_eq "fib(0) = 0"  $(fibonacci 0)   "0"
assert_eq "fib(1) = 1"  $(fibonacci 1)   "1"
assert_eq "fib(2) = 1"  $(fibonacci 2)   "1"
assert_eq "fib(5) = 5"  $(fibonacci 5)   "5"
assert_eq "fib(10)= 55" $(fibonacci 10)  "55"
assert_eq "fib(15)=610" $(fibonacci 15)  "610"

# ── String operations (built-in) ──────────────────────────────────────
suite "built-in string ops"
assert_eq "upper"          $("hello" | upper)           "HELLO"
assert_eq "lower"          $("WORLD" | lower)           "world"
assert_eq "trim both"      $("  hi  " | trim)            "hi"
assert_eq "len"            $("hello" | len)              "5"
assert_eq "reverse"        $("abcd" | reverse)           "dcba"
assert_eq "replace all"    $("aaa" | replace "a" "b")    "bbb"
assert_eq "startswith"     $("hello" | startswith "he")  "true"
assert_eq "endswith"       $("hello" | endswith "lo")    "true"
assert_eq "contains"       $("hello" | contains "ell")   "true"
assert_eq "repeat"         $("ab" | repeat 3)            "ababab"

# ── Array operations (built-in) ───────────────────────────────────────
suite "built-in array ops"
arr = [3, 1, 4, 1, 5, 9, 2, 6]
assert_eq "arr_len"     $(echo $arr | arr_len)           "8"
assert_eq "first"       $(echo $arr | first)             "3"
assert_eq "last"        $(echo $arr | last)              "6"
assert_eq "arr_min"     $(echo $arr | arr_min)           "1"
assert_eq "arr_max"     $(echo $arr | arr_max)           "9"
assert_eq "arr_sum"     $(echo $arr | arr_sum)           "31"
assert_true  "arr_contains 5"    $(echo $arr | arr_contains 5)
assert_false "arr_contains 99"   $(echo $arr | arr_contains 99)

# ── Summary ───────────────────────────────────────────────────────────
test_summary