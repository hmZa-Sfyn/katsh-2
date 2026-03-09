# KatSH v2.0 — Feature Reference

> A modern scripting shell with rich types, smart pipes, and beautiful errors.  
> Run scripts with `katsh script.ksh` or launch the interactive REPL with `katsh`.

---

## Table of Contents

1. [Variables](#1-variables)
2. [Strings](#2-strings)
3. [Numbers & Math](#3-numbers--math)
4. [Arrays](#4-arrays)
5. [Maps](#5-maps)
6. [Advanced Collections](#6-advanced-collections)
7. [Conditionals](#7-conditionals)
8. [Loops](#8-loops)
9. [Functions](#9-functions)
10. [Pattern Matching](#10-pattern-matching)
11. [Error Handling](#11-error-handling)
12. [Pipe Operators](#12-pipe-operators)
13. [Pipe Expressions `|>`](#13-pipe-expressions-)
14. [Shell Passthrough](#14-shell-passthrough)
15. [Background Execution `?(...)`](#15-background-execution-)
16. [Structs & Enums](#16-structs--enums)
17. [Switch](#17-switch)
18. [Defer, With, Goto](#18-defer-with-goto)
19. [When Guards](#19-when-guards)
20. [Throw / Raise](#20-throw--raise)
21. [Box Storage](#21-box-storage)
22. [Script Flags & Special Vars](#22-script-flags--special-vars)
23. [Built-in Commands](#23-built-in-commands)
24. [Error Messages](#24-error-messages)

---

## 1. Variables

Assign with `=`. No `$` on the left, always `$` to read.

```ksh
name = "Alice"
age  = 30
echo "Hello $name, you are $age"

# Readonly — cannot be reassigned
readonly PI = 3.14159

# Unset
unset name
```

**Compound assignment operators:**

```ksh
x = 10
x += 5    # x = 15
x -= 3    # x = 12
x *= 2    # x = 24
x /= 4    # x = 6
x **= 2   # x = 36  (power)
x++       # x = 37
x--       # x = 36
```

---

## 2. Strings

```ksh
s = "hello world"

# Interpolation inside double quotes
greeting = "Hi $name!"

# Single quotes — no interpolation (literal)
raw = 'cost is $5'

# Multiline (heredoc)
msg = """
  line one
  line two
"""
```

**String pipe ops** (use with `|` or `|>`):

| Op | What it does | Example |
|----|-------------|---------|
| `upper` | UPPERCASE | `"hi" \| upper` → `HI` |
| `lower` | lowercase | `"HI" \| lower` → `hi` |
| `trim` | strip whitespace | `"  hi  " \| trim` → `hi` |
| `length` | character count | `"hello" \| length` → `5` |
| `reverse` | reverse chars | `"abc" \| reverse` → `cba` |
| `replace old new` | find & replace | `"hi world" \| replace "world" "earth"` |
| `split delim` | split to array | `"a,b,c" \| split ","` |
| `join delim` | join array | `arr \| join ", "` |
| `contains text` | boolean check | `"hello" \| contains "ell"` |
| `startswith text` | prefix check | `"hello" \| startswith "he"` |
| `endswith text` | suffix check | `"hello" \| endswith "lo"` |
| `repeat N` | repeat string | `"ab" \| repeat 3` → `ababab` |
| `pad N` | pad to width | `"hi" \| pad 10` |
| `substr start len` | substring | `"hello" \| substr 1 3` → `ell` |
| `regex pattern` | regex match | `"foo123" \| regex "[0-9]+"` |

---

## 3. Numbers & Math

```ksh
a = 10
b = 3

sum  = $a + $b     # 13
diff = $a - $b     # 7
prod = $a * $b     # 30
quot = $a / $b     # 3.333...
mod  = $a % $b     # 1
pow  = $a ** $b    # 1000
```

**Math pipe ops:**

| Op | Example |
|----|---------|
| `add N` | `5 \| add 3` → `8` |
| `sub N` | `10 \| sub 4` → `6` |
| `mul N` | `3 \| mul 7` → `21` |
| `div N` | `20 \| div 4` → `5` |
| `abs` | `-5 \| abs` → `5` |
| `round` | `3.7 \| round` → `4` |
| `floor` | `3.9 \| floor` → `3` |
| `ceil` | `3.1 \| ceil` → `4` |
| `sqrt` | `9 \| sqrt` → `3` |
| `min N` | `5 \| min 3` → `3` |
| `max N` | `5 \| max 9` → `9` |
| `clamp lo hi` | `15 \| clamp 0 10` → `10` |

---

## 4. Arrays

```ksh
# Define
fruits = ["apple", "banana", "cherry"]

# Append
fruits[] = "date"

# Read by index (0-based, negatives count from end)
echo $fruits[0]    # apple
echo $fruits[-1]   # date

# Update
fruits[1] = "blueberry"

# Length
echo $fruits[len]  # 4

# Iterate
for item in $fruits {
  echo $item
}
```

**Array pipe ops:**

| Op | Example |
|----|---------|
| `arr_len` | `$fruits \| arr_len` → `4` |
| `arr_push val` | `$arr \| arr_push "x"` |
| `arr_pop` | `$arr \| arr_pop` |
| `arr_sort` | `$arr \| arr_sort` |
| `arr_reverse` | `$arr \| arr_reverse` |
| `arr_unique` | `$arr \| arr_unique` |
| `arr_join delim` | `$arr \| arr_join ", "` |
| `arr_contains val` | `$arr \| arr_contains "apple"` |
| `arr_index val` | `$arr \| arr_index "banana"` → `1` |
| `arr_slice start end` | `$arr \| arr_slice 1 3` |

---

## 5. Maps

Key-value store. Keys and values are strings.

```ksh
# Define
person = map { name: "Alice"  age: "30"  city: "Cairo" }

# Read
echo $person[name]      # Alice

# Set / update
person[email] = "alice@example.com"

# Delete a key
map_del person city

# Check key exists
if map_has $person age {
  echo "age is $person[age]"
}

# Iterate
for key val in $person {
  echo "$key = $val"
}
```

**Map pipe ops:**

| Op | Example |
|----|---------|
| `map_keys` | `$person \| map_keys` |
| `map_values` | `$person \| map_values` |
| `map_has key` | `$person \| map_has "name"` |
| `map_size` | `$person \| map_size` → `3` |
| `map_merge other` | `$a \| map_merge $b` |
| `map_invert` | swaps keys and values |

---

## 6. Advanced Collections

### Set — unique values only
```ksh
s = set { 1 2 3 2 1 }
s[] = 4
echo $s[has:2]    # true
```

### Stack — LIFO push/pop
```ksh
stk = stack {}
stack_push stk "first"
stack_push stk "second"
top = $(stack_pop stk)   # "second"
```

### Queue — FIFO enqueue/dequeue
```ksh
q = queue {}
queue_push q "a"
queue_push q "b"
first = $(queue_pop q)   # "a"
```

### Tuple — fixed-size, ordered record
```ksh
point = tuple(10, 20)
echo $point[0]    # 10
echo $point[1]    # 20
```

### Matrix — 2D numeric grid
```ksh
m = matrix(3, 3)         # 3×3 zeros
matrix_set m 0 0 42
val = $(matrix_get m 0 0)  # 42
```

---

## 7. Conditionals

```ksh
# if / elif / else
if $age >= 18 {
  echo "adult"
} elif $age >= 13 {
  echo "teen"
} else {
  echo "child"
}

# unless (opposite of if)
unless $logged_in {
  echo "please log in"
}

# Inline ternary assignment
label = if $score >= 60 then "pass" else "fail"

# Condition operators
# ==  !=  >  <  >=  <=
# &&  ||  !
# is_empty  is_number  is_file  is_dir
if is_empty $name {
  echo "name is blank"
}
```

---

## 8. Loops

```ksh
# for-in (array)
for item in $fruits {
  echo $item
}

# for range
for i in 1..10 {
  echo $i
}

# for range with step
for i in 0..20 step 5 {
  echo $i
}

# while
x = 0
while $x < 5 {
  echo $x
  x++
}

# do-while (runs at least once)
do {
  echo "tick"
  x++
} while $x < 3

# until (opposite of while)
until $x >= 10 {
  x++
}

# repeat N
repeat 3 {
  echo "hello"
}

# break / continue
for i in 1..10 {
  if $i == 5 { break }
  if $i % 2 == 0 { continue }
  echo $i
}
```

---

## 9. Functions

```ksh
# Define
func greet(name) {
  echo "Hello, $name!"
}

# Call
greet "Alice"

# Return a value
func add(a, b) {
  return $a + $b
}

result = add(3, 4)    # result = "7"
echo $result

# Multiple params with defaults
func greet(name, title="Dr.") {
  echo "Hello $title $name"
}

# Variadic — extra args in $_args
func sum(first) {
  total = $first
  for n in $_args { total += $n }
  return $total
}

total = sum(1 2 3 4 5)    # 15

# Recursion
func fib(n) {
  if $n <= 1 { return $n }
  a = fib($n - 1)
  b = fib($n - 2)
  return $a + $b
}
```

---

## 10. Pattern Matching

### match / case
```ksh
match $status {
  "ok":    echo "success"
  "err":   echo "failure"
  "warn":  echo "warning"
  default: echo "unknown: $status"
}

# Glob patterns
match $filename {
  "*.log":  echo "log file"
  "*.json": echo "json file"
  default:  echo "other"
}
```

### switch (strict equality, supports fallthrough)
```ksh
switch $day {
  "Mon": echo "Start of week"
  "Fri": echo "End of week"
  "Sat":
  "Sun": echo "Weekend"    # ← fallthrough from Sat
  default: echo "Midweek"
}
```

---

## 11. Error Handling

```ksh
# try / catch / finally
try {
  result = risky_operation()
} catch e {
  echo "caught: $e"
} finally {
  echo "always runs"
}

# Nested try
try {
  try {
    throw "inner error"
  } catch e {
    echo "inner caught: $e"
    throw "re-thrown: $e"
  }
} catch e {
  echo "outer caught: $e"
}

# $? — last exit code
run_command
if $? != 0 {
  echo "command failed"
}

# _error — last error message
echo $_error
```

---

## 12. Pipe Operators

Connect commands with `|`. Katsh pipes transform data — OS pipes (`figlet | lolcat`) are automatically detected and routed to the system shell.

```ksh
# Table pipes (work on command output)
ps | where cpu>10 | sort cpu desc | limit 5
ls  | grep ".go" | count
cat log.txt | grep "ERROR" | limit 20

# String pipes
echo "  Hello World  " | trim | lower      # "hello world"
echo "hello" | upper | repeat 3             # "HELLOHELLOHELLO"
cat words.txt | sort | unique

# OS pipes — transparently forwarded to bash
figlet "Hello" | lolcat
ls -la | grep go | head -10
curl -s https://api.example.com | jq .
```

**Table operators:**

| Op | Example |
|----|---------|
| `select col1,col2` | `ps \| select pid,name` |
| `where col=val` | `ps \| where name=bash` |
| `where col>val` | `ps \| where cpu>5` |
| `sort col [desc]` | `ps \| sort mem desc` |
| `limit N` | `ls \| limit 10` |
| `skip N` | `ls \| skip 5` |
| `count` | `ls \| count` |
| `unique col` | `ps \| unique name` |
| `reverse` | `ls \| reverse` |
| `grep text` | `ps \| grep python` |
| `fmt json\|csv\|tsv` | `ps \| fmt json` |
| `add col=val` | `ps \| add env=prod` |
| `rename old=new` | `ps \| rename pid=id` |

---

## 13. Pipe Expressions `|>`

Chain transformations inline in expressions. Each stage passes its result to the next.

```ksh
# Assign the final value
result = "  Hello World  " |> trim |> lower |> replace "world" "katsh"
echo $result    # "hello katsh"

# Math chain
val = 2 |> mul 10 |> add 5 |> sqrt
echo $val    # 5

# User functions in the chain
func double(n) { return $n * 2 }
big = 3 |> double |> double |> double    # 24

# Standalone (prints the result)
"hello" |> upper |> repeat 2    # HELLOHELLO
```

---

## 14. Shell Passthrough

Run native shell commands directly.

```ksh
# Prefix with bash!/zsh!/sh!
bash! git log --oneline -10
zsh!  source ~/.zshrc
sh!   echo $SHELL

# run — auto-detects shell
run curl -fsSL https://example.com

# capture — run and capture output to variable
output = capture git status --short

# $(...) — inline capture
branch = $(git rev-parse --abbrev-ref HEAD)
echo "on branch: $branch"

# Backtick — same as $(...)
user = `whoami`

# OS pipes — automatic, no prefix needed
figlet "KatSH" | lolcat
cat /etc/hosts | awk '{print $1}' | sort | uniq
```

---

## 15. Background Execution `?(...)`

Run a command asynchronously. Multiple `?(...)` on the same line start in **parallel** — KatSH collects all results before continuing.

```ksh
# Single background command
data = ?(curl -s https://api.example.com/users)

# Parallel fetches — both start at the same time
users  = ?(curl -s https://api.example.com/users)
orders = ?(curl -s https://api.example.com/orders)
# Results are ready here — whichever finished first waited for the other
echo "users: $users"
echo "orders: $orders"

# Inside strings
status = "fetched: ?(curl -s https://example.com/status)"

# Error handling
try {
  result = ?(some-command --that-might-fail)
} catch e {
  echo "background command failed: $e"
}
```

---

## 16. Structs & Enums

### struct
```ksh
# Define the type
struct Point { x y }

# Instantiate
p = Point(10, 20)
echo $p_x    # 10
echo $p_y    # 20

# Update a field
p_x = 99

# With defaults
struct Person { name  age="unknown"  role="user" }
alice = Person("Alice")
echo $alice_age    # unknown
```

### enum
```ksh
enum Color { Red Green Blue }
echo $Color_Red     # 0
echo $Color_Green   # 1
echo $Color_Blue    # 2

# Custom start value
enum Status { Pending=1  Active  Closed }
echo $Status_Active    # 2
```

---

## 17. Switch

Already covered in §10, but note the `fallthrough` keyword:

```ksh
switch $code {
  "200": echo "OK"
  "301":
  "302": echo "Redirect"   # 301 falls through to this
  "404": echo "Not Found"
  "500": echo "Server Error"
  default: echo "Unknown: $code"
}
```

---

## 18. Defer, With, Goto

### defer — run cleanup when the block exits
```ksh
defer echo "cleanup!"
defer rm /tmp/myfile.tmp

echo "doing work..."
# "cleanup!" and rm run here, in reverse order
```

### with — scoped variable binding
```ksh
with conn = open_db("mydb") {
  echo "using: $conn"
}
# $conn is unset here

# One-liner
with x = 42: echo "x is $x"
```

### goto / label — arbitrary control flow
```ksh
label retry:
  result = $(try_connect)
  if $result == "fail" {
    tries++
    goto retry when $tries < 5
  }
echo "connected after $tries tries"
```

---

## 19. When Guards

Add a condition to any command — runs only when condition is true.

```ksh
echo "big number!" when $n > 100
rm temp.log when -f temp.log
x++ when $x < 10
send_alert when $errors > 0
```

---

## 20. Throw / Raise

```ksh
func divide(a, b) {
  if $b == 0 { throw "cannot divide by zero" }
  return $a / $b
}

result = try {
  divide(10, 0)
} catch e {
  echo "Error: $e"    # Error: cannot divide by zero
  return -1
}

# raise is an alias for throw
raise "something went wrong"

# $_error holds the last thrown message
echo $_error
```

---

## 21. Box Storage

Store and recall command results within a session.

```ksh
# Store with auto-key
ps #=

# Store with named key
ps #=procs

# Retrieve
box get procs
box list
box show procs
box drop procs

# Query stored data
box get procs | where cpu>5 | sort cpu desc
```

---

## 22. Script Flags & Special Vars

### Run a script
```sh
katsh script.ksh
katsh -v script.ksh      # verbose: print each line
katsh -x script.ksh      # trace: print with + prefix
katsh -e script.ksh      # exit-on-error
katsh -n script.ksh      # dry-run: parse only
katsh -t script.ksh      # timing: show ms per line
```

### Special variables

| Variable | Contains |
|----------|---------|
| `$?` | Exit code of last command |
| `$_return` | Return value of last function call |
| `$_error` | Last error or thrown message |
| `$_args` | Extra variadic args passed to a function |
| `$_argc` | Number of args passed to a function |
| `$_i` | Current loop iteration index (in `for` loops) |
| `$0` | Script name |
| `$1` `$2` … | Positional script arguments |
| `$HOME` `$PATH` etc. | All OS environment variables |

---

## 23. Built-in Commands

KatSH ships ~60 built-in commands. A quick tour:

**Navigation & files**
```ksh
cd ~/projects       pwd         ls -la
pushd /tmp          popd        dirs
mkdir -p a/b/c      rm -rf tmp  cp src dst
mv old new          touch file  ln -s src dst
```

**Viewing & searching**
```ksh
cat file.txt        head -20 file    tail -f log
wc -l file          stat file        find . -name "*.go"
grep "error" log    diff a b
```

**Text processing**
```ksh
echo "hello"        printf "%s\n" hi   read name
sort file           uniq file          sed s/old/new/ file
awk '{print $1}'    cut -d: -f1        tr a-z A-Z
```

**System info**
```ksh
ps                  kill 1234          sleep 2
date                uptime             whoami
uname -a            hostname           id
```

**Network**
```ksh
curl https://example.com           ping google.com
nslookup example.com               ifconfig
```

**Hashing**
```ksh
echo "hello" | md5sum
echo "hello" | sha256sum
```

**Shell extras**
```ksh
alias ll = "ls -la"
unalias ll
aliases
history
export MY_VAR = "value"
vars              # show all variables
type echo         # show what a name resolves to
help              # built-in help
help echo         # help for a specific command
```

---

## 24. Error Messages

KatSH shows rich, located error messages with source highlighting:

```
  SyntaxError[E002] at line 12, col 5  unexpected token '}'
  ╭─ line 12
  │  if $x > 10 }
  │       ~~~~~~
  │       ^── here
  │
  ╰─ 💡 hint: missing opening '{' before this '}'
     fix :  if $x > 10 { ... }
```

### Error codes

| Code | Kind | Meaning |
|------|------|---------|
| E001 | CommandNotFound | Unknown command or typo |
| E002 | SyntaxError | Invalid syntax |
| E003 | TypeError | Wrong type for operation |
| E004 | RuntimeError | General runtime failure |
| E005 | UndefinedVariable | Variable used before assignment |
| E006 | DivisionByZero | `x / 0` or `x % 0` |
| E007 | ArgCountError | Wrong number of arguments |
| E008 | ReadonlyError | Reassigning a `readonly` variable |
| E009 | UnhandledThrow | `throw` not caught by any `try` block |
| E012 | BackgroundError | `?(cmd)` command failed |
| E013 | IndexError | Array index out of bounds or wrong type |

### Error is always suppressed inside `try`
```ksh
try {
  throw "oops"    # ← no print here
} catch e {
  echo $e         # "oops"  ← message arrives here intact
}
```

---

## Quick-Reference Card

```ksh
# Variables
x = 10 / name = "Alice" / readonly PI = 3.14

# Strings
"Hello $name" / 'literal $5' / upper lower trim replace split join

# Arrays
arr = ["a","b","c"] / arr[] = "d" / $arr[0] / $arr[-1] / $arr[len]

# Maps
m = map{k:v} / $m[key] / m[k]=v / map_has map_keys map_size

# Conditions
if $x > 0 { } elif { } else { }   /   unless $ok { }

# Loops
for i in 1..10 { }   /   while $x < 5 { }   /   repeat 3 { }

# Functions
func f(a,b) { return $a+$b }   /   result = f(3,4)

# Match / Switch
match $s { "a": ... default: ... }   /   switch $n { 1: ... }

# Errors
try { } catch e { echo $e } finally { }   /   throw "msg"

# Pipes (katsh)
ps | where cpu>5 | sort cpu desc | limit 10

# Pipes (OS — automatic passthrough)
figlet hello | lolcat   /   cat f | awk '{print $1}' | sort

# Pipe expressions
val = "hello" |> upper |> repeat 2

# Background
a = ?(cmd1)   b = ?(cmd2)   # run in parallel

# Shell passthrough
bash! git log --oneline   /   x = $(git rev-parse HEAD)

# Structs / Enums
struct Point { x y }   /   p = Point(1,2)   /   echo $p_x
enum Color { Red Green Blue }   /   echo $Color_Red

# Guards
echo "big" when $n > 100   /   rm tmp when -f tmp

# Defer / With / Goto
defer echo "bye"   /   with x=42 { echo $x }   /   goto label when cond

# Script flags
katsh -v -e -x script.ksh
```

---

*KatSH v2.0 — built with ❤️ in Go*
