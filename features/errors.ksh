#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  09_errors.ksh — Error handling: try/catch/finally, throw/raise
#
#  Topics:
#    try/catch · finally · throw · raise (alias) · $_error
#    re-throw · nested try · error type conventions
#    graceful degradation · retry pattern
# ─────────────────────────────────────────────────────────────────────

# ── Basic try/catch ───────────────────────────────────────────────────
println "=== basic try/catch ==="
try {
    throw "something went wrong"
    println "this line never runs"
} catch e {
    println "Caught: $e"
}
println ""

# ── finally always runs ───────────────────────────────────────────────
println "=== try/catch/finally ==="
func open_resource(name) {
    println "  Opening resource: $name"
    if $name == "bad": throw "resource not found: $name"
    println "  Using resource: $name"
}

for resource in ["good", "bad", "another"] {
    try {
        open_resource $resource
    } catch e {
        println "  Error: $e"
    } finally {
        println "  Cleanup done for: $resource"
    }
    println ""
}

# ── raise is an alias for throw ───────────────────────────────────────
println "=== raise (alias for throw) ==="
try {
    raise "raised error message"
} catch e {
    println "Caught raise: $e"
}
println ""

# ── Safe division function ────────────────────────────────────────────
println "=== safe division ==="
func safe_div(a, b) {
    if $b == 0: throw "DivisionByZero: cannot divide $a by zero"
    return $a / $b
}

for pair in [["10", "2"], ["7", "0"], ["100", "4"], ["9", "0"]] {
    a = $pair[0]
    b = $pair[1]
    try {
        result = safe_div($a, $b)
        println "  $a / $b = $result"
    } catch e {
        println "  $a / $b → Error: $e"
    }
}
println ""

# ── Structured error messages (convention) ───────────────────────────
println "=== structured error messages ==="
func parse_int(s) {
    unless $s | isnum: throw "TypeError: '$s' is not a number"
    return tonum $s
}

func validate_age(s) {
    n = parse_int($s)
    if $n < 0:   throw "ValueError: age cannot be negative"
    if $n > 150: throw "ValueError: age $n is unrealistically large"
    return $n
}

for input in ["25", "abc", "-5", "200", "42"] {
    try {
        age = validate_age($input)
        println "  '$input' → valid age: $age"
    } catch e {
        println "  '$input' → $e"
    }
}
println ""

# ── Nested try/catch ──────────────────────────────────────────────────
println "=== nested try ==="
func outer() {
    try {
        println "  outer: trying inner"
        try {
            throw "inner error"
        } catch inner_e {
            println "  inner caught: $inner_e"
            throw "outer error (caused by: $inner_e)"
        }
    } catch outer_e {
        println "  outer caught: $outer_e"
    }
}
outer
println ""

# ── $_error — last error message ─────────────────────────────────────
println "=== \$_error ==="
try {
    throw "test error for _error demo"
} catch e {
    println "Caught in catch: $e"
}
println "Last error stored in _error: $_error"
println ""

# ── Retry pattern ────────────────────────────────────────────────────
println "=== retry pattern ==="
attempt = 0
max_attempts = 3
success = false

func flaky_operation(attempt_num) {
    if $attempt_num < 3: throw "transient failure on attempt $attempt_num"
    return "success after $attempt_num attempts"
}

while $attempt < $max_attempts and not $success {
    attempt++
    try {
        result = flaky_operation($attempt)
        success = true
        println "  Succeeded: $result"
    } catch e {
        println "  Attempt $attempt failed: $e"
        if $attempt < $max_attempts {
            println "  Retrying..."
        }
    }
}

unless $success {
    println "  All $max_attempts attempts failed"
}
println ""

# ── Graceful degradation ──────────────────────────────────────────────
println "=== graceful degradation ==="
func get_config(key) {
    # Simulate: some keys exist, some don't
    configs = map { "host":"localhost" "port":"5432" "name":"mydb" }
    has = map_has configs $key
    unless $has: throw "ConfigError: key '$key' not found"
    return map_get configs $key
}

func get_config_or_default(key, default_val) {
    try {
        return get_config($key)
    } catch e {
        return $default_val
    }
}

host    = get_config_or_default "host" "127.0.0.1"
port    = get_config_or_default "port" "3306"
name    = get_config_or_default "name" "default"
timeout = get_config_or_default "timeout" "30"

println "  host=$host  port=$port  name=$name  timeout=$timeout"