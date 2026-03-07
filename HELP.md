# KatSH — Complete Reference

> **KatSH** is a structured, scriptable shell with typed values, advanced data
> types, and seamless bash/zsh passthrough.  
> Built on the **StructSH** engine · Version **0.4.0** · ~12 500 lines of Go

---

## Contents

1. [Quick Start](#1-quick-start)  
2. [Running Scripts (.ksh files)](#2-running-scripts-ksh-files)  
3. [Shell Passthrough — bash / zsh / sh](#3-shell-passthrough)  
4. [Variables & Expressions](#4-variables--expressions)  
5. [String Operations](#5-string-operations)  
6. [Array Operations](#6-array-operations)  
7. [Number Operations](#7-number-operations)  
8. [Pipe Operators `|`](#8-pipe-operators-)  
9. [Pipe Expression `|>`](#9-pipe-expression-)  
10. [Control Flow](#10-control-flow)  
11. [Functions](#11-functions)  
12. [Advanced Data Types](#12-advanced-data-types)  
13. [File & Directory Commands](#13-file--directory-commands)  
14. [Text Processing](#14-text-processing)  
15. [Process & System Commands](#15-process--system-commands)  
16. [Networking Commands](#16-networking-commands)  
17. [Box — Session Storage](#17-box--session-storage)  
18. [Aliases](#18-aliases)  
19. [History](#19-history)  
20. [Import / Export / Modules](#20-import--export--modules)  
21. [Error Reference](#21-error-reference)  
22. [Syntax Highlighting & Completion](#22-syntax-highlighting--completion)  
23. [Keyboard Shortcuts](#23-keyboard-shortcuts)  
24. [Complete Command Index](#24-complete-command-index)  

---

## 1. Quick Start

```bash
# Build from source
cd katsh/
go build -o katsh .

# Start the interactive REPL
./katsh

# Run a one-liner
./katsh -c "echo hello | upper"

# Run a script file
./katsh myscript.ksh

# Get help inside the REPL
help
help scripting
help datatypes
```

### Prompt

```
katsh ~/projects  ❯
       │           └── turns red if last command failed
       └── current directory (shortened)
```

---

## 2. Running Scripts (.ksh files)

### CLI flags

| Flag | Meaning |
|------|---------|
| `./katsh script.ksh` | Run a script file |
| `./katsh script.ksh a b c` | Run with positional args `$1 $2 $3` |
| `./katsh -e script.ksh` | Exit immediately on any non-zero exit code |
| `./katsh -x script.ksh` | Trace — print each line before executing |
| `./katsh -ex script.ksh` | Both `-e` and `-x` |
| `./katsh -n script.ksh` | Dry-run — parse and show lines, don't execute |
| `./katsh -c "cmd; cmd"` | Run inline command string |
| `./katsh --help` | Show usage |
| `./katsh --version` | Show version |

### Shebang

```sh
#!/usr/bin/env katsh
# My script
echo "Hello from KatSH"
```

### In-script option flags

```sh
#!/usr/bin/env katsh
# @set -e       exit on first error
# @set -x       trace mode (prints each line)
# @set -ex      both

x = 10
echo $x
```

### Positional arguments

```sh
#!/usr/bin/env katsh
echo "Script:    $0"       # script filename
echo "First arg: $1"
echo "Arg count: $#"
echo "All args:  $_args"
```

```bash
./katsh myscript.ksh hello world
# Script:    myscript.ksh
# First arg: hello
# Arg count: 2
# All args:  hello world
```

### Multi-line continuation

```sh
# Backslash continues a line
result = "start " \
         "middle " \
         "end"

curl -s \
     -H "Accept: application/json" \
     https://api.example.com/users
```

### Heredoc

```sh
message <<EOF
Dear $name,
Your order is ready.
EOF
echo $message
```

### Source another file

```sh
source ./utils.ksh              # relative to cwd
source $__script_dir/lib.ksh    # relative to current script
. ./helpers.ksh                 # POSIX dot syntax
```

### Special script variables

| Variable | Value |
|----------|-------|
| `$0` | Script filename |
| `$1` .. `$9` | Positional arguments |
| `$#` | Argument count |
| `$_args` | All args joined with spaces |
| `$__script_dir` | Directory of the running script |
| `$__script_file` | Absolute path of the running script |

---

## 3. Shell Passthrough

KatSH has two execution modes:

| Mode | Behaviour |
|------|-----------|
| **Capture** (default) | Output collected into a structured `Result` — can be piped, filtered, stored in Box |
| **Passthrough** | Real `stdin`/`stdout`/`stderr` connected — colours, pagers, interactive prompts, pipes all work |

### Passthrough syntax — all forms equivalent

```sh
bash! git log --oneline | head -20
zsh!  autoload -U compinit && compinit
sh!   for f in *.go; do wc -l "$f"; done
run   git log --oneline | grep feat
!     any command here
```

No quoting needed — the rest of the line is passed verbatim to the shell.

```sh
# Identical results:
bash! git diff HEAD~1 | grep "^+" | wc -l
run   git diff HEAD~1 | grep "^+" | wc -l
!     git diff HEAD~1 | grep "^+" | wc -l
```

### Explicit shell selection

```sh
bash! echo "I'm running in bash"
zsh!  print -P "%F{green}green text%f"
sh!   echo "POSIX sh"
ksh!  echo "Korn shell"
```

### Interactive shell sessions

```sh
bash         # drop into interactive bash
zsh          # drop into interactive zsh
bash!        # same (bare ! suffix)
zsh!         # same
```

### Native syntax (also routed through passthrough)

```sh
bash -c "git log | awk '{print $1}' | sort | uniq"
zsh  -c "print -P '%F{red}error%f'"
bash ./deploy.sh production us-east-1
zsh  ~/.config/zsh/init.zsh
```

### Capture output into a variable

```sh
# POSIX $() — works in any expression, including inside strings
branch = $(git branch --show-current)
hash   = $(git rev-parse --short HEAD)
count  = $(ls *.go | wc -l | tr -d ' ')

# Inside a string
msg    = "On branch $(git branch --show-current) — $(git log -1 --oneline)"
tag    = "v$(cat VERSION)-$(git rev-parse --short HEAD)"

# Backtick — same thing, original syntax
user   = `whoami`

# capture builtin — most explicit, stores stdout+stderr in a variable
capture result curl -sf https://api.example.com/health
echo $result

# Nested $()
info = "kernel: $(uname -r), user: $(whoami), home: $(echo $HOME)"
```

### Auto-passthrough (no prefix needed)

The following commands are **automatically** routed to passthrough mode:

```
Editors       vim  nvim  nano  emacs  micro  helix  hx  kak
Pagers        less  more  most
Monitors      htop  top  btop  glances  bpytop  iotop  iftop  nethogs  nload
Remote        ssh  telnet  mosh  ftp  sftp
Databases     mysql  psql  sqlite3  mongo  redis-cli  pgcli  mycli
REPLs         python  python3  node  ruby  irb  julia  R  lua  php  perl  ghci
Debuggers     gdb  lldb
Multiplexers  tmux  screen  zellij  byobu
TUI tools     ranger  nnn  mc  vifm  fzf  tig  lazygit  gitui
Media/mail    cmus  ncmpcpp  mutt  neomutt
Web           w3m  lynx
```

---

## 4. Variables & Expressions

### Assignment

```sh
name  = "Alice"
age   = 30
pi    = 3.14159
ready = true
```

### Compound assignment operators

```sh
x = 10
x += 5        # 15    — add
x -= 3        # 12    — subtract
x *= 2        # 24    — multiply
x /= 6        # 4     — divide
x %= 3        # 1     — modulo
x **= 8       # 1     — power (1^8=1)
x++           # 2     — increment
x--           # 1     — decrement
++x           # 2     — pre-increment
--x           # 1     — pre-decrement
```

### Arithmetic expressions

```sh
result = 3 + 4 * 2        # 11  (no precedence — left to right)
result = (3 + 4) * 2      # 14
result = 2 ** 10           # 1024
result = 17 % 5            # 2
```

### String interpolation

```sh
name = "World"
echo "Hello $name"                      # Hello World
echo "Length: ${#name}"                 # Length: 5
echo "Upper: $(echo $name | upper)"     # Upper: WORLD
echo "Default: ${unset_var:-fallback}"  # Default: fallback
echo "If set: ${name:+[${name}]}"       # If set: [World]
```

### Readonly / constant variables

```sh
MAX_SIZE = 100
readonly MAX_SIZE
MAX_SIZE = 200   # → E005 ReadonlyError
```

### Scoped binding with `with`

```sh
# Block scope
with x = 42 {
    echo $x      # 42
    with y = 99 {
        echo "$x $y"   # 42 99
    }
    # $y is gone
}
# $x is gone

# Inline scope
with name = "Bob": echo "Hello $name"
```

### Inline if expression (ternary)

```sh
label = if $score >= 90: "A"; else: "B"
msg   = if -f config.json: "found"; else: "missing"
```

### Special runtime variables

| Variable | Meaning |
|----------|---------|
| `$_return` | Return value of the last function call |
| `$_error` | Message from last `throw`/`raise` |
| `$_argc` | Argument count inside current function |
| `$_args` | Extra args beyond named params (in functions) |
| `$_i` | Loop counter in `repeat N` |

---

## 5. String Operations

Every string op works both as a **pipe stage** and a **standalone command**:

```sh
"hello world" | upper          # HELLO WORLD  (pipe)
upper "hello world"            # HELLO WORLD  (command)
x = "hello world" | upper      # store result
```

### Case & whitespace

| Command | Example | Result |
|---------|---------|--------|
| `upper` | `"hello" \| upper` | `HELLO` |
| `lower` | `"HELLO" \| lower` | `hello` |
| `title` | `"hello world" \| title` | `Hello World` |
| `trim` | `"  hi  " \| trim` | `hi` |
| `ltrim` | `"  hi  " \| ltrim` | `hi  ` |
| `rtrim` | `"  hi  " \| rtrim` | `  hi` |
| `strip` | same as `trim` | |

### Length & inspection

| Command | Example | Result |
|---------|---------|--------|
| `len` | `"hello" \| len` | `5` |
| `startswith` | `"hello" \| startswith "he"` | `true` |
| `endswith` | `"hello" \| endswith "lo"` | `true` |
| `contains` | `"hello" \| contains "ell"` | `true` |
| `isnum` | `"42" \| isnum` | `true` |
| `isalpha` | `"abc" \| isalpha` | `true` |
| `isalnum` | `"abc123" \| isalnum` | `true` |
| `isupper` | `"ABC" \| isupper` | `true` |
| `islower` | `"abc" \| islower` | `true` |
| `isspace` | `"   " \| isspace` | `true` |

### Transformation

| Command | Example | Result |
|---------|---------|--------|
| `reverse` | `"hello" \| reverse` | `olleh` |
| `repeat` | `"ab" \| repeat 3` | `ababab` |
| `replace` | `"foo foo" \| replace "foo" "bar"` | `bar bar` |
| `replace1` | `"foo foo" \| replace1 "foo" "bar"` | `bar foo` |
| `sub` | `"hello world" \| sub 6` | `world` |
| `sub` | `"hello world" \| sub 0 5` | `hello` |
| `pad` | `"hi" \| pad 10` | `hi        ` |
| `lpad` | `"hi" \| lpad 10` | `        hi` |
| `center` | `"hi" \| center 10` | `    hi    ` |

### Split & join

| Command | Example | Result |
|---------|---------|--------|
| `split` | `"a,b,c" \| split ","` | `["a","b","c"]` |
| `lines` | `"a\nb\nc" \| lines` | `["a","b","c"]` |
| `words` | `"hello world" \| words` | `["hello","world"]` |
| `chars` | `"abc" \| chars` | `["a","b","c"]` |
| `join` | `["a","b"] \| join ","` | `a,b` |
| `concat` | `"hello" \| concat " world"` | `hello world` |
| `prepend` | `"world" \| prepend "hello "` | `hello world` |

---

## 6. Array Operations

### Creating arrays

```sh
nums   = [1, 2, 3, 4, 5]
names  = ["Alice", "Bob", "Carol"]
empty  = []

# From command output (each line becomes an element)
files  = `ls *.go`
words  = "the quick brown fox" | split " "
```

### Access & mutation

```sh
echo $nums[0]       # 1  (zero-indexed)
echo $nums[-1]      # 5  (negative counts from end)
echo $nums.len      # 5  (or: arr_len $nums)

nums[0]  = 99       # set element
nums[]   = 6        # append
```

### Array pipe operations

| Command | Example | Result |
|---------|---------|--------|
| `first` | `$arr \| first` | first element |
| `last` | `$arr \| last` | last element |
| `nth` | `$arr \| nth 2` | element at index 2 |
| `slice` | `$arr \| slice 1 3` | elements 1 through 3 |
| `push` | `$arr \| push "x"` | copy with "x" appended |
| `pop` | `$arr \| pop` | copy without last element |
| `flatten` | `$arr \| flatten` | one-level flatten |
| `arr_len` | `$arr \| arr_len` | count of elements |
| `arr_sort` | `$arr \| arr_sort` | sorted copy |
| `arr_reverse` | `$arr \| arr_reverse` | reversed copy |
| `arr_uniq` | `$arr \| arr_uniq` | deduplicated copy |
| `arr_contains` | `$arr \| arr_contains "x"` | `true`/`false` |
| `arr_join` | `$arr \| arr_join ","` | joined string |
| `arr_map` | `$arr \| arr_map upper` | apply op to each element |
| `arr_filter` | `$arr \| arr_filter isnum` | keep matching elements |
| `arr_sum` | `$arr \| arr_sum` | numeric sum |
| `arr_min` | `$arr \| arr_min` | minimum value |
| `arr_max` | `$arr \| arr_max` | maximum value |
| `arr_avg` | `$arr \| arr_avg` | average value |

### Iterating over arrays

```sh
for name in $names {
    echo "Hello $name"
}

# With index via range
for i in range(0..$names.len) {
    echo "$i: $names[$i]"
}
```

---

## 7. Number Operations

### Pipe / command style

| Command | Example | Result |
|---------|---------|--------|
| `add` | `5 \| add 3` | `8` |
| `sub_n` | `10 \| sub_n 3` | `7` |
| `mul` | `4 \| mul 5` | `20` |
| `div` | `10 \| div 4` | `2.5` |
| `mod` | `10 \| mod 3` | `1` |
| `pow` | `2 \| pow 10` | `1024` |
| `abs` | `-5 \| abs` | `5` |
| `ceil` | `3.2 \| ceil` | `4` |
| `floor` | `3.8 \| floor` | `3` |
| `round` | `3.5 \| round` | `4` |
| `sqrt` | `16 \| sqrt` | `4` |
| `negate` | `5 \| negate` | `-5` |
| `hex` | `255 \| hex` | `ff` |
| `oct` | `8 \| oct` | `10` |
| `bin` | `10 \| bin` | `1010` |

### Type conversions

```sh
tonum "42"          # string → number
tostr 42            # number → string
toarray "a b c"     # string → array (splits on spaces)
typeof $x           # → "number" / "string" / "map" / "set" / ...
```

### Random numbers

```sh
random              # random integer
random 100          # random 0..99
random 10 50        # random 10..49
```

### bc — arbitrary precision

```sh
bc "scale=10; 22/7"     # 3.1428571428
bc "sqrt(2)"            # 1.4142135...
bc "2^64"               # 18446744073709551616
```

---

## 8. Pipe Operators `|`

Pipes transform structured `Result` objects (tables or text):

```sh
ls | select name size | where size > 1000 | sort size desc | limit 10
ps | where name ~ "node" | select pid name cpu mem
```

### Table transforms

| Operator | Usage | Description |
|----------|-------|-------------|
| `select` | `\| select col1 col2` | Keep only named columns |
| `where` | `\| where col > val` | Filter rows by condition |
| `where` | `\| where col ~ pattern` | Filter by regex match |
| `sort` | `\| sort col` | Sort ascending |
| `sort` | `\| sort col desc` | Sort descending |
| `limit` | `\| limit N` | Keep first N rows |
| `skip` | `\| skip N` | Skip first N rows |
| `head` | `\| head N` | Alias for `limit N` |
| `tail` | `\| tail N` | Keep last N rows |
| `grep` | `\| grep pattern` | Keep rows containing pattern |
| `count` | `\| count` | Count rows |
| `uniq` | `\| uniq col` | Deduplicate by column |
| `reverse` | `\| reverse` | Reverse row order |
| `fmt` | `\| fmt col=width` | Set column display width |
| `sum` | `\| sum col` | Sum a numeric column |
| `avg` | `\| avg col` | Average a numeric column |
| `min` | `\| min col` | Minimum of a column |
| `max` | `\| max col` | Maximum of a column |

### Where conditions

```sh
ps | where cpu > 5.0
ls | where size >= 1000000
ps | where name ~ "python"    # regex match
ls | where name !~ ".log"     # regex not-match
find | where name == "main.go"
```

### Storing pipe results in Box

```sh
ls | where name ~ ".go" #= go_files
ps | where cpu > 1.0    #= busy_procs
box get go_files
```

---

## 9. Pipe Expression `|>`

F#/Elixir-style value pipeline. Each stage receives the result of the previous:

```sh
# Basic chain
result = "  Hello World  " |> trim |> lower |> replace "world" "katsh"
# → "hello katsh"

# Numbers
x = 100 |> sqrt |> add 10 |> mul 2
# → 40

# Arrays
sorted_words = "fox the quick brown" |> split " " |> arr_sort |> arr_join ", "
# → "brown, fox, quick, the"

# Standalone (prints result)
"hello" |> upper |> reverse
# → OLLEH

# With user-defined functions
func double(n) { return $n * 2 }
result = 3 |> double |> double |> double
# → 24
```

---

## 10. Control Flow

### if / elif / else

```sh
# Inline
if $x > 10: echo "big"
if $x > 10: echo "big"; else: echo "small"

# Multi-branch
if $x > 100:
    echo "huge"
elif $x > 50:
    echo "large"
elif $x > 10:
    echo "medium"
else:
    echo "small"

# Block form
if $x > 10 {
    x -= 5
    echo "reduced: $x"
}
```

### unless

```sh
unless $ready: echo "not ready yet"
unless $x == 0 {
    echo "x is non-zero: $x"
}
```

### when (postfix guard)

```sh
echo "positive" when $n > 0
echo "even"     when $n % 2 == 0
rm tmp.txt      when -f tmp.txt
x++             when $x < 100
continue        when $item == ""
break           when $count >= 10
```

### match / case

```sh
match $status {
    "ok":      echo "success"
    "error":   echo "failed"
    "pending": echo "waiting"
    default:   echo "unknown: $status"
}

# Numeric comparisons
match $score {
    >=90: echo "A"
    >=80: echo "B"
    >=70: echo "C"
    *:    echo "F"
}

# Multiple values per case (pipe-separated)
match $day {
    "Sat" | "Sun": echo "weekend"
    *:             echo "weekday"
}
```

### switch

```sh
switch $color {
    "red":   echo "stop"
    "green": echo "go"
    "blue":  echo "slow down"
    default: echo "unknown"
}

# Fallthrough
switch $x {
    "a": echo "is a"; fallthrough
    "b": echo "is a or b"
    "c": echo "only c"
}
```

### for

```sh
# Over a literal list
for name in ["Alice", "Bob", "Carol"] {
    echo "Hello $name"
}

# Over a range
for i in range(1..10) {
    echo $i
}

# Over a variable array
for item in $myarray {
    echo $item
}

# Over command output (one line per iteration)
for file in `find -name "*.go"` {
    echo "processing $file"
}

# break / continue
for i in range(0..20) {
    continue when $i % 2 == 0    # skip even
    echo $i
    break when $i >= 9
}
```

### while

```sh
x = 0
while $x < 10 {
    x++
}

# Inline
while $running: sleep 1
```

### do-while / do-until

```sh
# Body always runs at least once
do {
    read -p "Enter password: " pw
} while $pw != "secret"

do {
    attempt++
    result = $(curl -sf $URL)
} until $result == "ok" or $attempt > 3
```

### repeat N

```sh
repeat 5 {
    echo "tick $_i"   # $_i = 0, 1, 2, 3, 4
}

repeat $count: echo "hello"
```

### try / catch / finally / throw

```sh
try {
    result = $(curl -sf https://api.example.com/data)
    if $result == "": throw "empty response"
    echo "got: $result"
} catch e {
    echo "Error: $e"
} finally {
    echo "request complete (always runs)"
}

# throw / raise (aliases)
throw "something went wrong"
raise "same as throw"

# Thrown message stored in $_error
try { throw "oops" } catch e { echo "caught: $e" }
```

### goto / label

```sh
func count_up() {
    x = 0
    label loop:
        x++
        echo $x
    goto loop when $x < 5
    echo "done"
}
count_up
# 1 2 3 4 5 done
```

Note: `goto` only works inside function or block bodies.

### && and ||

```sh
mkdir tmp && cd tmp
rm -rf dist && mkdir dist && echo "cleaned"
test -f config.json || echo "no config found"
grep -q "error" log.txt && echo "errors found!"
```

---

## 11. Functions

### Definition and calling

```sh
func greet(name) {
    println "Hello $name"
}
greet "Alice"
greet Alice          # quotes optional for single args

# Return a value
func square(n) {
    return $n * $n
}
result = square(7)   # → "49"
result = square 7    # same

# Return value also in $_return
square 7
echo $_return        # 49
```

### Multiple parameters

```sh
func add(a, b) {
    return $a + $b
}
echo $(add 3 7)      # 10

func clamp(val, lo, hi) {
    if $val < $lo: return $lo
    if $val > $hi: return $hi
    return $val
}
echo $(clamp 150 0 100)   # 100
```

### Extra / variadic arguments

```sh
func log(level) {
    # args beyond named params → $_args
    println "[$level] $_args"
}
log "INFO" "server started on port" 8080
# [INFO] server started on port 8080
```

### Local variables

```sh
func process(input) {
    local tmp = $input | upper | trim
    local count = $tmp | len
    return "$tmp ($count chars)"
}
```

### Defer

```sh
func with_cleanup() {
    defer echo "cleanup: done"
    defer rm -f /tmp/work.tmp

    echo "working..."
    # defers run LIFO when function exits
    # "cleanup: done" printed last even if error occurs
}
```

### Recursion

```sh
func factorial(n) {
    if $n <= 1: return 1
    prev = factorial($n - 1)
    return $n * $prev
}
echo $(factorial 10)    # 3628800

func fib(n) {
    if $n <= 1: return $n
    a = fib($n - 1)
    b = fib($n - 2)
    return $a + $b
}
echo $(fib 15)          # 610
```

### Exported functions (for import system)

```sh
func! public_utility(x) {    # ! = exported
    return $x | upper | trim
}
```

---

## 12. Advanced Data Types

### map — hash table

```sh
# Create
m = map { "name":"Alice"  "age":30  "city":"NYC" }
m = map { name=Alice age=30 city=NYC }   # = syntax

# Read
echo $m["name"]                  # Alice
val = map_get m "name"

# Write
map_set m "score" 100
map_set m "active" true

# Delete
map_del m "score"

# Check
has = map_has m "name"           # true / false

# Keys / values
keys = map_keys m                # sorted array of keys
vals = map_values m              # array of values (sorted by key)

# Merge (b overwrites a on conflict)
map_merge m extra                # merge extra into m in-place
map_merge m extra result         # store merged result in new var

# Inspect
map_len m                        # number of entries
map_show m                       # render as table
typeof m                         # "map"
```

### set — unique values with set algebra

```sh
s = set { 1 2 3 4 5 }
set_add    s 6
set_remove s 1
set_has    s 3               # true

a = set { 1 2 3 }
b = set { 2 3 4 }
set_union     a b r          # r = {1,2,3,4}
set_intersect a b r          # r = {2,3}
set_diff      a b r          # r = {1}   (in a but not b)

set_len  s
set_show s                   # render as table
```

### stack — LIFO

```sh
st = stack {}
stack_push st "task1"
stack_push st "task2"
stack_push st "task3"

top = stack_peek st          # "task3"  (no removal)
stack_pop st val             # val = "task3", removed
stack_pop st val             # val = "task2", removed
stack_len st                 # 1
stack_show st                # table (depth 0 = top)
```

### queue — FIFO

```sh
q = queue {}
enqueue q "job1"
enqueue q "job2"
enqueue q "job3"

front = queue_peek q         # "job1"
dequeue q item               # item = "job1", removed
dequeue q item               # item = "job2", removed
queue_len q                  # 1
queue_show q                 # table (position 0 = front)
```

### tuple — immutable ordered record

```sh
point = (10, 20)
color = (255, 128, 0, "orange")

x = tuple_get point 0        # 10
y = tuple_get point 1        # 20
n = tuple_get color 3        # "orange"

tuple_len color              # 4
tuple_show color             # table
```

### matrix — 2D numeric grid

```sh
# Create
A = matrix(3, 3)             # 3×3 zeros
I = matrix_identity 4        # 4×4 identity
B = matrix(2, 2, 1.0)        # 2×2 filled with 1.0
C = matrix(3, 3)

# Get / set individual cells
matrix_set A 0 0 5.0
matrix_set A 1 1 3.0
matrix_set A 2 2 7.0
val = matrix_get A 1 1       # "3"

# Operations
matrix_add A B result        # element-wise addition → result
matrix_mul A I result        # matrix multiplication → result
matrix_transpose A           # transpose in-place

# Stats
d = matrix_det A             # determinant (square matrices only)

# Inspect
matrix_show A                # render as table
typeof A                     # "matrix"
```

### typeof — inspect any value

```sh
typeof "hello"   # "string"
typeof 42        # "number"
typeof $m        # "map"
typeof $s        # "set"
typeof $st       # "stack"
typeof $q        # "queue"
typeof $t        # "tuple"
typeof $M        # "matrix"
typeof [1,2,3]   # "string"  (arrays are encoded strings)
```

### dt_show — display any data type as table

```sh
dt_show $m       # works for map, set, stack, queue, tuple, matrix
```

---

## 13. File & Directory Commands

### Navigation

```sh
cd /path/to/dir              # change directory
cd ..                        # up one level
cd -                         # previous directory
pwd                          # print working directory
pushd /tmp                   # push cwd, change to /tmp
popd                         # pop and return
dirs                         # show directory stack
```

### Listing

```sh
ls                           # list cwd
ls /path -la                 # long listing with hidden files
ll                           # ls -la shorthand
la                           # ls including hidden
tree                         # directory tree
tree -L 2 src/               # max depth 2
du                           # disk usage of cwd
du /path -d 1                # depth-limited usage
df                           # filesystem space
df -h                        # human-readable sizes
```

Piping listings:

```sh
ls | where name ~ ".go" | sort name
ls | sort size desc | head 10
find -name "*.go" | select name size | where size > 5000
```

### File CRUD

```sh
touch file.txt               # create / update timestamp
mkdir new_dir                # create directory
mkdir -p deep/nested/path    # create with parents
rm file.txt                  # delete file
rm -rf dir/                  # delete recursively
cp src.txt dst.txt           # copy file
cp -r srcdir/ dstdir/        # copy directory
mv old.txt new.txt           # move or rename
ln -s /target /link          # symbolic link
```

### Inspection

```sh
cat file.txt                 # show file contents
head file.txt                # first 10 lines
head -n 20 file.txt
tail file.txt                # last 10 lines
tail -n 20 file.txt
tail -f /var/log/app.log     # follow (passthrough mode)

wc file.txt                  # lines / words / bytes
wc -l file.txt               # lines only
stat file.txt                # metadata (size, dates, permissions)
file unknown.bin             # detect file type
readlink -f symlink          # resolve symlink chain
realpath relative/path       # absolute path
basename /a/b/c.txt          # c.txt
dirname  /a/b/c.txt          # /a/b
```

### find

```sh
find                         # all files recursively
find -name "*.go"            # by name pattern
find -name "*.go" -type f    # files only
find -type d                 # directories only
find -maxdepth 2             # limit depth
find -size +1M               # larger than 1 MB
find -newer reference.txt    # modified after reference
find -name "*.go" | select name size | sort size desc
```

### Archives

```sh
tar -czf archive.tar.gz dir/ # create gzip archive
tar -xzf archive.tar.gz      # extract
tar -tzf archive.tar.gz      # list contents
tar -xzf archive.tar.gz -C /target/  # extract to dir
gzip   file.txt              # compress → file.txt.gz
gunzip file.txt.gz           # decompress
zip -r archive.zip dir/
unzip archive.zip
unzip -l archive.zip         # list contents
```

### Hashing

```sh
md5sum    file.txt
sha1sum   file.txt
sha256sum file.txt
md5    "string value"        # hash a string directly
sha1   "string value"
sha256 "string value"
```

### Permissions

```sh
chmod 755 script.sh
chmod +x  script.sh
chmod -R 644 dir/
chown user:group file.txt
chown -R user:group dir/
```

---

## 14. Text Processing

### grep

```sh
grep "pattern"  file.txt
grep -i "case"  file.txt     # case-insensitive
grep -r "pat"   dir/         # recursive
grep -n "pat"   file.txt     # show line numbers
grep -v "pat"   file.txt     # invert — exclude matching
grep -c "pat"   file.txt     # count matches
grep -l "pat"   *.txt        # list files with matches
grep -E "a|b"   file.txt     # extended regex
grep -w "word"  file.txt     # whole word only
```

### sed

```sh
sed "s/old/new/g"   file.txt    # replace all occurrences
sed "s/old/new/"    file.txt    # replace first per line
sed -n "10,20p"     file.txt    # print lines 10-20
sed "/pattern/d"    file.txt    # delete matching lines
sed -i "s/a/b/g"    file.txt    # in-place edit
```

### awk

```sh
awk "{print $1}"              file.txt   # first field
awk -F: "{print $1, $3}"      /etc/passwd
awk "NR==1"                   file.txt   # first line only
awk "NR>=10 && NR<=20"        file.txt   # lines 10-20
awk "{sum+=$1} END{print sum}" nums.txt  # sum first column
awk "length($0) > 80"         file.txt   # lines > 80 chars
```

### cut / tr

```sh
cut -d: -f1         /etc/passwd    # field 1
cut -d, -f2,4       data.csv       # fields 2 and 4
cut -c1-10          file.txt       # chars 1-10
tr 'a-z' 'A-Z'                    # lowercase to upper
tr -d '\n'                         # delete newlines
tr -s ' '                          # squeeze spaces
```

### sort / uniq

```sh
sort file.txt                  # alphabetical
sort -n file.txt               # numeric
sort -r file.txt               # reverse
sort -k2 -t, file.csv          # by column 2, comma-separated
sort -u file.txt               # sort and deduplicate
uniq file.txt                  # remove consecutive duplicates
uniq -c file.txt               # count occurrences
uniq -d file.txt               # only show duplicates
uniq -u file.txt               # only show unique lines
```

### Other text commands

```sh
wc -l file.txt                 # count lines
wc -w file.txt                 # count words
wc -c file.txt                 # count bytes
rev file.txt                   # reverse each line
column -t file.txt             # align to columns
column -t -s, data.csv         # CSV to aligned table
nl file.txt                    # number lines
fold -w 80 file.txt            # wrap at 80 chars
shuf file.txt                  # shuffle lines randomly
paste file1.txt file2.txt      # merge files side by side
comm file1.txt file2.txt       # compare sorted files
diff old.txt new.txt           # show differences
diff -u old.txt new.txt        # unified diff format
diff -r dir1/ dir2/            # recursive diff
```

---

## 15. Process & System Commands

### Processes

```sh
ps                             # all processes (table)
ps | where name ~ "python"     # filter by name
ps | sort cpu desc | head 5    # top 5 CPU users
ps | sort mem desc | limit 10  # top 10 memory users

kill 1234                      # SIGTERM by PID
kill -9 1234                   # SIGKILL
kill -HUP 1234                 # SIGHUP (reload)
pgrep nginx                    # find PID by name
pkill nginx                    # kill by name

sleep 2                        # sleep 2 seconds
sleep 0.5                      # sleep 500ms

jobs                           # list background jobs
bg %1                          # resume job 1 in background
fg %1                          # bring job 1 to foreground
nohup cmd &                    # run immune to hangup
nice -n 10 cmd                 # lower priority
timeout 30 cmd                 # kill after 30s
```

### System info

```sh
uname -a                       # kernel info
uptime                         # uptime + load average
date                           # current date/time
date "+%Y-%m-%d %H:%M:%S"      # custom format
cal                            # calendar
cal 2025                       # full year calendar
hostname                       # system hostname
whoami                         # current username
id                             # user/group IDs
groups                         # current user's groups
who                            # logged-in users
w                              # who + what they're running
free -h                        # memory usage (human)
lscpu                          # CPU info
```

### Disk

```sh
df -h                          # filesystem space (human)
du -sh *                       # size of each item in cwd
du -sh /var --max-depth=1      # depth-limited
lsblk                          # block device tree
fdisk -l                       # partition table
blkid                          # block device UUIDs
mount                          # mounted filesystems
umount /mnt/usb                # unmount
```

### Services / logs

```sh
systemctl status nginx
systemctl start  nginx
systemctl stop   nginx
systemctl restart nginx
systemctl enable nginx

journalctl -f                  # follow system log (passthrough)
journalctl -n 100              # last 100 log entries
journalctl -u nginx            # logs for specific unit
journalctl --since "1 hour ago"

service apache2 status
service apache2 restart
```

---

## 16. Networking Commands

```sh
ping google.com                # ping (passthrough — Ctrl-C to stop)
ping -c 4 google.com           # only 4 pings

curl https://api.example.com   # GET request
curl -s  https://...           # silent
curl -o  file.zip https://...  # save to file
curl -H "Authorization: Bearer $TOKEN" https://...
curl -X POST -d '{"key":"val"}' https://...

wget https://example.com/file.zip
wget -O output.txt https://...
wget -q  https://...           # quiet

nslookup google.com
dig google.com
dig +short google.com A        # just the IP

ifconfig                       # interfaces (passthrough)
ip addr                        # addresses
ip route                       # routing table
ss -tlnp                       # TCP listening ports
netstat -tlnp                  # same, classic syntax
traceroute google.com          # (passthrough)
mtr google.com                 # combined trace (passthrough)

ssh user@host                  # SSH session (auto-passthrough)
ssh -p 2222 user@host
scp file.txt user@host:/path/
rsync -avz src/ user@host:dest/

openssl s_client -connect api.example.com:443
```

### HTTP builtins

```sh
# GET with auto-parsed JSON response
httpget https://api.example.com/users

# POST with JSON body
httppost https://api.example.com/users '{"name":"Alice","email":"a@b.com"}'

# jq JSON processor
httpget https://api.example.com/users | jq '.[0].name'
cat data.json | jq '.items[] | select(.active == true)'
```

---

## 17. Box — Session Storage

Box is an in-memory key-value store that persists for the duration of the session.

### Storing results

```sh
# Store pipeline results with #= operator
ls | where name ~ ".go" #= go_files
ps | where cpu > 1.0    #= busy_procs
find -name "*.log"      #= log_files

# Store a value directly
box set greeting "Hello World"
box set counter 0
```

### Reading

```sh
box get go_files         # retrieve stored result
box get greeting         # retrieve stored value
box                      # list all keys
```

### Manipulation

```sh
# All pipe operators work on retrieved results
box get go_files | sort name | limit 5
box get busy_procs | where name ~ "node"

# Tag entries with metadata
box tag go_files language go type source
box tag busy_procs category processes

# Rename
box rename go_files source_files

# Delete
box rm greeting
box clear                # remove all entries
```

### Persistence

```sh
box export backup.json   # save all entries to JSON
box import backup.json   # load entries from JSON
```

### Hashing stored values

```sh
box md5    data_key
box sha1   data_key
box sha256 data_key
```

---

## 18. Aliases

```sh
# Define
alias ll   = "ls -la"
alias gs   = "git status"
alias gc   = "git commit -m"
alias gp   = "git push"
alias py   = "python3"
alias k    = "kubectl"

# Use (arguments pass through)
ll
gs
gc "fix: improve performance"
k get pods -n production

# List all aliases
aliases

# Remove
unalias ll
unalias gs
```

---

## 19. History

```sh
history                  # show all history (table: index, command, exit_code)
history | tail 20        # last 20 commands
history | grep "git"     # search history
history | where exit_code != 0   # only failed commands
history | where command ~ "curl" # commands containing curl
```

History is saved to `~/.config/structsh/history.json` and loaded on startup.

---

## 20. Import / Export / Modules

### Import a script

```sh
import "./utils.ksh"                   # local file
import "github_user/repo/lib.ksh"      # GitHub (fetched, 24h cached)
import "https://example.com/lib.ksh"   # direct URL
```

### Export a variable to environment

```sh
export PATH     = "$PATH:/usr/local/bin"
export EDITOR   = "nvim"
export NODE_ENV = "production"
```

### Export a function (for the import system)

```sh
# The ! suffix marks a function as exported so importers can use it
func! str_slug(s) {
    return $s | lower | replace " " "-"
}
```

---

## 21. Error Reference

### Error display format

```
  TypeError[E003]  cannot apply `+=` to "hello" — not a number
  ╭─ line 12 in myscript.ksh
  │  total += "hello"
  │        ~~~~~~~~~~~
  │        ^── here
  │
  ╰─ 💡  hint: Use a numeric value. Try: tonum "hello"
     fix:  total += 42
```

### Error code table

| Code | Name | When it occurs | How to fix |
|------|------|----------------|------------|
| `E001` | `UnknownCommand` | Command not found in builtins or PATH | Check spelling · run `which cmd` |
| `E002` | `SyntaxError` | Parse failure | Check colons after `if`/`for` · match braces/brackets |
| `E003` | `TypeError` | Wrong type for operation (e.g. `+= "text"`) | Use `tonum` / `tostr` · check variable values |
| `E004` | `RuntimeError` | `throw`/`raise` or goto limit exceeded | Wrap in `try/catch` · check loop termination |
| `E005` | `ReadonlyError` | Assigning to a `readonly` variable | Use a different variable name |
| `E006` | `DivisionByZero` | Divide or modulo by zero | Check denominator before dividing |
| `E007` | `ArgumentError` | Wrong number of arguments to a function | Check function signature with `help` |
| `E010` | `FileNotFound` | Script file does not exist | Check path · run `ls dirname` |

### Common mistakes and fixes

```sh
# ✗ Missing colon after if condition
if $x > 10 echo "big"
# ✓ Add colon
if $x > 10: echo "big"

# ✗ = instead of == in condition
if $name = "Alice": echo "hi"
# ✓
if $name == "Alice": echo "hi"

# ✗ No spaces around operators
if $x>10: ...       # parse error
# ✓
if $x > 10: ...

# ✗ Forgetting $ on variable reference
echo name           # prints literal "name"
echo $name          # prints the value

# ✗ Complex bash pipeline without passthrough prefix
ls | sort -k5 -n | awk '{print $9}'   # awk not in PATH? or complex flags
# ✓ Use passthrough
bash! ls -la | sort -k5 -n | awk '{print $9}'

# ✗ $() not expanding inside single quotes
echo '$(whoami)'    # literal $(whoami)
echo "$(whoami)"    # ✓ expands inside double quotes
x = $(whoami)       # ✓ expands in assignment

# ✗ Array index out of bounds (silent — returns "")
arr = [1, 2, 3]
echo $arr[99]       # returns ""
# ✓ Check length first
if $arr.len > 5: echo $arr[5]

# ✗ goto at top level
goto somewhere      # Error: goto only works inside function/block body
# ✓ Wrap in a function

# ✗ Uncaught throw
throw "oops"        # exits with code 1
# ✓ Wrap in try/catch
try { throw "oops" } catch e { echo "handled: $e" }

# ✗ Trying to catch a shell exit code
bash! exit 1
# ✓ Check the return code
run exit 1
# shell exit codes don't integrate with try/catch (use run/shell for those)

# ✗ Using set as a variable (conflicts with `set` builtin)
set = [1, 2, 3]        # confusing — `set` is a builtin
# ✓ Use a different name
items = [1, 2, 3]

# ✗ Matrix dimension mismatch in multiplication
A = matrix(2, 3)
B = matrix(2, 3)
matrix_mul A B r   # error — cols(A)=3 must equal rows(B)=2
# ✓
B = matrix(3, 2)
matrix_mul A B r   # 2x3 × 3x2 = 2x2 ✓
```

---

## 22. Syntax Highlighting & Completion

### Colour scheme

| Colour | Meaning | Examples |
|--------|---------|---------|
| **Bold green** | Known katsh builtin | `ls` `echo` `map_get` `arr_sort` |
| **Green** | External command in PATH | `git` `docker` `make` |
| **Red + underline** | Unknown command (typo?) | `gti` `ecoh` |
| **Bold magenta** | Scripting keyword | `if` `for` `func` `match` `defer` |
| **Cyan** | Quoted string | `"hello"` `'world'` |
| **Yellow** | Numeric literal | `42` `3.14` |
| **Grey** | Operator / syntax | `\|` `=` `+` `{` `}` |

### Tab completion

| Context | What's completed |
|---------|-----------------|
| Start of line | All builtin names · alias names |
| After `$` | Variable names |
| After a space (argument) | Files and directories in cwd |
| After `box get/rm` | Box key names |

### Typo detection

When you type an unknown command, katsh finds the closest match:

```
  UnknownCommand[E001] unknown command "gti"
  ╰─ 💡 did you mean: git?
```

---

## 23. Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `←` / `→` | Move cursor left / right |
| `Ctrl+A` | Jump to start of line |
| `Ctrl+E` | Jump to end of line |
| `Ctrl+W` | Delete word to the left |
| `Ctrl+U` | Delete from cursor to start |
| `Ctrl+K` | Delete from cursor to end |
| `Ctrl+L` | Clear screen (keep current line) |
| `Ctrl+C` | Cancel current input |
| `Ctrl+D` | Exit shell (on empty line) |
| `↑` / `↓` | Browse command history |
| `Tab` | Complete command / path / variable |
| `Tab Tab` | Show all available completions |
| Paste | Bracketed-paste mode — multi-line paste works correctly |

---

## 24. Complete Command Index

### Navigation & filesystem
`cd` `pwd` `pushd` `popd` `dirs` `ls` `ll` `la` `tree` `du` `df`  
`cat` `head` `tail` `touch` `mkdir` `rmdir` `rm` `cp` `mv` `ln`  
`readlink` `realpath` `basename` `dirname` `mktemp` `mkfifo`  
`wc` `stat` `file` `find` `diff` `chmod` `chown`

### Text processing
`grep` `sed` `awk` `cut` `tr` `sort` `uniq` `tee` `split` `xargs`  
`nl` `fold` `expand` `unexpand` `column` `paste` `join` `comm` `shuf` `numfmt`  
`rev` `strings` `xxd` `od`

### Process & system
`ps` `kill` `sleep` `jobs` `bg` `fg` `nice` `timeout` `nohup`  
`pgrep` `pkill` `top` `lsof` `vmstat` `iostat`  
`uname` `uptime` `date` `cal` `hostname` `whoami` `id` `groups` `who` `w`  
`free` `lscpu` `lsusb` `lspci` `dmesg` `lsblk` `mount` `umount` `fdisk` `blkid`  
`journalctl` `systemctl` `service`

### Networking
`ping` `curl` `wget` `nslookup` `dig` `ifconfig` `ip`  
`ss` `netstat` `traceroute` `mtr` `openssl` `ssh` `scp` `rsync`  
`httpget` `httppost` `jq`

### Hashing & archives
`md5sum` `md5` `sha1sum` `sha1` `sha256sum` `sha256`  
`tar` `gzip` `gunzip` `zip` `unzip`

### Variables & environment
`set` `unset` `vars` `export` `import` `env` `printenv` `readonly`  
`declare` `typeset` `getopts` `eval` `exec` `test` `[`  
`read` `mapfile` `readarray` `source` `.`

### Shell passthrough
`run` `shell` `capture` `bash` `zsh` `sh` `fish` `ksh`  
Syntax: `bash!` `zsh!` `sh!` `!` (prefix)

### Scripting keywords
`if` `elif` `else` `fi` `for` `while` `do` `done` `until` `in` `range`  
`func` `return` `local` `break` `continue` `pass`  
`match` `case` `default` `switch` `fallthrough` `unless` `when`  
`try` `catch` `finally` `throw` `raise`  
`repeat` `enum` `struct` `defer` `with` `goto` `label`  
`and` `or` `not` `true` `false` `null` `nil` `readonly`  
`print` `println`

### Session
`box` `history` `alias` `unalias` `aliases` `which` `type`  
`help` `man` `watch` `clear` `exit` `quit`  
`echo` `printf` `yes` `seq` `base64` `bc` `factor` `random` `figlet` `notify`

### String operations (pipe + command)
`upper` `lower` `title` `trim` `ltrim` `rtrim` `strip`  
`len` `reverse` `repeat` `replace` `replace1` `sub` `sub_n` `pad` `lpad` `center`  
`startswith` `endswith` `contains` `isnum` `isalpha` `isalnum` `isspace` `isupper` `islower`  
`lines` `words` `chars` `split` `join` `concat` `prepend`

### Array operations
`first` `last` `nth` `slice` `push` `pop` `flatten`  
`arr_uniq` `arr_sort` `arr_reverse` `arr_len` `arr_join`  
`arr_contains` `arr_map` `arr_filter` `arr_sum` `arr_min` `arr_max` `arr_avg`

### Number operations
`add` `mul` `div` `mod` `pow` `sub_n`  
`abs` `ceil` `floor` `round` `sqrt` `negate`  
`hex` `oct` `bin` `tonum` `tostr` `toarray` `typeof`

### Map commands
`map_new` `map_set` `map_get` `map_del` `map_has`  
`map_keys` `map_values` `map_len` `map_show` `map_merge`

### Set commands
`set_new` `set_add` `set_remove` `set_has`  
`set_union` `set_intersect` `set_diff` `set_show` `set_len`

### Stack commands
`stack_new` `stack_push` `stack_pop` `stack_peek` `stack_len` `stack_show`

### Queue commands
`queue_new` `enqueue` `dequeue` `queue_peek` `queue_len` `queue_show`

### Tuple commands
`tuple_get` `tuple_len` `tuple_show`

### Matrix commands
`matrix_new` `matrix_get` `matrix_set` `matrix_add` `matrix_mul`  
`matrix_transpose` `matrix_det` `matrix_show` `matrix_identity`

### Type inspection
`typeof` `dt_show`

---

*KatSH · StructSH engine · 19 Go files · ~12 500 lines*
