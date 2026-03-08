#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  10_pipes.ksh — Pipe operators | and pipe expression |>
#
#  Topics:
#    | pipe for tables (select/where/sort/limit)
#    | pipe for text transformations
#    |> value pipeline (F#-style)
#    chaining with user functions
#    storing pipe results · #= operator
# ─────────────────────────────────────────────────────────────────────

# ── String value pipeline ─────────────────────────────────────────────
println "=== string | pipeline ==="
result = "  Hello, World!  " | trim | lower | replace "," "" | replace "!" ""
println "  result: '$result'"
println ""

# ── Array value pipeline ──────────────────────────────────────────────
println "=== array | pipeline ==="
words = "the quick brown fox jumps over the lazy dog" | words
println "  words:         $words"
println "  sorted:        $(echo $words | arr_sort)"
println "  unique sorted: $(echo $words | arr_uniq | arr_sort)"
println "  long words:    $(echo $words | arr_filter len)"    # len > 0 = truthy when > 3 chars
println ""

# ── Number | pipeline ─────────────────────────────────────────────────
println "=== number | pipeline ==="
println "  2 | pow 10 | add 24 | sqrt = $(2 | pow 10 | add 24 | sqrt)"
println "  255 | hex = $(255 | hex)"
println "  -42 | abs | sqrt = $(-42 | abs | sqrt)"
println ""

# ── |> pipe expression (F# / Elixir style) ───────────────────────────
println "=== |> value pipeline ==="

# Basic chain
r1 = "  HELLO WORLD  " |> trim |> lower |> replace "world" "katsh"
println "  r1 = '$r1'"

# Numeric pipeline
r2 = 144 |> sqrt |> mul 3 |> add 6 |> negate
println "  r2 = $r2"

# Array pipeline
r3 = "fox the quick brown" |> split " " |> arr_sort |> arr_join ", "
println "  r3 = '$r3'"

# String pipeline producing a number
r4 = "hello world katsh scripting rocks" |> words |> arr_len
println "  r4 = $r4 (word count)"
println ""

# ── User functions in |> ──────────────────────────────────────────────
println "=== user functions in |> ==="
func double(n)     { return $n * 2 }
func increment(n)  { return $n + 1 }
func square(n)     { return $n * $n }
func is_big(n)     { return $n > 100 }

r5 = 3 |> double |> increment |> square
println "  3 |> double |> increment |> square = $r5"

r6 = 5 |> square |> double |> increment
println "  5 |> square |> double |> increment = $r6"
println ""

# ── Building a text processing pipeline ──────────────────────────────
println "=== text processing pipeline ==="
func extract_words(text) {
    return $text | lower | replace "," "" | replace "." "" | words
}

func count_words(arr) {
    return $arr | arr_len
}

texts = [
    "The quick brown fox.",
    "Pack my box with five dozen liquor jugs.",
    "How vexingly quick daft zebras jump!",
]

for text in $texts {
    wcount = extract_words($text) | arr_len
    longest = extract_words($text) | arr_sort | last
    println "  '$text'"
    println "    words=$wcount  longest='$longest'"
}
println ""

# ── select / where / sort on command output ───────────────────────────
println "=== table pipe transforms ==="
# ps output is auto-parsed as a table
# (columns vary by OS — using array approach to demonstrate concept)

data = [
    "alice:developer:95000",
    "bob:designer:72000",
    "carol:developer:98000",
    "dave:manager:110000",
    "eve:designer:75000",
    "frank:developer:87000",
]

# Parse into a map array
println "  Developers earning > 90000:"
for row in $data {
    parts = $row | split ":"
    name   = $parts[0]
    role   = $parts[1]
    salary = tonum $parts[2]
    if $role == "developer" and $salary > 90000 {
        println "    $name: \$$salary"
    }
}
println ""

# ── #= store pipeline result in box ──────────────────────────────────
println "=== #= store result ==="
# Store for reuse later in the session:
# ls | where name ~ ".ksh" #= ksh_files
# box get ksh_files | sort name

# Demonstrate the concept with arrays:
top_scores = [88, 92, 75, 99, 83, 67, 95, 71, 88, 90]
sorted_top = $top_scores | arr_sort | arr_reverse | slice 0 4
println "  Top 5 scores: $sorted_top"
println ""

# ── Chained pipelines in assignments ──────────────────────────────────
println "=== complex pipeline chains ==="
raw_csv = "alice,30,eng\nbob,25,sales\ncarol,35,eng\ndave,28,eng\neve,32,sales"
rows    = $raw_csv | lines

eng_names = []
for row in $rows {
    parts = $row | split ","
    name  = $parts[0]
    dept  = $parts[2]
    eng_names[] = $name when $dept == "eng"
}

result_str = $eng_names | arr_sort | arr_join ", "
println "  Engineering team: $result_str"
println "  Count: $(echo $eng_names | arr_len)"