#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  14_passthrough.ksh — Shell passthrough: bash!, run, capture, $()
#
#  Topics:
#    bash!/zsh!/sh! · run · ! prefix
#    $() POSIX substitution · capture builtin
#    backtick subshell · combining with katsh features
#    auto-passthrough for interactive commands
# ─────────────────────────────────────────────────────────────────────

# ── $() POSIX command substitution ───────────────────────────────────
println "=== \$() command substitution ==="
whoami_out = $(whoami)
hostname_out = $(hostname)
println "User:     $whoami_out"
println "Hostname: $hostname_out"
println ""

# Inside strings
println "=== \$() inside strings ==="
msg = "Running as $(whoami) on $(hostname)"
println "$msg"
println ""

# Nested $()
kernel = $(uname -r)
println "Kernel: $kernel"
println ""

# ── capture builtin — stores stdout+stderr ────────────────────────────
println "=== capture builtin ==="
capture files_out ls /tmp
println "First line of /tmp: $(echo $files_out | lines | first)"
println ""

# capture into a variable and use it
capture date_str date "+%Y-%m-%d"
println "Today: $date_str"

capture uptime_str uptime
println "Uptime: $uptime_str"
println ""

# ── Backtick subshell (original katsh syntax) ─────────────────────────
println "=== backtick subshell ==="
home_dir = `echo $HOME`
shell_name = `basename $SHELL`
println "Home:  $home_dir"
println "Shell: $shell_name"
println ""

# ── bash! — passthrough with explicit bash ────────────────────────────
println "=== bash! passthrough ==="
# bash! runs the rest of the line as a bash command with real TTY
# Output goes straight to terminal (not captured)
bash! echo "Hello from bash!"
bash! echo "Bash version: $BASH_VERSION"
println ""

# ── run — passthrough using \$SHELL ──────────────────────────────────
println "=== run passthrough ==="
run echo "Hello from run (\$SHELL = $SHELL)"
run uname -s
println ""

# ── sh! — minimal POSIX shell ─────────────────────────────────────────
println "=== sh! passthrough ==="
sh! echo "Hello from /bin/sh"
println ""

# ── Pipes in passthrough commands ────────────────────────────────────
println "=== pipes inside passthrough ==="
bash! echo "one two three four five" | tr ' ' '\n' | sort | head -3
println ""
bash! seq 1 10 | awk '{sum+=$1} END{print "Sum 1-10:", sum}'
println ""

# ── Complex bash pipelines ────────────────────────────────────────────
println "=== complex bash pipeline ==="
bash! ls /usr/bin | sort | uniq | wc -l | xargs echo "Commands in /usr/bin:"
println ""

# ── Combining $() capture with katsh processing ───────────────────────
println "=== capture + katsh processing ==="
# Get process count, process with katsh math
proc_count = $(ps aux | wc -l | tr -d ' ')
println "Process count: $proc_count"
overhead_pct = $proc_count * 100 / 1000
println "Load estimate: ~$overhead_pct% of 1000"
println ""

# Get file sizes, sum them with katsh
capture size_output du -sh /usr/bin /usr/lib /etc 2>/dev/null
println "Disk usage:"
println "$size_output"
println ""

# ── Environment variables via bash ───────────────────────────────────
println "=== env vars from bash ==="
path_count = $(echo $PATH | tr ':' '\n' | wc -l | tr -d ' ')
println "PATH has $path_count entries"

first_path = $(echo $PATH | cut -d: -f1)
println "First PATH entry: $first_path"
println ""

# ── Git integration (if git is available) ────────────────────────────
println "=== git integration pattern ==="
func git_info() {
    # Safe: these return empty string if not in a git repo
    branch  = $(git branch --show-current 2>/dev/null)
    hash    = $(git rev-parse --short HEAD 2>/dev/null)
    changes = $(git status --porcelain 2>/dev/null | wc -l | tr -d ' ')

    if $branch == "" {
        println "  Not in a git repository"
        return
    }
    status_icon = if $changes > 0: "●"; else: "✓"
    println "  Branch: $branch"
    println "  Hash:   ${hash:-unknown}"
    println "  Status: $status_icon ($changes changed files)"
}
git_info
println ""

# ── System info collection pattern ───────────────────────────────────
println "=== collecting system info ==="
info = map {}
map_set info "os"      "$(uname -s)"
map_set info "arch"    "$(uname -m)"
map_set info "kernel"  "$(uname -r)"
map_set info "user"    "$(whoami)"
map_set info "shell"   "$(basename $SHELL)"
map_set info "home"    "$HOME"

println "System information:"
for key in $(map_keys info) {
    val = map_get info $key
    k_padded = $key | pad 10
    println "  $k_padded $val"
}