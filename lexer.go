package main

import (
	"fmt"
	"strings"
)

// ─────────────────────────────────────────────────────────────────────────────
//  Lexer — tokenises Katsh source with exact line + column tracking.
//
//  Token types:
//   WORD      bare word / identifier
//   NUMBER    numeric literal
//   STRING    quoted string (single or double)
//   BACKTICK  `...` subshell
//   OP        operator  = == != <= >= < > += -= *= /= %= ++ --
//   PUNCT     punctuation  { } ( ) [ ] : ; , . |
//   KEYWORD   if elif else for while func match case default
//              return break continue import export print do until
//   COMMENT   # or // or ///
//   EOF
// ─────────────────────────────────────────────────────────────────────────────

type TokKind int

const (
	TokWord TokKind = iota
	TokNumber
	TokString
	TokBacktick
	TokOp
	TokPunct
	TokKeyword
	TokComment
	TokEOF
	TokUnknown
)

func (k TokKind) String() string {
	return [...]string{
		"WORD", "NUMBER", "STRING", "BACKTICK", "OP",
		"PUNCT", "KEYWORD", "COMMENT", "EOF", "UNKNOWN",
	}[k]
}

// Token is one lexical unit.
type Token struct {
	Kind TokKind
	Text string
	Line int // 1-based
	Col  int // 0-based byte offset on the line
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%q)@%d:%d", t.Kind, t.Text, t.Line, t.Col)
}

// keywords is the set of reserved words.
var keywords = map[string]bool{
	"if": true, "elif": true, "else": true, "fi": true,
	"for": true, "while": true, "do": true, "done": true, "until": true,
	"func": true, "return": true, "in": true,
	"break": true, "continue": true,
	"match": true, "case": true, "default": true,
	"try": true, "catch": true, "throw": true,
	"import": true, "export": true,
	"print": true, "println": true,
	"true": true, "false": true, "null": true, "nil": true,
	"and": true, "or": true, "not": true,
	"range": true, "select": true,
}

// Lexer holds the lexer state.
type Lexer struct {
	src    []rune
	pos    int
	line   int
	col    int
	Tokens []Token
	Errors []LexError
}

type LexError struct {
	Msg  string
	Line int
	Col  int
}

// Lex tokenises the full source string and returns all tokens.
func Lex(source string) *Lexer {
	l := &Lexer{src: []rune(source), line: 1, col: 0}
	l.run()
	return l
}

// TokenAt returns the token at index i (or EOF token if out of range).
func (l *Lexer) TokenAt(i int) Token {
	if i < len(l.Tokens) {
		return l.Tokens[i]
	}
	return Token{Kind: TokEOF, Line: l.line, Col: l.col}
}

// ─────────────────────────────────────────────────────────────────────────────
//  Lexer internals
// ─────────────────────────────────────────────────────────────────────────────

func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

func (l *Lexer) peek2() rune {
	if l.pos+1 >= len(l.src) {
		return 0
	}
	return l.src[l.pos+1]
}

func (l *Lexer) advance() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	ch := l.src[l.pos]
	l.pos++
	if ch == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
	return ch
}

func (l *Lexer) startToken() (int, int) {
	return l.line, l.col
}

func (l *Lexer) emit(kind TokKind, text string, line, col int) {
	l.Tokens = append(l.Tokens, Token{Kind: kind, Text: text, Line: line, Col: col})
}

func (l *Lexer) run() {
	for l.pos < len(l.src) {
		l.scanOne()
	}
	l.emit(TokEOF, "", l.line, l.col)
}

