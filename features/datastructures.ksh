#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  31_data_structures.ksh — Linked list and graph with maps
#
#  Topics covered:
#    singly linked list (push/pop/insert/delete/traverse)
#    doubly linked list pattern
#    adjacency-list graph (add node/edge, BFS, DFS)
#    cycle detection · shortest path (BFS)
#    all using map as the backing store
# ─────────────────────────────────────────────────────────────────────

# ══════════════════════════════════════════════════════════════════════
#  SINGLY LINKED LIST
#  Storage: map with keys "node_N:val", "node_N:next", "head", "size"
# ══════════════════════════════════════════════════════════════════════
println "╔════════════════════════════════╗"
println "║       SINGLY LINKED LIST       ║"
println "╚════════════════════════════════╝"

func ll_new() {
    lst = map {}
    map_set lst "head" "null"
    map_set lst "size" 0
    map_set lst "_id"  0
    return $lst
}

func ll_push_front(lst, val) {
    id = tonum $(map_get lst "_id")
    id++
    map_set lst "_id" $id

    old_head = map_get lst "head"
    map_set lst "node_${id}:val"  $val
    map_set lst "node_${id}:next" $old_head
    map_set lst "head"            "node_$id"
    map_set lst "size"            (tonum $(map_get lst "size") + 1)
}

func ll_push_back(lst, val) {
    id = tonum $(map_get lst "_id")
    id++
    map_set lst "_id" $id
    map_set lst "node_${id}:val"  $val
    map_set lst "node_${id}:next" "null"
    map_set lst "size"            (tonum $(map_get lst "size") + 1)

    # Find tail
    cur = map_get lst "head"
    if $cur == "null" {
        map_set lst "head" "node_$id"
        return
    }
    while true {
        nxt = map_get lst "${cur}:next"
        if $nxt == "null": break
        cur = $nxt
    }
    map_set lst "${cur}:next" "node_$id"
}

func ll_to_array(lst) {
    result = []
    cur    = map_get lst "head"
    while $cur != "null" {
        val = map_get lst "${cur}:val"
        result[] = $val
        cur = map_get lst "${cur}:next"
    }
    return $result
}

func ll_pop_front(lst) {
    head = map_get lst "head"
    head == "null" && throw "ll_pop_front: list is empty"
    val  = map_get lst "${head}:val"
    nxt  = map_get lst "${head}:next"
    map_set lst "head" $nxt
    map_set lst "size" (tonum $(map_get lst "size") - 1)
    return $val
}

func ll_contains(lst, target) {
    cur = map_get lst "head"
    while $cur != "null" {
        val = map_get lst "${cur}:val"
        if $val == $target: return true
        cur = map_get lst "${cur}:next"
    }
    return false
}

# Demo
lst = ll_new()
for v in [10, 20, 30] { ll_push_back lst $v }
ll_push_front lst 5
ll_push_back  lst 40

println "List: $(ll_to_array $lst)"
println "Size: $(map_get $lst size)"
println "Contains 20: $(ll_contains $lst 20)"
println "Contains 99: $(ll_contains $lst 99)"

popped = ll_pop_front($lst)
println "Pop front: $popped"
println "After pop: $(ll_to_array $lst)"
println ""

# ══════════════════════════════════════════════════════════════════════
#  ADJACENCY-LIST GRAPH
#  Storage: map with keys "nodes", "edges:NODE" (comma-sep list)
# ══════════════════════════════════════════════════════════════════════
println "╔════════════════════════════════╗"
println "║            GRAPH               ║"
println "╚════════════════════════════════╝"

func graph_new() {
    g = map {}
    map_set g "nodes" ""
    return $g
}

func graph_add_node(g, node) {
    nodes = map_get g "nodes"
    if $nodes == "" {
        map_set g "nodes" $node
    } else {
        # Avoid duplicates
        $nodes ~ $node || map_set g "nodes" "$nodes,$node"
    }
    has_edges = map_has g "edges:$node"
    unless $has_edges: map_set g "edges:$node" ""
}

func graph_add_edge(g, from, to) {
    graph_add_node g $from
    graph_add_node g $to
    edges = map_get g "edges:$from"
    if $edges == "" {
        map_set g "edges:$from" $to
    } else {
        map_set g "edges:$from" "$edges,$to"
    }
}

func graph_neighbors(g, node) {
    raw = map_get g "edges:$node"
    if $raw == "": return []
    return $raw | split ","
}

