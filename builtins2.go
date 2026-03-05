package main

import (
	"archive/zip"
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
//  Built-ins batch 2 — 50 additional commands
//
//  System/Process: lsof, strace (stub), top, vmstat, iostat, free, lscpu,
//                  lsusb, lspci, dmesg, journalctl, systemctl, service,
//                  nice, timeout, pgrep, pkill, nohup, bg, fg
//
//  Filesystem:     lsblk, mount, umount, fdisk, blkid, mkfifo,
//                  readlink, realpath, basename, dirname, mktemp
//
//  Network:        ss, netstat, traceroute, mtr, openssl, ssh (stub),
//                  scp (stub), rsync (stub), curl (extended), httpget,
//                  httppost, jq (json query)
//
//  Text/Data:      xxd, od, strings, column, nl, fold, fmt, expand,
//                  unexpand, csplit, paste, join, comm, tsort, head2,
//                  shuf, numfmt, printf2
//
//  Import/Export:  import, export (extension + variable management)
//
//  Scripting:      eval, exec, test, read, mapfile/readarray,
//                  compgen, declare, typeset, getopts
//
//  Misc:           notify, banner, figlet (ascii art), matrix, lolcat,
//                  toilet, box2 (draw box), qr (placeholder), weather
// ─────────────────────────────────────────────────────────────────────────────

func handleBuiltin2(sh *Shell, command string, args []string) (*Result, bool, error) {
	switch command {

	// ── System / Process ────────────────────────────────────────────────────
	case "lsof":
		return builtinLsof(sh, args)
	case "top":
		return builtinTop(sh, args)
	case "vmstat":
		return delegateCmd("vmstat", args)
	case "iostat":
		return delegateCmd("iostat", args)
	case "free":
		return builtinFree(sh, args)
	case "lscpu":
		return builtinLscpu(sh, args)
	case "lsusb":
		return delegateCmd("lsusb", args)
	case "lspci":
		return delegateCmd("lspci", args)
	case "dmesg":
		return builtinDmesg(sh, args)
	case "journalctl":
		return delegateCmd("journalctl", args)
	case "systemctl":
		return delegateCmd("systemctl", args)
	case "service":
		return delegateCmd("service", args)
	case "nice":
		return builtinNice(sh, args)
	case "timeout":
		return builtinTimeout(sh, args)
	case "pgrep":
		return builtinPgrep(sh, args)
	case "pkill":
		return delegateCmd("pkill", args)
	case "nohup":
		return builtinNohup(sh, args)
	case "bg", "fg":
		return NewText(fmt.Sprintf("  %s job control not supported in this mode — use a real terminal multiplexer (tmux, screen)", command)), true, nil

	// ── Filesystem ──────────────────────────────────────────────────────────
	case "lsblk":
		return delegateCmd("lsblk", args)
	case "mount":
		return builtinMount(sh, args)
	case "umount":
		return delegateCmd("umount", args)
	case "fdisk":
		return delegateCmd("fdisk", args)
	case "blkid":
		return delegateCmd("blkid", args)
	case "mkfifo":
		return builtinMkfifo(sh, args)
	case "readlink":
		return builtinReadlink(sh, args)
	case "realpath":
		return builtinRealpath(sh, args)
	case "basename":
		return builtinBasename(sh, args)
	case "dirname":
		return builtinDirname(sh, args)
	case "mktemp":
		return builtinMktemp(sh, args)

	// ── Network (extended) ───────────────────────────────────────────────────
	case "ss", "netstat":
		return builtinSS(sh, args)
	case "traceroute", "tracert":
		return delegateCmd("traceroute", args)
	case "mtr":
		return delegateCmd("mtr", args)
	case "openssl":
		return delegateCmd("openssl", args)
	case "ssh":
		return delegateCmd("ssh", args)
	case "scp":
		return delegateCmd("scp", args)
	case "rsync":
		return delegateCmd("rsync", args)
	case "httpget":
		return builtinHttpGet(sh, args)
	case "httppost":
		return builtinHttpPost(sh, args)
	case "jq":
		return builtinJQ(sh, args)

	// ── Text / Data ──────────────────────────────────────────────────────────
	case "xxd":
		return builtinXxd(sh, args)
	case "od":
		return delegateCmd("od", args)
	case "strings":
		return builtinStrings(sh, args)
	case "column":
		return builtinColumn(sh, args)
	case "nl":
		return builtinNl(sh, args)
	case "fold":
		return builtinFold(sh, args)
	case "expand":
		return builtinExpand(sh, args)
	case "unexpand":
		return builtinUnexpand(sh, args)
	case "paste":
		return builtinPaste(sh, args)
	case "join":
		return delegateCmd("join", args)
	case "comm":
		return builtinComm(sh, args)
	case "shuf":
		return builtinShuf(sh, args)
	case "numfmt":
		return builtinNumfmt(sh, args)

	// ── Import / Export extension system ────────────────────────────────────
	case "import":
		return builtinImport(sh, args)
	case "export":
		return builtinExport(sh, args)

	// ── Scripting helpers ────────────────────────────────────────────────────
	case "eval":
		return builtinEval(sh, args)
	case "exec":
		return builtinExec(sh, args)
	case "test", "[":
		return builtinTest(sh, args)
	case "read":
		return builtinRead(sh, args)
	case "mapfile", "readarray":
		return builtinMapfile(sh, args)
	case "declare", "typeset":
		return builtinDeclare(sh, args)
	case "getopts":
		return NewText("getopts: use structured args parsing in Katsh funcs instead"), true, nil

	// ── Fun / misc ───────────────────────────────────────────────────────────
	case "figlet", "toilet":
		return builtinFiglet(sh, args)
	case "matrix":
		return builtinMatrix(sh, args)
	case "lolcat":
		return builtinLolcat(sh, args)
	case "banner2", "drawbox":
		return builtinDrawbox(sh, args)
	case "notify":
		return builtinNotify(sh, args)
	}
	return nil, false, nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  System / Process
// ─────────────────────────────────────────────────────────────────────────────

func builtinLsof(sh *Shell, args []string) (*Result, bool, error) {
	cmd := exec.Command("lsof", args...)
	cmd.Dir = sh.cwd
	out, err := cmd.Output()
	if err != nil {
		// fallback: show open files via /proc on Linux
		if runtime.GOOS == "linux" {
			return parseProcFds(), true, nil
		}
		return nil, true, err
	}
	cols := []string{"command", "pid", "user", "fd", "type", "device", "size", "node", "name"}
	var rows []Row
	lines := strings.Split(string(out), "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 9 {
			continue
		}
		rows = append(rows, Row{
			"command": fields[0], "pid": fields[1], "user": fields[2],
			"fd": fields[3], "type": fields[4], "device": fields[5],
			"size": fields[6], "node": fields[7], "name": strings.Join(fields[8:], " "),
		})
	}
	return NewTable(cols, rows), true, nil
}

func parseProcFds() *Result {
	cols := []string{"pid", "fd", "target"}
	var rows []Row
	procs, _ := filepath.Glob("/proc/*/fd/*")
	seen := 0
	for _, p := range procs {
		if seen > 200 {
			break
		}
		target, err := os.Readlink(p)
		if err != nil {
			continue
		}
		parts := strings.Split(p, "/")
		if len(parts) < 5 {
			continue
		}
		rows = append(rows, Row{"pid": parts[2], "fd": parts[4], "target": target})
		seen++
	}
	return NewTable(cols, rows)
}

func builtinTop(sh *Shell, args []string) (*Result, bool, error) {
	// Snapshot top processes without interactive mode
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("top", "-l", "1", "-stats", "pid,command,cpu,mem")
	} else {
		cmd = exec.Command("top", "-b", "-n", "1")
	}
	cmd.Dir = sh.cwd
	out, err := cmd.Output()
	if err != nil {
		return nil, true, err
	}
	lines := strings.Split(string(out), "\n")
	cols := []string{"pid", "user", "cpu%", "mem%", "command"}
	var rows []Row
	inTable := false
	for _, line := range lines {
		if strings.Contains(line, "PID") {
			inTable = true
			continue
		}
		if !inTable {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		rows = append(rows, Row{
			"pid": fields[0], "user": fields[1],
			"cpu%": fields[2], "mem%": fields[5],
			"command": fields[len(fields)-1],
		})
		if len(rows) >= 20 {
			break
		}
	}
	return NewTable(cols, rows), true, nil
}

func builtinFree(sh *Shell, args []string) (*Result, bool, error) {
	cmd := exec.Command("free", args...)
	out, err := cmd.Output()
	if err != nil {
		return delegateCmd("vm_stat", args)
	}
	cols := []string{"type", "total", "used", "free", "shared", "buff/cache", "available"}
	var rows []Row
	for _, line := range strings.Split(string(out), "\n")[1:] {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		row := Row{"type": strings.TrimSuffix(fields[0], ":")}
		colNames := []string{"total", "used", "free", "shared", "buff/cache", "available"}
		for i, f := range fields[1:] {
			if i < len(colNames) {
				row[colNames[i]] = f
			}
		}
		rows = append(rows, row)
	}
	return NewTable(cols, rows), true, nil
}

func builtinLscpu(sh *Shell, args []string) (*Result, bool, error) {
	if runtime.GOOS == "linux" {
		out, err := os.ReadFile("/proc/cpuinfo")
		if err == nil {
			cols := []string{"field", "value"}
			var rows []Row
			seen := map[string]bool{}
			for _, line := range strings.Split(string(out), "\n") {
				if !strings.Contains(line, ":") {
					continue
				}
				parts := strings.SplitN(line, ":", 2)
				key := strings.TrimSpace(parts[0])
				if seen[key] {
					continue
				}
				seen[key] = true
				rows = append(rows, Row{"field": key, "value": strings.TrimSpace(parts[1])})
			}
			return NewTable(cols, rows), true, nil
		}
	}
	return delegateCmd("lscpu", args)
}

func builtinDmesg(sh *Shell, args []string) (*Result, bool, error) {
	cmd := exec.Command("dmesg", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, true, err
	}
	cols := []string{"timestamp", "level", "message"}
	var rows []Row
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		ts, msg, lv := "", line, "info"
		if strings.HasPrefix(line, "[") {
			end := strings.Index(line, "]")
			if end > 0 {
				ts = line[1:end]
				msg = strings.TrimSpace(line[end+1:])
			}
		}
		if strings.Contains(msg, "error") || strings.Contains(msg, "Error") {
			lv = "error"
		}
		if strings.Contains(msg, "warn") || strings.Contains(msg, "Warn") {
			lv = "warn"
		}
		rows = append(rows, Row{"timestamp": ts, "level": lv, "message": truncStr(msg, 80)})
		if len(rows) >= 100 {
			break
		}
	}
	return NewTable(cols, rows), true, nil
}

