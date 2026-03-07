#!/usr/bin/env katsh
# =============================================================================
# 05_functions.ksh - Functions and Advanced Features
# =============================================================================
# This script teaches you about:
# - Creating and calling functions
# - Return values
# - Variable arguments
# - Import/Export
# - Error handling
# - Importing modules
#
# RUN THIS SCRIPT: ./katsh examples/05_functions.ksh
# =============================================================================

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                       FUNCTION BASICS                                   ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo "=== FUNCTION BASICS ==="

# Define a simple function
func greet(name) {
    echo "Hello, $name!"
}

# Call the function
greet "Alice"
greet "Bob"

# Function with return value
func add(a, b) {
    return $a + $b
}

result = add 3 7
echo "3 + 7 = $result"

# Function that returns a value
func multiply(x, y) {
    result = $x * $y
    return $result
}

product = multiply 4 5
echo "4 * 5 = $product"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    VARIABLE ARGUMENTS                                   ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== VARIABLE ARGUMENTS ==="

# Functions can accept extra arguments into $_args
func log(level, msg) {
    echo "[$level] $msg"
    echo "  Extra args: $_args"
}

log "INFO" "Application started"
log "ERROR" "Connection failed" "timeout" "retry=3"

# Use $_argc to check argument count
func count_args() {
    echo "Number of args: $_argc"
}

count_args "one"
count_args "one" "two" "three"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    SPECIAL VARIABLES IN FUNCTIONS                       ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== SPECIAL FUNCTION VARIABLES ==="

# $_return - return value of last function call
func get_value() {
    return 42
}
get_value
echo "Return value: $_return"

# $_error - last error message
# (set when throw/raise is used)
# try { throw "test error" } catch e: echo $e

# $_args - extra arguments beyond named params
func show_all(first) {
    echo "First: $first"
    echo "Rest: $_args"
}
show_all "alpha" "beta" "gamma"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    RECURSIVE FUNCTIONS                                  ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== RECURSIVE FUNCTIONS ==="

# Factorial function
func factorial(n) {
    if $n <= 1:
        return 1
    prev = factorial `echo $n - 1 | bc`
    return $n * $prev
}

echo "Factorial of 5: $(factorial 5)"
echo "Factorial of 3: $(factorial 3)"

# Fibonacci
func fib(n) {
    if $n <= 1:
        return $n
    return fib($n - 1) + fib($n - 2)
}

echo "Fibonacci(6): $(fib 6)"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    IMPORT / EXPORT                                      ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== IMPORT / EXPORT ==="

# Export a variable to the OS environment
# export MY_VAR="hello"
# echo $MY_VAR  # In shell, this would be $MY_VAR

# Export a function (available to child processes)
# export func my_function

# List exported variables
# export

# Import scripts
# import "./helpers.ksh"
# import "https://example.com/lib.ksh"
# import "username/repo/path/script.ksh"

echo "Note: Import statements would load external scripts"
echo "Example: import './utils.ksh'"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    PIPE OPERATORS                                        ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== PIPE OPERATORS ==="

# Pipe operators transform command output
# Commands like: select, where, sort, limit, grep, etc.

# Example: ls | select name | where name~.go | sort name
# This would: list files, keep name column, filter .go files, sort

echo "Sample data processing:"
echo "ls | select name,size | sort size desc | limit 5"
echo "(Shows 5 largest files)"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    BOX STORAGE                                         ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== BOX STORAGE ==="

# Store command results with #=
ls #=myfiles
# box get myfiles

echo "Box commands:"
echo "  ls #=key        - Store result with key"
echo "  box get key     - Retrieve stored result"
echo "  box             - List all stored results"
echo "  box rm key      - Delete stored result"
echo "  box rename old new"
echo "  box tag key tagname"
echo "  box search query"
echo "  box export file.json"
echo "  box import file.json"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    ALIASES                                              ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== ALIASES ==="

# Create shortcuts for commands
# alias ll='ls -la'
# alias gs='git status'

echo "Alias commands:"
echo "  alias name='command'"
echo "  unalias name"
echo "  aliases"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    SCRIPT ARGUMENTS                                     ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== SCRIPT ARGUMENTS ==="

echo "Script name: $0"
echo "First arg: $1"
echo "Second arg: $2"
echo "Arg count: $#"
echo "All args: $_args"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    SPECIAL SCRIPT VARIABLES                             ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== SPECIAL SCRIPT VARIABLES ==="

echo "\$__script_dir - Directory of running script"
echo "\$__script_file - Full path to running script"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    SHELL PASSTHROUGH                                    ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== SHELL PASSTHROUGH ==="

# Run bash/zsh commands directly
# bash! git status
# zsh! echo "ZSH here"
# sh! for i in *.txt; do echo $i; done

# Run in passthrough mode (real stdin/stdout)
# run git log --oneline
# ! any command

echo "Note: Use ! prefix to run shell commands"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    PRACTICAL FUNCTION EXAMPLES                          ║
# ╚══════════════════════════════════════════════════════════════════════════╝

echo ""
echo "=== PRACTICAL FUNCTION EXAMPLES ==="

# Validation function
func is_valid_email(email) {
    # Simple check - contains @
    if $(echo $email | contains "@"):
        return true
    return false
}

if is_valid_email "test@example.com":
    echo "Valid email!"
else:
    echo "Invalid email"

# String utility functions
func to_uppercase(text) {
    return $(echo $text | upper)
}

func capitalize(text) {
    return $(echo $text | title)
}

echo "Upper: $(to_uppercase hello)"
echo "Title: $(capitalize hello world)"

# Calculator function
func calculate(a, op, b) {
    match $op {
        case "+": return $(echo $a + $b | bc)
        case "-": return $(echo $a - $b | bc)
        case "*": return $(echo $a \* $b | bc)
        case "/": return $(echo "scale=2; $a / $b" | bc)
        default: return "Error"
    }
}

echo "5 + 3 = $(calculate 5 "+" 3)"
echo "10 - 4 = $(calculate 10 "-" 4)"
echo "6 * 7 = $(calculate 6 "*" 7)"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                         PRACTICE EXERCISES                               ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Try these:
# 1. Create a function that checks if a number is prime
# 2. Create a function that converts Celsius to Fahrenheit
# 3. Create a function that finds the maximum in an array
# 4. Create a function that reverses a string
# 5. Create a function that generates a random number between min and max

echo ""
echo "=== Functions Tutorial Complete! ==="
