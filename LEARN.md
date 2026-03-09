# KatSH Feature Reference

---

## Variables & Assignment

### Simple Assignment
Assign any value — string, number, or expression — to a named variable.
```sh
name = "Alice"
x = 42
```

### Compound Operators
In-place math with `+=` `-=` `*=` `/=` `%=` `**=`.
```sh
x += 10
score **= 2
```

### Increment / Decrement
Shorthand `++` and `--` operators, prefix or suffix form.
```sh
x++
--count
```

### Multi-Assign
Assign multiple variables in one line from a list or array.
```sh
a, b, c = 1, 2, 3
first, rest... = $arr
```

### Readonly Variables
Declare a variable that cannot be reassigned; violations raise E008.
```sh
readonly PI = 3.14159
PI = 0  # → ReadonlyError
```

### Null Coalescing `??`
Use a fallback value when a variable is empty, null, or undefined.
```sh
port = $PORT ?? 8080
name = $username ?? "guest"
```

---

## Strings

### Interpolation
Embed variables inside double-quoted strings with `$var` or `${var}`.
```sh
echo "Hello $name"
echo "Count: ${#items}"
```

### String Repeat `* N`
Repeat a string N times using the `*` operator.
```sh
line = "-" * 40
echo "ha" * 3   # hahaha
```

### Multiline Strings `""" """`
Triple-quoted string literals that preserve newlines.
```sh
msg = """
  line one
  line two
"""
```

### String Ops via `|>`
Chain string transformations: `upper`, `lower`, `trim`, `split`, `reverse`, etc.
```sh
result = "  hello  " |> trim |> upper
words  = "a,b,c" |> split ","
```

---

## Numbers & Math

### Arithmetic
Standard `+` `-` `*` `/` `%` and power `**` with correct precedence.
```sh
area = $w * $h
hyp  = ($a ** 2 + $b ** 2) ** 0.5
```

### Numeric Helpers
Built-in `abs`, `sign`, `clamp`, `round`, `floor`, `ceil`, `sqrt`, `log`, `pow`.
```sh
abs -7          # 7
clamp $n 0 100
```

### Base Conversion
Convert integers to hex, octal, or binary representations.
```sh
hex 255         # ff
bin 10          # 1010
```

### Type Casts
Cast values with `int()`, `float()`, `str()`, `bool()`, `len()`.
```sh
n = int("42")
s = str($pi)
```

---

## Control Flow

### if / elif / else
Standard conditional branching; supports inline and block forms.
```sh
if $x > 0: echo "positive"
if $x > 10 { echo "big" } elif $x > 0 { echo "small" } else { echo "zero" }
```

### unless
Executes the body only when the condition is **false**.
```sh
unless $logged_in: echo "please log in"
unless -f config.json { echo "missing config" }
```

### Ternary Expression `? :`
Inline conditional expression returning one of two values.
```sh
label = $score >= 90 ? "A" : "B"
msg   = ($x > 0) ? "positive" : "non-positive"
```

### match / case
Pattern matching with wildcards, operators, and `|` alternatives.
```sh
match $status {
  200: echo "OK"
  404 | 410: echo "not found"
  >= 500: echo "server error"
  default: echo "other"
}
```

### switch
Strict equality switch with optional `fallthrough` support.
```sh
switch $lang {
  "go":     echo "gopher"
  "python": echo "snake"
  default:  echo "unknown"
}
```

### when guard
Suffix a command with a condition; runs only if the condition is true.
```sh
echo "big number" when $n > 1000
x++ when $x < 10
```

---

## Loops

### for … in
Iterate over arrays, ranges, command output, or space-separated lists.
```sh
for item in $list: echo $item
for i in range 1 10: echo $i
```

### while
Loop while a condition is true; `break` and `continue` are supported.
```sh
while $x < 100 { x *= 2 }
while -f lock.pid { sleep 1 }
```

### do … while / until
Execute the body at least once, then check the condition.
```sh
do { x++ } while $x < 10
do { read line } until $line == "quit"
```

### repeat N
Run a block exactly N times; `_i` holds the current index.
```sh
repeat 5 { echo "tick $_i" }
repeat $count { process_item }
```

### loop / forever *(new)*
Infinite loop — use `break` or `break when cond` to exit. `_i` counts iterations.
```sh
loop { x++; break when $x >= 10 }
forever { poll; sleep 1; break when $done == "true" }
```