func builtinNice(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: nice [-n N] <command> [args...]"), true, nil
	}
	return delegateCmd("nice", args)
}

func builtinTimeout(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) < 2 {
		return NewText("usage: timeout <seconds> <command> [args...]"), true, nil
	}
	return delegateCmd("timeout", args)
}

func builtinPgrep(sh *Shell, args []string) (*Result, bool, error) {
	cmd := exec.Command("pgrep", args...)
	out, err := cmd.Output()
	if err != nil {
		return NewText("no matching processes"), true, nil
	}
	cols := []string{"pid"}
	var rows []Row
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			rows = append(rows, Row{"pid": line})
		}
	}
	return NewTable(cols, rows), true, nil
}

func builtinNohup(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: nohup <command> [args...]"), true, nil
	}
	return delegateCmd("nohup", args)
}

// ─────────────────────────────────────────────────────────────────────────────
//  Filesystem
// ─────────────────────────────────────────────────────────────────────────────

func builtinMount(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) >= 2 {
		return delegateCmd("mount", args)
	}
	// Show mounts as table
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		cmd := exec.Command("mount")
		out, e2 := cmd.Output()
		if e2 != nil {
			return nil, true, e2
		}
		return NewText(string(out)), true, nil
	}
	cols := []string{"device", "mountpoint", "type", "options"}
	var rows []Row
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		rows = append(rows, Row{"device": fields[0], "mountpoint": fields[1], "type": fields[2], "options": fields[3]})
	}
	return NewTable(cols, rows), true, nil
}

