#!/usr/bin/env katsh
# =============================================================================
# 06_all_features.ksh - Complete Script Using ALL Features
# =============================================================================
# This script demonstrates EVERY feature of Katsh scripting in one file.
# It serves as a comprehensive reference and practical example.
#
# FEATURES DEMONSTRATED:
# ✓ Variables & Types
# ✓ String Operations
# ✓ Array Operations
# ✓ Arithmetic
# ✓ If/Elif/Else
# ✓ Match/Case
# ✓ For Loops
# ✓ While Loops
# ✓ Repeat Loops
# ✓ Functions
# ✓ Try/Catch/Finally
# ✓ Pipe Operators
# ✓ Box Storage
# ✓ Import (conceptual)
# ✓ Subshell Capture
# ✓ String Interpolation
#
# RUN THIS SCRIPT: ./katsh examples/06_all_features.ksh
# =============================================================================

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                          HEADER                                          ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║     Katsh Complete Script - All Features Demo                     ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
echo ""

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  1. VARIABLES - Basic types, assignment, arithmetic                     ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo "══════════════════════════════════════════════════════════════════════"
echo "  1. VARIABLES & TYPES"
echo "══════════════════════════════════════════════════════════════════════"

# String variables
name = "Alice"
age = 25
height = 5.8
is_student = true
nickname = 'None'

echo "Name: $name"
echo "Age: $age"
echo "Height: $height"
echo "Student: $is_student"

# Arithmetic
x = 10
x += 5        # 15
x *= 2        # 30
x -= 10       # 20
echo "Calculated x: $x"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  2. STRING INTERPOLATION & OPERATIONS                                   ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  2. STRING OPERATIONS"
echo "══════════════════════════════════════════════════════════════════════"

# Interpolation
greeting = "Hello, $name!"
echo "Greeting: $greeting"

# Advanced interpolation
default_name = "${name:-Guest}"
echo "Default: $default_name"

# String length
echo "Name length: ${#name}"

# String transformations
text = "  Hello World  "
echo "Original: '$text'"
echo "Upper: '$(echo $text | upper)'"
echo "Lower: '$(echo "HELLO" | lower)'"
echo "Trimmed: '$(echo $text | trim)'"

# Split and join
data = "apple,banana,cherry" | split ","
echo "Split: $data"
echo "Join: $(echo $data | join " - ")"

# Repeat
line = "─"x50
echo $line

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  3. ARRAY OPERATIONS                                                     ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  3. ARRAYS"
echo "══════════════════════════════════════════════════════════════════════"

# Create array
fruits = ["apple", "banana", "cherry", "date", "elderberry"]
echo "Fruits: $fruits"

# Access elements
echo "First: ${fruits[0]}"
echo "Last: ${fruits[-1]}"
echo "Length: $fruits.len"

# Array operations
echo "Sorted: $(echo $fruits | arr_sort)"
echo "Reversed: $(echo $fruits | arr_reverse)"
echo "First item: $(echo $fruits | first)"
echo "Last item: $(echo $fruits | last)"

# Map and filter (conceptual)
nums = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
echo "Numbers: $nums"
echo "Sum: $(echo $nums | arr_sum)"
echo "Min: $(echo $nums | arr_min)"
echo "Max: $(echo $nums | arr_max)"
echo "Avg: $(echo $nums | arr_avg)"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  4. CONTROL FLOW - If/Elif/Else                                         ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  4. IF / ELIF / ELSE"
echo "══════════════════════════════════════════════════════════════════════"

# Simple if
score = 85
if $score >= 90:
    grade = "A"
elif $score >= 80:
    grade = "B"
elif $score >= 70:
    grade = "C"
elif $score >= 60:
    grade = "D"
else:
    grade = "F"

echo "Score: $score => Grade: $grade"

# Block syntax
age = 25
if $age >= 18 {
    echo "Adult"
    if $age >= 65:
        echo "Senior citizen"
} else {
    echo "Minor"
}

# Ternary
status = if $age >= 18: "adult"; else: "minor"
echo "Status: $status"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  5. MATCH / CASE                                                        ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  5. MATCH / CASE"
echo "══════════════════════════════════════════════════════════════════════"

# Simple match
color = "green"
match $color {
    case "red":   echo "Stop!"
    case "green": echo "Go!"
    case "yellow": echo "Caution!"
    case *:       echo "Unknown"
}