---

## Functions

### func definition
Define a named function with optional parameters and a return value.
```sh
func greet(name) { echo "Hello $name" }
func add(a, b) { return $a + $b }
```

### Function call & return
Call a function and capture its return value in a variable.
```sh
greet "Alice"
result = add(3, 4)
```

### Exported functions `!`
Mark a function as exported; visible in subshells.
```sh
func deploy!() { echo "deploying..." }
```

### defer
Schedule a command to run when the enclosing function exits.
```sh
func process() {
  defer echo "cleanup done"
  echo "working..."
}
```

### with scoped binding
Bind a variable for the duration of a block, then restore the old value.
```sh
with x = 99 { echo "x is $x" }
with tmp = $(mktemp) { process $tmp }
```

---

## Error Handling

### try / catch / finally
Catch thrown errors; `finally` always runs.
```sh
try { risky_cmd } catch e { echo "caught: $e" }
try { open $f } catch err { echo $err } finally { cleanup }
```

### throw / raise
Throw a custom error message; propagates to enclosing `try`.
```sh
throw "file not found"
raise $errMsg
```

### assert
Fail with a clear message if a condition is false.
```sh
assert $x > 0 "x must be positive"
assert -f config.json "config file required"
```

---

## Arrays

### Array literals & indexing
Create arrays with `[a, b, c]`; access with `$arr[N]` (negative indices work).
```sh
colors = ["red", "green", "blue"]
echo $colors[0]   # red
echo $colors[-1]  # blue
```

### arr_push / arr_pop
Append to or remove from the end of an array in-place.
```sh
arr_push $stack "item"
last = arr_pop $stack
```

### arr_shift / arr_unshift
Remove from or prepend to the front of an array.
```sh
first = arr_shift $queue
arr_unshift $list "new_first"
```

### arr_sort / arr_reverse / arr_unique
Sort (numeric or lexicographic), reverse, or deduplicate an array.
```sh
sorted = arr_sort $nums
deduped = arr_unique $tags
```

### arr_filter
Keep only items matching an inline condition using `$_` or `$_item`.
```sh
big = arr_filter $nums "$_ > 10"
strs = arr_filter $data "$_ is not empty"
```

### arr_map
Transform each element with an expression; result available as `$_`.
```sh
doubled = arr_map $nums "$_ * 2"
upper   = arr_map $names "upper $_"
```

### arr_find / arr_contains
Find the first matching item or check membership.
```sh
found = arr_find $list "$_ starts with A"
arr_contains $tags "urgent"   # true/false
```

### arr_sum / arr_min / arr_max / arr_len
Numeric aggregates and length for numeric arrays.
```sh
total = arr_sum $prices
echo arr_len $items
```

### arr_join / arr_flatten / arr_zip / arr_chunk
Combine, flatten nested, zip two arrays, or split into fixed-size chunks.
```sh
csv = arr_join $fields ","
pairs = arr_zip $keys $vals
```

### `|?` inline filter *(new)*
Filter an array in a single expression using `$_` as the element.
```sh
evens  = $nums |? $_ % 2 == 0
starts = $names |? $_ starts with "A"
```

---

## Maps

### Map literals & access
Create maps with `map{ key=val }` and access with `$m["key"]`.
```sh
config = map{ host="localhost" port=5432 }
echo $config["host"]
```

### map_keys / map_values / map_size
Retrieve all keys, all values, or the number of entries.
```sh
keys = map_keys $config
map_size $config     # 2
```

### map_has / map_del
Check key existence or remove a key in-place.
```sh
map_has $config "host"   # true
map_del $config "port"
```

### map_merge / map_entries
Merge two maps (second wins on conflicts) or get key-value pairs.
```sh
merged  = map_merge $defaults $overrides
entries = map_entries $config
```

---

## Advanced Data Types

### set
Unordered collection of unique values with membership operations.
```sh
s = set{"a" "b" "c"}
set_add $s "d"
```

### stack
LIFO collection with `stack_push` / `stack_pop`.
```sh
stk = stack{}
stack_push $stk "item"
top = stack_pop $stk
```

### queue
FIFO collection with `queue_enqueue` / `queue_dequeue`.
```sh
q = queue{}
queue_enqueue $q "task"
job = queue_dequeue $q
```