func builtinMkfifo(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: mkfifo <name...>"), true, nil
	}
	cols := []string{"name", "status"}
	var rows []Row
	for _, name := range args {
		p := resolvePath(sh.cwd, name)
		err := exec.Command("mkfifo", p).Run()
		st := "created"
		if err != nil {
			st = err.Error()
		}
		rows = append(rows, Row{"name": name, "status": st})
	}
	return NewTable(cols, rows), true, nil
}

func builtinReadlink(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: readlink <path...>"), true, nil
	}
	cols := []string{"path", "target"}
	var rows []Row
	for _, p := range args {
		target, err := os.Readlink(resolvePath(sh.cwd, p))
		if err != nil {
			target = "not a symlink: " + err.Error()
		}
		rows = append(rows, Row{"path": p, "target": target})
	}
	return NewTable(cols, rows), true, nil
}

func builtinRealpath(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: realpath <path...>"), true, nil
	}
	cols := []string{"path", "realpath"}
	var rows []Row
	for _, p := range args {
		rp, err := filepath.Abs(resolvePath(sh.cwd, p))
		if err != nil {
			rp = err.Error()
		}
		rows = append(rows, Row{"path": p, "realpath": rp})
	}
	return NewTable(cols, rows), true, nil
}

func builtinBasename(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: basename <path> [suffix]"), true, nil
	}
	base := filepath.Base(args[0])
	if len(args) > 1 {
		base = strings.TrimSuffix(base, args[1])
	}
	return NewText(base), true, nil
}

func builtinDirname(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: dirname <path>"), true, nil
	}
	return NewText(filepath.Dir(args[0])), true, nil
}

func builtinMktemp(sh *Shell, args []string) (*Result, bool, error) {
	pattern := "Katsh-*"
	dir := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-d":
			dir = os.TempDir()
		case "-p":
			if i+1 < len(args) {
				dir = args[i+1]
				i++
			}
		default:
			pattern = args[i]
		}
	}
	if dir == "" {
		f, err := os.CreateTemp("", pattern)
		if err != nil {
			return nil, true, err
		}
		f.Close()
		return NewText(f.Name()), true, nil
	}
	p, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return nil, true, err
	}
	return NewText(p), true, nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  Network (extended)
// ─────────────────────────────────────────────────────────────────────────────

