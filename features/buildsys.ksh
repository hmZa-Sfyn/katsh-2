#!/usr/bin/env katsh
# ─────────────────────────────────────────────────────────────────────
#  33_build_system.ksh — Mini task runner / build system
#
#  Topics covered:
#    task registry with maps · dependency resolution (topological sort)
#    task execution with timing · skip if up-to-date
#    before/after hooks · parallel task grouping (simulated)
#    colored output · build graph visualization
# ─────────────────────────────────────────────────────────────────────

# ── ANSI helpers ──────────────────────────────────────────────────────
func col(code, s) { return "\033[${code}m${s}\033[0m" }
func green(s)  { return col "32" $s }
func yellow(s) { return col "33" $s }
func red(s)    { return col "31" $s }
func cyan(s)   { return col "36" $s }
func bold(s)   { return col "1"  $s }
func grey(s)   { return col "90" $s }
func dim(s)    { return col "2"  $s }

# ── Task registry ─────────────────────────────────────────────────────
tasks_desc = map {}     # name → description
tasks_deps = map {}     # name → comma-sep deps
tasks_done = set {}     # set of completed task names
tasks_skipped = set {}
build_errors = []

func task(name, desc, deps) {
    map_set tasks_desc $name $desc
    map_set tasks_deps $name $deps
}

# ── Topological sort (Kahn's algorithm) ──────────────────────────────
func topo_sort(all_tasks) {
    in_degree = map {}
    adj       = map {}

    for t in $all_tasks {
        map_set in_degree $t 0
        map_set adj       $t ""
    }

    for t in $all_tasks {
        deps_raw = map_get tasks_deps $t
        if $deps_raw == "": continue
        deps = $deps_raw | split ","
        for dep in $deps {
            dep = $dep | trim
            cur_deg = tonum $(map_get in_degree $t)
            map_set in_degree $t ($cur_deg + 1)
            cur_adj = map_get adj $dep
            if $cur_adj == "" {
                map_set adj $dep $t
            } else {
                map_set adj $dep "$cur_adj,$t"
            }
        }
    }

    q = queue {}
    for t in $all_tasks {
        if $(map_get in_degree $t) == 0: enqueue q $t
    }

    order = []
    while $(queue_len q) > 0 {
        dequeue q t
        order[] = $t

        neighbors_raw = map_get adj $t
        if $neighbors_raw != "" {
            for neighbor in $($neighbors_raw | split ",") {
                neighbor = $neighbor | trim
                deg = tonum $(map_get in_degree $neighbor) - 1
                map_set in_degree $neighbor $deg
                if $deg == 0: enqueue q $neighbor
            }
        }
    }
    return $order
}

# ── Task runner ───────────────────────────────────────────────────────
func ljust(s, w) { return tostr($s) | pad  $w }
func rjust(s, w) { return tostr($s) | lpad $w }

tasks_timing = map {}

func run_task(name) {
    set_has tasks_done $name && {
        println "  $(grey "skip $name (already done)")"
        return
    }

    # Run dependencies first
    deps_raw = map_get tasks_deps $name
    if $deps_raw != "" {
        for dep in $($deps_raw | split ",") {
            dep = $dep | trim
            run_task $dep
        }
    }

    # Execute the task
    desc = map_get tasks_desc $name
    printf "  %-20s " "$(cyan $name)"

    start_ms = $(date +%s%3N 2>/dev/null)
    if $start_ms == "": start_ms = 0

    try {
        # Dispatch to task function
        match $name {
            "clean":      task_clean
            "install":    task_install
            "generate":   task_generate
            "compile":    task_compile
            "lint":       task_lint
            "test":       task_test
            "test:unit":  task_test_unit
            "test:e2e":   task_test_e2e
            "bundle":     task_bundle
            "minify":     task_minify
            "docker:build": task_docker_build
            "docker:push":  task_docker_push
            "deploy":     task_deploy
            default:      throw "no handler for task '$name'"
        }

        end_ms = $(date +%s%3N 2>/dev/null)
        if $end_ms == "": end_ms = 0
        elapsed = $end_ms - $start_ms

        set_add tasks_done $name
        map_set tasks_timing $name "${elapsed}ms"
        println "$(green '✓')  $(grey "${elapsed}ms")"
    } catch e {
        build_errors[] = "$name: $e"
        println "$(red '✗')  $e"
        throw "TaskFailed: $name"
    }
}

