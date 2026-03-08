#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  11_maps.ksh — Maps (hash tables)
#
#  Topics:
#    creation · get/set/del · has · keys/values
#    iteration · merge · nested maps · word frequency
#    building a simple in-memory "database"
# ─────────────────────────────────────────────────────────────────────

# ── Creating maps ─────────────────────────────────────────────────────
println "=== creating maps ==="
person = map { "name":"Alice" "age":30 "city":"NYC" "active":true }
println "person: $(map_show person)"
println ""

# = syntax also works
config = map { host=localhost port=5432 dbname=myapp }
println "config keys: $(map_keys config)"
println ""

# ── get/set/del ───────────────────────────────────────────────────────
println "=== get / set / del ==="
println "name: $(map_get person name)"
println "city: $(map_get person city)"

# Bracket syntax also works in expressions
println "age via \$person[\"age\"]: $person["age"]"

map_set person "email" "alice@example.com"
map_set person "score" 99
println "After set — keys: $(map_keys person)"

map_del person "active"
println "After del 'active' — keys: $(map_keys person)"
println ""

# ── has ───────────────────────────────────────────────────────────────
println "=== map_has ==="
for key in ["name", "email", "phone", "city", "score"] {
    has = map_has person $key
    println "  has '$key': $has"
}
println ""

# ── keys / values ─────────────────────────────────────────────────────
println "=== keys and values ==="
keys = map_keys person
vals = map_values person
println "keys:   $keys"
println "values: $vals"
println "len:    $(map_len person)"
println ""

# ── Iterating over a map ──────────────────────────────────────────────
println "=== iterating ==="
for key in $(map_keys person) {
    val = map_get person $key
    println "  $key = $val"
}
println ""

# ── merge ─────────────────────────────────────────────────────────────
println "=== merge ==="
base    = map { "timeout":30 "retries":3 "debug":false }
overrides = map { "timeout":60 "debug":true "log_level":"verbose" }
map_merge base overrides merged
println "base:      $(map_keys base)"
println "overrides: $(map_keys overrides)"
println "merged:"
for k in $(map_keys merged) {
    println "  $k = $(map_get merged $k)"
}
println ""

# ── Word frequency counter ────────────────────────────────────────────
println "=== word frequency ==="
text = "the cat sat on the mat the cat sat"
words = $text | words
freq  = map {}

for word in $words {
    current = map_get freq $word
    if $current == "" {
        map_set freq $word 1
    } else {
        map_set freq $word ($current + 1)
    }
}

println "Word frequencies:"
for word in $(map_keys freq) {
    count = map_get freq $word
    bar   = "#" | repeat $count
    println "  $word: $count $bar"
}
println ""

# ── Nested maps (simulated — maps of string-encoded values) ───────────
println "=== simple in-memory record store ==="
# We'll use map + key conventions to simulate records

db = map {}
id = 1

func db_insert(m, name, age, role) {
    map_set m "user_$id:name" $name
    map_set m "user_$id:age"  $age
    map_set m "user_$id:role" $role
}

db_insert db "Alice"   30 "engineer"  ; id++
db_insert db "Bob"     25 "designer"  ; id++
db_insert db "Carol"   35 "engineer"  ; id++
db_insert db "Dave"    28 "manager"   ; id++

println "All records:"
for i in range(1..4) {
    name = map_get db "user_$i:name"
    age  = map_get db "user_$i:age"
    role = map_get db "user_$i:role"
    println "  [$i] $name, age=$age, role=$role"
}
println ""

println "Engineers only:"
for i in range(1..4) {
    role = map_get db "user_$i:role"
    if $role == "engineer" {
        name = map_get db "user_$i:name"
        println "  $name"
    }
}
println ""

# ── Environment-variable style config map ────────────────────────────
println "=== config map with defaults ==="
env_config = map {
    "APP_HOST"    = "0.0.0.0"
    "APP_PORT"    = "8080"
    "APP_ENV"     = "development"
    "APP_DEBUG"   = "true"
    "DB_HOST"     = "localhost"
    "DB_PORT"     = "5432"
    "DB_NAME"     = "app"
    "DB_SSL"      = "false"
}

func getenv(m, key, default_val) {
    val = map_get m $key
    if $val == "": return $default_val
    return $val
}

host   = getenv env_config "APP_HOST"  "127.0.0.1"
port   = getenv env_config "APP_PORT"  "3000"
ssl    = getenv env_config "DB_SSL"    "true"
region = getenv env_config "AWS_REGION" "us-east-1"

println "  APP_HOST:   $host"
println "  APP_PORT:   $port"
println "  DB_SSL:     $ssl"
println "  AWS_REGION: $region (default)"