func builtinSS(sh *Shell, args []string) (*Result, bool, error) {
	// Try system ss/netstat first
	name := "ss"
	if _, err := exec.LookPath("ss"); err != nil {
		name = "netstat"
	}
	cmd := exec.Command(name, append([]string{"-tunapl"}, args...)...)
	out, err := cmd.Output()
	if err != nil {
		// Pure Go fallback: show TCP/UDP listeners
		return builtinSSGo()
	}
	cols := []string{"proto", "state", "local", "remote", "pid"}
	var rows []Row
	for _, line := range strings.Split(string(out), "\n")[1:] {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		rows = append(rows, Row{
			"proto": fields[0], "state": fields[1],
			"local": fields[4], "remote": fields[5],
			"pid": func() string {
				if len(fields) > 6 {
					return fields[6]
				}
				return "-"
			}(),
		})
	}
	return NewTable(cols, rows), true, nil
}

func builtinSSGo() (*Result, bool, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, true, err
	}
	cols := []string{"interface", "address", "flags"}
	var rows []Row
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			rows = append(rows, Row{
				"interface": iface.Name,
				"address":   addr.String(),
				"flags":     iface.Flags.String(),
			})
		}
	}
	return NewTable(cols, rows), true, nil
}

func builtinHttpGet(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: httpget <url>"), true, nil
	}
	url := args[0]
	resp, err := http.Get(url)
	if err != nil {
		return nil, true, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}

	cols := []string{"status", "content-type", "size", "body"}
	ct := resp.Header.Get("Content-Type")
	row := Row{
		"status":       strconv.Itoa(resp.StatusCode) + " " + resp.Status,
		"content-type": ct,
		"size":         strconv.Itoa(len(body)),
		"body":         truncStr(string(body), 200),
	}
	return NewTable(cols, []Row{row}), true, nil
}

func builtinHttpPost(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) < 2 {
		return NewText("usage: httppost <url> <body>"), true, nil
	}
	url := args[0]
	body := strings.NewReader(args[1])
	ct := "application/json"
	for i := 0; i < len(args); i++ {
		if args[i] == "-H" && i+1 < len(args) {
			if strings.HasPrefix(args[i+1], "Content-Type:") {
				ct = strings.TrimSpace(strings.TrimPrefix(args[i+1], "Content-Type:"))
			}
		}
	}
	resp, err := http.Post(url, ct, body)
	if err != nil {
		return nil, true, err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return NewTable([]string{"status", "body"}, []Row{{"status": resp.Status, "body": truncStr(string(out), 500)}}), true, nil
}

func builtinJQ(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) < 2 {
		return NewText("usage: jq <query> <file>"), true, nil
	}
	query := args[0]
	file := resolvePath(sh.cwd, args[1])
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, true, err
	}

	// Very basic jq: support .field, .[N], .[], keys, length
	var parsed interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, true, fmt.Errorf("invalid JSON: %v", err)
	}
	result := jqEval(query, parsed)
	out, _ := json.MarshalIndent(result, "  ", "  ")
	return NewText(string(out)), true, nil
}

func jqEval(query string, data interface{}) interface{} {
	query = strings.TrimSpace(query)
	if query == "." {
		return data
	}
	if query == "keys" {
		if m, ok := data.(map[string]interface{}); ok {
			var keys []string
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			result := make([]interface{}, len(keys))
			for i, k := range keys {
				result[i] = k
			}
			return result
		}
	}
	if query == "length" {
		switch v := data.(type) {
		case []interface{}:
			return len(v)
		case map[string]interface{}:
			return len(v)
		case string:
			return len(v)
		}
	}
	if strings.HasPrefix(query, ".") {
		field := query[1:]
		// Array index: .[N]
		if strings.HasPrefix(field, "[") && strings.HasSuffix(field, "]") {
			n, err := strconv.Atoi(field[1 : len(field)-1])
			if err == nil {
				if arr, ok := data.([]interface{}); ok && n < len(arr) {
					return arr[n]
				}
			}
			return nil
		}
		// .[] iterate
		if field == "[]" {
			return data
		}
		// .field
		if m, ok := data.(map[string]interface{}); ok {
			if strings.Contains(field, ".") {
				parts := strings.SplitN(field, ".", 2)
				return jqEval("."+parts[1], m[parts[0]])
			}
			return m[field]
		}
	}
	return data
}

// ─────────────────────────────────────────────────────────────────────────────
//  Text / Data
// ─────────────────────────────────────────────────────────────────────────────

func builtinXxd(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: xxd <file>"), true, nil
	}
	data, err := os.ReadFile(resolvePath(sh.cwd, args[0]))
	if err != nil {
		return nil, true, err
	}
	cols := []string{"offset", "hex", "ascii"}
	var rows []Row
	for i := 0; i < len(data); i += 16 {
		end := i + 16
		if end > len(data) {
			end = len(data)
		}
		chunk := data[i:end]
		var hexParts, asciiParts []string
		for _, b := range chunk {
			hexParts = append(hexParts, fmt.Sprintf("%02x", b))
			if b >= 32 && b < 127 {
				asciiParts = append(asciiParts, string(rune(b)))
			} else {
				asciiParts = append(asciiParts, ".")
			}
		}
		rows = append(rows, Row{
			"offset": fmt.Sprintf("%08x", i),
			"hex":    strings.Join(hexParts, " "),
			"ascii":  strings.Join(asciiParts, ""),
		})
		if len(rows) > 512 {
			break
		}
	}
	return NewTable(cols, rows), true, nil
}

