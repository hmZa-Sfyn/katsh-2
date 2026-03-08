#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  26_templates.ksh — String templates and report generation
#
#  Topics covered:
#    multi-line string building · sprintf-style formatting
#    table rendering with pad/lpad/center · banner/box drawing
#    template substitution with map · generating Markdown/HTML/CSV
#    column-aligned reports · summary cards
# ─────────────────────────────────────────────────────────────────────

# ── sprintf-style formatting helpers ─────────────────────────────────
func rjust(s, w)  { return tostr($s) | lpad  $w }    # right-align
func ljust(s, w)  { return tostr($s) | pad   $w }    # left-align
func cjust(s, w)  { return tostr($s) | center $w }   # center
func hr(ch, w)    { return $ch | repeat $w }          # horizontal rule

# ── Column-aligned table printer ──────────────────────────────────────
println "=== column-aligned table ==="

func print_table(headers, rows, widths) {
    # Header row
    line = ""
    for i in range(0..($headers | arr_len)) {
        h = $headers[$i]
        w = tonum $widths[$i]
        line = $line | concat "$(ljust $h $w)  "
    }
    println "$line"

    # Separator
    sep = ""
    for w in $widths {
        sep = $sep | concat "$(hr '─' $w)  "
    }
    println "$sep"

    # Data rows
    for row in $rows {
        cols = $row | split "|"
        line2 = ""
        for i in range(0..($cols | arr_len)) {
            c = $cols[$i]
            w = tonum $widths[$i]
            line2 = $line2 | concat "$(ljust $c $w)  "
        }
        println "$line2"
    }
}

headers = ["Name", "Department", "Salary", "Yrs"]
widths  = ["15", "14", "9", "4"]
rows    = [
    "Alice|Engineering|$95,000|5",
    "Bob|Design|$72,000|3",
    "Carol|Engineering|$105,000|8",
    "Dave|Management|$115,000|10",
    "Eve|Design|$78,000|4",
]

print_table $headers $rows $widths
println ""

# ── Right-aligned numeric columns ─────────────────────────────────────
println "=== right-aligned numbers ==="
data = [
    ["Revenue",    "12,847,392"],
    ["Expenses",    "9,231,050"],
    ["Net Profit",  "3,616,342"],
    ["Tax (22%)",     "795,595"],
    ["After Tax",   "2,820,747"],
]

println "$(ljust 'Item' 15) $(rjust 'Amount ($)' 14)"
println "$(hr '─' 15) $(hr '─' 14)"
for row in $data {
    parts = $row | split "|"
    label = $parts[0]
    amount = $parts[1]
    println "$(ljust $label 15) $(rjust $amount 14)"
}
println "$(hr '═' 15) $(hr '═' 14)"
println ""

# ── Banner / box drawing ───────────────────────────────────────────────
println "=== box / banner drawing ==="
func draw_box(title, lines_arr) {
    max_w = $title | len
    for line in $lines_arr {
        l = $line | len
        max_w = if $l > $max_w: $l; else: $max_w
    }
    inner_w = $max_w + 2
    border  = hr "─" $inner_w

    println "┌${border}┐"
    println "│ $(cjust $title $max_w) │"
    println "├${border}┤"
    for line in $lines_arr {
        println "│ $(ljust $line $max_w) │"
    }
    println "└${border}┘"
}

draw_box "Deployment Summary" [
    "Environment : production",
    "Version     : v2.1.4",
    "Status      : ✓ Success",
    "Duration    : 4m 12s",
    "Deployed by : alice",
]
println ""

# ── Template substitution with map ────────────────────────────────────
println "=== template substitution ==="
func render(template, vars) {
    result = $template
    for key in $(map_keys vars) {
        val    = map_get vars $key
        result = $result | replace "{{$key}}" $val
    }
    return $result
}

email_template = "Dear {{name}},

Your order #{{order_id}} has been {{status}}.
Item: {{item}} × {{qty}}
Total: \${{total}}

Thank you for shopping with {{store}}.
— The {{store}} Team"

vars = map {
    name     = Alice
    order_id = ORD-8821
    status   = shipped
    item     = "Mechanical Keyboard"
    qty      = 1
    total    = "149.99"
    store    = "TechShop"
}

rendered = render($email_template, $vars)
println "$rendered"
println ""

# ── Generating Markdown ────────────────────────────────────────────────
println "=== Markdown generation ==="
func md_header(level, text) {
    prefix = "#" | repeat $level
    return "$prefix $text"
}
func md_bold(text)   { return "**$text**" }
func md_code(text)   { return "\`$text\`" }
func md_link(text, url) { return "[$text]($url)" }
func md_li(text)     { return "- $text" }
func md_table_row(cols, widths) {
    row = "|"
    for i in range(0..($cols | arr_len)) {
        c = $cols[$i]
        w = tonum $widths[$i]
        row = $row | concat " $(ljust $c $w) |"
    }
    return $row
}

report = []
report[] = "$(md_header 1 'KatSH Feature Report')"
report[] = ""
report[] = "$(md_header 2 'Data Types')"
report[] = ""
report[] = "$(md_li "$(md_bold 'string') — UTF-8 text with rich pipe operations")"
report[] = "$(md_li "$(md_bold 'number') — IEEE 754 float with math ops")"
report[] = "$(md_li "$(md_bold 'array')  — ordered list, zero-indexed")"
report[] = "$(md_li "$(md_bold 'map')    — key-value hash table")"
report[] = "$(md_li "$(md_bold 'set')    — unique value collection")"
report[] = ""
report[] = "$(md_header 2 'Quick Example')"
report[] = ""
report[] = "$(md_code 'x = "hello world" |> trim |> upper |> replace "WORLD" "KATSH"')"
report[] = ""
report[] = "See $(md_link 'full docs' 'https://github.com/example/katsh') for more."

for line in $report {
    println "$line"
}
println ""

# ── Generating CSV ─────────────────────────────────────────────────────
println "=== CSV generation ==="
func csv_quote(s) {
    # Quote field if it contains comma or quote
    if $s ~ "," or $s ~ '"' {
        escaped = $s | replace '"' '""'
        return '"$escaped"'
    }
    return $s
}

func csv_row(fields) {
    quoted = $fields | arr_map csv_quote
    return $quoted | arr_join ","
}

csv_lines = []
csv_lines[] = csv_row(["Name", "Role", "Salary", "Start Date"])
csv_lines[] = csv_row(["Alice Smith", "Engineer", "95000", "2020-03-15"])
csv_lines[] = csv_row(["Bob, Jr.", "Designer", "72000", "2021-07-01"])
csv_lines[] = csv_row(['Carol "CJ" Jones', "Manager", "115000", "2019-01-10"])

for row in $csv_lines {
    println "$row"
}
println ""

# ── Progress / status summary card ────────────────────────────────────
println "=== status summary card ==="
func status_icon(ok) {
    return if $ok: "✅"; else: "❌"
}

checks = [
    ["Database connection", true],
    ["Cache server",        true],
    ["Email service",       false],
    ["Storage bucket",      true],
    ["API gateway",         false],
]

draw_box "System Health Check" []
println "┌────────────────────────────────────┐"
pass_count = 0
fail_count = 0
for check in $checks {
    name = $check[0]
    ok   = $check[1]
    icon = status_icon($ok)
    if $ok: pass_count++; else: fail_count++
    println "│  $icon  $(ljust $name 28) │"
}
println "├────────────────────────────────────┤"
total = $pass_count + $fail_count
println "│  $(ljust "Passed: $pass_count / $total" 34) │"
println "└────────────────────────────────────┘"