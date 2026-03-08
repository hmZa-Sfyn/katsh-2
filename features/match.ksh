#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  08_match_switch.ksh — match/case and switch
#
#  Topics:
#    match with string values · numeric comparisons · wildcard *
#    multiple values per case (|) · switch · fallthrough
#    match as expression · match in functions
# ─────────────────────────────────────────────────────────────────────

# ── Basic match — string values ───────────────────────────────────────
println "=== string match ==="
for day in ["Monday", "Wednesday", "Saturday", "Sunday", "Thursday"] {
    match $day {
        "Monday":                println "  $day: Start of work week"
        "Friday":                println "  $day: TGIF!"
        "Saturday" | "Sunday":  println "  $day: Weekend 🎉"
        default:                 println "  $day: Midweek"
    }
}
println ""

# ── match with numeric comparisons ────────────────────────────────────
println "=== numeric match ==="
func letter_grade(score) {
    match $score {
        >=90: return "A"
        >=80: return "B"
        >=70: return "C"
        >=60: return "D"
        *:    return "F"
    }
}

for s in [100, 95, 85, 74, 63, 55, 0] {
    g = letter_grade($s)
    println "  $s → $g"
}
println ""

# ── match with wildcard * (catch-all, like else) ──────────────────────
println "=== HTTP status codes ==="
func describe_status(code) {
    match $code {
        200: return "OK"
        201: return "Created"
        204: return "No Content"
        301: return "Moved Permanently"
        302: return "Found"
        400: return "Bad Request"
        401: return "Unauthorized"
        403: return "Forbidden"
        404: return "Not Found"
        500: return "Internal Server Error"
        502: return "Bad Gateway"
        503: return "Service Unavailable"
        *:   return "Unknown status"
    }
}

for code in [200, 201, 404, 500, 302, 418] {
    msg = describe_status($code)
    println "  $code: $msg"
}
println ""

# ── match with multiple values per arm ────────────────────────────────
println "=== file extension classifier ==="
func classify_file(ext) {
    match $ext {
        "go" | "rs" | "c" | "cpp" | "java":
            return "compiled language source"
        "py" | "rb" | "js" | "ts" | "ksh":
            return "scripting language source"
        "jpg" | "jpeg" | "png" | "gif" | "webp":
            return "image"
        "mp3" | "flac" | "wav" | "ogg":
            return "audio"
        "mp4" | "mkv" | "avi" | "mov":
            return "video"
        "json" | "yaml" | "toml" | "xml":
            return "config/data"
        "md" | "txt" | "rst":
            return "documentation"
        default:
            return "unknown"
    }
}

files = ["main.go", "script.py", "photo.jpg", "song.mp3",
         "video.mkv", "config.json", "README.md", "binary.exe"]
for f in $files {
    ext   = $f | split "." | last
    kind  = classify_file($ext)
    println "  $f → $kind"
}
println ""

# ── switch — like match but always runs first matching case ───────────
println "=== switch ==="
for season in ["spring", "summer", "autumn", "winter", "monsoon"] {
    switch $season {
        "spring":  println "  $season: warm and blooming 🌸"
        "summer":  println "  $season: hot and sunny ☀️"
        "autumn":  println "  $season: cool and colorful 🍂"
        "winter":  println "  $season: cold and snowy ❄️"
        default:   println "  $season: not a standard season"
    }
}
println ""

# ── switch with fallthrough ───────────────────────────────────────────
println "=== fallthrough ==="
for level in ["error", "warn", "info", "debug"] {
    println "  Level '$level' activates:"
    switch $level {
        "debug":   println "    - debug output"; fallthrough
        "info":    println "    - info output"; fallthrough
        "warn":    println "    - warn output"; fallthrough
        "error":   println "    - error output"
    }
}
println ""

# ── match as a value in an expression ─────────────────────────────────
println "=== match as expression ==="
langs = ["go", "python", "javascript", "rust", "katsh", "cobol"]
for lang in $langs {
    verdict = match $lang {
        "go" | "rust":     "blazing fast"
        "python":          "batteries included"
        "javascript":      "runs everywhere"
        "katsh":           "structured and elegant"
        *:                 "classic"
    }
    println "  $lang: $verdict"
}