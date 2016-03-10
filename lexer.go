package main

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	eof = -1
)
const (
	ERR = iota
	EOF
	VAR
	NUM
	LP
	RP
	PLUS
	MINUS
	MUL
	DIV
	REM
	EXP
	ASSIGN
	SEMI
	LB
	RB
	COMMENT
	COMMA
	CHAR
	STR
	EQ
	NE
	LE
	LT
	GE
	GT
	Keyword
	Return
	Else
	If
	Do
	For
	While
	Char
	Int
)

var key = map[string]int{
	"return": Return,
	"else":   Else,
	"if":     If,
	"do":     Do,
	"for":    For,
	"while":  While,
	"char":   Char,
	"int":    Int,
}

type lexer struct {
	input string
	pos   int
	start int
	width int
}

func newLexer(input string) *lexer {
	return &lexer{
		input: input,
	}
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += l.width
	return r
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) error(msg string) *Token {
	return &Token{
		Type: ERR,
		Text: msg,
	}
}

func (l *lexer) emit(typ int) *Token {
	t := &Token{typ, l.start, l.input[l.start:l.pos]}
	if typ == EOF {
		t.Text = "EOF"
	}
	l.start = l.pos
	return t
}

func (l *lexer) jump(pos int) {
	l.start = pos
	l.pos = pos
}

func (l *lexer) nextToken() *Token {
	for c := l.next(); c != eof; c = l.next() {
		switch c {
		case ' ', '\t', '\n', '\r':
			l.ws()
		case '+':
			return l.emit(PLUS)
		case '-':
			return l.emit(MINUS)
		case '*':
			return l.emit(MUL)
		case '/':
			if next := l.peek(); next == '*' {
				l.next()
				if i := strings.Index(l.input[l.pos:], "*/"); i > 0 {
					l.pos += i + 2
					l.start = l.pos
					break
				} else {
					return l.error("unclosed comment")
				}
			}
			return l.emit(DIV)
		case '%':
			return l.emit(REM)
		case '^':
			return l.emit(EXP)
		case '=':
			if next := l.peek(); next == '=' {
				l.next()
				return l.emit(EQ)
			}
			return l.emit(ASSIGN)
		case '!':
			if next := l.peek(); next == '=' {
				l.next()
				return l.emit(NE)
			}
		case '<':
			if next := l.peek(); next == '=' {
				l.next()
				return l.emit(LE)
			} else {
				return l.emit(LT)
			}
		case '>':
			if next := l.peek(); next == '=' {
				l.next()
				return l.emit(GE)
			} else {
				return l.emit(GT)
			}
		case '"':
			next := l.next()
			for ; next != '"' && next != '\n' && next != eof; next = l.next() {
			}
			if next == '\n' || next == eof {
				return l.error("string literal unterminated")
			} else {
				return l.emit(STR)
			}
		case '\'':
			next := l.next()
			for ; next != '\'' && next != '\n' && next != eof; next = l.next() {
			}
			if next == '\n' || next == eof {
				return l.error("char literal unterminated")
			} else {
				return l.emit(CHAR)
			}
		case ',':
			return l.emit(COMMA)
		case '(':
			return l.emit(LP)
		case ')':
			return l.emit(RP)
		case ';':
			return l.emit(SEMI)
		case '{':
			return l.emit(LB)
		case '}':
			return l.emit(RB)
		default:
			if unicode.IsLetter(c) || c == '_' {
				return varOrKey(l.collect(0))
			} else if unicode.IsDigit(c) {
				return l.collect(1)
			} else {
				return l.error(fmt.Sprintf("invalid character: %c", c))
			}
		}
	}
	return l.emit(EOF)
}

func varOrKey(tok *Token) *Token {
	if key[tok.Text] > Keyword {
		tok.Type = key[tok.Text]
	}
	return tok
}

func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func (l *lexer) collect(typ int) *Token {
	predict := [2]func(r rune) bool{
		isAlphaNumeric,
		unicode.IsDigit,
	}
	for n := l.next(); predict[typ](n); n = l.next() {
	}
	l.backup()
	if typ == 0 {
		return l.emit(VAR)
	} else {
		return l.emit(NUM)
	}
}

func (l *lexer) ws() {
	for c := l.next(); c == ' ' || c == '\t' || c == '\n' || c == '\r'; c = l.next() {
	}
	l.backup()
	l.start = l.pos
}

type Token struct {
	Type int
	Pos  int
	Text string
}

func (t Token) String() string {
	switch {
	case t.Type == EOF:
		return "EOF"
	case t.Type == ERR:
		return t.Text
	case t.Type > Keyword:
		return fmt.Sprintf("<%s>", t.Text)
	case len(t.Text) > 10:
		return fmt.Sprintf("%.10q...", t.Text)
	}
	return fmt.Sprintf("%q", t.Text)
}
