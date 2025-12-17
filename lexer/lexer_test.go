package lexer

import (
	"testing"

	"github.com/donmahallem/go_gcode/token"
)

func TestNextToken(t *testing.T) {
	input := `
M104 S75 ; comment
{if temp > 50}
  M106
{endif}
G1 X{10+20}
`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.NEWLINE, "\n"},
		{token.IDENT, "M104"},
		{token.IDENT, "S75"},
		{token.COMMENT, "; comment"},
		{token.NEWLINE, "\n"},
		{token.LBRACE, "{"},
		{token.IF, "if"},
		{token.IDENT, "temp"},
		{token.GT, ">"},
		{token.INT, "50"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "M106"},
		{token.NEWLINE, "\n"},
		{token.LBRACE, "{"},
		{token.ENDIF, "endif"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "G1"},
		{token.IDENT, "X"},
		{token.LBRACE, "{"},
		{token.INT, "10"},
		{token.PLUS, "+"},
		{token.INT, "20"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\n"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestNextToken_Complex(t *testing.T) {
	input := `
{if filament_type[0] == "PLA"}
G1 X10.5
{elsif val != 10}
{else}
{endif}
`
	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.NEWLINE, "\n"},
		{token.LBRACE, "{"},
		{token.IF, "if"},
		{token.IDENT, "filament_type"},
		{token.LBRACKET, "["},
		{token.INT, "0"},
		{token.RBRACKET, "]"},
		{token.EQ, "=="},
		{token.STRING, "PLA"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\n"},
		{token.IDENT, "G1"},
		{token.IDENT, "X10.5"},
		{token.NEWLINE, "\n"},
		{token.LBRACE, "{"},
		{token.ELSIF, "elsif"},
		{token.IDENT, "val"},
		{token.NOT_EQ, "!="},
		{token.INT, "10"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\n"},
		{token.LBRACE, "{"},
		{token.ELSE, "else"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\n"},
		{token.LBRACE, "{"},
		{token.ENDIF, "endif"},
		{token.RBRACE, "}"},
		{token.NEWLINE, "\n"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}