# Multiple values
day = "Saturday"
match $day {
    case "Saturday"|"Sunday": echo "Weekend!"
    case *: echo "Weekday"
}

# Comparison operators
score = 88
match $score {
    case >= 90: echo "Excellent"
    case >= 80: echo "Good"
    case >= 70: echo "Average"
    default:    echo "Needs work"
}

# Pattern matching
filename = "report.pdf"
match $filename {
    case "*.txt": echo "Text file"
    case "*.pdf": echo "PDF document"
    case "*.ksh": echo "Katsh script"
    case "*.go":  echo "Go source"
    case *:       echo "Unknown type"
}

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  6. FOR LOOPS                                                           ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  6. FOR LOOPS"
echo "══════════════════════════════════════════════════════════════════════"

# Range loop
echo "Count 0-4:"
for i in range(0, 5):
    echo "  $i"

# Array iteration
echo "Names:"
for n in ["Alice", "Bob", "Carol"]:
    echo "  Hello, $n!"

# With break
echo "With break at 3:"
for i in range(0, 10) {
    if $i == 3:
        break
    echo "  $i"
}

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  7. WHILE LOOPS                                                         ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  7. WHILE LOOPS"
echo "══════════════════════════════════════════════════════════════════════"

count = 0
while $count < 3:
    echo "Count: $count"
    count++

# Block syntax
x = 0
while $x < 3 {
    echo "x = $x"
    x++
}

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  8. REPEAT LOOPS                                                        ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  8. REPEAT LOOPS"
echo "══════════════════════════════════════════════════════════════════════"

repeat 3:
    echo "Repeat!"

# With index
repeat 5 {
    echo "Index: $_i"
}

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  9. FUNCTIONS                                                           ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  9. FUNCTIONS"
echo "══════════════════════════════════════════════════════════════════════"

# Simple function
func say_hello(person) {
    echo "Hello, $person!"
}
say_hello "World"

# Function with return
func add(a, b) {
    return $a + $b
}
result = add 5 3
echo "5 + 3 = $result"

# Function with multiple operations
func calculate_area(radius) {
    # Area = π * r² (approximated)
    result = $(echo "scale=2; 3.14159 * $radius * $radius" | bc)
    return $result
}
area = calculate_area 5
echo "Area of radius 5: $area"

# Variadic function
func greet_all(greeting) {
    echo "$greeting, $_args!"
}
greet_all "Hello" "Alice" "Bob" "Carol"

# Recursive function (factorial)
func factorial(n) {
    if $n <= 1:
        return 1
    prev = $(factorial $(echo $n - 1 | bc))
    return $(echo "$n * $prev" | bc)
}
echo "Factorial of 5: $(factorial 5)"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  10. TRY / CATCH                                                        ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  10. TRY / CATCH"
echo "══════════════════════════════════════════════════════════════════════"

# Try-catch
try {
    echo "Trying safe operation..."
    x = 10 / 2
    echo "Result: $x"
} catch e {
    echo "Error caught: $e"
} finally {
    echo "Finally block executed"
}

# Simulated error handling
try {
    echo "Simulating error..."
    if true:
        throw "Test error!"
} catch error_msg {
    echo "Caught: $error_msg"
} finally {
    echo "Cleanup complete"
}

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  11. SUBSHELL CAPTURE                                                   ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  11. SUBSHELL CAPTURE"
echo "══════════════════════════════════════════════════════════════════════"

# Backtick syntax
current_date = `date +%Y-%m-%d`
echo "Date: $current_date"

# $() syntax
current_time = $(date +%H:%M:%S)
echo "Time: $current_time"

# In string
message = "Running on $(hostname) at $(date +%H:%M)"
echo "Message: $message"

# Command in variable
file_count = $(ls | count)
echo "Files in current dir: $file_count"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  12. PIPE OPERATORS (Conceptual)                                        ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  12. PIPE OPERATORS"
echo "══════════════════════════════════════════════════════════════════════"

echo "These operators transform command output:"
echo "  cmd | select col1,col2   - Keep columns"
echo "  cmd | where col=val      - Filter rows"
echo "  cmd | sort col desc      - Sort"
echo "  cmd | limit N            - First N rows"
echo "  cmd | grep text          - Search"
echo "  cmd | unique             - Deduplicate"
echo "  cmd | fmt json|csv       - Format output"