func graph_nodes(g) {
    raw = map_get g "nodes"
    if $raw == "": return []
    return $raw | split ","
}

# ── BFS (breadth-first search) ────────────────────────────────────────
func bfs(g, start) {
    visited = set {}
    order   = []
    q       = queue {}

    enqueue q $start
    set_add visited $start

    while $(queue_len q) > 0 {
        dequeue q node
        order[] = $node
        for neighbor in $(graph_neighbors $g $node) {
            unless $(set_has visited $neighbor) {
                set_add visited $neighbor
                enqueue q $neighbor
            }
        }
    }
    return $order
}

# ── DFS (depth-first search, iterative) ──────────────────────────────
func dfs(g, start) {
    visited = set {}
    order   = []
    stk     = stack {}

    stack_push stk $start

    while $(stack_len stk) > 0 {
        stack_pop stk node
        unless $(set_has visited $node) {
            set_add visited $node
            order[] = $node
            for neighbor in $(graph_neighbors $g $node) {
                stack_push stk $neighbor
            }
        }
    }
    return $order
}

# ── Shortest path via BFS ─────────────────────────────────────────────
func shortest_path(g, start, end) {
    if $start == $end: return [$start]
    visited = set {}
    prev    = map {}
    q       = queue {}

    enqueue q $start
    set_add visited $start
    found = false

    while $(queue_len q) > 0 and not $found {
        dequeue q node
        for neighbor in $(graph_neighbors $g $node) {
            unless $(set_has visited $neighbor) {
                set_add visited $neighbor
                map_set prev $neighbor $node
                enqueue q $neighbor
                if $neighbor == $end {
                    found = true
                    break
                }
            }
        }
    }

    unless $found: return []

    # Reconstruct path
    path = []
    cur  = $end
    while $cur != "" {
        path[] = $cur
        cur    = map_get prev $cur
    }
    return $path | arr_reverse
}

# Build example graph
#   A - B - C
#   |   |   |
#   D - E - F
#       |
#       G

g = graph_new()
graph_add_edge g "A" "B"
graph_add_edge g "A" "D"
graph_add_edge g "B" "C"
graph_add_edge g "B" "E"
graph_add_edge g "C" "F"
graph_add_edge g "D" "E"
graph_add_edge g "E" "F"
graph_add_edge g "E" "G"

println "Graph nodes: $(graph_nodes $g)"
println ""

println "BFS from A:  $(bfs $g 'A')"
println "DFS from A:  $(dfs $g 'A')"
println ""

println "Shortest paths:"
for pair in [["A","G"], ["A","F"], ["D","C"], ["G","A"], ["B","G"]] {
    from = $pair[0]
    to   = $pair[1]
    path = shortest_path($g, $from, $to)
    if $(echo $path | arr_len) > 0 {
        joined = $path | arr_join " → "
        println "  $from → $to : $joined  ($(echo $path | arr_len) steps)"
    } else {
        println "  $from → $to : no path"
    }
}
println ""

# ── Cycle detection (DFS coloring) ────────────────────────────────────
println "── Cycle Detection ──────────────────"
func has_cycle_util(g, node, visited, rec_stack) {
    map_set visited   $node "true"
    map_set rec_stack $node "true"

    for neighbor in $(graph_neighbors $g $node) {
        in_vis   = map_get visited   $neighbor
        in_stack = map_get rec_stack $neighbor

        if $in_vis != "true" {
            if $(has_cycle_util $g $neighbor $visited $rec_stack) == "true" {
                return true
            }
        } elif $in_stack == "true" {
            return true
        }
    }
    map_set rec_stack $node "false"
    return false
}

func has_cycle(g) {
    visited    = map {}
    rec_stack  = map {}
    for node in $(graph_nodes $g) {
        if $(map_get visited $node) != "true" {
            if $(has_cycle_util $g $node $visited $rec_stack) == "true" {
                return true
            }
        }
    }
    return false
}

# Undirected graph above — no directed cycle
println "Undirected graph has cycle: $(has_cycle $g)"

# Add a directed cycle: G → B
cyclic_g = graph_new()
graph_add_edge cyclic_g "A" "B"
graph_add_edge cyclic_g "B" "C"
graph_add_edge cyclic_g "C" "A"    # cycle!
println "Directed cycle A→B→C→A has cycle: $(has_cycle $cyclic_g)"