func (l *Lexer) scanOne() {
	ch := l.peek()

	// Whitespace — skip but don't emit
	if ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n' {
		l.advance()
		return
	}

	line, col := l.startToken()

	// Line comment  # or //  or  ///
	if ch == '#' && l.peek2() != '=' {
		var sb strings.Builder
		for l.peek() != '\n' && l.peek() != 0 {
			sb.WriteRune(l.advance())
		}
		l.emit(TokComment, sb.String(), line, col)
		return
	}
	if ch == '/' && l.peek2() == '/' {
		var sb strings.Builder
		for l.peek() != '\n' && l.peek() != 0 {
			sb.WriteRune(l.advance())
		}
		l.emit(TokComment, sb.String(), line, col)
		return
	}

	// String literals
	if ch == '"' || ch == '\'' {
		l.scanString(ch, line, col)
		return
	}

	// Backtick subshell
	if ch == '`' {
		l.scanBacktick(line, col)
		return
	}

	// Numbers
	if ch >= '0' && ch <= '9' {
		l.scanNumber(line, col)
		return
	}

	// Identifiers / keywords
	if isIdentStart(ch) {
		l.scanWord(line, col)
		return
	}

	// Two-char operators
	ch2 := l.peek2()
	switch {
	case ch == '=' && ch2 == '=':
		l.advance()
		l.advance()
		l.emit(TokOp, "==", line, col)
		return
	case ch == '!' && ch2 == '=':
		l.advance()
		l.advance()
		l.emit(TokOp, "!=", line, col)
		return
	case ch == '<' && ch2 == '=':
		l.advance()
		l.advance()
		l.emit(TokOp, "<=", line, col)
		return
	case ch == '>' && ch2 == '=':
		l.advance()
		l.advance()
		l.emit(TokOp, ">=", line, col)
		return
	case ch == '+' && ch2 == '=':
		l.advance()
		l.advance()
		l.emit(TokOp, "+=", line, col)
		return
	case ch == '-' && ch2 == '=':
		l.advance()
		l.advance()
		l.emit(TokOp, "-=", line, col)
		return
	case ch == '*' && ch2 == '=':
		l.advance()
		l.advance()
		l.emit(TokOp, "*=", line, col)
		return
	case ch == '/' && ch2 == '=':
		l.advance()
		l.advance()
		l.emit(TokOp, "/=", line, col)
		return
	case ch == '%' && ch2 == '=':
		l.advance()
		l.advance()
		l.emit(TokOp, "%=", line, col)
		return
	case ch == '+' && ch2 == '+':
		l.advance()
		l.advance()
		l.emit(TokOp, "++", line, col)
		return
	case ch == '-' && ch2 == '-':
		l.advance()
		l.advance()
		l.emit(TokOp, "--", line, col)
		return
	case ch == '&' && ch2 == '&':
		l.advance()
		l.advance()
		l.emit(TokOp, "&&", line, col)
		return
	case ch == '|' && ch2 == '|':
		l.advance()
		l.advance()
		l.emit(TokOp, "||", line, col)
		return
	case ch == '-' && ch2 == '>':
		l.advance()
		l.advance()
		l.emit(TokOp, "->", line, col)
		return
	case ch == '=' && ch2 == '>':
		l.advance()
		l.advance()
		l.emit(TokOp, "=>", line, col)
		return
	case ch == '.' && ch2 == '.':
		l.advance()
		l.advance()
		l.emit(TokOp, "..", line, col)
		return
	case ch == '#' && ch2 == '=':
		l.advance()
		l.advance()
		// consume optional key
		var sb strings.Builder
		sb.WriteString("#=")
		for isIdentChar(l.peek()) {
			sb.WriteRune(l.advance())
		}
		l.emit(TokOp, sb.String(), line, col)
		return
	}

	// Single-char
	switch ch {
	case '=', '<', '>', '+', '-', '*', '/', '%', '!', '~', '^', '&':
		l.advance()
		l.emit(TokOp, string(ch), line, col)
	case '{', '}', '(', ')', '[', ']', ':', ';', ',', '.', '|', '@', '$':
		l.advance()
		l.emit(TokPunct, string(ch), line, col)
	default:
		l.advance()
		l.emit(TokUnknown, string(ch), line, col)
		l.Errors = append(l.Errors, LexError{
			Msg:  fmt.Sprintf("unexpected character %q", ch),
			Line: line,
			Col:  col,
		})
	}
}

func (l *Lexer) scanString(quote rune, line, col int) {
	l.advance() // consume opening quote
	var sb strings.Builder
	sb.WriteRune(quote)
	closed := false
	for l.pos < len(l.src) {
		ch := l.peek()
		if ch == '\\' {
			l.advance()
			esc := l.advance()
			switch esc {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			default:
				sb.WriteRune(esc)
			}
			continue
		}
		if ch == quote {
			l.advance()
			sb.WriteRune(quote)
			closed = true
			break
		}
		sb.WriteRune(l.advance())
	}
	if !closed {
		l.Errors = append(l.Errors, LexError{
			Msg:  fmt.Sprintf("unterminated string starting with %q", quote),
			Line: line,
			Col:  col,
		})
	}
	l.emit(TokString, sb.String(), line, col)
}

