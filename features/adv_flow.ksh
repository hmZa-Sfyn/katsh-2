#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  15_advanced_flow.ksh — goto/label, enum, struct, defer, with
#
#  Topics:
#    goto/label inside functions · enum pattern · struct pattern
#    defer (LIFO cleanup) · with block scoping
#    readonly constants · switch + fallthrough
# ─────────────────────────────────────────────────────────────────────

# ════════════════════════════════════════════════════════════════════
#  GOTO / LABEL
# ════════════════════════════════════════════════════════════════════
println "=== goto / label ==="

# Simple counting loop with goto
func count_to(limit) {
    n = 0
    label top:
        n++
        println "  $n"
    goto top when $n < $limit
}
count_to 5
println ""

# goto for retry logic
func fetch_with_retry(max) {
    attempt = 0
    label try_again:
        attempt++
        success = attempt >= 3    # simulates success on 3rd attempt
        if $success {
            println "  Succeeded on attempt $attempt"
            return $attempt
        }
        println "  Attempt $attempt failed, retrying..."
    goto try_again when $attempt < $max
    println "  All $max attempts failed"
    return -1
}
fetch_with_retry 5
println ""

# goto for state machine
func simple_fsm() {
    state = "start"
    data  = ["a", "b", "STOP", "c", "d"]
    idx   = 0

    label check_state:
        item = $data[$idx]
        idx++

        match $state {
            "start": {
                println "  start → processing '$item'"
                state = "running"
            }
            "running": {
                if $item == "STOP" {
                    state = "done"
                } else {
                    println "  running → got '$item'"
                }
            }
        }
    goto check_state when $state != "done" and $idx < $data.len
    println "  FSM reached state: $state"
}
simple_fsm
println ""

# ════════════════════════════════════════════════════════════════════
#  ENUM PATTERN
# ════════════════════════════════════════════════════════════════════
println "=== enum pattern ==="

# KatSH enums are named constant groups
enum Direction { NORTH SOUTH EAST WEST }
enum Color     { RED GREEN BLUE YELLOW }
enum Status    { PENDING RUNNING DONE FAILED }

println "Direction.NORTH = $Direction.NORTH"
println "Color.BLUE      = $Color.BLUE"
println "Status.DONE     = $Status.DONE"
println ""

# Use enum values in match
func describe_dir(d) {
    match $d {
        $Direction.NORTH: return "going up"
        $Direction.SOUTH: return "going down"
        $Direction.EAST:  return "going right"
        $Direction.WEST:  return "going left"
        *:                return "unknown direction"
    }
}

for dir in [$Direction.NORTH, $Direction.EAST, $Direction.WEST] {
    println "  $dir → $(describe_dir $dir)"
}
println ""

# ════════════════════════════════════════════════════════════════════
#  STRUCT PATTERN
# ════════════════════════════════════════════════════════════════════
println "=== struct pattern ==="

struct Point { x y }
struct Rect  { x y width height }
struct Person { name age email role }

# Constructors return maps
p1    = Point(10, 20)
p2    = Point(30, 40)
rect  = Rect(5, 5, 100, 60)
alice = Person("Alice", 30, "alice@example.com", "engineer")

println "p1: x=$p1.x  y=$p1.y"
println "p2: x=$p2.x  y=$p2.y"
println "rect: ${rect.width}×${rect.height} at (${rect.x},${rect.y})"
println "alice: $alice.name, age $alice.age, role $alice.role"
println ""

# Methods via functions that take struct maps
func distance(a, b) {
    dx = $b.x - $a.x
    dy = $b.y - $a.y
    return $(($dx * $dx) + ($dy * $dy) | sqrt)
}

func rect_area(r) {
    return $r.width * $r.height
}

func rect_contains(r, p) {
    return $p.x >= $r.x and $p.x <= ($r.x + $r.width) \
       and $p.y >= $r.y and $p.y <= ($r.y + $r.height)
}

d = distance($p1, $p2)
println "distance(p1, p2) = $d"
println "rect area = $(rect_area $rect)"
println "rect contains p1: $(rect_contains $rect $p1)"
println "rect contains p2: $(rect_contains $rect $p2)"
println ""

# ════════════════════════════════════════════════════════════════════
#  DEFER
# ════════════════════════════════════════════════════════════════════
println "=== defer (LIFO cleanup) ==="

func open_files() {
    defer println "  [defer 4] all files closed"
    defer println "  [defer 3] flushed buffers"
    defer println "  [defer 2] closed file B"
    defer println "  [defer 1] closed file A"

    println "  opened file A"
    println "  opened file B"
    println "  doing work..."
}
open_files
println ""

# Defer with error handling
func risky_operation(fail) {
    defer println "  cleanup ran (fail=$fail)"
    if $fail: throw "operation failed"
    println "  operation succeeded"
}

println "risky(false):"
risky_operation false
println ""

println "risky(true):"
try {
    risky_operation true
} catch e {
    println "  caught: $e"
}
println ""

# ════════════════════════════════════════════════════════════════════
#  WITH — block scoping
# ════════════════════════════════════════════════════════════════════
println "=== with (block scoping) ==="

outer = "I am outer"
with temp = "I am temp" {
    println "  inside with: temp='$temp'"
    println "  inside with: outer='$outer'"
    with nested = "I am nested" {
        println "  nested: temp='$temp'  nested='$nested'"
    }
    # $nested is gone here
}
# $temp is gone here
println "  after with: outer='$outer'"
println "  after with: temp='${temp:-<gone>}'"
println ""

# with for temporary configuration
func process_with_config(env) {
    with debug = $env == "dev" {
        with log_prefix = if $debug: "[DEBUG]"; else: "[INFO]" {
            println "  $log_prefix Processing in $env mode"
            println "  $log_prefix Debug enabled: $debug"
        }
    }
}
process_with_config "dev"
process_with_config "prod"