func builtinStrings(sh *Shell, args []string) (*Result, bool, error) {
	minLen := 4
	files := args
	if len(args) > 1 && args[0] == "-n" {
		fmt.Sscanf(args[1], "%d", &minLen)
		files = args[2:]
	}
	if len(files) == 0 {
		return NewText("usage: strings [-n N] <file>"), true, nil
	}
	cols := []string{"file", "offset", "string"}
	var rows []Row
	for _, f := range files {
		data, err := os.ReadFile(resolvePath(sh.cwd, f))
		if err != nil {
			continue
		}
		var cur strings.Builder
		start := 0
		for i, b := range data {
			if b >= 32 && b < 127 {
				if cur.Len() == 0 {
					start = i
				}
				cur.WriteRune(rune(b))
			} else {
				if cur.Len() >= minLen {
					rows = append(rows, Row{"file": f, "offset": fmt.Sprintf("%d", start), "string": cur.String()})
				}
				cur.Reset()
			}
		}
		if cur.Len() >= minLen {
			rows = append(rows, Row{"file": f, "offset": fmt.Sprintf("%d", start), "string": cur.String()})
		}
	}
	return NewTable(cols, rows), true, nil
}

func builtinColumn(sh *Shell, args []string) (*Result, bool, error) {
	sep := "\t"
	files := args
	for i := 0; i < len(args); i++ {
		if args[i] == "-s" && i+1 < len(args) {
			sep = args[i+1]
			i++
		} else if args[i] == "-t" { /* tab-align */
		} else {
			files = args[i:]
			break
		}
	}
	if len(files) == 0 {
		return NewText("usage: column [-s sep] [-t] <file>"), true, nil
	}
	data, err := os.ReadFile(resolvePath(sh.cwd, files[0]))
	if err != nil {
		return nil, true, err
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return NewText(""), true, nil
	}
	// Auto-detect columns from first line
	header := strings.Split(strings.TrimSpace(lines[0]), sep)
	cols := make([]string, len(header))
	for i, h := range header {
		if h == "" {
			cols[i] = fmt.Sprintf("col%d", i+1)
		} else {
			cols[i] = h
		}
	}
	var rows []Row
	for _, line := range lines[1:] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, sep)
		row := Row{}
		for i, col := range cols {
			if i < len(fields) {
				row[col] = fields[i]
			} else {
				row[col] = ""
			}
		}
		rows = append(rows, row)
	}
	return NewTable(cols, rows), true, nil
}

func builtinNl(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: nl <file>"), true, nil
	}
	data, err := os.ReadFile(resolvePath(sh.cwd, args[0]))
	if err != nil {
		return nil, true, err
	}
	cols := []string{"line", "text"}
	var rows []Row
	for i, line := range strings.Split(string(data), "\n") {
		rows = append(rows, Row{"line": strconv.Itoa(i + 1), "text": line})
	}
	return NewTable(cols, rows), true, nil
}

func builtinFold(sh *Shell, args []string) (*Result, bool, error) {
	width := 80
	files := args
	for i := 0; i < len(args); i++ {
		if (args[i] == "-w" || args[i] == "-b") && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &width)
			i++
		} else {
			files = args[i:]
			break
		}
	}
	if len(files) == 0 {
		return NewText("usage: fold [-w N] <file>"), true, nil
	}
	data, err := os.ReadFile(resolvePath(sh.cwd, files[0]))
	if err != nil {
		return nil, true, err
	}
	var sb strings.Builder
	for _, line := range strings.Split(string(data), "\n") {
		for len(line) > width {
			sb.WriteString(line[:width] + "\n")
			line = line[width:]
		}
		sb.WriteString(line + "\n")
	}
	return NewText(sb.String()), true, nil
}

func builtinExpand(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: expand <file>"), true, nil
	}
	data, err := os.ReadFile(resolvePath(sh.cwd, args[0]))
	if err != nil {
		return nil, true, err
	}
	return NewText(strings.ReplaceAll(string(data), "\t", "    ")), true, nil
}

func builtinUnexpand(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: unexpand <file>"), true, nil
	}
	data, err := os.ReadFile(resolvePath(sh.cwd, args[0]))
	if err != nil {
		return nil, true, err
	}
	return NewText(strings.ReplaceAll(string(data), "    ", "\t")), true, nil
}

