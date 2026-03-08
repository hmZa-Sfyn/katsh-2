#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  12_collections.ksh — Sets, Stacks, Queues, Tuples
#
#  Topics:
#    set creation · membership · algebra (union/intersect/diff)
#    stack LIFO · push/pop/peek
#    queue FIFO · enqueue/dequeue/peek
#    tuple creation · indexed access
#    real-world patterns for each
# ─────────────────────────────────────────────────────────────────────

# ════════════════════════════════════════════════════════════════════
#  SETS
# ════════════════════════════════════════════════════════════════════
println "╔════════════════════════════╗"
println "║           SETS             ║"
println "╚════════════════════════════╝"

# ── Creating sets ─────────────────────────────────────────────────────
primes = set { 2 3 5 7 11 13 17 19 }
evens  = set { 2 4 6 8 10 12 14 16 18 20 }

println "primes: $(set_show primes)"
println ""

# ── Add / remove / membership ─────────────────────────────────────────
languages = set { "go" "python" "rust" "javascript" }
set_add    languages "katsh"
set_add    languages "go"       # duplicate — ignored
set_remove languages "javascript"

println "languages: $(set_len languages) items"
for lang in ["go", "katsh", "java", "rust"] {
    has = set_has languages $lang
    println "  has '$lang': $has"
}
println ""

# ── Set algebra ───────────────────────────────────────────────────────
a = set { 1 2 3 4 5 }
b = set { 3 4 5 6 7 }

set_union     a b u
set_intersect a b i
set_diff      a b d
set_diff      b a d2

println "a = {1,2,3,4,5}  b = {3,4,5,6,7}"
println "union:        $(set_len u) items"
println "intersect:    $(set_len i) items"
println "diff (a-b):   $(set_len d) items"
println "diff (b-a):   $(set_len d2) items"
println ""

# ── Real-world: deduplicating tags ────────────────────────────────────
println "=== deduplicating tags ==="
all_tags = set {}
posts = [
    "go scripting tools automation",
    "go performance concurrency",
    "scripting automation shell tools",
    "rust performance systems",
]
for post in $posts {
    for tag in $(echo $post | words) {
        set_add all_tags $tag
    }
}
println "All unique tags ($(set_len all_tags)):"
println "  $(set_show all_tags)"
println ""

# ════════════════════════════════════════════════════════════════════
#  STACKS  (LIFO)
# ════════════════════════════════════════════════════════════════════
println "╔════════════════════════════╗"
println "║          STACKS            ║"
println "╚════════════════════════════╝"

st = stack {}
for item in ["first", "second", "third", "fourth"] {
    stack_push st $item
    println "  push '$item' → depth=$(stack_len st)"
}
println ""
println "  peek: $(stack_peek st)"
println ""

println "  Popping:"
while $(stack_len st) > 0 {
    stack_pop st val
    println "  pop → '$val' (remaining: $(stack_len st))"
}
println ""

# ── Real-world: undo stack ────────────────────────────────────────────
println "=== undo stack ==="
undo_stack = stack {}
state      = "initial"

func apply_action(s, new_state) {
    stack_push $s $state   # save current state
    state = $new_state
    println "  applied → state='$state'  undo depth=$(stack_len $s)"
}

func undo(s) {
    if $(stack_len $s) == 0 {
        println "  nothing to undo"
        return
    }
    stack_pop $s prev
    state = $prev
    println "  undone  → state='$state'  undo depth=$(stack_len $s)"
}

apply_action undo_stack "step-1"
apply_action undo_stack "step-2"
apply_action undo_stack "step-3"
undo undo_stack
undo undo_stack
undo undo_stack
undo undo_stack   # nothing to undo
println ""

# ════════════════════════════════════════════════════════════════════
#  QUEUES  (FIFO)
# ════════════════════════════════════════════════════════════════════
println "╔════════════════════════════╗"
println "║          QUEUES            ║"
println "╚════════════════════════════╝"

q = queue {}
for job in ["job-1", "job-2", "job-3", "job-4"] {
    enqueue q $job
    println "  enqueue '$job' → size=$(queue_len q)"
}
println ""
println "  peek (next): $(queue_peek q)"
println ""

println "  Processing:"
while $(queue_len q) > 0 {
    dequeue q job
    println "  process '$job' (remaining: $(queue_len q))"
}
println ""

# ── Real-world: task scheduler ───────────────────────────────────────
println "=== task scheduler ==="
task_queue = queue {}

func schedule(q, task) {
    enqueue $q $task
    println "  scheduled: $task"
}

func process_next(q) {
    if $(queue_len $q) == 0: return
    dequeue $q task
    println "  processing: $task"
}

schedule task_queue "send-welcome-email"
schedule task_queue "generate-thumbnail"
schedule task_queue "update-search-index"
schedule task_queue "send-notification"
println ""

println "  Running tasks:"
while $(queue_len task_queue) > 0 {
    process_next task_queue
}
println ""

# ════════════════════════════════════════════════════════════════════
#  TUPLES
# ════════════════════════════════════════════════════════════════════
println "╔════════════════════════════╗"
println "║          TUPLES            ║"
println "╚════════════════════════════╝"

# Tuples are immutable ordered records
point   = (10, 20)
rgb     = (255, 128, 0)
person2 = ("Alice", 30, "engineer", true)

println "point:   $(tuple_show point)"
println "rgb:     $(tuple_show rgb)"
println "person2: $(tuple_show person2)"
println ""

# Access by index
x    = tuple_get point 0
y    = tuple_get point 1
name = tuple_get person2 0
age  = tuple_get person2 1

println "point.x = $x,  point.y = $y"
println "person2.name = $name,  person2.age = $age"
println ""

# ── Real-world: structured return values ─────────────────────────────
println "=== returning multiple values via tuple ==="
func min_max(arr) {
    mn = $arr | arr_min
    mx = $arr | arr_max
    return ($mn, $mx)
}

func stats(arr) {
    s = $arr | arr_sum
    n = $arr | arr_len
    a = $s / $n
    mn = $arr | arr_min
    mx = $arr | arr_max
    return ($s, $n, $a, $mn, $mx)
}

data = [42, 17, 85, 3, 60, 29, 78, 51]
result = stats($data)

total = tuple_get result 0
count = tuple_get result 1
avg   = tuple_get result 2
min_v = tuple_get result 3
max_v = tuple_get result 4

println "data: $data"
println "  total=$total  count=$count  avg=$avg  min=$min_v  max=$max_v"