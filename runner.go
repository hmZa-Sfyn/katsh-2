package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
//  Script runner — executes .ksh files and inline command strings
//
//  Features:
//   • Shebang stripping  (#!/usr/bin/env katsh)
//   • In-script @set flags  (# @set -e  /  # @set -x)
//   • Positional arguments  $0 $1 $2 ... $# $_args
//   • Multi-line continuation  with backslash \
//   • Heredoc  <<EOF ... EOF
//   • Source / import  relative to script directory
//   • Exit code propagation
//   • Trace mode  (+command echo before each line)
//   • Dry-run mode  (parse and show but don't execute)
//   • Error context  (file:line in error messages)
//   • Timing  (--time flag or KATSH_TIME=1 env var)
// ─────────────────────────────────────────────────────────────────────────────

// ScriptOptions controls script execution behaviour.
type ScriptOptions struct {
	ExitOnError bool   // -e: exit on first non-zero exit code
	Trace       bool   // -x: print each logical line before executing
	DryRun      bool   // -n: parse but don't execute
	Verbose     bool   // -v: print raw source lines as read
	InlineCmd   string // -c: inline command string (no file)
	Timing      bool   // show execution time per line
}

// RunScript reads and executes a .ksh script file.
// scriptArgs are bound to $1, $2, ... inside the script.
func RunScript(sh *Shell, path string, scriptArgs []string, opts ScriptOptions) int {
	// Resolve path
	absPath, err := filepath.Abs(path)
	if err != nil || !fileExists(absPath) {
		// Try with .ksh extension if not found
		if !strings.HasSuffix(path, ".ksh") {
			if fileExists(absPath + ".ksh") {
				absPath = absPath + ".ksh"
			} else {
				fmt.Fprintf(os.Stderr, "\n  %sFileNotFound[E010]%s  script not found: %q\n  ╰─ check the path with: ls %s\n\n",
					ansiRed+ansiBold, ansiReset, path, filepath.Dir(absPath))
				return 127
			}
		} else {
			fmt.Fprintf(os.Stderr, "\n  %sFileNotFound[E010]%s  script not found: %q\n  ╰─ check the path with: ls %s\n\n",
				ansiRed+ansiBold, ansiReset, path, filepath.Dir(absPath))
			return 127
		}
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "katsh: cannot read %q: %v\n", absPath, err)
		return 1
	}

	// Set script dir as cwd context for relative imports
	scriptDir := filepath.Dir(absPath)
	origCwd := sh.cwd
	if sh.cwd == origCwd {
		// Leave cwd as-is but store script dir for relative path resolution
		sh.vars["__script_dir"] = scriptDir
		sh.vars["__script_file"] = absPath
	}

	return runLines(sh, string(data), absPath, scriptArgs, opts)
}

// RunInline executes a string of commands (from -c flag).
func RunInline(sh *Shell, cmd string, opts ScriptOptions) int {
	return runLines(sh, cmd, "<inline>", nil, opts)
}