func builtinPaste(sh *Shell, args []string) (*Result, bool, error) {
	sep := "\t"
	files := args
	for i := 0; i < len(args); i++ {
		if args[i] == "-d" && i+1 < len(args) {
			sep = args[i+1]
			i++
		} else {
			files = args[i:]
			break
		}
	}
	if len(files) < 2 {
		return NewText("usage: paste [-d sep] <file1> <file2>..."), true, nil
	}
	var allLines [][]string
	maxLen := 0
	for _, f := range files {
		data, err := os.ReadFile(resolvePath(sh.cwd, f))
		if err != nil {
			return nil, true, err
		}
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		allLines = append(allLines, lines)
		if len(lines) > maxLen {
			maxLen = len(lines)
		}
	}
	var sb strings.Builder
	for i := 0; i < maxLen; i++ {
		parts := make([]string, len(allLines))
		for j, lines := range allLines {
			if i < len(lines) {
				parts[j] = lines[i]
			}
		}
		sb.WriteString(strings.Join(parts, sep) + "\n")
	}
	return NewText(sb.String()), true, nil
}

func builtinComm(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) < 2 {
		return NewText("usage: comm <file1> <file2>"), true, nil
	}
	readLines := func(path string) []string {
		data, _ := os.ReadFile(resolvePath(sh.cwd, path))
		return strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	}
	lines1 := readLines(args[0])
	lines2 := readLines(args[1])
	set1 := map[string]bool{}
	for _, l := range lines1 {
		set1[l] = true
	}
	set2 := map[string]bool{}
	for _, l := range lines2 {
		set2[l] = true
	}
	cols := []string{"only_in_1", "only_in_2", "in_both"}
	var rows []Row
	both := map[string]bool{}
	for _, l := range lines1 {
		if set2[l] {
			both[l] = true
		}
	}
	for k := range both {
		rows = append(rows, Row{"only_in_1": "", "only_in_2": "", "in_both": k})
	}
	for _, l := range lines1 {
		if !both[l] {
			rows = append(rows, Row{"only_in_1": l, "only_in_2": "", "in_both": ""})
		}
	}
	for _, l := range lines2 {
		if !both[l] {
			rows = append(rows, Row{"only_in_1": "", "only_in_2": l, "in_both": ""})
		}
	}
	return NewTable(cols, rows), true, nil
}

func builtinShuf(sh *Shell, args []string) (*Result, bool, error) {
	n := -1
	files := args
	for i := 0; i < len(args); i++ {
		if args[i] == "-n" && i+1 < len(args) {
			fmt.Sscanf(args[i+1], "%d", &n)
			i++
		} else {
			files = args[i:]
			break
		}
	}
	if len(files) == 0 {
		return NewText("usage: shuf [-n N] <file>"), true, nil
	}
	data, err := os.ReadFile(resolvePath(sh.cwd, files[0]))
	if err != nil {
		return nil, true, err
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	rand.Shuffle(len(lines), func(i, j int) { lines[i], lines[j] = lines[j], lines[i] })
	if n >= 0 && n < len(lines) {
		lines = lines[:n]
	}
	return NewText(strings.Join(lines, "\n")), true, nil
}

func builtinNumfmt(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: numfmt [--to=iec|si] <number...>"), true, nil
	}
	to := "iec"
	nums := args
	for i, a := range args {
		if strings.HasPrefix(a, "--to=") {
			to = a[5:]
			nums = args[i+1:]
			break
		}
	}
	cols := []string{"input", "formatted"}
	var rows []Row
	for _, n := range nums {
		f, err := strconv.ParseFloat(n, 64)
		if err != nil {
			rows = append(rows, Row{"input": n, "formatted": "error"})
			continue
		}
		var formatted string
		switch to {
		case "iec":
			formatted = fmtBytes(int64(f))
		case "si":
			switch {
			case f >= 1e9:
				formatted = fmt.Sprintf("%.1fG", f/1e9)
			case f >= 1e6:
				formatted = fmt.Sprintf("%.1fM", f/1e6)
			case f >= 1e3:
				formatted = fmt.Sprintf("%.1fK", f/1e3)
			default:
				formatted = n
			}
		default:
			formatted = n
		}
		rows = append(rows, Row{"input": n, "formatted": formatted})
	}
	return NewTable(cols, rows), true, nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  Import / Export — Extension System
//
//  import "file.ksh"              run a local script file
//  import "https://..."           fetch remote script and run it (cached)
//  import "github.com/user/pkg"   fetch from GitHub raw
//  export func|var name           mark as exported (available to child shells)
//  export NAME=value              set env variable (standard shell export)
// ─────────────────────────────────────────────────────────────────────────────

// extensionCacheDir returns the directory for cached extensions.
func extensionCacheDir() string {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.TempDir()
	}
	dir := filepath.Join(home, ".config", "Katsh", "extensions")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func builtinImport(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: import <file|url|github-path>"), true, nil
	}
	path := args[0]
	// Strip surrounding quotes
	path = strings.Trim(path, `"'`)

	var scriptContent []byte
	var err error
	var sourceName string

	switch {
	case strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://"):
		// Remote URL — check cache first
		cacheKey := strings.NewReplacer("/", "_", ":", "-", "?", "_").Replace(path)
		cachePath := filepath.Join(extensionCacheDir(), cacheKey+".ksh")

		// Use cache if fresh (< 24h)
		if info, e := os.Stat(cachePath); e == nil && time.Since(info.ModTime()) < 24*time.Hour {
			scriptContent, err = os.ReadFile(cachePath)
			sourceName = "cached:" + cacheKey
		} else {
			// Fetch
			resp, e := http.Get(path)
			if e != nil {
				return nil, true, fmt.Errorf("import: failed to fetch %s: %v", path, e)
			}
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				return nil, true, fmt.Errorf("import: HTTP %d for %s", resp.StatusCode, path)
			}
			scriptContent, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, true, err
			}
			_ = os.WriteFile(cachePath, scriptContent, 0600)
			sourceName = "remote:" + path
		}

	case strings.Contains(path, "/") && !strings.HasPrefix(path, ".") && !strings.HasPrefix(path, "/"):
		// GitHub shorthand: "user/repo/path" → raw.githubusercontent.com
		url := "https://raw.githubusercontent.com/" + path
		resp, e := http.Get(url)
		if e != nil {
			return nil, true, fmt.Errorf("import: cannot fetch from GitHub: %v", e)
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return nil, true, fmt.Errorf("import: GitHub returned HTTP %d", resp.StatusCode)
		}
		scriptContent, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, true, err
		}
		sourceName = "github:" + path

	default:
		// Local file
		localPath := resolvePath(sh.cwd, path)
		scriptContent, err = os.ReadFile(localPath)
		if err != nil {
			return nil, true, fmt.Errorf("import: cannot read %s: %v", path, err)
		}
		sourceName = "file:" + path
	}

	// Execute script line by line
	lines := strings.Split(string(scriptContent), "\n")
	executed := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		code := sh.execLine(line)
		if code != 0 {
			return nil, true, fmt.Errorf("import: %s: error at: %s (exit %d)", sourceName, line, code)
		}
		executed++
	}

	cols := []string{"source", "lines_executed", "funcs_defined", "vars_set"}
	row := Row{
		"source":         sourceName,
		"lines_executed": strconv.Itoa(executed),
		"funcs_defined":  strconv.Itoa(len(sh.funcs)),
		"vars_set":       strconv.Itoa(len(sh.vars)),
	}
	return NewTable(cols, []Row{row}), true, nil
}