# Example (commented out as it requires actual data)
# ls -la | select name,size | where size>1000 | sort size desc | limit 5

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  13. BOOLEAN OPERATORS & TESTS                                          ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  13. BOOLEAN OPERATORS & TESTS"
echo "══════════════════════════════════════════════════════════════════════"

# Logical operators
a = true
b = false

if $a and not $b:
    echo "Logic: true AND NOT false = true"

if $a or $b:
    echo "Logic: true OR false = true"

# Short-circuit
echo "Testing && (and):"
-f README.md && echo "  README.md exists!"

echo "Testing || (or):"
-f nonexistent.txt || echo "  File doesn't exist!"

# Test flags
if -d examples: echo "examples is a directory"
if -f README.md: echo "README.md is a file"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  14. BOX STORAGE (Conceptual)                                           ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  14. BOX STORAGE"
echo "══════════════════════════════════════════════════════════════════════"

echo "Store command results with #= syntax:"
echo "  ls #=files"
echo "  box get files"
echo "  box"
echo "  box rm files"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  15. IMPORT / EXPORT (Conceptual)                                       ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  15. IMPORT / EXPORT"
echo "══════════════════════════════════════════════════════════════════════"

echo "Import scripts:"
echo "  import './utils.ksh'"
echo "  import 'https://example.com/lib.ksh'"
echo "  import 'user/repo/script.ksh'"

echo "Export to environment:"
echo "  export PATH=\$PATH:/new/path"
echo "  export VAR=value"
echo "  export func my_function"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  16. SPECIAL VARIABLES                                                   ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  16. SPECIAL VARIABLES"
echo "══════════════════════════════════════════════════════════════════════"

echo "Script info:"
echo "  \$0 = script name"
echo "  \$1, \$2 = positional args"
echo "  \$# = arg count"
echo "  \$_args = all args"
echo "  \$__script_dir = script directory"

echo "Runtime:"
echo "  \$_return = last function return"
echo "  \$_error = last error message"
echo "  \$_i = loop index in repeat"
echo "  \$_args = extra function args"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║  17. PRACTICAL EXAMPLE - Student Management System                     ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  17. PRACTICAL EXAMPLE: Student Records"
echo "══════════════════════════════════════════════════════════════════════"

# Define students as arrays
students = [
    {"name": "Alice", "score": 92},
    {"name": "Bob", "score": 78},
    {"name": "Carol", "score": 85},
    {"name": "David", "score": 67},
    {"name": "Eve", "score": 91}
]

# Process each student
echo "Student Grades:"
echo "--------------"

for student in $students {
    # This is a conceptual example - actual table processing uses pipes
    name = $student["name"]
    score = $student["score"]

    # Determine grade using match
    match $score {
        case >= 90: grade = "A (Excellent)"
        case >= 80: grade = "B (Good)"
        case >= 70: grade = "C (Average)"
        case >= 60: grade = "D (Pass)"
        default:    grade = "F (Fail)"
    }

    echo "  $name: $score => $grade"
}

# Calculate average (conceptual)
# In practice: ps aux | select user,cpu | where ...

echo ""
echo "Summary:"
echo "  Highest score: 92 (Alice)"
echo "  Lowest score: 67 (David)"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                          FOOTER                                         ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "══════════════════════════════════════════════════════════════════════"
echo "  ✓ ALL SCRIPING FEATURES DEMONSTRATED!"
echo "══════════════════════════════════════════════════════════════════════"
echo ""
echo "This script covered:"
echo "  • Variables & Data Types"
echo "  • Strings & String Operations"
echo "  • Arrays & Array Operations"
echo "  • Arithmetic Operations"
echo "  • If / Elif / Else"
echo "  • Match / Case"
echo "  • For Loops"
echo "  • While Loops"
echo "  • Repeat Loops"
echo "  • Functions"
echo "  • Try / Catch / Finally"
echo "  • Subshell Capture"
echo "  • Pipe Operators"
echo "  • Boolean Logic"
echo "  • Box Storage"
echo "  • Import / Export"
echo ""
echo "╔════════════════════════════════════════════════════════════════════╗"
echo "║                    HAPPY SCRIPTIING! 🎉                           ║"
echo "╚════════════════════════════════════════════════════════════════════╝"
