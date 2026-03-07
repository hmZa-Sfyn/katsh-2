#!/usr/bin/env katsh
# =============================================================================
# 03_arrays.ksh - Array Operations
# =============================================================================
# This script teaches you how to work with arrays in Katsh.
# Arrays store multiple values in an ordered list.
#
# RUN THIS SCRIPT: ./katsh examples/03_arrays.ksh
# =============================================================================

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                         ARRAY BASICS                                     ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Create arrays using square brackets
fruits = ["apple", "banana", "cherry"]
echo "Fruits: $fruits"

# Empty array
empty = []
echo "Empty array: $empty"

# Array from command output (each line becomes an element)
# Note: In practice, you'd use backticks or $()
# files = `ls`

# Array from string split
colors = "red,green,blue" | split ","
echo "Colors from string: $colors"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    ACCESSING ARRAY ELEMENTS                              ║
# ╚══════════════════════════════════════════════════════════════════════════╝

numbers = [10, 20, 30, 40, 50]

echo ""
echo "Array: $numbers"
echo "First element (index 0): ${numbers[0]}"
echo "Second element (index 1): ${numbers[1]}"
echo "Last element (index -1): ${numbers[-1]}"
echo "Second to last (index -2): ${numbers[-2]}"

# Array length
echo "Array length: $(echo $numbers | arr_len)"

# Also using .len property
echo "Length via .len: $numbers.len"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    MODIFYING ARRAYS                                     ║
# ╚══════════════════════════════════════════════════════════════════════════╝

items = ["first", "second", "third"]

echo ""
echo "Original: $items"

# Append element
items[] = "fourth"
echo "After append: $items"

# Insert at specific index
items[0] = "new-first"
echo "After index assignment: $items"

# Remove last element (using slice)
items = $items | slice 0 -1
echo "After removing last: $items"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    ARRAY PIPE OPERATIONS                                ║
# ╚══════════════════════════════════════════════════════════════════════════╝

nums = [5, 2, 8, 1, 9, 3, 7, 4, 6]

echo ""
echo "Original array: $nums"

# Get first/last
echo "First: $(echo $nums | first)"
echo "Last: $(echo $nums | last)"
echo "Nth (index 2): $(echo $nums | nth 2)"

# Slice (start, end)
echo "Slice (1 to 4): $(echo $nums | slice 1 4)"

# Push (add to end)
echo "Push 10: $(echo $nums | push 10)"

# Pop (remove last)
echo "Pop: $(echo $nums | pop)"

# Sort
echo "Sorted: $(echo $nums | arr_sort)"

# Reverse
echo "Reversed: $(echo $nums | arr_reverse)"

# Unique (remove duplicates)
dupes = [1, 2, 2, 3, 3, 3, 4]
echo "With duplicates: $dupes"
echo "Unique: $(echo $dupes | arr_uniq)"

# Contains check
echo "Contains 5: $(echo $nums | arr_contains 5)"
echo "Contains 99: $(echo $nums | arr_contains 99)"

# Join array into string
echo "Join with '-': $(echo $nums | arr_join "-")"

# Map (apply operation to each element)
echo "Double each: $(echo $nums | arr_map 'mul 2')"

# Filter (keep matching elements)
echo "Even numbers: $(echo $nums | arr_filter 'isnum' | arr_filter 'mod 2')"

# Sum, Min, Max
echo "Sum: $(echo $nums | arr_sum)"
echo "Min: $(echo $nums | arr_min)"
echo "Max: $(echo $nums | arr_max)"
echo "Average: $(echo $nums | arr_avg)"

# Flatten nested arrays
nested = [[1, 2], [3, 4], [5]]
echo "Nested: $nested"
echo "Flattened: $(echo $nested | flatten)"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    ITERATING OVER ARRAYS                                 ║
# ╚══════════════════════════════════════════════════════════════════════════╝

names = ["Alice", "Bob", "Carol"]

echo ""
echo "--- For Loop Iteration ---"

# Simple for loop
for name in $names:
    echo "Hello, $name!"

# For loop with block
echo ""
for person in $names {
    if $person == "Bob": echo "Special: $person"
    else:echo "Hello: $person"
}

# Using range
echo ""
echo "--- Range-based For Loop ---"
for i in range(0, 5):
    echo "Index: $i"

# For loop over command output
echo ""
echo "--- Iterate command output ---"
# for line in `ls examples`:
#     echo "File: $line"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    ASSOCIATIVE ARRAYS (MAPS)                             ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Katsh supports key-value pairs via special notation
# Note: The syntax may vary, here's how you can simulate:

# Using string keys in certain contexts
echo ""
echo "--- Working with key-value like data ---"

# Create table-like data
data = [
    {"name": "Alice", "age": "25"},
    {"name": "Bob", "age": "30"}
]
echo "Data: $data"

# Access via pipe (when data is from command)
# ps aux | select user,pid,cmd | where user=root

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    PRACTICAL EXAMPLES                                   ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Example 1: Process files
echo ""
echo "--- Example: Process file extensions ---"
files = ["report.pdf", "data.csv", "image.png", "script.ksh", "video.mp4"]

# Extract extensions
for f in $files:
    ext = $f | sub $(echo $f | len | sub 4)
    echo "File: $f => Extension: $ext"

# Example 2: Filter and transform
echo ""
echo "--- Example: Filter numbers ---"
data = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]

# Note: Complex filtering syntax
echo "Even numbers:"
# In practice, you'd use where pipe operator

# Example 3: Build a menu
echo ""
echo "--- Example: Build a menu ---"
options = ["Option 1", "Option 2", "Option 3", "Quit"]

for i in range(0, $options.len):
    idx = $i | add 1
    echo "$idx. $options[$i]"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                         PRACTICE EXERCISES                               ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Try these:
# 1. Create an array of numbers and calculate the average
# 2. Find the largest number in an array
# 3. Remove all duplicates from an array
# 4. Reverse an array without using reverse
# 5. Create a for loop that prints the index alongside the value

echo ""
echo "=== Arrays Tutorial Complete! ==="