func builtinExport(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		// List all exported vars
		cols := []string{"name", "value"}
		var rows []Row
		for _, env := range os.Environ() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				rows = append(rows, Row{"name": parts[0], "value": parts[1]})
			}
		}
		return NewTable(cols, rows), true, nil
	}

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			// export NAME=value
			parts := strings.SplitN(arg, "=", 2)
			name, val := parts[0], parts[1]
			sh.setVar(name, val)
			_ = os.Setenv(name, val)
		} else if fn, ok := sh.funcs[arg]; ok {
			// export func — mark as exported
			fn.Exported = true
			fmt.Printf("  %s✔ func %q marked as exported%s\n", ansiGreen, arg, ansiReset)
		} else if v, ok := sh.vars[arg]; ok {
			// export var
			_ = os.Setenv(arg, v)
			fmt.Printf("  %s✔ exported %q = %q%s\n", ansiGreen, arg, v, ansiReset)
		} else {
			_ = os.Setenv(arg, "")
		}
	}
	return nil, true, nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  Scripting helpers
// ─────────────────────────────────────────────────────────────────────────────

func builtinEval(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return nil, true, nil
	}
	code := sh.execLine(strings.Join(args, " "))
	if code != 0 {
		return nil, true, fmt.Errorf("eval exited with code %d", code)
	}
	return nil, true, nil
}

func builtinExec(sh *Shell, args []string) (*Result, bool, error) {
	if len(args) == 0 {
		return NewText("usage: exec <command> [args...]"), true, nil
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = sh.cwd
	err := cmd.Run()
	if err != nil {
		return nil, true, err
	}
	return nil, true, nil
}

func builtinTest(sh *Shell, args []string) (*Result, bool, error) {
	// test expression (simplified)
	// Flatten args and eval as condition
	expr := strings.Join(args, " ")
	expr = strings.TrimSuffix(expr, "]")
	if sh.evalCond(expr) {
		return nil, true, nil
	}
	return nil, true, fmt.Errorf("test: false")
}

func builtinRead(sh *Shell, args []string) (*Result, bool, error) {
	varName := "REPLY"
	prompt := ""
	for i := 0; i < len(args); i++ {
		if args[i] == "-p" && i+1 < len(args) {
			prompt = args[i+1]
			i++
		} else if args[i] == "-r" { /* raw mode */
		} else {
			varName = args[i]
		}
	}
	if prompt != "" {
		fmt.Print("  " + prompt)
	}
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		sh.setVar(varName, scanner.Text())
	}
	return nil, true, nil
}

func builtinMapfile(sh *Shell, args []string) (*Result, bool, error) {
	varName := "MAPFILE"
	if len(args) > 0 {
		varName = args[len(args)-1]
	}
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	sh.setVar(varName, sh.makeArray(lines))
	return nil, true, nil
}