### tuple
Immutable ordered collection; access by numeric index.
```sh
point = tuple(10, 20)
echo $point[0]   # 10
```

### matrix
2D numeric grid with row/column access.
```sh
m = matrix(3, 3)
echo $m[0][1]
```

---

## Range

### range
Generate a numeric sequence as an array; supports start, end, and step.
```sh
for i in range 1 11: echo $i
r = range 0 100 5   # [0,5,10,...,95]
```

---

## Type System

### typeof / kindof
Print the runtime type of any value: number, string, array, map, bool, null…
```sh
typeof $x
kindof $config   # map
```

### isnull / isnil / isnone
Check whether a value is empty or a null sentinel.
```sh
isnull $result
isnil $optional
```

### Comparisons: `in`, `not in`
Test set membership against an array or space-separated list.
```sh
if $role in ["admin","mod"]: allow
if $cmd not in $blocked: run
```

### Comparisons: `starts with`, `ends with`, `contains`
Natural-language string tests usable in any condition.
```sh
if $path starts with "/home": echo "user dir"
if $email contains "@": validate
```

### Comparisons: `is empty`, `is not empty`
Test whether a value is blank or non-blank.
```sh
if $name is empty: error "name required"
if $list is not empty: process
```

---

## Shell Integration

### Command execution
Run any system command directly; katsh falls back to PATH lookup.
```sh
ls -la
git status
```

### Capture output `$(...)`
Capture stdout of a command or subshell into a variable.
```sh
now = $(date +%s)
files = $(ls *.go)
```

### Background commands `?(...)`
Run a command in the background; `?()` waits and returns output.
```sh
result = ?(curl https://api.example.com/data)
?(heavy_job) &
```

### Shell passthrough `bash!` / `sh!`
Drop into raw bash or sh for a single command, bypassing katsh parsing.
```sh
bash! for f in *.txt; do echo $f; done
sh! export VAR=value
```

### Pipes `|`
Pipe between system commands and katsh's built-in pipe ops.
```sh
cat log.txt | grep "ERROR" | count
ls | upper | sort
```

### `| field N` *(new)*
Select the Nth whitespace-separated token from a string (1-based).
```sh
user   = $(whoami) | field 1
second = "one two three" | field 2   # two
```

### `&&` / `||`
Run the right side only on success (`&&`) or failure (`||`).
```sh
build && deploy
test || echo "tests failed"
```

---

## Output & Formatting

### print / println
Print a value with `  ` indentation; `println` adds a newline.
```sh
print "loading..."
println "done"
```

### format *(new)*
Printf-style formatted output or assignment using `%s %d %f %x` verbs.
```sh
format "Hello %s, age %d" $name $age
msg = format "%.2f%%" $pct
```

### box / table display
Render a value in a styled box (used automatically for data types).
```sh
box_show $config
table_show $results
```

---

## OOP-lite

### struct
Define a named record type with optional default field values.
```sh
struct Point { x y }
struct User { name age="unknown" }
```

### Struct instantiation
Create an instance; fields become `varname_field` variables.
```sh
p = Point(10, 20)
echo $p_x   # 10
```

### enum
Define named integer constants; members stored as `EnumName_Member`.
```sh
enum Color { Red Green Blue }
echo $Color_Red   # 0
```

---

## Pipe Expressions `|>`

### Chained transforms
Apply a sequence of string/number operations without intermediate variables.
```sh
result = "  Hello World  " |> trim |> lower |> split " "
total  = 42 |> add 8 |> mul 2
```

### User functions as pipe stages
Any user-defined function can act as a pipe stage.
```sh
func double(x) { return $x * 2 }
result = 5 |> double |> double   # 20
```

---

## goto / label

### goto + label
Jump to a named label inside a function body; supports conditional form.
```sh
func count() {
  label top:
  echo $_i; _i++
  goto top when $_i < 5
}
```

---

## Misc

### pass
No-op placeholder; useful inside empty branches or stubs.
```sh
if $debug: pass
func stub() { pass }
```

### local
Declare a variable as local to the current block (prevents global leak).
```sh
func process() { local tmp = "working" }
```

### Zip two arrays
Combine two arrays element-wise into an array of pairs.
```sh
pairs = zip $keys $values
```

### String pad / center / mask
Pad strings left/right, center them, or mask sensitive content.
```sh
padded  = $label | pad_right 20
masked  = $token | mask
```