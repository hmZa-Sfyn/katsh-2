#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  36_config.ksh — Layered configuration management
#
#  Topics covered:
#    config priority: defaults → file → env vars → CLI args
#    INI/env file parser · schema validation
#    required vs optional keys · type coercion
#    config accessor with dot notation
#    masked output for secrets · environment profiles
# ─────────────────────────────────────────────────────────────────────

# ══════════════════════════════════════════════════════════════════════
#  LAYER 1: Hard-coded defaults
# ══════════════════════════════════════════════════════════════════════
func make_defaults() {
    d = map {}
    # App
    map_set d "app.name"     "myapp"
    map_set d "app.version"  "0.0.0"
    map_set d "app.host"     "0.0.0.0"
    map_set d "app.port"     "8080"
    map_set d "app.env"      "development"
    map_set d "app.debug"    "false"
    map_set d "app.log_level" "info"
    # Database
    map_set d "db.host"     "localhost"
    map_set d "db.port"     "5432"
    map_set d "db.name"     "app_dev"
    map_set d "db.user"     "postgres"
    map_set d "db.password" ""
    map_set d "db.pool_min" "2"
    map_set d "db.pool_max" "10"
    map_set d "db.ssl"      "false"
    # Cache
    map_set d "cache.host" "localhost"
    map_set d "cache.port" "6379"
    map_set d "cache.ttl"  "3600"
    # Security
    map_set d "security.jwt_secret"    ""
    map_set d "security.allowed_hosts" "*"
    map_set d "security.cors_origin"   "*"
    return $d
}

# ══════════════════════════════════════════════════════════════════════
#  LAYER 2: Parse .env / INI file content
# ══════════════════════════════════════════════════════════════════════
func parse_env_file(content) {
    result = map {}
    for line in $(echo $content | lines) {
        line = $line | trim
        # Skip blank lines and comments
        continue when $line == ""
        continue when $line | startswith "#"
        continue when $line | startswith ";"
        # Handle KEY=VALUE
        if $line ~ "=" {
            eqpos = 0
            chars = $line | chars
            for i in range(0..($chars | arr_len)) {
                if $chars[$i] == "=" {
                    eqpos = $i
                    break
                }
            }
            key = $line | sub 0 $eqpos | trim | lower | replace "_" "."
            val = $line | sub ($eqpos + 1) | trim
            # Strip surrounding quotes
            if ($val | startswith '"') and ($val | endswith '"') {
                val = $val | sub 1 ($val | len - 2)
            }
            map_set result $key $val
        }
    }
    return $result
}

# ══════════════════════════════════════════════════════════════════════
#  LAYER 3: Read from environment variables (simulated)
# ══════════════════════════════════════════════════════════════════════
func read_env_vars(prefix) {
    result = map {}
    # Simulate env vars with APP_ prefix
    env_vars = map {
        "APP_PORT"           = "9090"
        "APP_ENV"            = "production"
        "APP_DEBUG"          = "false"
        "APP_LOG_LEVEL"      = "warn"
        "DB_HOST"            = "db.production.internal"
        "DB_NAME"            = "app_prod"
        "DB_PASSWORD"        = "s3cr3t!"
        "DB_SSL"             = "true"
        "SECURITY_JWT_SECRET" = "jwt-super-secret-key-do-not-share"
    }
    for env_key in $(map_keys env_vars) {
        if $env_key | startswith $prefix {
            val     = map_get env_vars $env_key
            cfg_key = $env_key | sub ($prefix | len) | lower | replace "_" "."
            map_set result $cfg_key $val
        }
    }
    return $result
}

# ══════════════════════════════════════════════════════════════════════
#  MERGE: Apply layers in priority order
# ══════════════════════════════════════════════════════════════════════
func merge_config(base, override_map) {
    for key in $(map_keys override_map) {
        val = map_get override_map $key
        map_set base $key $val
    }
    return $base
}

