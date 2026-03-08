#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  25_arguments.ksh — Positional args, flag parsing, validation
#
#  Run: katsh 25_arguments.ksh [options...]
#
#  Topics covered:
#    $0 $1 $2 ... $# $_args · required vs optional args
#    flag parsing (--flag value, -f value, --switch)
#    usage/help output · default values
#    subcommand dispatch · variadic args
#
#  Usage examples:
#    katsh 25_arguments.ksh deploy production v1.2.3
#    katsh 25_arguments.ksh --help
#    katsh 25_arguments.ksh greet --name Alice --loud
# ─────────────────────────────────────────────────────────────────────

# ── Basic positional arguments ────────────────────────────────────────
println "=== positional argument variables ==="
println "  \$0 (script name):  $0"
println "  \$1 (first arg):    ${1:-<not given>}"
println "  \$2 (second arg):   ${2:-<not given>}"
println "  \$# (arg count):    $#"
println "  \$_args (all args): ${_args:-<none>}"
println ""

# ── Require a minimum number of arguments ─────────────────────────────
func require(n, usage) {
    if $# < $n {
        println "Usage: katsh $0 $usage"
        println "  Got $# argument(s), need at least $n"
        throw "ArgumentError"
    }
}

# ── Usage / help ──────────────────────────────────────────────────────
func print_usage() {
    println "Usage: katsh $0 <command> [options]"
    println ""
    println "Commands:"
    println "  greet  <name>                    Print a greeting"
    println "  add    <a> <b>                   Add two numbers"
    println "  deploy <env> <version> [--dry]   Deploy to environment"
    println "  upper  <word...>                 Uppercase all words"
    println ""
    println "Options:"
    println "  --help, -h     Show this help"
    println "  --loud         Use bold uppercase output"
    println "  --dry          Dry-run mode (no real actions)"
    println ""
    println "Examples:"
    println "  katsh $0 greet Alice"
    println "  katsh $0 add 3 7"
    println "  katsh $0 deploy production v1.2.3"
    println "  katsh $0 deploy staging v2.0.0 --dry"
}

# ── Simple flag parser ────────────────────────────────────────────────
# Parses $_args into a map of flags and a list of positional args
func parse_args() {
    flags    = map {}
    positional = []
    args_arr = $_args | words

    i = 0
    while $i < ($args_arr | arr_len) {
        arg = $args_arr[$i]
        if $arg | startswith "--" {
            key = $arg | sub 2    # strip --
            # peek at next token: if it doesn't start with - it's the value
            next_i = $i + 1
            next = $args_arr[$next_i]
            if $next != "" and not ($next | startswith "-") {
                map_set flags $key $next
                i += 2
            } else {
                map_set flags $key "true"
                i++
            }
        } elif $arg | startswith "-" {
            key = $arg | sub 1    # strip -
            next_i = $i + 1
            next = $args_arr[$next_i]
            if $next != "" and not ($next | startswith "-") {
                map_set flags $key $next
                i += 2
            } else {
                map_set flags $key "true"
                i++
            }
        } else {
            positional[] = $arg
            i++
        }
    }
    # Return as a tuple: (flags_map_key, positional_arr_key)
    box set "__flags__"    "$(map_keys flags | arr_join ',')"
    box set "__pos__"      "$(positional | arr_join ',')"
    return $flags
}

# ── Subcommand: greet ─────────────────────────────────────────────────
func cmd_greet(flags, args) {
    name = $args[0]
    name = ${name:-World}
    loud = map_get flags "loud"

    if $loud == "true" {
        println "  HELLO, $(echo $name | upper)!"
    } else {
        println "  Hello, $name!"
    }
}

# ── Subcommand: add ───────────────────────────────────────────────────
func cmd_add(flags, args) {
    a = $args[0]
    b = $args[1]
    if $a == "" or $b == "": throw "add requires two numbers"
    $a | isnum || throw "add: '$a' is not a number"
    $b | isnum || throw "add: '$b' is not a number"
    result = tonum($a) + tonum($b)
    println "  $a + $b = $result"
}

# ── Subcommand: deploy ────────────────────────────────────────────────
func cmd_deploy(flags, args) {
    env     = $args[0]
    version = $args[1]
    dry     = map_get flags "dry"

    $env     != "" || throw "deploy: env is required"
    $version != "" || throw "deploy: version is required"

    println "  Environment: $env"
    println "  Version:     $version"
    if $dry == "true" {
        println "  [DRY RUN] would deploy now — skipping"
    } else {
        println "  Deploying..."
        println "  Done ✓"
    }
}

# ── Subcommand: upper ─────────────────────────────────────────────────
func cmd_upper(flags, args) {
    if $args | arr_len == 0: throw "upper: provide at least one word"
    for word in $args {
        println "  $(echo $word | upper)"
    }
}

# ── Demo: simulate running with various argument sets ─────────────────
println "=== simulating argument dispatch ==="
println ""

func simulate(desc, cmd, args_arr, flags_map) {
    println "▶  $desc"
    try {
        match $cmd {
            "help":   print_usage
            "greet":  cmd_greet $flags_map $args_arr
            "add":    cmd_add   $flags_map $args_arr
            "deploy": cmd_deploy $flags_map $args_arr
            "upper":  cmd_upper  $flags_map $args_arr
            default:  println "  Unknown command: '$cmd' (try --help)"
        }
    } catch e {
        println "  Error: $e"
    }
    println ""
}

# Build flag maps for each demo call
f_empty  = map {}
f_loud   = map { "loud":"true" }
f_dry    = map { "dry":"true" }

simulate "greet World (no args)"  "greet"  []            $f_empty
simulate "greet Alice"             "greet"  ["Alice"]     $f_empty
simulate "greet Alice --loud"      "greet"  ["Alice"]     $f_loud
simulate "add 3 7"                 "add"    ["3","7"]     $f_empty
simulate "add abc 5"               "add"    ["abc","5"]   $f_empty
simulate "add (missing second)"    "add"    ["10"]        $f_empty
simulate "deploy production v1.0"  "deploy" ["production","v1.0.3"] $f_empty
simulate "deploy staging --dry"    "deploy" ["staging","v2.0.0-rc1"] $f_dry
simulate "deploy (missing args)"   "deploy" []            $f_empty
simulate "upper foo bar baz"       "upper"  ["foo","bar","baz"] $f_empty
simulate "unknown command"         "oops"   []            $f_empty
simulate "help"                    "help"   []            $f_empty

# ── Variadic: collect remaining args ─────────────────────────────────
println "=== variadic arg pattern ==="
func sum_all() {
    # $_args receives everything beyond named params
    total = 0
    for n in $($_args | words) {
        $n | isnum || continue
        total = $total + tonum($n)
    }
    return $total
}

println "  sum_all is called inside a function — $_args holds extra args"

func demo_sum(label) {
    result = sum_all()
    println "  $label → sum = $result"
}
# Simulate: in real usage these come from $1 $2 $3 ...
println "  (variadic sum is best called from command line with real \$_args)"