// runLines is the core execution engine for both file and inline scripts.
func runLines(sh *Shell, src, srcName string, scriptArgs []string, opts ScriptOptions) int {
	// ── Scan for in-script @set flags ──────────────────────────────────────
	for _, rawLine := range strings.SplitN(src, "\n", 10) {
		trimmed := strings.TrimSpace(rawLine)
		if strings.HasPrefix(trimmed, "# @set ") {
			flag := strings.TrimSpace(trimmed[7:])
			switch flag {
			case "-e":  opts.ExitOnError = true
			case "-x":  opts.Trace = true
			case "-ex", "-xe": opts.ExitOnError = true; opts.Trace = true
			case "-n":  opts.DryRun = true
			}
		}
	}

	// ── Set positional arguments ────────────────────────────────────────────
	sh.vars["0"] = srcName
	sh.vars["#"] = strconv.Itoa(len(scriptArgs))
	sh.vars["_args"] = strings.Join(scriptArgs, " ")
	for i, a := range scriptArgs {
		sh.vars[strconv.Itoa(i+1)] = a
	}

	// ── Parse logical lines (handle continuation and heredocs) ─────────────
	logicalLines, lineNums := parseLogicalLines(src)

	// ── Print script banner in trace mode ──────────────────────────────────
	if opts.Trace {
		fmt.Printf("\n  %s+ running: %s%s\n", ansiGrey, srcName, ansiReset)
		if len(scriptArgs) > 0 {
			fmt.Printf("  %s+ args: %s%s\n\n", ansiGrey, strings.Join(scriptArgs, " "), ansiReset)
		}
	}

	start := time.Now()
	lastCode := 0

	// ── Store script name for error reporting ──────────────────────────────
	sh.currentFile = filepath.Base(srcName)

	// ── Execute each logical line ───────────────────────────────────────────
	for i, line := range logicalLines {
		lineNo := lineNums[i]
		line = strings.TrimSpace(line)

		// Skip blanks and pure comments
		if line == "" || isComment(line) { continue }

		// Update location for error messages
		sh.currentLine = lineNo
		sh.currentSrc  = line

		// Verbose mode: print raw line as read
		if opts.Verbose {
			fmt.Printf("  %s%s:%d%s  %s\n", ansiGrey, filepath.Base(srcName), lineNo, ansiReset, line)
		}

		// Trace mode: print line with + prefix before executing
		if opts.Trace {
			fmt.Printf("  %s+[%d]%s  %s%s%s\n",
				ansiDarkCyan, lineNo, ansiReset,
				ansiWhite, line, ansiReset)
		}

		// Dry-run: don't execute, just show
		if opts.DryRun { continue }

		// Execute
		lineStart := time.Now()
		code := sh.execLine(line)
		elapsed := time.Since(lineStart)

		if opts.Timing && elapsed > time.Millisecond {
			fmt.Printf("  %s  [%.1fms]%s\n", ansiGrey, float64(elapsed.Nanoseconds())/1e6, ansiReset)
		}

		if code != 0 {
			lastCode = code
			if opts.ExitOnError {
				// Rich exit-on-error message with file, line, source
				fmt.Fprintf(os.Stderr, "\n  %s✘ %s:%d — script aborted (exit code %d)%s\n",
					ansiRed+ansiBold, filepath.Base(srcName), lineNo, code, ansiReset)
				fmt.Fprintf(os.Stderr, "  %s│  %s%s\n",
					ansiGrey, line, ansiReset)
				fmt.Fprintf(os.Stderr, "  %s╰─ add 'try { } catch e { }' to handle this error%s\n\n",
					ansiYellow, ansiReset)
				return code
			}
		} else {
			lastCode = 0
		}
	}

	if opts.Trace {
		fmt.Printf("\n  %s+ done in %.1fms (exit %d)%s\n\n",
			ansiGrey, float64(time.Since(start).Nanoseconds())/1e6, lastCode, ansiReset)
	}

	return lastCode
}

// ─── Logical line parser ──────────────────────────────────────────────────────
// Handles:
//   • Shebang lines  #!/...
//   • Pure comment lines  # ...
//   • Continuation lines  cmd \<newline>next
//   • Heredocs  <<EOF ... EOF  (collapsed to single line)
//   • Semicolon-separated statements on one physical line

