package token

type TokenType string

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT  = "IDENT" // add, foobar, x, y, ...
	INT    = "INT"   // 1343456
	FLOAT  = "FLOAT" // 1.23
	STRING = "STRING"

	// Operators and delimiters
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"

	LT = "<"
	GT = ">"

	EQ     = "=="
	NOT_EQ = "!="
	OR     = "||"
	AND    = "&&"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	IF    = "IF"
	ELSE  = "ELSE"
	ELSIF = "ELSIF" // or ELSEIF
	ENDIF = "ENDIF"
	TRUE  = "TRUE"
	FALSE = "FALSE"

	// GCode specific
	COMMENT = "COMMENT"
	TEXT    = "TEXT" // Raw GCode text
	NEWLINE = "NEWLINE"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

var keywords = map[string]TokenType{
	"if":    IF,
	"else":  ELSE,
	"elsif": ELSIF,
	"endif": ENDIF,
	"true":  TRUE,
	"false": FALSE,
}

func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
