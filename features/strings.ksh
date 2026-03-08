#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  02_strings.ksh — String operations
#
#  Topics:
#    case · trim · length · search · transform · split/join
#    reverse · repeat · pad · sub · replace · concat
# ─────────────────────────────────────────────────────────────────────

s = "  Hello, World!  "

# ── Case & whitespace ─────────────────────────────────────────────────
println "Original:  '$s'"
println "upper:     '$(echo $s | upper)'"
println "lower:     '$(echo $s | lower)'"
println "title:     '$(echo $s | title)'"
println "trim:      '$(echo $s | trim)'"
println "ltrim:     '$(echo $s | ltrim)'"
println "rtrim:     '$(echo $s | rtrim)'"
println ""

# ── Inspecting strings ────────────────────────────────────────────────
clean = $s | trim
println "Cleaned: '$clean'"
println "len:         $($clean | len)"
println "startswith H: $($clean | startswith 'H')"
println "endswith !:   $($clean | endswith '!')"
println "contains Wor: $($clean | contains 'Wor')"
println "isnum:        $(echo $clean | isnum)"
println ""

# Type checks on different strings
for val in ["42", "3.14", "hello", "ABC", "abc", "  "] {
    n   = $val | isnum
    a   = $val | isalpha
    up  = $val | isupper
    lo  = $val | islower
    sp  = $val | isspace
    println "'$val'  isnum=$n  isalpha=$a  isupper=$up  islower=$lo  isspace=$sp"
}
println ""

# ── Transformation ────────────────────────────────────────────────────
word = "hello"
println "reverse:      $(echo $word | reverse)"
println "repeat 4:     $(echo $word | repeat 4)"
println "pad 15:       '$(echo $word | pad 15)'"
println "lpad 15:      '$(echo $word | lpad 15)'"
println "center 15:    '$(echo $word | center 15)'"
println ""

sentence = "the cat sat on the mat"
println "replace 'the' → 'a':   $(echo $sentence | replace 'the' 'a')"
println "replace1 'the' → 'a':  $(echo $sentence | replace1 'the' 'a')"
println ""

# sub(start) — from index
# sub(start, length) — slice
url = "https://example.com/api/users"
println "sub(8):     $(echo $url | sub 8)"
println "sub(0, 5):  $(echo $url | sub 0 5)"
println ""

# ── Split, join, concat ───────────────────────────────────────────────
csv = "Alice,Bob,Carol,Dave"
names = $csv | split ","
println "split:  $names"
println "join -: $(echo $names | arr_join '-')"
println "first:  $(echo $names | first)"
println "last:   $(echo $names | last)"
println ""

a = "Hello"
b = " World"
println "concat:  $(echo $a | concat $b)"
println "prepend: $(echo $b | prepend $a)"
println ""

# ── lines / words / chars ─────────────────────────────────────────────
text = "line one\nline two\nline three"
lines_arr = $text | lines
println "lines:  $lines_arr"
println "count:  $(echo $lines_arr | arr_len)"

sentence2 = "the quick brown fox"
println "words:  $(echo $sentence2 | words)"
println "chars of 'cat': $(echo 'cat' | chars)"