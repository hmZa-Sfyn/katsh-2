#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  29_etl_pipeline.ksh — Extract, Transform, Load pipeline
#
#  Topics covered:
#    multi-stage pipeline design · data validation at each step
#    map-based records · aggregate/group-by patterns
#    filter/transform/reduce · pipeline error handling
#    output formatting (CSV, table, summary)
# ─────────────────────────────────────────────────────────────────────

# ══════════════════════════════════════════════════════════════════════
#  STAGE 1: EXTRACT — parse raw input data
# ══════════════════════════════════════════════════════════════════════

# Simulated raw data (CSV string)
RAW_DATA = "id,name,department,salary,start_date,active
1,Alice Smith,Engineering,95000,2020-03-15,true
2,Bob Jones,Design,72000,2021-07-01,true
3,Carol White,Engineering,105000,2019-01-10,true
4,Dave Brown,Management,115000,2018-06-20,true
5,Eve Davis,Design,78000,2022-02-28,true
6,Frank Wilson,Engineering,88000,2021-11-05,true
7,Grace Lee,Management,98000,2020-08-12,true
8,Hank Moore,Engineering,91000,2023-01-20,true
9,Iris Taylor,Design,69000,2022-09-01,false
10,Jack Anderson,Engineering,,2021-04-15,true"

func extract(raw_csv) {
    println "[EXTRACT] Parsing CSV input..."
    all_lines = $raw_csv | lines
    header    = $all_lines[0] | split ","
    records   = []

    for i in range(1..($all_lines | arr_len)) {
        line = $all_lines[$i]
        line = $line | trim
        continue when $line == ""
        fields = $line | split ","
        rec = map {}
        for j in range(0..($header | arr_len)) {
            key = $header[$j]
            val = $fields[$j]
            map_set rec $key $val
        }
        records[] = $rec
    }
    println "[EXTRACT] Loaded $(echo $records | arr_len) raw records"
    return $records
}

# ══════════════════════════════════════════════════════════════════════
#  STAGE 2: VALIDATE — check data quality
# ══════════════════════════════════════════════════════════════════════

func validate_record(rec) {
    id     = map_get rec "id"
    name   = map_get rec "name"
    dept   = map_get rec "department"
    salary = map_get rec "salary"
    active = map_get rec "active"

    $id     != "" || throw "missing id"
    $name   != "" || throw "missing name (id=$id)"
    $dept   != "" || throw "missing department (id=$id)"
    $salary != "" || throw "missing salary (id=$id)"
    $salary | isnum || throw "salary not numeric: '$salary' (id=$id)"
    tonum($salary) > 0 || throw "salary must be positive (id=$id)"
    ($active == "true" or $active == "false") || throw "active must be bool (id=$id)"
}

func validate(records) {
    println "[VALIDATE] Checking $(echo $records | arr_len) records..."
    valid   = []
    invalid = []
    errors  = []

    for rec in $records {
        try {
            validate_record $rec
            valid[] = $rec
        } catch e {
            id = map_get rec "id"
            errors[] = "Record $id: $e"
            invalid[] = $rec
        }
    }

    if $(echo $errors | arr_len) > 0 {
        println "[VALIDATE] ⚠️  $(echo $errors | arr_len) invalid record(s):"
        for err in $errors {
            println "             $err"
        }
    }
    println "[VALIDATE] $(echo $valid | arr_len) valid records"
    return $valid
}

# ══════════════════════════════════════════════════════════════════════
#  STAGE 3: TRANSFORM — enrich and reshape
# ══════════════════════════════════════════════════════════════════════

func transform_record(rec) {
    name    = map_get rec "name"
    salary  = tonum $(map_get rec "salary")
    dept    = map_get rec "department"
    start   = map_get rec "start_date"

    # Derive: first/last name
    parts = $name | split " "
    map_set rec "first_name" $parts[0]
    map_set rec "last_name"  $parts[1]

    # Derive: salary band
    band = match $salary {
        >=110000: "Executive"
        >=95000:  "Senior"
        >=80000:  "Mid"
        *:        "Junior"
    }
    map_set rec "salary_band" $band

    # Derive: annual bonus (10% senior+, 5% others)
    bonus_pct = if $salary >= 95000: 10; else: 5
    bonus     = $salary * $bonus_pct / 100
    map_set rec "bonus"     $bonus
    map_set rec "bonus_pct" "$bonus_pct%"

    # Derive: start year
    start_year = $start | sub 0 4
    map_set rec "start_year" $start_year

    return $rec
}

func transform(records) {
    println "[TRANSFORM] Enriching $(echo $records | arr_len) records..."
    transformed = []
    for rec in $records {
        enriched = transform_record($rec)
        transformed[] = $enriched
    }
    println "[TRANSFORM] Done"
    return $transformed
}

