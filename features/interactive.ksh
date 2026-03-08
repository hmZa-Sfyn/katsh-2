#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  38_interactive.ksh — Interactive scripts: menus, prompts, spinners
#
#  Topics covered:
#    read for user input · confirmation prompts (y/n)
#    numbered menus · password input masking
#    spinner/progress animation · form filling
#    multi-step wizards · input validation loops
#
#  Note: interactive features require a real terminal.
#        This script demonstrates the PATTERNS — prompts are
#        simulated with preset values when run non-interactively.
# ─────────────────────────────────────────────────────────────────────

# ── Detect if running interactively ──────────────────────────────────
INTERACTIVE = $(test -t 0 && echo "true" || echo "false" 2>/dev/null)
if $INTERACTIVE == "": INTERACTIVE = false
println "Running interactively: $INTERACTIVE"
println ""

# ── Simulated input for non-interactive demos ─────────────────────────
# In a real script you'd use: read -p "prompt" varname
# Here we simulate responses

func prompt_demo(prompt_text, default_val) {
    if $INTERACTIVE == "true" {
        # Real interactive read (passthrough)
        bash! read -p "$prompt_text: " _answer
        answer = $(bash! echo "$_answer")
    } else {
        answer = $default_val
        println "$prompt_text: $answer  $(dim '[simulated]')"
    }
    return $answer
}

func dim(s) { return "\033[2m$s\033[0m" }
func bold(s) { return "\033[1m$s\033[0m" }
func green(s) { return "\033[32m$s\033[0m" }
func red(s)   { return "\033[31m$s\033[0m" }
func yellow(s){ return "\033[33m$s\033[0m" }
func cyan(s)  { return "\033[36m$s\033[0m" }

# ── Pattern: confirm (y/n prompt) ─────────────────────────────────────
println "=== confirm prompt ==="
func confirm(question, default_yes) {
    hint = if $default_yes: "[Y/n]"; else: "[y/N]"
    if $INTERACTIVE == "true" {
        bash! read -p "$question $hint: " _yn
        yn = $($_yn | lower | trim)
    } else {
        yn = if $default_yes: "y"; else: "n"
        println "$question $hint: $yn  $(dim '[simulated]')"
    }
    if $yn == "": yn = if $default_yes: "y"; else: "n"
    return $yn == "y" or $yn == "yes"
}

if $(confirm "Do you want to continue?" true) {
    println "  → User said yes"
} else {
    println "  → User said no"
}
println ""

# ── Pattern: numbered menu ────────────────────────────────────────────
println "=== numbered menu ==="
func show_menu(title, options) {
    println "  $(bold $title)"
    println "  $(dim '─────────────────────────')"
    for i in range(0..($options | arr_len)) {
        n = $i + 1
        println "  $(cyan $n). $options[$i]"
    }
    println "  $(dim '─────────────────────────')"
}

func menu_select(title, options, default_choice) {
    show_menu $title $options
    if $INTERACTIVE == "true" {
        bash! read -p "  Enter choice [1-$(echo $options | arr_len)]: " _choice
        choice = $(bash! echo "$_choice" | tr -d '[:space:]')
    } else {
        choice = $default_choice
        println "  Enter choice [1-$(echo $options | arr_len)]: $choice  $(dim '[simulated]')"
    }
    idx = tonum($choice) - 1
    if $idx < 0 or $idx >= ($options | arr_len) {
        println "  Invalid choice: $choice"
        return ""
    }
    return $options[$idx]
}

env_options = ["development", "staging", "production", "local"]
selected = menu_select "Select environment:" $env_options "2"
println "  Selected: $(green $selected)"
println ""

# ── Pattern: validated input loop ────────────────────────────────────
println "=== validated input loop ==="
func prompt_validated(prompt_text, validator_fn, err_msg, default_val, max_attempts) {
    attempts = 0
    while $attempts < $max_attempts {
        attempts++
        val = prompt_demo "$prompt_text" $default_val
        is_valid = $validator_fn($val)
        if $is_valid: return $val
        println "  $(red '✗') $err_msg (attempt $attempts / $max_attempts)"
    }
    throw "max attempts reached"
}

func is_nonempty(s) { return ($s | trim | len) > 0 }
func is_valid_port(s) {
    $s | isnum || return false
    n = tonum $s
    return $n >= 1 and $n <= 65535
}

simulated_inputs = map { "name":"Alice" "port":"8080" "email":"alice@example.com" }

println "Name:"
try {
    name = prompt_validated "  Enter your name" is_nonempty "Name cannot be empty" "Alice" 3
    println "  Got: $(green $name)"
} catch e { println "  $(red $e)" }
println ""

println "Port:"
try {
    port = prompt_validated "  Enter port number" is_valid_port "Must be 1-65535" "8080" 3
    println "  Got: $(green $port)"
} catch e { println "  $(red $e)" }
println ""

# ── Pattern: multi-step wizard ────────────────────────────────────────
println "=== multi-step wizard ==="
func wizard_step(step_num, total, title) {
    pct = $step_num * 100 / $total
    bar = "█" | repeat $step_num
    spc = "░" | repeat ($total - $step_num)
    println ""
    println "  $(bold "Step $step_num/$total: $title")"
    println "  [$bar$spc] $pct%"
    println "  $(dim '─────────────────────────────────────')"
}

func run_wizard() {
    data = map {}
    total_steps = 4

    # Step 1: Project name
    wizard_step 1 $total_steps "Project Setup"
    project_name = prompt_demo "  Project name" "my-awesome-app"
    map_set data "name" $project_name
    println "  ✓ Project: $project_name"

    # Step 2: Language
    wizard_step 2 $total_steps "Language"
    lang = menu_select "  Choose language:" ["go", "python", "node", "rust"] "1"
    map_set data "language" $lang
    println "  ✓ Language: $lang"

    # Step 3: Features
    wizard_step 3 $total_steps "Features"
    features = []
    for feat in ["auth", "database", "cache", "api"] {
        enabled = confirm "  Enable $feat?" true
        if $enabled: features[] = $feat
    }
    map_set data "features" "$features"
    println "  ✓ Features: $features"

    # Step 4: Confirmation
    wizard_step 4 $total_steps "Review & Confirm"
    println "  $(bold 'Summary:')"
    for key in $(map_keys data) {
        val = map_get data $key
        println "    $(key | pad 12) = $val"
    }
    println ""
    confirmed = confirm "  Create project?" true

    if $confirmed {
        println ""
        println "  $(green '✓ Project created successfully!')"
        println "  $(dim "  → $project_name ($lang) with $features")"
    } else {
        println "  $(yellow '⚠ Cancelled')"
    }
    return $data
}

run_wizard
println ""

# ── Pattern: spinner (simulated) ──────────────────────────────────────
println "=== spinner animation ==="
func spinner(label, steps) {
    frames = ["⠋","⠙","⠹","⠸","⠼","⠴","⠦","⠧","⠇","⠏"]
    for i in range(0..$steps) {
        frame = $frames[$i % ($frames | arr_len)]
        printf "\r  $frame  $label... "
        # sleep 0.05  # uncomment in real use
    }
    printf "\r  $(green '✓')  $label done!       \n"
}

spinner "Installing dependencies" 20
spinner "Compiling source code"   15
spinner "Running tests"           25
spinner "Building Docker image"   30
println ""
println "All tasks complete ✓"