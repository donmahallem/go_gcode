package lexer

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/donmahallem/go_gcode/token"
)

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
	line         int
	column       int

	mode       int
	braceStack []byte // Stack to track nested delimiters
}

// TokenSource represents an object that can produce tokens and be reset.
// Implemented by both the in-memory Lexer and the StreamLexer.
type TokenSource interface {
	NextToken() token.Token
	Reset()
}

// StreamLexer tokenizes input from an io.Reader without materializing the whole
// input string. It implements TokenSource for parser compatibility.
type StreamLexer struct {
	r          *bufio.Reader
	ch         byte
	line       int
	column     int
	mode       int
	braceStack []byte
}

const (
	MODE_DATA = iota
	MODE_CODE
)

func New(input string) *Lexer {
	l := &Lexer{
		input:      input,
		line:       1,
		column:     0,
		mode:       MODE_DATA,
		braceStack: make([]byte, 0),
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}

	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
	l.column += 1
}

func (l *Lexer) Reset() {
	l.position = 0
	l.readPosition = 0
	l.ch = 0
	l.line = 1
	l.column = 0
	l.mode = MODE_DATA
	l.braceStack = make([]byte, 0)
	l.readChar()
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// --- StreamLexer helpers ---

func NewFromReader(r io.Reader) *StreamLexer {
	l := &StreamLexer{
		r:          bufio.NewReader(r),
		line:       1,
		column:     0,
		mode:       MODE_DATA,
		braceStack: make([]byte, 0),
	}
	l.readChar()
	return l
}

func (l *StreamLexer) readChar() {
	if l.ch == '\n' {
		l.line++
		l.column = 0
	}
	b, err := l.r.ReadByte()
	if err != nil {
		l.ch = 0
	} else {
		l.ch = b
	}
	l.column += 1
}

func (l *StreamLexer) Reset() {
	l.mode = MODE_DATA
	l.braceStack = make([]byte, 0)
}

func (l *StreamLexer) peekChar() byte {
	b, err := l.r.Peek(1)
	if err != nil || len(b) == 0 {
		return 0
	}
	return b[0]
}

func (l *StreamLexer) skipWhitespaceGCode() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *StreamLexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		if l.ch == '\n' {
			l.line++
			l.column = 0
		}
		l.readChar()
	}
}