func (l *Lexer) scanBacktick(line, col int) {
	l.advance() // consume `
	var sb strings.Builder
	sb.WriteRune('`')
	closed := false
	depth := 1
	for l.pos < len(l.src) {
		ch := l.peek()
		if ch == '`' {
			depth--
			l.advance()
			sb.WriteRune('`')
			if depth == 0 {
				closed = true
				break
			}
		} else {
			sb.WriteRune(l.advance())
		}
	}
	if !closed {
		l.Errors = append(l.Errors, LexError{
			Msg:  "unterminated backtick expression",
			Line: line,
			Col:  col,
		})
	}
	l.emit(TokBacktick, sb.String(), line, col)
}

func (l *Lexer) scanNumber(line, col int) {
	var sb strings.Builder
	for ch := l.peek(); (ch >= '0' && ch <= '9') || ch == '.' || ch == '_'; ch = l.peek() {
		if ch == '_' {
			l.advance()
			continue
		}
		sb.WriteRune(l.advance())
	}
	// Optional suffix e.g. 1k 1M
	if ch := l.peek(); ch == 'k' || ch == 'K' || ch == 'm' || ch == 'M' || ch == 'g' || ch == 'G' {
		sb.WriteRune(l.advance())
	}
	l.emit(TokNumber, sb.String(), line, col)
}

func (l *Lexer) scanWord(line, col int) {
	var sb strings.Builder
	// Handle $VAR
	if l.peek() == '$' {
		sb.WriteRune(l.advance())
	}
	for isIdentChar(l.peek()) {
		sb.WriteRune(l.advance())
	}
	text := sb.String()
	kind := TokWord
	if keywords[strings.ToLower(text)] {
		kind = TokKeyword
	}
	l.emit(kind, text, line, col)
}

func isIdentStart(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == '$'
}

func isIdentChar(ch rune) bool {
	return isIdentStart(ch) || (ch >= '0' && ch <= '9')
}

// ─────────────────────────────────────────────────────────────────────────────
//  Lex-based error position finder
//
//  Given source code and a character/pattern to locate, returns (line, col).
// ─────────────────────────────────────────────────────────────────────────────

// FindErrorPos scans tokens to find the best match for a syntax problem.
// Returns (line, col, excerpt).
func FindErrorPos(source, pattern string) (line, col int, excerpt string) {
	lx := Lex(source)
	// Look for matching token text
	for _, tok := range lx.Tokens {
		if strings.Contains(tok.Text, pattern) {
			return tok.Line, tok.Col, buildExcerpt(source, tok.Line, tok.Col)
		}
	}
	// Fallback: scan lex errors
	for _, le := range lx.Errors {
		return le.Line, le.Col, buildExcerpt(source, le.Line, le.Col)
	}
	return 1, 0, source
}

// buildExcerpt extracts a single source line by line number.
func buildExcerpt(source string, line, col int) string {
	lines := strings.Split(source, "\n")
	if line-1 < len(lines) {
		return lines[line-1]
	}
	return source
}

// LexAndLocate locates the first lex error or unexpected token in source.
// Returns a ShellError with precise position info.
func LexAndLocate(source string) *ShellError {
	lx := Lex(source)
	if len(lx.Errors) == 0 {
		return nil
	}
	le := lx.Errors[0]
	excerpt := buildExcerpt(source, le.Line, le.Col)
	return &ShellError{
		Code:    "E002",
		Kind:    "SyntaxError",
		Message: le.Msg,
		Source:  excerpt,
		Line:    le.Line,
		Col:     le.Col,
		Hint:    "Check for mismatched quotes, brackets, or unknown characters",
	}
}

// ParseLocate scans a script body for common syntax problems and returns
// the ShellError with exact line+col from the lexer.
func ParseLocate(source, problem string) *ShellError {
	lx := Lex(source)
	lines := strings.Split(source, "\n")

	// Search tokens for the problem keyword
	for _, tok := range lx.Tokens {
		if strings.EqualFold(tok.Text, problem) || strings.Contains(tok.Text, problem) {
			excerpt := ""
			if tok.Line-1 < len(lines) {
				excerpt = lines[tok.Line-1]
			}
			return &ShellError{
				Code:    "E002",
				Kind:    "SyntaxError",
				Message: fmt.Sprintf("unexpected %q", tok.Text),
				Source:  excerpt,
				Line:    tok.Line,
				Col:     tok.Col,
			}
		}
	}
	return errSyntax(problem, source, 0)
}
