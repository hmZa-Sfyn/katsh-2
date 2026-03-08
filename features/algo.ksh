#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  18_algorithms.ksh — Classic algorithms implemented in KatSH
#
#  Topics:
#    bubble sort · insertion sort · quicksort
#    binary search · linear search
#    GCD / LCM · prime sieve
#    palindrome · anagram
#    Levenshtein distance (edit distance)
# ─────────────────────────────────────────────────────────────────────

# ── Bubble sort ───────────────────────────────────────────────────────
println "=== bubble sort ==="
func bubble_sort(arr) {
    n = $arr | arr_len
    for i in range(0..$n) {
        for j in range(0..($n - $i - 2)) {
            a = tonum $arr[$j]
            b = tonum $arr[$j + 1]
            if $a > $b {
                # swap
                tmp       = $arr[$j]
                arr[$j]   = $arr[$j + 1]
                arr[$j+1] = $tmp
            }
        }
    }
    return $arr
}

unsorted = [64, 34, 25, 12, 22, 11, 90, 45, 7]
println "  unsorted: $unsorted"
sorted = bubble_sort($unsorted)
println "  sorted:   $sorted"
println ""

# ── Insertion sort ───────────────────────────────────────────────────
println "=== insertion sort ==="
func insertion_sort(arr) {
    n = $arr | arr_len
    for i in range(1..$n) {
        key = tonum $arr[$i]
        j = $i - 1
        while $j >= 0 and tonum($arr[$j]) > $key {
            arr[$j + 1] = $arr[$j]
            j--
        }
        arr[$j + 1] = $key
    }
    return $arr
}

data = [5, 2, 8, 1, 9, 3, 7, 4, 6]
println "  unsorted: $data"
sorted2 = insertion_sort($data)
println "  sorted:   $sorted2"
println ""

# ── Binary search ─────────────────────────────────────────────────────
println "=== binary search ==="
func binary_search(arr, target) {
    lo = 0
    hi = ($arr | arr_len) - 1
    while $lo <= $hi {
        mid = ($lo + $hi) / 2
        val = tonum $arr[$mid]
        if $val == $target: return $mid
        if $val < $target:  lo = $mid + 1
        else:               hi = $mid - 1
    }
    return -1
}

sorted3 = [2, 5, 8, 12, 16, 23, 38, 42, 55, 72, 91]
for target in [23, 1, 91, 42, 100, 2] {
    idx = binary_search($sorted3, $target)
    if $idx >= 0 {
        println "  $target found at index $idx"
    } else {
        println "  $target not found"
    }
}
println ""

# ── GCD and LCM ──────────────────────────────────────────────────────
println "=== GCD and LCM ==="
func gcd(a, b) {
    while $b != 0 {
        tmp = $b
        b   = $a % $b
        a   = $tmp
    }
    return $a
}

func lcm(a, b) {
    g = gcd($a, $b)
    return $a * $b / $g
}

pairs = [["12","8"], ["48","18"], ["100","75"], ["17","13"]]
for pair in $pairs {
    a = tonum $pair[0]
    b = tonum $pair[1]
    g = gcd($a, $b)
    l = lcm($a, $b)
    println "  gcd($a, $b) = $g    lcm($a, $b) = $l"
}
println ""

# ── Sieve of Eratosthenes ─────────────────────────────────────────────
println "=== prime sieve (up to 60) ==="
func sieve(limit) {
    # Use a map as a boolean array
    is_prime = map {}
    for i in range(2..$limit) {
        map_set is_prime $i true
    }
    p = 2
    while $p * $p <= $limit {
        if $(map_get is_prime $p) == "true" {
            multiple = $p * $p
            while $multiple <= $limit {
                map_set is_prime $multiple false
                multiple = $multiple + $p
            }
        }
        p++
    }
    primes_arr = []
    for i in range(2..$limit) {
        if $(map_get is_prime $i) == "true" {
            primes_arr[] = $i
        }
    }
    return $primes_arr
}

primes_list = sieve(60)
println "  $primes_list"
println "  Count: $(echo $primes_list | arr_len)"
println ""

# ── Palindrome check ──────────────────────────────────────────────────
println "=== palindrome check ==="
func is_palindrome(s) {
    cleaned = $s | lower | replace " " ""
    return $cleaned == ($cleaned | reverse)
}

words2 = ["racecar", "hello", "level", "world", "kayak",
          "noon", "katsh", "madam", "Never Odd Or Even"]
for w in $words2 {
    p = is_palindrome($w)
    mark = if $p: "✓"; else: "✗"
    println "  $mark '$w'"
}
println ""

# ── Anagram check ─────────────────────────────────────────────────────
println "=== anagram check ==="
func normalize(s) {
    return $s | lower | replace " " "" | chars | arr_sort | arr_join ""
}

func is_anagram(a, b) {
    return $(normalize $a) == $(normalize $b)
}

pairs2 = [
    ["listen",  "silent"],
    ["hello",   "world"],
    ["triangle","integral"],
    ["cat",     "dog"],
    ["dusty",   "study"],
]
for pair in $pairs2 {
    a = $pair[0]
    b = $pair[1]
    p = is_anagram($a, $b)
    mark = if $p: "✓ anagram"; else: "✗ not anagram"
    println "  '$a' ↔ '$b': $mark"
}
println ""

# ── Levenshtein edit distance ─────────────────────────────────────────
println "=== edit distance (Levenshtein) ==="
func edit_distance(s1, s2) {
    c1 = $s1 | chars
    c2 = $s2 | chars
    m = $c1 | arr_len
    n = $c2 | arr_len

    # Build DP table using a map (row * (n+1) + col as key)
    dp = map {}
    for i in range(0..$m) {
        map_set dp "$i:0" $i
    }
    for j in range(0..$n) {
        map_set dp "0:$j" $j
    }

    for i in range(1..$m) {
        for j in range(1..$n) {
            ch1 = $c1[$i - 1]
            ch2 = $c2[$j - 1]
            if $ch1 == $ch2 {
                val = map_get dp "$(($i-1)):$(($j-1))"
            } else {
                v1 = tonum $(map_get dp "$(($i-1)):$j")
                v2 = tonum $(map_get dp "$i:$(($j-1))")
                v3 = tonum $(map_get dp "$(($i-1)):$(($j-1))")
                min1 = if $v1 < $v2: $v1; else: $v2
                val  = if $min1 < $v3: $min1 + 1; else: $v3 + 1
            }
            map_set dp "$i:$j" $val
        }
    }
    return $(map_get dp "$m:$n")
}

pairs3 = [
    ["kitten",  "sitting"],
    ["saturday","sunday"],
    ["abc",     "abc"],
    ["",        "hello"],
    ["katsh",   "bash"],
]
for pair in $pairs3 {
    a = $pair[0]
    b = $pair[1]
    d = edit_distance($a, $b)
    println "  '$a' → '$b': distance = $d"
}