func (l *StreamLexer) readComment() token.Token {
	startCol := l.column
	// consume ';'
	l.readChar()
	var sb strings.Builder
	sb.WriteByte(';')
	for l.ch != '\n' && l.ch != 0 {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	return token.Token{Type: token.COMMENT, Literal: sb.String(), Line: l.line, Column: startCol}
}

func (l *StreamLexer) readIdentifier() string {
	var sb strings.Builder
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '.' {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	return sb.String()
}

func (l *StreamLexer) readGCodeIdentifier() string {
	var sb strings.Builder
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '.' || l.ch == '-' {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	return sb.String()
}

func (l *StreamLexer) readNumber() string {
	var sb strings.Builder
	for isDigit(l.ch) {
		sb.WriteByte(l.ch)
		l.readChar()
	}
	return sb.String()
}

func (l *StreamLexer) readString() string {
	l.readChar()
	var sb strings.Builder
	for {
		if l.ch == '"' || l.ch == 0 {
			break
		}
		sb.WriteByte(l.ch)
		l.readChar()
	}
	return sb.String()
}

func (l *StreamLexer) newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *StreamLexer) pushStack(b byte) {
	l.braceStack = append(l.braceStack, b)
}

func (l *StreamLexer) popStack(expected byte) bool {
	if len(l.braceStack) == 0 {
		return false
	}
	top := l.braceStack[len(l.braceStack)-1]
	if top == expected {
		l.braceStack = l.braceStack[:len(l.braceStack)-1]
		return true
	}
	return false
}

func (l *StreamLexer) nextTokenInCode() token.Token {
	l.skipWhitespace()

	var tok token.Token
	startLine := l.line
	startCol := l.column

	switch l.ch {
	case '{':
		tok = l.newToken(token.LBRACE, l.ch)
		l.pushStack('{')
	case '}':
		if l.popStack('{') {
			tok = l.newToken(token.RBRACE, l.ch)
			if len(l.braceStack) == 0 {
				l.mode = MODE_DATA
			}
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	case '[':
		tok = l.newToken(token.LBRACKET, l.ch)
		l.pushStack('[')
	case ']':
		if l.popStack('[') {
			tok = l.newToken(token.RBRACKET, l.ch)
			if len(l.braceStack) == 0 {
				l.mode = MODE_DATA
			}
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	case '(':
		tok = l.newToken(token.LPAREN, l.ch)
	case ')':
		tok = l.newToken(token.RPAREN, l.ch)
	case '+':
		tok = l.newToken(token.PLUS, l.ch)
	case '-':
		tok = l.newToken(token.MINUS, l.ch)
	case '/':
		tok = l.newToken(token.SLASH, l.ch)
	case '*':
		tok = l.newToken(token.ASTERISK, l.ch)
	case '<':
		tok = l.newToken(token.LT, l.ch)
	case '>':
		tok = l.newToken(token.GT, l.ch)
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = l.newToken(token.ASSIGN, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = l.newToken(token.BANG, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.OR, Literal: literal}
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.AND, Literal: literal}
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		l.readChar() // consume closing quote
		tok.Line = startLine
		tok.Column = startCol
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.Line = startLine
			tok.Column = startCol
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT // Simplified, could be FLOAT
			tok.Literal = l.readNumber()
			// Check for dot for float
			if l.ch == '.' {
				l.readChar() // consume dot
				tok.Literal += "." + l.readNumber()
				tok.Type = token.FLOAT
			}
			tok.Line = startLine
			tok.Column = startCol
			return tok
		} else {
			tok = l.newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	tok.Line = startLine
	tok.Column = startCol
	return tok
}

// NextToken implements tokenization for streaming lexer.
func (l *StreamLexer) NextToken() token.Token {
	if l.mode == MODE_DATA {
		l.skipWhitespaceGCode()

		if l.ch == 0 {
			return token.Token{Type: token.EOF, Literal: "", Line: l.line, Column: l.column}
		}

		if l.ch == '\n' {
			tok := token.Token{Type: token.NEWLINE, Literal: "\n", Line: l.line, Column: l.column}
			l.readChar()
			return tok
		}

		if l.ch == ';' {
			return l.readComment()
		}

		if l.ch == '{' || l.ch == '[' {
			l.mode = MODE_CODE
			return l.nextTokenInCode()
		}

		if isLetter(l.ch) {
			startLine := l.line
			startCol := l.column
			literal := l.readGCodeIdentifier()
			return token.Token{Type: token.IDENT, Literal: literal, Line: startLine, Column: startCol}
		}

		if isDigit(l.ch) || l.ch == '.' {
			startLine := l.line
			startCol := l.column
			literal := l.readNumber()
			var tokType token.TokenType = token.INT
			if l.ch == '.' {
				l.readChar() // consume dot
				literal += "." + l.readNumber()
				tokType = token.FLOAT
			}
			return token.Token{Type: tokType, Literal: literal, Line: startLine, Column: startCol}
		}

		if l.ch == '-' {
			tok := token.Token{Type: token.MINUS, Literal: "-", Line: l.line, Column: l.column}
			l.readChar()
			return tok
		}

		// Fallback for unknown chars in DATA mode (maybe treat as text or illegal?)
		// For now, let's treat as ILLEGAL or just consume as char
		tok := l.newToken(token.ILLEGAL, l.ch)
		l.readChar()
		return tok
	}

	// MODE_CODE
	return l.nextTokenInCode()
}

func (l *Lexer) NextToken() token.Token {
	if l.mode == MODE_DATA {
		l.skipWhitespaceGCode()

		if l.ch == 0 {
			return token.Token{Type: token.EOF, Literal: "", Line: l.line, Column: l.column}
		}

		if l.ch == '\n' {
			tok := token.Token{Type: token.NEWLINE, Literal: "\n", Line: l.line, Column: l.column}
			l.readChar()
			return tok
		}

		if l.ch == ';' {
			return l.readComment()
		}

		if l.ch == '{' || l.ch == '[' {
			l.mode = MODE_CODE
			return l.nextTokenInCode()
		}

		if isLetter(l.ch) {
			startLine := l.line
			startCol := l.column
			literal := l.readGCodeIdentifier()
			return token.Token{Type: token.IDENT, Literal: literal, Line: startLine, Column: startCol}
		}

		if isDigit(l.ch) || l.ch == '.' {
			startLine := l.line
			startCol := l.column
			literal, tokType := l.readNumberWithFraction()
			return token.Token{Type: tokType, Literal: literal, Line: startLine, Column: startCol}
		}

		if l.ch == '-' {
			tok := token.Token{Type: token.MINUS, Literal: "-", Line: l.line, Column: l.column}
			l.readChar()
			return tok
		}

		// Fallback for unknown chars in DATA mode (maybe treat as text or illegal?)
		// For now, let's treat as ILLEGAL or just consume as char
		tok := newToken(token.ILLEGAL, l.ch)
		l.readChar()
		return tok
	}

	return l.nextTokenInCode()
}

func (l *Lexer) skipWhitespaceGCode() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) nextTokenInCode() token.Token {
	l.skipWhitespace()

	var tok token.Token
	startLine := l.line
	startCol := l.column

	switch l.ch {
	case '{':
		tok = newToken(token.LBRACE, l.ch)
		l.pushStack('{')
	case '}':
		if l.popStack('{') {
			tok = newToken(token.RBRACE, l.ch)
			if len(l.braceStack) == 0 {
				l.mode = MODE_DATA
			}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '[':
		tok = newToken(token.LBRACKET, l.ch)
		l.pushStack('[')
	case ']':
		if l.popStack('[') {
			tok = newToken(token.RBRACKET, l.ch)
			if len(l.braceStack) == 0 {
				l.mode = MODE_DATA
			}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case '=':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.EQ, Literal: literal}
		} else {
			tok = newToken(token.ASSIGN, l.ch)
		}
	case '!':
		if l.peekChar() == '=' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.NOT_EQ, Literal: literal}
		} else {
			tok = newToken(token.BANG, l.ch)
		}
	case '|':
		if l.peekChar() == '|' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.OR, Literal: literal}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '&':
		if l.peekChar() == '&' {
			ch := l.ch
			l.readChar()
			literal := string(ch) + string(l.ch)
			tok = token.Token{Type: token.AND, Literal: literal}
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		l.readChar() // consume closing quote
		tok.Line = startLine
		tok.Column = startCol
		return tok
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.Line = startLine
			tok.Column = startCol
			return tok
		} else if isDigit(l.ch) || l.ch == '.' {
			tok.Literal, tok.Type = l.readNumberWithFraction()
			tok.Line = startLine
			tok.Column = startCol
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}

	l.readChar()
	tok.Line = startLine
	tok.Column = startCol
	return tok
}

func (l *Lexer) readComment() token.Token {
	startPos := l.position
	startCol := l.column
	// consume ';'
	l.readChar()
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	literal := l.input[startPos:l.position]
	return token.Token{Type: token.COMMENT, Literal: literal, Line: l.line, Column: startCol}
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '.' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readGCodeIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '.' || l.ch == '-' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// readNumberWithFraction reads an integer or float (supports leading '.' like ".5").
// Returns the literal and the corresponding token type. A lone dot (".")
// is treated as ILLEGAL.
func (l *Lexer) readNumberWithFraction() (string, token.TokenType) {
	var buf bytes.Buffer

	// Leading dot (e.g., .5)
	if l.ch == '.' {
		buf.WriteByte('.')
		l.readChar() // consume dot
		if !isDigit(l.ch) {
			// lone dot is illegal
			return buf.String(), token.ILLEGAL
		}
		for isDigit(l.ch) {
			buf.WriteByte(l.ch)
			l.readChar()
		}
		return buf.String(), token.FLOAT
	}

	// Integer part
	for isDigit(l.ch) {
		buf.WriteByte(l.ch)
		l.readChar()
	}

	// Fractional part
	if l.ch == '.' {
		buf.WriteByte('.')
		l.readChar() // consume dot
		// allow trailing dot (e.g., "123.") to be considered FLOAT for compatibility
		for isDigit(l.ch) {
			buf.WriteByte(l.ch)
			l.readChar()
		}
		return buf.String(), token.FLOAT
	}

	return buf.String(), token.INT
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func (l *Lexer) pushStack(b byte) {
	l.braceStack = append(l.braceStack, b)
}

func (l *Lexer) popStack(expected byte) bool {
	if len(l.braceStack) == 0 {
		return false
	}
	top := l.braceStack[len(l.braceStack)-1]
	if top == expected {
		l.braceStack = l.braceStack[:len(l.braceStack)-1]
		return true
	}
	return false
}
