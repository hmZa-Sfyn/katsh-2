#!/usr/bin/env katsh
# =============================================================================
# 01_variables.ksh - Variables and Basic Operations
# =============================================================================
# This script teaches you how to work with variables in Katsh.
# Variables are used to store data that you can use later.
#
# RUN THIS SCRIPT: ./katsh examples/01_variables.ksh
# =============================================================================

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                           VARIABLES BASICS                               ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# -----------------------------------------------------------------------------
# 1. BASIC ASSIGNMENT
# Variables store values. Use = to assign.
# -----------------------------------------------------------------------------

# Strings - wrap in double quotes for interpolation
name = "Alice"
echo "My name is: $name"

# Numbers (integers)
age = 25
echo "I am $age years old"

# Booleans (true/false)
is_student = true
echo "Am I a student? $is_student"

# -----------------------------------------------------------------------------
# 2. STRING INTERPOLATION
# Double quotes allow variable expansion ($name becomes the value)
# Single quotes keep everything literal (no interpolation)
# -----------------------------------------------------------------------------

greeting = "Hello, $name!"        # Interpolated: Hello, Alice!
echo $greeting

raw_string = 'No $interpolation here'  # Literal: No $interpolation here
echo $raw_string

# Advanced interpolation with defaults
unset_var = "fallback"
echo "Default: ${unset_var:-default}"      # fallback (variable is set)
echo "Missing: ${does_not_exist:-missing}" # missing (variable not set)

# String length
echo "Length of name: ${#name}"            # 5

# Conditional expansion
echo "If set: ${name:+Hi $name}"           # Hi Alice (name is set)
echo "If not set: ${missing:+value}"       # (empty, missing not set)

# -----------------------------------------------------------------------------
# 3. ARITHMETIC OPERATORS
# Katsh supports arithmetic operations
# -----------------------------------------------------------------------------

x = 10
echo "x = $x"

# Compound assignment
x += 5
echo "x += 5 => $x"    # 15

x -= 3
echo "x -= 3 => $x"    # 12

x *= 2
echo "x *= 2 => $x"    # 24

x /= 6
echo "x /= 6 => $x"    # 4

x %= 3
echo "x %= 3 => $x"    # 1

x **= 2
echo "x **= 2 => $x"   # 1 (1 to the power of 2)

# Increment/Decrement
x = 5
x++
echo "x++ => $x"       # 6

x--
echo "x-- => $x"       # 5

++x
echo "++x => $x"       # 6

# Direct arithmetic in expressions
result = 3 + 4 * 2
echo "3 + 4 * 2 = $result"  # 14 (left to right, no precedence)

result = (3 + 4) * 2
echo "(3 + 4) * 2 = $result"  # 14

result = 2 ** 10
echo "2 ** 10 = $result"    # 1024 (power)

result = 17 % 5
echo "17 % 5 = $result"     # 2 (modulo/remainder)

# -----------------------------------------------------------------------------
# 4. SUBSHELL CAPTURE
# Capture command output into a variable using backticks or $()
# -----------------------------------------------------------------------------

# Using backticks (original syntax)
current_dir = `pwd`
echo "Current directory: $current_dir"

# Using $() (POSIX syntax, can be nested)
file_count = $(ls | count)
echo "Number of files in current dir: $file_count"

# Inside a string
message = "Running on $(hostname)"
echo $message

# -----------------------------------------------------------------------------
# 5. TERNARY / INLINE IF
# Conditionally assign a value in one line
# -----------------------------------------------------------------------------

score = 85
grade = if $score >= 90: "A"; else: if $score >= 80: "B"; else: "C"
echo "Score: $score => Grade: $grade"

# Simpler form
status = if $age >= 18: "adult"; else: "minor"
echo "Age $age => $status"

# Using test flags
file = "test.txt"
result = if -f $file: "exists"; else: "missing"
echo "File $file: $result"

# -----------------------------------------------------------------------------
# 6. BOOLEAN OPERATORS
# && (and), || (or), not (not)
# -----------------------------------------------------------------------------

a = true
b = false

if $a and $b:
    echo "Both are true"
else:
    echo "Not both true"

if $a or $b:
    echo "At least one is true"

if not $b:
    echo "b is false"

# Short-circuit with commands
# && runs next command only if previous succeeded
# || runs next command only if previous failed

echo "Testing &&:"
-f README.md && echo "README.md exists!"

echo "Testing ||:"
-f nonexistent.txt || echo "File doesn't exist!"

# -----------------------------------------------------------------------------
# 7. TEST FLAGS
# Check file types and string properties
# -----------------------------------------------------------------------------

# File tests
if -f README.md: echo "README.md is a regular file"
if -d examples: echo "examples is a directory"
if -z "": echo "Empty string is zero-length"
if -n "hello": echo "Non-empty string has content"

# -----------------------------------------------------------------------------
# 8. SCOPED VARIABLES WITH 'with'
# Create temporary scoped variables
# -----------------------------------------------------------------------------

with x = 42 {
    echo "Inside scope: x = $x"
    with y = 99 {
        echo "Nested: x = $x, y = $y"
    }
}
# x and y are not accessible here

# Inline scope
with name = "Bob": echo "Hello $name"

# -----------------------------------------------------------------------------
# 9. READONLY VARIABLES
# Make variables unchangeable
# -----------------------------------------------------------------------------

MAX_SIZE = 100
readonly MAX_SIZE
# MAX_SIZE = 200  # This would cause an error!

echo "MAX_SIZE is readonly: $MAX_SIZE"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                         PRACTICE EXERCISES                                ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Try these on your own:
# 1. Create variables for first name and last name, combine them
# 2. Use arithmetic to calculate area of a circle (π * r²)
# 3. Capture the current date into a variable
# 4. Use a ternary to check if a number is even or odd
# 5. Test if a file exists and print appropriate message

echo ""
echo "=== Variables Tutorial Complete! ==="