func builtinDeclare(sh *Shell, args []string) (*Result, bool, error) {
	// declare -a name (array), declare -i name (integer), declare -x name (export)
	for _, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			if _, ok := sh.vars[arg]; !ok {
				sh.setVar(arg, "")
			}
		}
	}
	return nil, true, nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  Fun / ASCII art
// ─────────────────────────────────────────────────────────────────────────────

func builtinFiglet(sh *Shell, args []string) (*Result, bool, error) {
	text := strings.Join(args, " ")
	if text == "" {
		text = "Katsh"
	}
	// Try system figlet/toilet first
	for _, prog := range []string{"figlet", "toilet"} {
		if _, err := exec.LookPath(prog); err == nil {
			out, _ := exec.Command(prog, text).Output()
			return NewText(string(out)), true, nil
		}
	}
	// Fallback: simple big letter banner using # characters
	var sb strings.Builder
	sb.WriteString("\n")
	for _, line := range bigTextLines(text) {
		sb.WriteString("  " + line + "\n")
	}
	sb.WriteString("\n")
	return NewText(sb.String()), true, nil
}

func bigTextLines(text string) []string {
	// Very simple 3-line ASCII art
	top := ""
	mid := ""
	bot := ""
	for _, ch := range strings.ToUpper(text) {
		switch ch {
		case ' ':
			top += "    "
			mid += "    "
			bot += "    "
		default:
			top += "###  "
			mid += " #   "
			bot += "###  "
		}
	}
	return []string{top, mid, bot}
}

func builtinMatrix(sh *Shell, args []string) (*Result, bool, error) {
	cols := 40
	rows := 12
	chars := "ｱｲｳｴｵｶｷｸｹｺｻｼｽｾｿﾀﾁﾂﾃ0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	runes := []rune(chars)
	var sb strings.Builder
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			ch := runes[r.Intn(len(runes))]
			if j%3 == 0 {
				sb.WriteString(ansiGreen)
			} else {
				sb.WriteString(ansiDarkGreen)
			}
			sb.WriteRune(ch)
		}
		sb.WriteString(ansiReset + "\n")
	}
	return NewText(sb.String()), true, nil
}

func builtinLolcat(sh *Shell, args []string) (*Result, bool, error) {
	if _, err := exec.LookPath("lolcat"); err == nil {
		cmd := exec.Command("lolcat", args...)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		_ = cmd.Run()
		return nil, true, nil
	}
	// Fallback: rainbow colors
	text := strings.Join(args, " ")
	if text == "" {
		text = "Katsh is awesome!"
	}
	colors := []string{ansiRed, ansiYellow, ansiGreen, ansiCyan, ansiBlue, ansiMagenta}
	var sb strings.Builder
	for i, ch := range text {
		sb.WriteString(colors[i%len(colors)])
		sb.WriteRune(ch)
	}
	sb.WriteString(ansiReset)
	return NewText(sb.String()), true, nil
}

func builtinDrawbox(sh *Shell, args []string) (*Result, bool, error) {
	text := strings.Join(args, " ")
	if text == "" {
		text = "Hello from Katsh"
	}
	width := len(text) + 4
	top := "╔" + strings.Repeat("═", width) + "╗"
	mid := "║  " + text + "  ║"
	bot := "╚" + strings.Repeat("═", width) + "╝"
	return NewText(fmt.Sprintf("\n  %s\n  %s\n  %s\n", top, mid, bot)), true, nil
}

func builtinNotify(sh *Shell, args []string) (*Result, bool, error) {
	title := "Katsh"
	msg := strings.Join(args, " ")
	if msg == "" {
		msg = "Done!"
	}
	// Try system notify
	for _, prog := range []string{"notify-send", "osascript", "terminal-notifier"} {
		if _, err := exec.LookPath(prog); err == nil {
			switch prog {
			case "notify-send":
				exec.Command("notify-send", title, msg).Run()
			case "osascript":
				exec.Command("osascript", "-e", fmt.Sprintf(`display notification "%s" with title "%s"`, msg, title)).Run()
			}
			return NewText(okMsg("notification sent")), true, nil
		}
	}
	return NewText(fmt.Sprintf("  🔔 %s: %s", title, msg)), true, nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  Helper: delegate to system command
// ─────────────────────────────────────────────────────────────────────────────

func delegateCmd(name string, args []string) (*Result, bool, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return nil, true, fmt.Errorf("%s: command not found on this system", name)
	}
	cmd := exec.Command(path, args...)
	cmd.Stdin = os.Stdin
	out, err := cmd.CombinedOutput()
	if err != nil && len(out) == 0 {
		return nil, true, err
	}
	return NewText(strings.TrimRight(string(out), "\n")), true, nil
}

// ─────────────────────────────────────────────────────────────────────────────
//  ZIP helper used by the archive commands in builtins.go
// ─────────────────────────────────────────────────────────────────────────────

func zipDir(src, dst string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(src, path)
		fw, err := w.Create(rel)
		if err != nil {
			return err
		}
		rf, err := os.Open(path)
		if err != nil {
			return err
		}
		defer rf.Close()
		_, err = io.Copy(fw, rf)
		return err
	})
}
