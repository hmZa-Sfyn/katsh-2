#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  37_regex.ksh — Regex and text search patterns
#
#  Topics covered:
#    ~ (match) and !~ (not-match) operators
#    regex in where clauses (pipe filtering)
#    grep/sed/awk via passthrough for complex patterns
#    extracting groups via bash · URL/email/IP parsers
#    text tokenizer · templating with regex
# ─────────────────────────────────────────────────────────────────────

# ── Basic ~ and !~ operators ──────────────────────────────────────────
println "=== ~ regex match operator ==="
strings = [
    "hello world",
    "foo123bar",
    "alice@example.com",
    "192.168.1.42",
    "GET /api/users HTTP/1.1",
    "ERROR: connection refused",
    "2024-03-15",
    "just plain text",
]

println "Strings matching '[0-9]' (contain a digit):"
for s in $strings {
    if $s ~ "[0-9]": println "  ✓ '$s'"
}
println ""

println "Strings NOT matching '@' (no @-sign):"
for s in $strings {
    if $s !~ "@": println "  ✓ '$s'"
}
println ""

# ── Pattern classifier ────────────────────────────────────────────────
println "=== pattern classifier ==="
func classify_string(s) {
    if $s ~ "^[0-9]+$":                    return "integer"
    if $s ~ "^[0-9]+\.[0-9]+$":            return "float"
    if $s ~ "@" and $s ~ "\.":             return "email"
    if $s ~ "^https?://":                  return "url"
    if $s ~ "^[0-9]{1,3}\.[0-9]{1,3}":    return "ip-address"
    if $s ~ "^[0-9]{4}-[0-9]{2}-[0-9]{2}": return "date"
    if $s ~ "^#[0-9a-fA-F]{6}$":          return "hex-color"
    if $s ~ "^[A-Z_][A-Z0-9_]+$":         return "constant"
    return "text"
}

samples = [
    "42", "3.14", "alice@example.com", "https://katsh.io",
    "192.168.0.1", "2024-03-15", "#FF5733", "MAX_SIZE", "hello world",
]
for s in $samples {
    kind = classify_string($s)
    println "  $(tostr $s | pad 25) → $kind"
}
println ""

# ── URL parser ────────────────────────────────────────────────────────
println "=== URL parsing ==="
func parse_url(url) {
    parts = map {}
    # Protocol
    if $url ~ "^https://" {
        map_set parts "protocol" "https"
        rest = $url | sub 8
    } elif $url ~ "^http://" {
        map_set parts "protocol" "http"
        rest = $url | sub 7
    } else {
        map_set parts "protocol" "unknown"
        rest = $url
    }
    # Split host from path
    slash_pos = 0
    chars = $rest | chars
    for i in range(0..($chars | arr_len)) {
        if $chars[$i] == "/" {
            slash_pos = $i
            break
        }
    }
    if $slash_pos > 0 {
        host = $rest | sub 0 $slash_pos
        path = $rest | sub $slash_pos
    } else {
        host = $rest
        path = "/"
    }
    # Split host:port
    if $host ~ ":" {
        hp = $host | split ":"
        map_set parts "host" $hp[0]
        map_set parts "port" $hp[1]
    } else {
        map_set parts "host" $host
        map_set parts "port" $(if $(map_get parts "protocol") == "https": "443"; else: "80")
    }
    # Split path and query
    if $path ~ "\?" {
        pq = $path | split "?"
        map_set parts "path"  $pq[0]
        map_set parts "query" $pq[1]
    } else {
        map_set parts "path"  $path
        map_set parts "query" ""
    }
    return $parts
}

urls = [
    "https://api.example.com/v1/users?page=2&limit=10",
    "http://localhost:8080/health",
    "https://blog.katsh.io/posts/scripting-guide",
]
for url in $urls {
    p = parse_url($url)
    println "  URL: $url"
    println "    protocol=$(map_get $p protocol)  host=$(map_get $p host)  port=$(map_get $p port)"
    println "    path=$(map_get $p path)  query=$(map_get $p query)"
    println ""
}

# ── Email validator with parts ────────────────────────────────────────
println "=== email parsing ==="
func parse_email(e) {
    $e ~ "@" || throw "not an email: $e"
    parts = $e | split "@"
    local_part  = $parts[0]
    domain      = $parts[1]
    domain ~ "\." || throw "domain has no TLD: $domain"
    dp    = $domain | split "."
    tld   = $dp | last
    result = map {}
    map_set result "local"  $local_part
    map_set result "domain" $domain
    map_set result "tld"    $tld
    return $result
}

emails = ["alice@example.com", "bob.jones@company.co.uk",
          "carol+filter@test.io", "notanemail", "bad@nodot"]
for e in $emails {
    try {
        p = parse_email($e)
        println "  ✓ $e → local=$(map_get $p local)  domain=$(map_get $p domain)  tld=$(map_get $p tld)"
    } catch err {
        println "  ✗ $e → $err"
    }
}
println ""

# ── Text tokenizer ────────────────────────────────────────────────────
println "=== simple tokenizer ==="
func tokenize(source) {
    tokens = []
    words  = $source | words
    for w in $words {
        kind = match {
            $w ~ "^[0-9]+$":                 "NUMBER"
            $w ~ "^\".*\"$" or $w ~ "^'.*'$": "STRING"
            $w ~ "^[+\-*/=<>!]+$":            "OPERATOR"
            $w ~ "^[a-z][a-zA-Z0-9_]*$":      "IDENT"
            $w ~ "^[A-Z][A-Z0-9_]*$":          "CONST"
            *:                                 "UNKNOWN"
        }
        tokens[] = "$kind($w)"
    }
    return $tokens
}

source_snippet = "x = foo + 42 BAR_CONST / \"hello\" <= 100"
println "  Source: '$source_snippet'"
println "  Tokens:"
for tok in $(tokenize $source_snippet) {
    println "    $tok"
}
println ""

# ── Grep via bash! for complex patterns ──────────────────────────────
println "=== grep via bash! ==="
sample_log = "2024-01-15 INFO  server started
2024-01-15 ERROR failed to connect: timeout
2024-01-15 WARN  memory at 89%
2024-01-15 ERROR database query failed
2024-01-15 INFO  request processed in 45ms
2024-01-15 DEBUG cache miss key=user:42"

# Write to temp file
tmp = $(mktemp)
bash! printf '%s\n' "$sample_log" > "$tmp"

println "Error lines:"
bash! grep "ERROR" "$tmp"
println ""

println "Lines with numbers:"
bash! grep -E "[0-9]+" "$tmp"
println ""

println "Lines NOT containing INFO or DEBUG:"
bash! grep -vE "INFO|DEBUG" "$tmp"
println ""

bash! rm -f "$tmp"

# ── sed for text transformation ───────────────────────────────────────
println "=== sed transforms ==="
text = "The quick brown Fox jumps over the Lazy Dog"
tmp2 = $(mktemp)
bash! echo "$text" > "$tmp2"

bash! sed 's/[A-Z]/(&)/g' "$tmp2"         # wrap capitals in parens
bash! sed 's/\b\w\{4\}\b/[&]/g' "$tmp2"  # bracket 4-letter words
println ""

bash! rm -f "$tmp2"