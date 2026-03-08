#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  19_box.ksh — Box session storage
#
#  Topics:
#    box set / get · #= pipe storage · list all keys
#    tag / rename / rm / clear
#    box export / import · hashing stored data
#    using box as a cache · building a session "database"
# ─────────────────────────────────────────────────────────────────────

# ── Basic set / get ───────────────────────────────────────────────────
println "=== basic box set / get ==="
box set greeting    "Hello from Box"
box set counter     42
box set server_host "api.example.com"
box set debug_mode  true

println "greeting:    $(box get greeting)"
println "counter:     $(box get counter)"
println "server_host: $(box get server_host)"
println "debug_mode:  $(box get debug_mode)"
println ""

# ── List all keys ─────────────────────────────────────────────────────
println "=== list all box keys ==="
box
println ""

# ── #= operator — store pipeline result ──────────────────────────────
println "=== #= pipe storage ==="

# Store command output
find $__script_dir -name "*.ksh" -type f #= all_scripts

println "Stored as 'all_scripts':"
box get all_scripts
println ""

# ── Retrieve and further process stored results ───────────────────────
println "=== retrieve + process stored result ==="
# (Uncomment in a real session — box results persist between commands)
# box get all_scripts | sort name | limit 5

# Demonstrate with array:
nums = [42, 17, 85, 3, 60, 29, 78, 51, 95, 12]
box set raw_nums "$nums"

stored = $(box get raw_nums)
println "Stored:        $stored"
println "Sorted:        $(echo $stored | arr_sort)"
println "Top 3:         $(echo $stored | arr_sort | arr_reverse | slice 0 2)"
println ""

# ── Tagging ───────────────────────────────────────────────────────────
println "=== tagging ==="
box set user_data   "[alice,bob,carol]"
box set prod_config "host=prod.example.com port=443"
box set dev_config  "host=localhost port=8080"

box tag user_data   type=data   source=api    env=all
box tag prod_config type=config env=production
box tag dev_config  type=config env=development

box
println ""

# ── Rename ────────────────────────────────────────────────────────────
println "=== rename ==="
box set temp_result "temporary value"
box rename temp_result final_result
println "renamed: $(box get final_result)"
println ""

# ── Hashing ──────────────────────────────────────────────────────────
println "=== hashing stored values ==="
box set important_data "user:alice:score:99:active:true"
println "md5:    $(box md5    important_data)"
println "sha1:   $(box sha1   important_data)"
println "sha256: $(box sha256 important_data)"
println ""

# ── Box as a cache ────────────────────────────────────────────────────
println "=== box as computation cache ==="
func fib_cached(n) {
    cache_key = "fib_$n"

    # Check cache first
    cached = $(box get $cache_key 2>/dev/null)
    if $cached != "" {
        return $cached
    }

    # Compute
    if $n <= 1 {
        result = $n
    } else {
        a = fib_cached($n - 1)
        b = fib_cached($n - 2)
        result = $a + $b
    }

    # Store in cache
    box set $cache_key $result
    return $result
}

println "Computing Fibonacci numbers with caching:"
for i in range(0..15) {
    f = fib_cached($i)
    printf "  fib(%2d) = %d\n" $i $f
}
println ""
println "Cache entries created:"
box
println ""

# ── Export / import ───────────────────────────────────────────────────
println "=== export / import ==="
export_path = "/tmp/katsh_box_demo.json"
box export $export_path
println "Exported to: $export_path"
file_size = $(wc -c < $export_path | tr -d ' ')
println "File size: $file_size bytes"
println ""

# Clear and reimport
box clear
println "After clear:"
box
println ""

box import $export_path
println "After import:"
box
println ""

bash! rm -f "$export_path"
println "Cleaned up export file"
println ""

# ── Box as a session "database" ───────────────────────────────────────
println "=== box as session database ==="
# Simulate storing multiple user records
for user in ["alice:engineer:95000", "bob:designer:72000", "carol:manager:110000"] {
    parts = $user | split ":"
    name  = $parts[0]
    role  = $parts[1]
    sal   = $parts[2]
    box set "user:$name:role"   $role
    box set "user:$name:salary" $sal
}

println "All users from box:"
for name in ["alice", "bob", "carol"] {
    role = $(box get "user:$name:role")
    sal  = $(box get "user:$name:salary")
    println "  $name: $role, \$$sal"
}
println ""

# Query: find engineers
println "Engineers:"
for name in ["alice", "bob", "carol"] {
    role = $(box get "user:$name:role")
    println "  $name" when $role == "engineer"
}
println ""

# ── Remove specific keys ──────────────────────────────────────────────
println "=== cleanup ==="
box rm greeting
box rm counter
println "Removed 'greeting' and 'counter'"
println "Remaining keys:"
box