func parseLogicalLines(src string) (lines []string, lineNumbers []int) {
	scanner := bufio.NewScanner(strings.NewReader(src))
	physLine := 0
	var pending strings.Builder
	pendingStart := 0
	inHeredoc := false
	heredocDelim := ""
	var heredocBuf strings.Builder
	heredocStart := 0

	flush := func() {
		if pending.Len() > 0 {
			lines = append(lines, pending.String())
			lineNumbers = append(lineNumbers, pendingStart)
			pending.Reset()
		}
	}

	for scanner.Scan() {
		physLine++
		raw := scanner.Text()

		// ── Heredoc collection ────────────────────────────────────────────
		if inHeredoc {
			if strings.TrimSpace(raw) == heredocDelim {
				// End of heredoc — emit as a single string variable assignment
				// The line that started the heredoc is already in pending
				heredocContent := heredocBuf.String()
				// Escape newlines for inline use
				escaped := strings.ReplaceAll(heredocContent, "\n", "\\n")
				pending.WriteString(" \"" + escaped + "\"")
				flush()
				inHeredoc = false
				heredocBuf.Reset()
			} else {
				if heredocBuf.Len() > 0 { heredocBuf.WriteString("\n") }
				heredocBuf.WriteString(raw)
			}
			continue
		}

		// ── Shebang ───────────────────────────────────────────────────────
		if physLine == 1 && strings.HasPrefix(raw, "#!") { continue }

		// ── Blank and comment lines ───────────────────────────────────────
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || isComment(trimmed) {
			flush()
			continue
		}

		// ── Line continuation: backslash at end ───────────────────────────
		if strings.HasSuffix(raw, "\\") {
			if pending.Len() == 0 { pendingStart = physLine }
			pending.WriteString(strings.TrimSuffix(raw, "\\"))
			pending.WriteString(" ")
			continue
		}

		// ── Start heredoc: line ends with  <<DELIM ────────────────────────
		if idx := findHeredoc(raw); idx >= 0 {
			delim := strings.TrimSpace(raw[idx+2:])
			delim = strings.Trim(delim, `"'`)
			if pending.Len() == 0 { pendingStart = physLine }
			pending.WriteString(raw[:idx])
			inHeredoc = true
			heredocDelim = delim
			heredocStart = physLine
			_ = heredocStart
			continue
		}

		// ── Normal line: append to pending and flush ──────────────────────
		if pending.Len() == 0 { pendingStart = physLine }
		pending.WriteString(raw)
		flush()
	}

	// Flush any remaining pending content
	if pending.Len() > 0 {
		lines = append(lines, pending.String())
		lineNumbers = append(lineNumbers, pendingStart)
	}

	return lines, lineNumbers
}

// findHeredoc returns the index of << in s (outside quotes), or -1.
func findHeredoc(s string) int {
	inQ := false
	qCh := byte(0)
	for i := 0; i < len(s)-1; i++ {
		if inQ {
			if s[i] == qCh { inQ = false }
			continue
		}
		if s[i] == '"' || s[i] == '\'' { inQ = true; qCh = s[i]; continue }
		if s[i] == '<' && s[i+1] == '<' {
			// Make sure it's not <<<
			if i+2 < len(s) && s[i+2] == '<' { continue }
			return i
		}
	}
	return -1
}

// isComment returns true for pure comment lines (# but not #=).
func isComment(line string) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "#!") { return true } // shebang
	if strings.HasPrefix(trimmed, "# @") { return true } // @set directives
	if strings.HasPrefix(trimmed, "#=") { return false } // box store operator
	if strings.HasPrefix(trimmed, "#") { return true }
	if strings.HasPrefix(trimmed, "//") { return true }
	return false
}

// fileExists returns true if path is an existing regular file.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// ─── Run command ──────────────────────────────────────────────────────────────
// Exposed so builtins (source, import) can reuse the same runner.

// SourceFile sources a file relative to sh.cwd using the current options.
func SourceFile(sh *Shell, path string) int {
	abs := path
	if !filepath.IsAbs(path) {
		// Try relative to script dir first, then cwd
		scriptDir := sh.vars["__script_dir"]
		if scriptDir != "" {
			candidate := filepath.Join(scriptDir, path)
			if fileExists(candidate) {
				abs = candidate
			}
		}
		if abs == path {
			abs = filepath.Join(sh.cwd, path)
		}
	}
	return RunScript(sh, abs, nil, ScriptOptions{})
}
