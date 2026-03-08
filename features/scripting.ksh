#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  16_scripting_patterns.ksh — Real scripting patterns
#
#  Topics:
#    argument parsing ($1/$# validation)
#    config file loading
#    retry with exponential backoff
#    progress indicators
#    color output helpers
#    dry-run mode
#    logging (levels, timestamps)
# ─────────────────────────────────────────────────────────────────────

# ── ANSI color helpers ────────────────────────────────────────────────
func red(s)    { return "\033[31m$s\033[0m" }
func green(s)  { return "\033[32m$s\033[0m" }
func yellow(s) { return "\033[33m$s\033[0m" }
func blue(s)   { return "\033[34m$s\033[0m" }
func bold(s)   { return "\033[1m$s\033[0m" }
func grey(s)   { return "\033[90m$s\033[0m" }

# ── Logging with levels and timestamps ───────────────────────────────
func ts() {
    return $(date "+%H:%M:%S")
}

func log_info(msg) {
    t = ts()
    println "$(grey $t) $(green "[INFO]") $msg"
}

func log_warn(msg) {
    t = ts()
    println "$(grey $t) $(yellow "[WARN]") $msg"
}

func log_error(msg) {
    t = ts()
    println "$(grey $t) $(red "[ERROR]") $msg"
}

func log_debug(msg) {
    unless $DEBUG: return
    t = ts()
    println "$(grey $t) $(blue "[DEBUG]") $msg"
}

DEBUG = true
log_info  "Server starting..."
log_warn  "Memory usage at 78%"
log_error "Connection refused: db:5432"
log_debug "Cache miss for key 'user:42'"
DEBUG = false
log_debug "This won't print (DEBUG=false)"
println ""

# ── Argument validation pattern ───────────────────────────────────────
println "=== argument validation ==="
func require_args(min_count, usage_str) {
    if $# < $min_count {
        println "$(red 'Error:')"
        println "  $(bold Usage): $usage_str"
        println "  Got $# argument(s), need at least $min_count"
        throw "ArgumentError: insufficient arguments"
    }
}

func check_not_empty(val, name) {
    if $val == "": throw "ArgumentError: '$name' cannot be empty"
}

# Simulate calling with args
func deploy(env, version) {
    try {
        require_args 2 "deploy <env> <version>"
        check_not_empty $env     "env"
        check_not_empty $version "version"
    } catch e {
        log_error "Validation failed: $e"
        return 1
    }

    log_info "Deploying version $(bold $version) to $(bold $env)"
    return 0
}

deploy "production" "v1.2.3"
try { deploy "" "v1.0.0" } catch e { log_error "$e" }
try { deploy "staging" "" } catch e { log_error "$e" }
println ""

# ── Config file loading pattern ───────────────────────────────────────
println "=== config loading ==="
func load_config(config_map, content) {
    for line in $(echo $content | lines) {
        line = $line | trim
        # Skip empty lines and comments
        continue when $line == ""
        continue when $line | startswith "#"
        # Parse KEY=VALUE
        if $line ~ "=" {
            parts = $line | split "="
            key   = $parts[0] | trim
            val   = $parts[1] | trim
            map_set $config_map $key $val
        }
    }
}

# Simulate a config file content
config_content = "
# App configuration
APP_HOST=0.0.0.0
APP_PORT=8080
APP_ENV=production

# Database
DB_HOST=db.internal
DB_PORT=5432
DB_NAME=myapp
DB_POOL=10

# Feature flags
FEATURE_DARK_MODE=true
FEATURE_BETA=false
"

cfg = map {}
load_config cfg $config_content

println "Loaded config:"
for key in $(map_keys cfg) {
    val = map_get cfg $key
    println "  $key = $val"
}
println ""

# ── Retry with exponential backoff ───────────────────────────────────
println "=== retry with backoff ==="
func with_retry(max_attempts, fn_name) {
    attempt  = 0
    delay    = 1

    while $attempt < $max_attempts {
        attempt++
        log_debug "Attempt $attempt / $max_attempts"

        try {
            # call the function by name
            result = $fn_name()
            log_info "Succeeded on attempt $attempt"
            return $result
        } catch e {
            if $attempt >= $max_attempts {
                log_error "All $max_attempts attempts failed: $e"
                throw $e
            }
            log_warn "Attempt $attempt failed: $e — retrying in ${delay}s"
            # sleep $delay   # commented out for fast demo
            delay = $delay * 2    # exponential backoff
        }
    }
}

attempt_counter = 0
func unreliable() {
    attempt_counter++
    if $attempt_counter < 3: throw "service unavailable"
    return "data loaded successfully"
}

DEBUG = true
try {
    r = with_retry 5 unreliable
    log_info "Result: $r"
} catch e {
    log_error "Gave up: $e"
}
DEBUG = false
println ""

# ── Dry-run mode ──────────────────────────────────────────────────────
println "=== dry-run mode ==="
DRY_RUN = true

func run_cmd(cmd) {
    if $DRY_RUN {
        println "  $(grey '[dry-run]') would run: $(yellow $cmd)"
        return 0
    }
    # Actually run: run $cmd
    log_info "Running: $cmd"
}

run_cmd "git add -A"
run_cmd "git commit -m 'release v2.0'"
run_cmd "git push origin main"
run_cmd "kubectl apply -f deploy.yaml"

DRY_RUN = false
println ""

# ── Progress bar ─────────────────────────────────────────────────────
println "=== progress indicator ==="
func progress_bar(current, total, width) {
    pct   = $current * 100 / $total
    filled = $current * $width / $total
    empty  = $width - $filled
    bar    = "█" | repeat $filled
    space  = "░" | repeat $empty
    printf "\r  [${bar}${space}] ${pct}%%"
}

total_items = 20
println "Processing $total_items items..."
for i in range(1..20) {
    progress_bar $i $total_items 30
    # sleep 0.05   # commented out for fast demo
}
println ""    # newline after progress bar
log_info "All $total_items items processed"
println ""

# ── Pipeline with validation ──────────────────────────────────────────
println "=== validated data pipeline ==="
func validate_email(e) {
    unless $e ~ "@": throw "InvalidEmail: '$e' missing @"
    unless $e ~ "\\.": throw "InvalidEmail: '$e' missing domain"
    return $e
}

func validate_age_num(a) {
    unless $a | isnum: throw "InvalidAge: '$a' is not a number"
    n = tonum $a
    if $n < 0 or $n > 150: throw "InvalidAge: $n out of range"
    return $n
}

raw_users = [
    "alice:alice@example.com:28",
    "bob:notanemail:25",
    "carol:carol@test.org:abc",
    "dave:dave@company.io:35",
]

valid = []
errors2 = []

for row in $raw_users {
    parts = $row | split ":"
    name  = $parts[0]
    email = $parts[1]
    age   = $parts[2]

    try {
        validate_email $email
        validate_age_num $age
        valid[] = "$name ($email, age $age)"
        log_info "Valid user: $name"
    } catch e {
        errors2[] = "$name: $e"
        log_warn "Skipped: $name — $e"
    }
}

println ""
println "Valid users ($(echo $valid | arr_len)):"
for u in $valid { println "  ✓ $u" }
println "Errors ($(echo $errors2 | arr_len)):"
for e in $errors2 { println "  ✗ $e" }