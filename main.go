package main

import (
	"fmt"
	"os"
)

// ─────────────────────────────────────────────────────────────────────────────
//  KatSH — Structured Shell entry point
//
//  Usage:
//    katsh                        interactive REPL
//    katsh script.ksh             run script file
//    katsh -e  script.ksh         exit on first error
//    katsh -x  script.ksh         trace mode
//    katsh -ex script.ksh         both flags
//    katsh -n  script.ksh         dry run (parse, don't execute)
//    katsh -c  "cmd; cmd"         inline commands
//    katsh script.ksh arg1 arg2   pass positional args ($1 $2 ...)
// ─────────────────────────────────────────────────────────────────────────────

func main() {
	args := os.Args[1:]

	opts := ScriptOptions{}
	remaining := []string{}

	i := 0
	for i < len(args) {
		a := args[i]
		switch a {
		case "-e":
			opts.ExitOnError = true
		case "-x":
			opts.Trace = true
		case "-ex", "-xe":
			opts.ExitOnError = true
			opts.Trace = true
		case "-n":
			opts.DryRun = true
		case "-v":
			opts.Verbose = true
		case "-c":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "katsh: -c requires a command string")
				os.Exit(1)
			}
			opts.InlineCmd = args[i]
		case "--":
			remaining = append(remaining, args[i+1:]...)
			i = len(args)
			continue
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		case "--version":
			fmt.Println("KatSH v0.4.0 (StructSH engine)")
			os.Exit(0)
		default:
			remaining = append(remaining, a)
		}
		i++
	}

	sh := NewShell()

	if opts.InlineCmd != "" {
		os.Exit(RunInline(sh, opts.InlineCmd, opts))
	}

	if len(remaining) > 0 {
		os.Exit(RunScript(sh, remaining[0], remaining[1:], opts))
	}

	sh.Run()
}

func printUsage() {
	fmt.Print(`
  KatSH — Structured Shell  v0.4.0

  Usage:
    katsh                         interactive REPL
    katsh script.ksh              run script file
    katsh -e  script.ksh          exit on first non-zero command
    katsh -x  script.ksh          trace mode (print each line before running)
    katsh -ex script.ksh          -e and -x together
    katsh -n  script.ksh          dry-run (parse only, no execution)
    katsh -c  "cmd1; cmd2"        run inline command string
    katsh script.ksh arg1 arg2    $1 $2 ... set from command line args

  Shebang line in script:
    #!/usr/bin/env katsh

  In-script option comments:
    # @set -e    exit on error
    # @set -x    trace mode

`)
}