# ══════════════════════════════════════════════════════════════════════
#  SCHEMA VALIDATION
# ══════════════════════════════════════════════════════════════════════
func validate_config(cfg) {
    errors = []

    # Required keys
    required = ["app.host", "app.port", "db.host", "db.name", "db.user"]
    for key in $required {
        val = map_get cfg $key
        if $val == "": errors[] = "Required key '$key' is missing"
    }

    # Numeric keys
    numerics = ["app.port", "db.port", "db.pool_min", "db.pool_max", "cache.port", "cache.ttl"]
    for key in $numerics {
        val = map_get cfg $key
        if $val != "" and not ($val | isnum) {
            errors[] = "Key '$key' must be numeric, got '$val'"
        }
    }

    # Port ranges
    for key in ["app.port", "db.port", "cache.port"] {
        val = map_get cfg $key
        if $val | isnum {
            n = tonum $val
            if $n < 1 or $n > 65535 {
                errors[] = "Key '$key' port $n out of range (1-65535)"
            }
        }
    }

    # Boolean keys
    booleans = ["app.debug", "db.ssl"]
    for key in $booleans {
        val = map_get cfg $key
        if $val != "" and $val != "true" and $val != "false" {
            errors[] = "Key '$key' must be true/false, got '$val'"
        }
    }

    # Production requires secrets
    env = map_get cfg "app.env"
    if $env == "production" {
        jwt = map_get cfg "security.jwt_secret"
        if $jwt == "": errors[] = "security.jwt_secret required in production"
        db_pass = map_get cfg "db.password"
        if $db_pass == "": errors[] = "db.password required in production"
    }

    return $errors
}

# ══════════════════════════════════════════════════════════════════════
#  ACCESSOR with masking for secrets
# ══════════════════════════════════════════════════════════════════════
_secret_keys = ["db.password", "security.jwt_secret"]

func cfg_get(cfg, key) {
    return map_get cfg $key
}

func cfg_display(cfg, key) {
    val = map_get cfg $key
    is_secret = $_secret_keys | arr_contains $key
    if $is_secret and $val != "" {
        return "***" | repeat ($val | len)
    }
    return $val
}

# ══════════════════════════════════════════════════════════════════════
#  BUILD CONFIG
# ══════════════════════════════════════════════════════════════════════
println "╔══════════════════════════════════════╗"
println "║     Configuration Management          ║"
println "╚══════════════════════════════════════╝"
println ""

# Simulate a .env file
env_file_content = "
# App settings
APP_NAME=myapp
APP_VERSION=2.1.4

# Override port for staging
APP_PORT=8443
"

println "=== Layer 1: Defaults ==="
cfg = make_defaults()
println "  Loaded $(map_len cfg) default keys"
println ""

println "=== Layer 2: .env file ==="
file_cfg = parse_env_file($env_file_content)
file_keys = map_keys file_cfg
println "  Parsed $(echo $file_keys | arr_len) keys from .env:"
for k in $file_keys { println "  ↳ $k = $(map_get file_cfg $k)" }
cfg = merge_config($cfg, $file_cfg)
println ""

println "=== Layer 3: Environment variables (APP_ prefix) ==="
env_cfg  = read_env_vars "APP_"
db_cfg   = read_env_vars "DB_"
sec_cfg  = read_env_vars "SECURITY_"
map_merge env_cfg db_cfg  combined_env
map_merge combined_env sec_cfg env_final
env_keys = map_keys env_final
println "  Read $(echo $env_keys | arr_len) env var overrides"
cfg = merge_config($cfg, $env_final)
println ""

println "=== Validation ==="
validation_errors = validate_config($cfg)
if $(echo $validation_errors | arr_len) > 0 {
    println "  ❌ Validation failed:"
    for e in $validation_errors { println "  → $e" }
} else {
    println "  ✅ All validation checks passed"
}
println ""

println "=== Final Configuration ==="
sections = ["app", "db", "cache", "security"]
for section in $sections {
    println "  [$section]"
    for key in $(map_keys cfg) {
        if $key | startswith "$section." {
            display_val = cfg_display($cfg, $key)
            short_key   = $key | sub ($section | len + 1)
            println "    $(short_key | pad 18) = $display_val"
        }
    }
}
println ""

println "=== Profile: production ==="
println "  app.env    = $(cfg_get $cfg 'app.env')"
println "  app.port   = $(cfg_get $cfg 'app.port')"
println "  app.debug  = $(cfg_get $cfg 'app.debug')"
println "  db.host    = $(cfg_get $cfg 'db.host')"
println "  db.ssl     = $(cfg_get $cfg 'db.ssl')"
println "  db.password (masked) = $(cfg_display $cfg 'db.password')"
println "  jwt_secret (masked)  = $(cfg_display $cfg 'security.jwt_secret')"