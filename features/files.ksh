#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  17_files.ksh — File reading, parsing, and transformation
#
#  Topics:
#    cat / head / tail · wc · find · grep
#    line-by-line processing · CSV parsing
#    file inventory · extension grouping
#    diff · stat · size formatting
# ─────────────────────────────────────────────────────────────────────

script_dir = $__script_dir
parent_dir = $__script_dir | sub 0 ($(__script_dir | len) - ($(__script_dir | split "/" | last | len) + 1))

# ── List scripts in examples directory ───────────────────────────────
println "=== listing script files ==="
println "Looking in: $script_dir"
println ""

# Find all .ksh files
for f in `find $script_dir -name "*.ksh" -type f` {
    name = $f | split "/" | last
    println "  $name"
}
println ""

# ── Counting lines in files ───────────────────────────────────────────
println "=== line counts ==="
total_lines = 0
for f in `find $script_dir -name "*.ksh" -type f` {
    name   = $f | split "/" | last
    count  = $(wc -l < $f | tr -d ' ')
    total_lines = $total_lines + $count
    padded = $name | pad 35
    num    = tostr $count | lpad 5
    println "  $padded $num lines"
}
println "  $("─" | repeat 40)"
total_padded = "TOTAL" | pad 35
num_total = tostr $total_lines | lpad 5
println "  $total_padded $num_total lines"
println ""

# ── head / tail ───────────────────────────────────────────────────────
println "=== head of this file ==="
head -n 5 $__script_file
println ""

println "=== first script's shebang line ==="
first_script = "$script_dir/01_variables.ksh"
head -n 1 $first_script
println ""

# ── Searching inside files ────────────────────────────────────────────
println "=== find all 'func' definitions across examples ==="
func_defs = $(grep -h "^func " $script_dir/*.ksh | sort)
println "$func_defs"
println ""

# Count unique function names
func_count = $(grep -h "^func " $script_dir/*.ksh | wc -l | tr -d ' ')
println "Total functions defined: $func_count"
println ""

# ── Word frequency in scripts ─────────────────────────────────────────
println "=== most used keywords in examples ==="
keyword_counts = map {}
keywords = ["if", "for", "func", "return", "println", "match", "try", "while", "catch"]

for kw in $keywords {
    count = $(grep -h "\b$kw\b" $script_dir/*.ksh 2>/dev/null | wc -l | tr -d ' ')
    map_set keyword_counts $kw $count
}

# Show sorted by count
pairs = []
for kw in $(map_keys keyword_counts) {
    n = map_get keyword_counts $kw
    pairs[] = "$n:$kw"
}
sorted = $pairs | arr_sort | arr_reverse
for pair in $sorted {
    parts = $pair | split ":"
    n   = $parts[0]
    kw  = $parts[1]
    bar = "▪" | repeat ($n / 3 + 1)
    padded = $kw | pad 10
    println "  $padded $n $bar"
}
println ""

# ── Simulated CSV parsing ─────────────────────────────────────────────
println "=== CSV parsing ==="

# Write a temp CSV to demonstrate
csv_content = "name,department,salary,years
Alice,Engineering,95000,5
Bob,Design,72000,3
Carol,Engineering,105000,8
Dave,Management,115000,10
Eve,Design,78000,4
Frank,Engineering,88000,6
Grace,Management,98000,7"

# Parse header
lines_arr = $csv_content | lines
header = $lines_arr[0] | split ","
println "Columns: $header"
println ""

# Parse rows
rows = []
total_salary = 0
dept_counts  = map {}

for i in range(1..6) {
    row  = $lines_arr[$i] | split ","
    name = $row[0]
    dept = $row[1]
    sal  = tonum $row[2]
    yrs  = tonum $row[3]

    total_salary = $total_salary + $sal

    # Count by department
    cur = map_get dept_counts $dept
    if $cur == "": cur = 0
    map_set dept_counts $dept ($cur + 1)

    rows[] = "$name|$dept|$sal|$yrs"
}

# Summary
avg_salary = $total_salary / 6
println "Total employees: 6"
println "Total salary:    \$$total_salary"
println "Average salary:  \$$avg_salary"
println ""

println "By department:"
for dept in $(map_keys dept_counts) {
    count = map_get dept_counts $dept
    println "  $dept: $count employees"
}
println ""

# High earners
println "High earners (>90k):"
for row in $rows {
    parts = $row | split "|"
    name  = $parts[0]
    dept  = $parts[1]
    sal   = tonum $parts[2]
    if $sal > 90000 {
        println "  $name ($dept): \$$sal"
    }
}
println ""

# ── File extension inventory ──────────────────────────────────────────
println "=== file extension inventory ==="
ext_counts = map {}

# Count file extensions in /usr/bin (or similar)
for f in `find $script_dir -type f` {
    name = $f | split "/" | last
    has_dot = $name | contains "."
    if $has_dot {
        ext = $name | split "." | last
    } else {
        ext = "(no extension)"
    }
    cur = map_get ext_counts $ext
    if $cur == "": cur = 0
    map_set ext_counts $ext ($cur + 1)
}

println "Script directory file types:"
for ext in $(map_keys ext_counts) {
    count = map_get ext_counts $ext
    println "  .$ext: $count"
}
println ""

# ── Temp file pattern ────────────────────────────────────────────────
println "=== temp file usage pattern ==="
tmp = $(mktemp)
echo "line 1: hello" > $tmp
echo "line 2: world" >> $tmp
echo "line 3: katsh" >> $tmp

println "Temp file: $tmp"
println "Contents:"
bash! cat "$tmp"
println ""
line_count = $(wc -l < $tmp | tr -d ' ')
println "Lines: $line_count"
bash! rm "$tmp"
println "Temp file cleaned up."