# ── Task implementations (simulated) ─────────────────────────────────
func task_clean()       { println -n "" }   # noop — instant
func task_install()     {
    # Simulate installing deps
    packages = ["lodash@4.17.21", "axios@1.6.0", "express@4.18.2"]
    for pkg in $packages { println "    $(grey "  ↳ $pkg")" }
}
func task_generate()    { println -n "" }
func task_compile()     {
    files = ["src/main.go", "src/server.go", "src/handlers.go", "src/models.go"]
    for f in $files { println "    $(grey "  ↳ compiling $f")" }
}
func task_lint()        {
    warnings = 0
    println "    $(grey "  ↳ 0 errors, $warnings warnings")"
}
func task_test_unit()   {
    println "    $(grey "  ↳ 47 passed, 0 failed")"
}
func task_test_e2e()    {
    println "    $(grey "  ↳ 12 passed, 0 failed")"
}
func task_test()        { println -n "" }
func task_bundle()      {
    println "    $(grey "  ↳ output: dist/app.js (142kb)")"
}
func task_minify()      {
    println "    $(grey "  ↳ output: dist/app.min.js (48kb)")"
}
func task_docker_build() {
    println "    $(grey "  ↳ image: myapp:v2.1.4 (287MB)")"
}
func task_docker_push()  {
    println "    $(grey "  ↳ pushed: registry.io/myapp:v2.1.4")"
}
func task_deploy()      {
    println "    $(grey "  ↳ deployed to production k8s cluster")"
    println "    $(grey "  ↳ 3 pods rolling updated")"
}

# ── Register all tasks ────────────────────────────────────────────────
task "clean"        "Remove build artifacts"           ""
task "install"      "Install dependencies"             ""
task "generate"     "Code generation"                  "install"
task "compile"      "Compile source code"              "generate,install"
task "lint"         "Run linter"                       "install"
task "test:unit"    "Run unit tests"                   "compile"
task "test:e2e"     "Run end-to-end tests"             "compile"
task "test"         "Run all tests"                    "test:unit,test:e2e"
task "bundle"       "Bundle assets"                    "compile"
task "minify"       "Minify bundle"                    "bundle"
task "docker:build" "Build Docker image"               "test,minify"
task "docker:push"  "Push image to registry"           "docker:build"
task "deploy"       "Deploy to production"             "docker:push"

all_tasks = ["clean","install","generate","compile","lint",
             "test:unit","test:e2e","test","bundle","minify",
             "docker:build","docker:push","deploy"]

# ── Visualize build graph ─────────────────────────────────────────────
println "$(bold 'Build Graph')"
println "$(grey '─────────────────────────────────────────')"
for t in $all_tasks {
    deps = map_get tasks_deps $t
    desc = map_get tasks_desc $t
    if $deps != "" {
        println "  $(cyan $t)  $(grey "← $deps")"
    } else {
        println "  $(cyan $t)  $(grey "(no deps)")"
    }
}
println ""

# ── Run full build ────────────────────────────────────────────────────
println "$(bold 'Running: deploy')"
println "$(grey '─────────────────────────────────────────')"
start_total = $(date +%s%3N 2>/dev/null)
if $start_total == "": start_total = 0

try {
    run_task "deploy"
    end_total = $(date +%s%3N 2>/dev/null)
    if $end_total == "": end_total = 0
    elapsed_total = $end_total - $start_total

    println ""
    println "$(grey '─────────────────────────────────────────')"
    println "$(green '✓ Build succeeded')  $(grey "(${elapsed_total}ms total)")"
    println ""
    println "$(bold 'Task timings:')"
    for t in $all_tasks {
        if set_has tasks_done $t {
            timing = map_get tasks_timing $t
            println "  $(ljust $t 20) $timing"
        }
    }
} catch e {
    println ""
    println "$(red '✗ Build failed:')"
    for err in $build_errors {
        println "  $(red '•') $err"
    }
}