# ══════════════════════════════════════════════════════════════════════
#  STAGE 4: AGGREGATE — compute summaries
# ══════════════════════════════════════════════════════════════════════

func aggregate(records) {
    println "[AGGREGATE] Computing department summaries..."
    dept_salary = map {}
    dept_count  = map {}
    dept_bonus  = map {}

    for rec in $records {
        dept   = map_get rec "department"
        salary = tonum $(map_get rec "salary")
        bonus  = tonum $(map_get rec "bonus")

        cur_sal = map_get dept_salary $dept
        cur_cnt = map_get dept_count  $dept
        cur_bon = map_get dept_bonus  $dept
        if $cur_sal == "": cur_sal = 0
        if $cur_cnt == "": cur_cnt = 0
        if $cur_bon == "": cur_bon = 0

        map_set dept_salary $dept ($cur_sal + $salary)
        map_set dept_count  $dept ($cur_cnt + 1)
        map_set dept_bonus  $dept ($cur_bon + $bonus)
    }

    # Build summary records
    summaries = []
    for dept in $(map_keys dept_salary) {
        total = tonum $(map_get dept_salary $dept)
        count = tonum $(map_get dept_count  $dept)
        bonus = tonum $(map_get dept_bonus  $dept)
        avg   = $total / $count

        s = map {}
        map_set s "department"  $dept
        map_set s "headcount"   $count
        map_set s "total_salary" $total
        map_set s "avg_salary"   $avg
        map_set s "total_bonus"  $bonus
        summaries[] = $s
    }
    return $summaries
}

# ══════════════════════════════════════════════════════════════════════
#  STAGE 5: LOAD — output results
# ══════════════════════════════════════════════════════════════════════

func ljust(s, w) { return tostr($s) | pad $w }
func rjust(s, w) { return tostr($s) | lpad $w }

func load_employee_report(records) {
    println ""
    println "╔══════════════════════════════════════════════════════════════════╗"
    println "║                   EMPLOYEE REPORT                               ║"
    println "╠══════════════════════════════════════════════════════════════════╣"
    println "║ $(ljust 'Name' 18) $(ljust 'Dept' 14) $(rjust 'Salary' 8) $(ljust 'Band' 10) $(rjust 'Bonus' 8) ║"
    println "╠══════════════════════════════════════════════════════════════════╣"

    for rec in $records {
        name   = map_get rec "name"
        dept   = map_get rec "department"
        salary = map_get rec "salary"
        band   = map_get rec "salary_band"
        bonus  = map_get rec "bonus"
        println "║ $(ljust $name 18) $(ljust $dept 14) $(rjust $salary 8) $(ljust $band 10) $(rjust $bonus 8) ║"
    }
    println "╚══════════════════════════════════════════════════════════════════╝"
}

func load_dept_summary(summaries) {
    println ""
    println "╔══════════════════════════════════════════╗"
    println "║           DEPARTMENT SUMMARY             ║"
    println "╠══════════════════════════════════════════╣"
    println "║ $(ljust 'Department' 14) $(rjust 'HC' 3) $(rjust 'Avg Salary' 11) $(rjust 'Total Bonus' 11) ║"
    println "╠══════════════════════════════════════════╣"
    for s in $summaries {
        dept  = map_get s "department"
        hc    = map_get s "headcount"
        avg   = map_get s "avg_salary"
        bonus = map_get s "total_bonus"
        println "║ $(ljust $dept 14) $(rjust $hc 3) $(rjust $avg 11) $(rjust $bonus 11) ║"
    }
    println "╚══════════════════════════════════════════╝"
}

# ══════════════════════════════════════════════════════════════════════
#  RUN THE PIPELINE
# ══════════════════════════════════════════════════════════════════════
println "═══════════════════════════════════════"
println "  ETL Pipeline — HR Data Processing"
println "═══════════════════════════════════════"
println ""

raw        = extract($RAW_DATA)
validated  = validate($raw)
enriched   = transform($validated)
summaries  = aggregate($enriched)

load_employee_report $enriched
load_dept_summary    $summaries

# Grand totals
total_headcount    = $enriched | arr_len
total_payroll      = 0
total_bonus_amount = 0
for rec in $enriched {
    total_payroll      = $total_payroll      + tonum $(map_get rec "salary")
    total_bonus_amount = $total_bonus_amount + tonum $(map_get rec "bonus")
}
avg_payroll = $total_payroll / $total_headcount

println ""
println "  Total headcount:  $total_headcount"
println "  Total payroll:    \$$total_payroll"
println "  Average salary:   \$$avg_payroll"
println "  Total bonuses:    \$$total_bonus_amount"
println ""
println "[PIPELINE] Complete ✓"