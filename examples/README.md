# Katsh Scripting Examples

A comprehensive collection of example scripts to learn Katsh scripting language.

## Quick Start

Run any script:
```bash
./katsh examples/01_variables.ksh
```

## Available Examples

| File | Description | What You'll Learn |
|------|-------------|-------------------|
| [`01_variables.ksh`](01_variables.ksh) | Variables & Basic Operations | Variables, types, arithmetic, interpolation |
| [`02_strings.ksh`](02_strings.ksh) | String Operations | Case, split, join, trim, transform |
| [`03_arrays.ksh`](03_arrays.ksh) | Array Operations | Arrays, access, iteration, map/filter |
| [`04_control_flow.ksh`](04_control_flow.ksh) | Control Flow | If/elif/else, match/case, loops |
| [`05_functions.ksh`](05_functions.ksh) | Functions & Advanced | Functions, return, recursion, import |
| [`06_all_features.ksh`](06_all_features.ksh) | Complete Demo | All features combined in one script |

## Learning Path

### 1. Start with Variables
Learn the basics of storing and manipulating data:
- String, number, and boolean variables
- Arithmetic operations (+, -, *, /, %, **)
- String interpolation
- Subshell capture

### 2. Master Strings
Strings are fundamental - learn to manipulate them:
- Case conversion (upper, lower, title)
- Trimming and padding
- Split and join
- Search and replace

### 3. Work with Arrays
Handle collections of data:
- Create and access arrays
- Array methods (sort, reverse, unique)
- Iteration with for loops
- Map and filter operations

### 4. Control Flow
Make your scripts smart:
- Conditional statements (if/elif/else)
- Pattern matching (match/case)
- Loops (for, while, repeat)
- Try/catch error handling

### 5. Functions
Create reusable code:
- Define and call functions
- Return values
- Variable arguments
- Recursive functions

### 6. Complete Example
See everything working together in one comprehensive script.

## Running Scripts

### From Command Line
```bash
./katsh examples/01_variables.ksh
```

### With Arguments
```bash
./katsh examples/05_functions.ksh arg1 arg2
```

### With Flags
```bash
./katsh -x examples/01_variables.ksh   # Trace mode
./katsh -e examples/01_variables.ksh   # Exit on error
./katsh -n examples/01_variables.ksh   # Dry run (parse only)
```

### As Executable
Add shebang and make executable:
```bash
#!/usr/bin/env katsh
echo "Hello"
```

```bash
chmod +x examples/01_variables.ksh
./examples/01_variables.ksh
```

### Inside Katsh REPL
```bash
./katsh
# Then type:
import ./examples/01_variables.ksh
```

## Key Concepts Summary

### Variables
```ksh
name = "Alice"
age = 25
```

### Strings
```ksh
greeting = "Hello, $name!"
uppercase = "hello" | upper
```

### Arrays
```ksh
fruits = ["apple", "banana"]
first = $fruits[0]
```

### Loops
```ksh
for i in range(0, 10):
    echo $i
```

### Functions
```ksh
func add(a, b) {
    return $a + $b
}
```

### Match/Case
```ksh
match $value {
    case 1: echo "one"
    case *: echo "other"
}
```

## Need More Help?

- Check [`README.md`](../README.md) for full documentation
- Run `help` inside the Katsh REPL
- Check [`HELP.md`](../HELP.md) for complete reference

## License

MIT License - Feel free to use these examples in your own projects!
