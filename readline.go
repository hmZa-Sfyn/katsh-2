package main

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

// isPasteStart returns true when b starts with the bracketed-paste-start sequence
// ESC [ 2 0 0 ~  (6 bytes: 0x1b 0x5b 0x32 0x30 0x30 0x7e)
func isPasteStart(b []byte) bool {
	return len(b) >= 6 &&
		b[0] == 0x1b && b[1] == '[' &&
		b[2] == '2' && b[3] == '0' && b[4] == '0' && b[5] == '~'
}

// findPasteEnd returns the index in buf where ESC[201~ begins, or -1 if not found.
func findPasteEnd(buf []byte) int {
	end := []byte{0x1b, '[', '2', '0', '1', '~'}
	for i := 0; i <= len(buf)-6; i++ {
		match := true
		for j := 0; j < 6; j++ {
			if buf[i+j] != end[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// ─────────────────────────────────────────────────────────────────────────────
//  Readline — full raw-mode line editor
//
//  Features:
//   • ↑↓  history navigation
//   • ←→  cursor movement (character)
//   • Ctrl-A/E  home / end
//   • Ctrl-W    delete previous word
//   • Ctrl-U    clear whole line
//   • Ctrl-K    kill to end of line
//   • Ctrl-L    clear screen
//   • Ctrl-C    cancel line
//   • Ctrl-D    EOF / exit when line empty
//   • Tab       completion (commands, paths, $vars, box keys)
//   • Paste     bracketed paste (Shift+Ctrl+V / middle-click / right-click paste)
//   • Live syntax highlighting while you type
// ─────────────────────────────────────────────────────────────────────────────

// Readline reads one edited line. Returns (line, eof).
func (sh *Shell) Readline(prompt string) (string, bool) {
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		// fallback — plain read
		fmt.Print(prompt)
		var line string
		fmt.Scanln(&line)
		return line, false
	}
	defer term.Restore(fd, oldState)

	// Enable bracketed paste mode: terminal wraps pastes in ESC[200~ ... ESC[201~
	// This prevents pasted text being interpreted as keystrokes.
	fmt.Print("\x1b[?2004h")
	defer fmt.Print("\x1b[?2004l") // disable on exit

	fmt.Print(prompt)

	var buf []rune
	cursor := 0
	histIdx := -1
	savedLine := ""

	// insertRunes inserts a slice of runes at cursor position.
	insertRunes := func(runes []rune) {
		newBuf := make([]rune, len(buf)+len(runes))
		copy(newBuf, buf[:cursor])
		copy(newBuf[cursor:], runes)
		copy(newBuf[cursor+len(runes):], buf[cursor:])
		buf = newBuf
		cursor += len(runes)
	}

	// Redraw current line with syntax highlighting
	redraw := func() {
		pLen := visibleLen(prompt)
		// Go to start of this line, jump past prompt, clear rest
		fmt.Printf("\r\033[%dC\033[K", pLen)
		fmt.Print(highlightInput(string(buf)))
		// Reposition cursor
		col := pLen + runeVisualWidth(buf[:cursor])
		fmt.Printf("\r\033[%dC", col)
	}

	for {
		raw := make([]byte, 256)
		n, err := os.Stdin.Read(raw)
		if err != nil || n == 0 {
			return string(buf), true
		}
		b := raw[:n]

		// ── Bracketed paste: ESC [ 2 0 0 ~ ──────────────────────────────────
		// Modern terminals send  \x1b[200~<text>\x1b[201~  when you paste.
		if isPasteStart(b) {
			// Accumulate everything until we see the closing ESC[201~
			var pasteBuf []byte
			// The payload may already be partly in b after the 6-byte header
			after := b[6:] // bytes after \x1b[200~
			pasteBuf = append(pasteBuf, after...)
			for {
				if pasteEnd := findPasteEnd(pasteBuf); pasteEnd >= 0 {
					pasteBuf = pasteBuf[:pasteEnd]
					break
				}
				chunk := make([]byte, 512)
				nr, er := os.Stdin.Read(chunk)
				if er != nil || nr == 0 {
					break
				}
				pasteBuf = append(pasteBuf, chunk[:nr]...)
			}
			// Convert to runes, replacing \r/\n with space (single-line paste)
			// but preserving the full text including newlines for multi-line
			text := string(pasteBuf)
			// Replace \r\n and bare \r with \n for consistency
			text = strings.ReplaceAll(text, "\r\n", "\n")
			text = strings.ReplaceAll(text, "\r", "\n")
			// For single-line input: replace newlines with spaces so the
			// whole paste lands on one line. Multi-line pastes execute each line.
			lines := strings.Split(text, "\n")
			if len(lines) == 1 {
				insertRunes([]rune(lines[0]))
				redraw()
			} else {
				// Multi-line paste: complete first line immediately, then
				// re-queue remaining lines (execute them in sequence).
				firstLine := string(buf[:cursor]) + lines[0] + string(buf[cursor:])
				fmt.Print("\r\n")
				term.Restore(fd, oldState)
				fmt.Print("\x1b[?2004l")
				return firstLine, false
			}
			continue
		}

		// ── Escape / CSI sequences ───────────────────────────────────────────
		if b[0] == 0x1b {
			if n >= 3 && b[1] == '[' {
				switch b[2] {
				case 'A': // ↑ up — history prev
					if len(sh.history) == 0 {
						continue
					}
					if histIdx == -1 {
						savedLine = string(buf)
						histIdx = len(sh.history) - 1
					} else if histIdx > 0 {
						histIdx--
					}
					buf = []rune(sh.history[histIdx].Raw)
					cursor = len(buf)
					redraw()

				case 'B': // ↓ down — history next
					if histIdx == -1 {
						continue
					}
					if histIdx < len(sh.history)-1 {
						histIdx++
						buf = []rune(sh.history[histIdx].Raw)
					} else {
						histIdx = -1
						buf = []rune(savedLine)
					}
					cursor = len(buf)
					redraw()

				case 'C': // → right
					if cursor < len(buf) {
						cursor++
						redraw()
					}

				case 'D': // ← left
					if cursor > 0 {
						cursor--
						redraw()
					}

				case 'H': // Home
					cursor = 0
					redraw()

				case 'F': // End
					cursor = len(buf)
					redraw()

				case '3': // Delete key (ESC [ 3 ~)
					if n >= 4 && b[3] == '~' && cursor < len(buf) {
						buf = append(buf[:cursor], buf[cursor+1:]...)
						redraw()
					}

				case '1': // ESC [ 1 ~ = Home
					if n >= 4 && b[3] == '~' {
						cursor = 0
						redraw()
					}
				case '4': // ESC [ 4 ~ = End
					if n >= 4 && b[3] == '~' {
						cursor = len(buf)
						redraw()
					}
				}
			}
			continue
		}

		// ── Control characters ───────────────────────────────────────────────
		switch b[0] {
		case 0x0d, 0x0a: // Enter
			fmt.Print("\r\n")
			return string(buf), false

		case 0x7f, 0x08: // Backspace
			if cursor > 0 {
				buf = append(buf[:cursor-1], buf[cursor:]...)
				cursor--
				redraw()
			}

		case 0x01: // Ctrl-A — home
			cursor = 0
			redraw()

		case 0x05: // Ctrl-E — end
			cursor = len(buf)
			redraw()

		case 0x0b: // Ctrl-K — kill to end
			buf = buf[:cursor]
			redraw()

		case 0x15: // Ctrl-U — kill whole line
			buf = buf[:0]
			cursor = 0
			redraw()

		case 0x17: // Ctrl-W — delete prev word
			end := cursor
			for cursor > 0 && buf[cursor-1] == ' ' {
				cursor--
			}
			for cursor > 0 && buf[cursor-1] != ' ' {
				cursor--
			}
			buf = append(buf[:cursor], buf[end:]...)
			redraw()

		case 0x0c: // Ctrl-L — clear screen
			fmt.Print("\033[2J\033[H")
			fmt.Print(prompt)
			fmt.Print(highlightInput(string(buf)))
			col := visibleLen(prompt) + runeVisualWidth(buf[:cursor])
			fmt.Printf("\r\033[%dC", col)

		case 0x04: // Ctrl-D — EOF if empty
			if len(buf) == 0 {
				fmt.Print("\r\n")
				return "", true
			}

		case 0x03: // Ctrl-C — cancel line
			fmt.Print("^C\r\n")
			buf = buf[:0]
			cursor = 0
			histIdx = -1
			fmt.Print(prompt)

		case 0x09: // Tab — completion
			opts := sh.completionOptions(string(buf), cursor)
			switch len(opts) {
			case 0:
				// nothing
			case 1:
				// complete the word
				word := lastWord(string(buf[:cursor]))
				rest := string(buf[cursor:])
				completed := string(buf[:cursor-len([]rune(word))]) + opts[0]
				if !strings.HasSuffix(opts[0], "/") {
					completed += " "
				}
				buf = []rune(completed + rest)
				cursor = len([]rune(completed))
				redraw()
			default:
				// show options below
				fmt.Print("\r\n")
				for i, o := range opts {
					fmt.Printf("  %s%s%s", ansiCyan, o, ansiReset)
					if i < len(opts)-1 {
						fmt.Print("  ")
					}
				}
				fmt.Print("\r\n")
				fmt.Print(prompt)
				fmt.Print(highlightInput(string(buf)))
				col := visibleLen(prompt) + runeVisualWidth(buf[:cursor])
				fmt.Printf("\r\033[%dC", col)
			}

		default:
			// Printable / multibyte UTF-8 — insert at cursor
			ch, size := utf8.DecodeRune(b[:n])
			if ch != utf8.RuneError || size > 1 {
				insertRunes([]rune{ch})
				redraw()
			}
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
//  Live syntax highlighting
// ─────────────────────────────────────────────────────────────────────────────

// highlightInput colorizes a command line string for display.
func highlightInput(line string) string {
	if strings.TrimSpace(line) == "" {
		return line
	}
	spans := spanize(line)
	var out strings.Builder
	cmdPos := 0 // which non-space token are we on

	for _, sp := range spans {
		if sp.space {
			out.WriteString(sp.text)
			continue
		}
		t := sp.text

		switch {
		// Comments
		case strings.HasPrefix(t, "##") || strings.HasPrefix(t, "//") || strings.HasPrefix(t, "# ") || t == "#":
			out.WriteString(ansiGrey + t + ansiReset)
			continue

		// Backtick subshell
		case strings.HasPrefix(t, "`"):
			out.WriteString(ansiMagenta + t + ansiReset)
			cmdPos++
			continue

		// String literals
		case len(t) >= 2 && ((t[0] == '"' && t[len(t)-1] == '"') || (t[0] == '\'' && t[len(t)-1] == '\'')):
			out.WriteString(ansiYellow + t + ansiReset)
			cmdPos++
			continue

		// Box-store operator
		case strings.HasPrefix(t, "#="):
			out.WriteString(ansiYellow + t + ansiReset)
			continue

		// Variable reference
		case strings.HasPrefix(t, "$"):
			out.WriteString(ansiDarkCyan + t + ansiReset)
			cmdPos++
			continue

		// Flags
		case strings.HasPrefix(t, "-") && len(t) > 1:
			out.WriteString(ansiCyan + t + ansiReset)
			continue

		// Numbers
		case isNumericStr(t):
			out.WriteString(ansiDarkYellow + t + ansiReset)
			cmdPos++
			continue

		// Operators
		case isSyntaxOp(t):
			out.WriteString(ansiGrey + t + ansiReset)
			continue

		// Keywords
		case isKeyword(t):
			out.WriteString(ansiBold + ansiMagenta + t + ansiReset)
			cmdPos++
			continue
		}

		// First token (command name) and first token after |
		if cmdPos == 0 || sp.afterPipe {
			if isBuiltin(t) {
				out.WriteString(ansiBold + ansiGreen + t + ansiReset)
			} else if findInPath(t) != "" {
				out.WriteString(ansiGreen + t + ansiReset)
			} else {
				// Unknown command — red underline
				out.WriteString(ansiRed + ansiUnderline + t + ansiReset)
			}
		} else {
			out.WriteString(ansiWhite + t + ansiReset)
		}
		cmdPos++
	}
	return out.String()
}

type span struct {
	text      string
	space     bool
	afterPipe bool
}

func spanize(s string) []span {
	var spans []span
	var cur strings.Builder
	inQuote := false
	quoteChar := rune(0)
	inBacktick := false
	prevWasPipe := false

	flush := func(isSpace bool) {
		if cur.Len() == 0 {
			return
		}
		t := cur.String()
		spans = append(spans, span{text: t, space: isSpace, afterPipe: prevWasPipe && !isSpace})
		if !isSpace {
			prevWasPipe = (t == "|")
		}
		cur.Reset()
	}

	for _, ch := range s {
		switch {
		case inBacktick:
			cur.WriteRune(ch)
			if ch == '`' {
				inBacktick = false
				flush(false)
			}
		case inQuote:
			cur.WriteRune(ch)
			if ch == quoteChar {
				inQuote = false
				flush(false)
			}
		case ch == '`':
			flush(false)
			inBacktick = true
			cur.WriteRune(ch)
		case ch == '"' || ch == '\'':
			flush(false)
			inQuote = true
			quoteChar = ch
			cur.WriteRune(ch)
		case ch == ' ' || ch == '\t':
			flush(false)
			cur.WriteRune(ch)
			flush(true)
		default:
			cur.WriteRune(ch)
		}
	}
	flush(false)
	return spans
}

func isKeyword(t string) bool {
	switch strings.ToLower(t) {
	// Control flow
	case "if", "elif", "else", "fi",
		"for", "while", "do", "done", "until",
		"in", "range", "break", "continue",
		// Functions
		"func", "return", "local",
		// Values
		"true", "false", "null", "nil",
		// Logic
		"and", "or", "not",
		// Output
		"print", "println", "pass",
		// Match/switch
		"match", "case", "default", "switch", "fallthrough",
		// Conditional shortcuts
		"unless", "when",
		// Error handling
		"try", "catch", "finally", "throw", "raise",
		// Loops
		"repeat",
		// Type declarations
		"enum", "struct",
		// Scoping
		"defer", "with",
		// Jumping
		"goto", "label",
		// Module
		"import", "export", "readonly",
		// Data type constructors (highlighted as keywords)
		"map", "set", "stack", "queue", "tuple", "matrix":
		return true
	}
	return false
}

func isSyntaxOp(t string) bool {
	switch t {
	// Pipe / logical
	case "|", "||", "&&", "|>",
		// Comparison
		"==", "!=", ">=", "<=", ">", "<", "~=", "!~", "~",
		// Assignment / compound
		"=", "+=", "-=", "*=", "/=", "%=", "**=",
		// Arithmetic
		"+", "-", "*", "/", "%", "**",
		// Increment / decrement
		"++", "--",
		// Range / spread
		"..", "...",
		// Grouping / structure
		"{", "}", "(", ")", "[", "]",
		// Separators
		":", ";", ",",
		// Arrow / map
		"->", "=>", ":=",
		// String concat
		".",
		// Redirect
		">>", "2>":
		return true
	}
	return false
}

func isNumericStr(t string) bool {
	if t == "" {
		return false
	}
	dot := false
	for i, ch := range t {
		if i == 0 && (ch == '-' || ch == '+') {
			continue
		}
		if ch == '.' {
			if dot {
				return false
			}
			dot = true
			continue
		}
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// ─────────────────────────────────────────────────────────────────────────────
//  Tab completion
// ─────────────────────────────────────────────────────────────────────────────

func (sh *Shell) completionOptions(line string, cursor int) []string {
	prefix := line[:cursor]
	word := lastWord(prefix)
	toks := strings.Fields(prefix)
	isCmd := len(toks) == 0 || (len(toks) == 1 && !strings.HasSuffix(prefix, " "))

	var opts []string

	if isCmd {
		for _, b := range allBuiltinNames() {
			if strings.HasPrefix(b, word) {
				opts = append(opts, b)
			}
		}
		for name := range sh.aliases {
			if strings.HasPrefix(name, word) {
				opts = append(opts, name)
			}
		}
	} else if strings.HasPrefix(word, "$") {
		pfx := word[1:]
		for k := range sh.vars {
			if strings.HasPrefix(k, pfx) {
				opts = append(opts, "$"+k)
			}
		}
	} else {
		// Path completion
		dir := sh.cwd
		filePfx := word
		if idx := strings.LastIndex(word, "/"); idx >= 0 {
			dir = resolvePath(sh.cwd, word[:idx+1])
			filePfx = word[idx+1:]
		}
		entries, err := os.ReadDir(dir)
		if err == nil {
			for _, e := range entries {
				name := e.Name()
				if strings.HasPrefix(name, filePfx) {
					pfxBase := word[:len(word)-len(filePfx)]
					if e.IsDir() {
						name += "/"
					}
					opts = append(opts, pfxBase+name)
				}
			}
		}
		// Box key completion after "box get/rm/rename/tag"
		if len(toks) >= 2 && toks[0] == "box" {
			for _, k := range sh.box.Keys() {
				if strings.HasPrefix(k, word) {
					opts = append(opts, k)
				}
			}
		}
	}
	return opts
}

func lastWord(s string) string {
	if strings.HasSuffix(s, " ") {
		return ""
	}
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return ""
	}
	return parts[len(parts)-1]
}

func allBuiltinNames() []string {
	return []string{
		"select", "where", "bash!", "zsh!", "sh!", "!zsh", "!bash", "!sh", "skip", "limit", "rename", "count",
		// ── Navigation ──────────────────────────────────────────────────────
		"cd", "pwd", "pushd", "popd", "dirs",
		// ── Listing ─────────────────────────────────────────────────────────
		"ls", "ll", "la", "tree", "du", "df",
		// ── File operations ─────────────────────────────────────────────────
		"cat", "head", "tail", "touch", "mkdir", "rmdir", "rm", "cp", "mv", "ln",
		"readlink", "realpath", "basename", "dirname", "mktemp", "mkfifo",
		// ── Inspection ───────────────────────────────────────────────────────
		"wc", "stat", "file", "find", "diff",
		// ── Text processing ──────────────────────────────────────────────────
		"grep", "sed", "awk", "cut", "tr", "sort", "uniq", "tee", "split", "xargs",
		"nl", "fold", "expand", "unexpand", "column", "paste", "join", "comm", "shuf",
		"numfmt", "rev", "strings", "xxd", "od",
		// ── Permissions ──────────────────────────────────────────────────────
		"chmod", "chown",
		// ── Process management ───────────────────────────────────────────────
		"ps", "kill", "sleep", "jobs", "nice", "timeout", "pgrep", "pkill", "nohup",
		"bg", "fg", "top", "lsof", "vmstat", "iostat",
		// ── System info ──────────────────────────────────────────────────────
		"uname", "uptime", "date", "cal", "hostname", "whoami", "id", "groups", "who", "w",
		"free", "lscpu", "lsusb", "lspci", "dmesg", "lsblk", "mount", "umount",
		"fdisk", "blkid", "journalctl", "systemctl", "service",
		// ── Networking ───────────────────────────────────────────────────────
		"ping", "curl", "wget", "nslookup", "dig", "ifconfig", "ip",
		"ss", "netstat", "traceroute", "tracert", "mtr", "openssl",
		"ssh", "scp", "rsync", "httpget", "httppost", "jq",
		// ── Hashing / archives ───────────────────────────────────────────────
		"md5sum", "md5", "sha1sum", "sha1", "sha256sum", "sha256",
		"tar", "gzip", "gunzip", "zip", "unzip",
		// ── Text generation / math ───────────────────────────────────────────
		"echo", "printf", "yes", "seq", "base64", "bc", "factor", "random",
		// ── Variables / env ──────────────────────────────────────────────────
		"set", "unset", "vars", "export", "import", "env", "printenv",
		"readonly", "declare", "typeset", "getopts",
		// ── Identification ────────────────────────────────────────────────────
		"alias", "unalias", "aliases", "which", "type",
		// ── Scripting helpers ─────────────────────────────────────────────────
		"eval", "exec", "test", "[", "read", "mapfile", "readarray", "source", ".",
		"true", "false", "pass", "local", "break", "continue", "return",
		// ── Shell passthrough ─────────────────────────────────────────────────
		"run", "shell", "capture", "bash", "zsh", "sh", "fish", "ksh", "dash",
		// ── Session ───────────────────────────────────────────────────────────
		"box", "history", "clear", "help", "man", "watch", "exit", "quit",
		// ── Fun / visual ──────────────────────────────────────────────────────
		"figlet", "toilet", "lolcat", "banner2", "drawbox", "notify",
		// ── Scripting keywords ────────────────────────────────────────────────
		"if", "elif", "else", "fi",
		"for", "while", "do", "done", "until",
		"in", "range",
		"func", "return",
		"match", "case", "default", "switch", "fallthrough",
		"unless", "when",
		"try", "catch", "finally", "throw", "raise",
		"repeat",
		"enum", "struct",
		"defer", "with",
		"goto", "label",
		"and", "or", "not",
		"null", "nil",
		"print", "println",
		// ── String operations (pipe ops + standalone) ──────────────────────
		"upper", "lower", "title",
		"trim", "ltrim", "rtrim", "strip",
		"len", "reverse", "replace", "replace1",
		"sub", "sub_n", "pad", "lpad", "center",
		"startswith", "endswith", "contains",
		"isnum", "isalpha", "isalnum", "isspace", "isupper", "islower",
		"lines", "words", "chars", "concat", "prepend",
		// ── Array operations ──────────────────────────────────────────────────
		"first", "last", "nth", "slice", "push", "pop", "flatten",
		"arr_uniq", "arr_sort", "arr_reverse", "arr_len", "arr_join",
		"arr_contains", "arr_map", "arr_filter",
		"arr_sum", "arr_min", "arr_max", "arr_avg",
		// ── Number operations ─────────────────────────────────────────────────
		"add", "mul", "div", "mod", "pow",
		"abs", "ceil", "floor", "round", "sqrt", "negate",
		"hex", "oct", "bin",
		"tonum", "tostr", "toarray",
		// ── Type inspection ───────────────────────────────────────────────────
		"typeof", "dt_show",
		// ── Map commands ──────────────────────────────────────────────────────
		"map_new", "map_set", "map_get", "map_del", "map_has",
		"map_keys", "map_values", "map_len", "map_show", "map_merge",
		// ── Set commands ──────────────────────────────────────────────────────
		"set_new", "set_add", "set_remove", "set_has",
		"set_union", "set_intersect", "set_diff", "set_show", "set_len",
		// ── Stack commands ────────────────────────────────────────────────────
		"stack_new", "stack_push", "stack_pop", "stack_peek", "stack_len", "stack_show",
		// ── Queue commands ────────────────────────────────────────────────────
		"queue_new", "enqueue", "dequeue", "queue_peek", "queue_len", "queue_show",
		// ── Tuple commands ────────────────────────────────────────────────────
		"tuple_get", "tuple_len", "tuple_show",
		// ── Matrix commands ───────────────────────────────────────────────────
		"matrix_new", "matrix_get", "matrix_set", "matrix_add", "matrix_mul",
		"matrix_transpose", "matrix_det", "matrix_show", "matrix_identity",
	}
}

// ─────────────────────────────────────────────────────────────────────────────
//  Visual width helpers
// ─────────────────────────────────────────────────────────────────────────────

// visibleLen returns terminal column width of a string, skipping ANSI escapes.
func visibleLen(s string) int {
	n := 0
	esc := false
	for _, ch := range s {
		if esc {
			if ch == 'm' {
				esc = false
			}
			continue
		}
		if ch == '\033' {
			esc = true
			continue
		}
		n++
	}
	return n
}

// runeVisualWidth returns display width of a rune slice (approx 1 per rune).
func runeVisualWidth(r []rune) int { return len(r) }
