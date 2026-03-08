#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  03_arrays.ksh — Arrays and array operations
#
#  Topics:
#    creation · access · mutation · slice · search
#    sort · filter · map · reduce (sum/avg/min/max)
#    iteration · arr_join · flatten
# ─────────────────────────────────────────────────────────────────────

# ── Creating arrays ───────────────────────────────────────────────────
fruits  = ["apple", "banana", "cherry", "date", "elderberry"]
nums    = [5, 3, 8, 1, 9, 2, 7, 4, 6]
mixed   = ["hello", 42, true, "world"]
empty   = []

println "fruits: $fruits"
println "nums:   $nums"
println "mixed:  $mixed"
println ""

# ── Access by index ───────────────────────────────────────────────────
println "fruits[0]:  $fruits[0]"    # first
println "fruits[2]:  $fruits[2]"    # third
println "fruits[-1]: $fruits[-1]"   # last (negative index)
println "fruits[-2]: $fruits[-2]"   # second-to-last
println "length:     $fruits.len"
println ""

# ── Mutation ──────────────────────────────────────────────────────────
fruits[0] = "APPLE"
println "After fruits[0]='APPLE': $fruits"

fruits[] = "fig"                    # append
println "After append 'fig': $fruits"
println ""

# ── Pipe operations ───────────────────────────────────────────────────
println "first:    $(echo $fruits | first)"
println "last:     $(echo $fruits | last)"
println "nth 2:    $(echo $fruits | nth 2)"
println "arr_len:  $(echo $fruits | arr_len)"
println ""

# ── Sorting ───────────────────────────────────────────────────────────
println "nums unsorted: $nums"
println "arr_sort:      $(echo $nums | arr_sort)"
println "arr_reverse:   $(echo $nums | arr_reverse)"
println "sorted desc:"
sorted = $nums | arr_sort | arr_reverse
println "  $sorted"
println ""

# ── Deduplication ────────────────────────────────────────────────────
dups = [3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5]
println "with dups:    $dups"
println "arr_uniq:     $(echo $dups | arr_uniq)"
println ""

# ── Search ────────────────────────────────────────────────────────────
println "contains 'banana':  $(echo $fruits | arr_contains 'banana')"
println "contains 'mango':   $(echo $fruits | arr_contains 'mango')"
println ""

# ── Slice ─────────────────────────────────────────────────────────────
println "slice 1 3: $(echo $fruits | slice 1 3)"  # indices 1,2,3
println ""

# ── push / pop ────────────────────────────────────────────────────────
stack = ["a", "b", "c"]
println "push 'd': $(echo $stack | push 'd')"
println "pop:      $(echo $stack | pop)"
println "original: $stack"   # unchanged — pipe returns new value
println ""

# ── Map: apply operation to every element ────────────────────────────
words = ["hello", "world", "katsh"]
println "arr_map upper:  $(echo $words | arr_map upper)"
println "arr_map reverse: $(echo $words | arr_map reverse)"
println ""

# ── Filter: keep elements matching predicate ──────────────────────────
values = ["42", "hello", "99", "world", "7", "foo"]
println "original:       $values"
println "arr_filter isnum:   $(echo $values | arr_filter isnum)"
println "arr_filter isalpha: $(echo $values | arr_filter isalpha)"
println ""

# ── Numeric reductions ────────────────────────────────────────────────
scores = [85, 92, 78, 95, 88, 73, 90]
println "scores: $scores"
println "sum:    $(echo $scores | arr_sum)"
println "avg:    $(echo $scores | arr_avg)"
println "min:    $(echo $scores | arr_min)"
println "max:    $(echo $scores | arr_max)"
println ""

# ── Join ─────────────────────────────────────────────────────────────
tags = ["go", "shell", "cli", "tools"]
println "arr_join ', ': $(echo $tags | arr_join ', ')"
println "arr_join ' | ': $(echo $tags | arr_join ' | ')"
println ""

# ── flatten ───────────────────────────────────────────────────────────
# Arrays can be created from lines of command output
files = `echo "a.go\nb.go\nc.go"`
println "from command output: $files"
println ""

# ── Iterating ─────────────────────────────────────────────────────────
println "Iterating fruits:"
for fruit in $fruits {
    n = $fruit | len
    println "  $fruit ($n chars)"
}
println ""

println "Iterating with index:"
i = 0
for score in $scores {
    i++
    println "  [$i] $score"
}