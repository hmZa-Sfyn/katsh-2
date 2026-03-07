#!/usr/bin/env katsh
# =============================================================================
# 02_strings.ksh - String Operations
# =============================================================================
# This script teaches you how to work with strings in Katsh.
# Strings can be manipulated using various built-in operations.
#
# RUN THIS SCRIPT: ./katsh examples/02_strings.ksh
# =============================================================================

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                        STRING BASICS                                     ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Create strings with quotes
single = 'Single quotes - no interpolation'
double = "Double quotes - $single works"

echo "Single: $single"
echo "Double: $double"

# Empty string
empty = ""
echo "Empty string length: ${#empty}"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    CASE & WHITESPACE OPERATIONS                          ║
# ╚══════════════════════════════════════════════════════════════════════════╝

text = "  Hello World  "

echo "Original: '$text'"
echo "Upper:    '$(echo $text | upper)'"
echo "Lower:    '$(echo "HELLO" | lower)'"
echo "Title:    '$(echo "hello world" | title)'"
echo "Trim:     '$(echo $text | trim)'"
echo "Ltrim:    '$(echo $text | ltrim)'"
echo "Rtrim:    '$(echo $text | rtrim)'"

# Using pipe syntax (command style)
result = "hello" | upper
echo "Pipe to upper: $result"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    LENGTH & INSPECTION                                   ║
# ╚══════════════════════════════════════════════════════════════════════════╝

str = "Hello World"

echo "String: '$str'"
echo "Length: ${#str}"
echo "Starts with 'Hello': $(echo $str | startswith "Hello")"
echo "Ends with 'World': $(echo $str | endswith "World")"
echo "Contains 'lo Wo': $(echo $str | contains "lo Wo")"
echo "Is numeric '123': $(echo "123" | isnum)"
echo "Is alpha 'abc': $(echo "abc" | isalpha)"
echo "Is alnum 'abc123': $(echo "abc123" | isalnum)"
echo "Is upper 'ABC': $(echo "ABC" | isupper)"
echo "Is lower 'abc': $(echo "abc" | islower)"
echo "Is space '   ': $(echo "   " | isspace)"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    TRANSFORMATION OPERATIONS                            ║
# ╚══════════════════════════════════════════════════════════════════════════╝

str = "Hello World"

echo ""
echo "--- Transformation Operations ---"
echo "Reverse: $(echo $str | reverse)"
echo "Repeat 'ab' 3x: $(echo "ab" | repeat 3)"
echo "Replace 'o' with 'X': $(echo $str | replace "o" "X")"
echo "Replace first 'o' only: $(echo $str | replace1 "o" "X")"
echo "Sub from position 6: $(echo $str | sub 6)"
echo "Sub position 0, length 5: $(echo $str | sub 0 5)"
echo "Pad to width 20: '$(echo "hi" | pad 20)'"
echo "Left pad to 10: '$(echo "hi" | lpad 10)'"
echo "Center in 20: '$(echo "hi" | center 20)'"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    SPLIT & JOIN OPERATIONS                               ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Split a string into an array
csv = "apple,banana,cherry"
fruits = $csv | split ","
echo ""
echo "CSV: $csv"
echo "Split into array: $fruits"
echo "First fruit: $(echo $fruits | first)"
echo "Last fruit: $(echo $fruits | last)"

# Split by newlines
poem = "line one
line two
line three"
lines = $poem | lines
echo ""
echo "Poem lines: $lines"
echo "Number of lines: $(echo $lines | arr_len)"

# Split into words
sentence = "the quick brown fox"
words = $sentence | words
echo ""
echo "Words: $words"

# Split into characters
text = "abc"
chars = $text | chars
echo "Characters: $chars"

# Join array into string
arr = ["a", "b", "c"]
joined = $arr | join ","
echo ""
echo "Join array with comma: $joined"

# Concatenate strings
hello = "Hello"
world = "World"
combined = $hello . " " . $world
echo "Concatenate: $combined"

# Prepend
greeting = "world" | prepend "Hello "
echo "Prepend: $greeting"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    REPEATING STRINGS                                    ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Repeat character N times using interpolation
divider = "─"x60
echo ""
echo "Divider line:"
echo $divider

star_line = "*"x20
echo $star_line

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    HEREDOC STRINGS                                       ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Heredocs allow multi-line strings
name = "Alice"
message <<EOF
Dear $name,
This is a multi-line string.
You can include variables like $name here.
Best regards
EOF

echo ""
echo "Heredoc message:"
echo $message

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    ESCAPE SEQUENCES                                     ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Using printf for escape sequences
echo ""
printf "Tab: [\t]\n"
printf "Newline: [first\nsecond]\n"
printf "Quote: [\"quoted\"]\n"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                    PRACTICAL EXAMPLES                                   ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Example 1: Clean user input
user_input = "  ALICE  "
cleaned = $user_input | trim | lower
echo ""
echo "Input: '$user_input' => Cleaned: '$cleaned'"

# Example 2: Build a file path
dir = "/home/user"
file = "document.TXT"
ext = "txt"
path = $dir . "/" . $file
# Change extension
new_file = $file | replace ".TXT" ".pdf"
echo "Original file: $file => New file: $new_file"

# Example 3: Check filename
filename = "report_2024.pdf"
echo "Filename: $filename"
echo "  - Has .pdf: $(echo $filename | endswith ".pdf")"
echo "  - Starts with report: $(echo $filename | startswith "report")"

# ╔══════════════════════════════════════════════════════════════════════════╗
# ║                         PRACTICE EXERCISES                               ║
# ╚══════════════════════════════════════════════════════════════════════════╝

# Try these:
# 1. Take a sentence and reverse it
# 2. Count the number of words in a sentence
# 3. Convert a name to title case
# 4. Extract the extension from a filename
# 5. Create a formatted table using repeated characters

echo ""
echo "=== Strings Tutorial Complete! ==="
