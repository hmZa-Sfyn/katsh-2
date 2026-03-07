#!/usr/bin/env katsh
# =============================================================================
# 04_control_flow.ksh - Control Flow
# =============================================================================
# This script teaches you about control flow in Katsh:
# - If / Elif / Else
# - Match / Case
# - For loops
# - While / Do-While
# - Repeat
# - Try / Catch
#
# RUN THIS SCRIPT: ./katsh examples/04_control_flow.ksh
# =============================================================================

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                       IF / ELIF / ELSE                                   ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo "=== IF / ELIF / ELSE ==="

# Simple if
x = 10
if $x > 5:
    echo "x is greater than 5"

# If-else
age = 17
if $age >= 18:
    echo "Adult"
else:
    echo "Minor"

# If-elif-else
score = 75
if $score >= 90:
    echo "Grade: A"
elif $score >= 80:
    echo "Grade: B"
elif $score >= 70:
    echo "Grade: C"
elif $score >= 60:
    echo "Grade: D"
else:
    echo "Grade: F"

# Block syntax (with curly braces)
score = 85
if $score >= 90 {
    echo "A - Excellent!"
} elif $score >= 80 {
    echo "B - Good job!"
} else {
    echo "Keep trying!"
}

# Unless (execute if condition is FALSE)
name = ""
unless $name == "":
    echo "Hello, $name!"
echo "(unless skipped because name is empty)"

# Inline ternary (already covered, but here it is again)
status = if $age >= 18: "adult"; else: "minor"
echo "Status: $status"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                       MATCH / CASE                                       ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== MATCH / CASE ==="

# Simple match
color = "red"
match $color {
    case "red":   echo "Stop!"
    case "green": echo "Go!"
    case "yellow": echo "Caution!"
    case *:       echo "Unknown signal"
}

# Multiple values in one case
color = "orange"
match $color {
    case "red"|"orange": echo "Caution - warm color"
    case "green"|"blue": echo "Cool color"
    case *:              echo "Other color"
}

# Comparison operators in cases
score = 85
match $score {
    case >= 90: echo "Excellent!"
    case >= 80: echo "Good"
    case >= 70: echo "Average"
    case >= 60: echo "Below average"
    default:    echo "Needs improvement"
}

# Glob/pattern matching
filename = "report_2024.pdf"
match $filename {
    case "*.txt": echo "Text file"
    case "*.pdf": echo "PDF document"
    case "*.go":  echo "Go source code"
    case "*.ksh": echo "Katsh script"
    case "test_*": echo "Test file"
    case *:      echo "Unknown type"
}

# Match with variable binding
day = "Monday"
match $day {
    case "Saturday"|"Sunday": weekend = true; echo "Weekend!"
    case *: weekend = false; echo "Weekday"
}
echo "Is weekend: $weekend"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                       FOR LOOPS                                         ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== FOR LOOPS ==="

# Simple range loop
echo "Counting 0-4:"
for i in range(0, 5):
    echo "  $i"

# Range with .. operator
echo "Counting 1-5:"
for i in range(1..5):
    echo "  $i"

# Loop through array
echo "Names:"
for name in ["Alice", "Bob", "Carol"]:
    echo "  Hello, $name!"

# Loop through command output
# Note: This captures command output
echo "Current .ksh files:"
for f in range(0, 3):
    echo "  File $f"

# For loop with block and break
echo "Loop with break at 3:"
for i in range(0, 10) {
    if $i == 3:
        break
    echo "  $i"
}

# For loop with continue
echo "Loop with continue (skip 3):"
for i in range(0, 5) {
    if $i == 3:
        continue
    echo "  $i"
}

# Using $_i for index in repeat (see below)

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                       WHILE LOOPS                                        ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== WHILE LOOPS ==="

# Simple while
x = 0
while $x < 3:
    echo "While: x = $x"
    x++

# While with block
x = 0
while $x < 3 {
    echo "While block: x = $x"
    x++
}

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                       DO-WHILE / DO-UNTIL                               ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== DO-WHILE / DO-UNTIL ==="

# Do-while (always runs at least once)
x = 0
do {
    echo "Do-while: x = $x"
    x++
} while $x < 3

# Do-until (runs until condition is true)
x = 0
do {
    echo "Do-until: x = $x"
    x++
} until $x >= 3

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                       REPEAT LOOPS                                       ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== REPEAT LOOPS ==="

# Repeat N times
repeat 3:
    echo "Hello!"

# Repeat with block (access $_i for index)
repeat 5 {
    echo "Iteration: $_i"
}

# Calculate sum using repeat
sum = 0
repeat 5 {
    sum = $sum + $_i
}
echo "Sum of 0+1+2+3+4 = $sum"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                       TRY / CATCH                                        ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo " === TRY / CATCH ==="

# Basic try-catch
try {
    echo "Trying something safe..."
    result = 10 / 2
    echo "Result: $result"
} catch {
    echo "This won't run - no error!"
}

# Try-catch with error
try {
    # This would cause an error in real scenario
    # For demo, we'll just show the syntax
    echo "Trying potentially failing operation..."
    # Simulate an error condition
    x = 0
    if $x == 0:
        # Using throw to simulate error
        throw "Division by zero!"
} catch e {
    echo "Caught error: $e"
} finally {
    echo "Finally block always runs"
}

# Try without catch (just finally)
try {
    echo "Trying..."
} finally {
    echo "Cleanup code in finally"
}

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    NESTED CONTROL FLOW                                  ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== NESTED CONTROL FLOW ==="

# Nested loops with break
for i in range(0, 5) {
    for j in range(0, 5) {
        if $j == 2:
            break
        echo "i=$i, j=$j"
    }
}

# Match inside a for loop
fruits = ["apple", "banana", "cherry", "date"]
for fruit in $fruits {
    match $fruit {
        case "apple": echo "$fruit - keep doctor away"
        case "banana": echo "$fruit - energy boost"
        case "cherry": echo "$fruit - sweet treat"
        case *: echo "$fruit - unknown"
    }
}

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    PRACTICAL EXAMPLES                                   ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== PRACTICAL EXAMPLE: Number Guessing ==="

# Simulate a simple game
target = 7
guess = 3

if $guess < $target:
    echo "Too low!"
elif $guess > $target:
    echo "Too high!"
else:
    echo "Correct!"

echo ""
echo "=== PRACTICAL EXAMPLE: Grade Calculator ==="

grade_score = 88
match $grade_score {
    case >= 90: grade = "A"; desc = "Excellent"
    case >= 80: grade = "B"; desc = "Good"
    case >= 70: grade = "C"; desc = "Average"
    case >= 60: grade = "D"; desc = "Pass"
    default:    grade = "F"; desc = "Fail"
}
echo "Score: $grade_score => Grade: $grade ($desc)"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                         PRACTICE EXERCISES                             ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Try these:
# 1. Create a for loop that calculates factorial
# 2. Use match to convert day number to day name
# 3. Create a while loop that finds first number divisible by 7
# 4. Use nested loops to print a multiplication table
# 5. Create a simple menu system using match

echo ""
echo "=== Control Flow Tutorial Complete! ==="
