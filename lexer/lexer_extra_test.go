package lexer

import (
	"testing"

	"github.com/donmahallem/go_gcode/token"
)

func TestNumberParsingEdgeCases(t *testing.T) {
	cases := []struct {
		input    string
		expected []struct {
			tt  token.TokenType
			lit string
		}
	}{
		{".", []struct {
			tt  token.TokenType
			lit string
		}{{token.ILLEGAL, "."}, {token.EOF, ""}}},
		{".5", []struct {
			tt  token.TokenType
			lit string
		}{{token.FLOAT, ".5"}, {token.EOF, ""}}},
		{"123.", []struct {
			tt  token.TokenType
			lit string
		}{{token.FLOAT, "123."}, {token.EOF, ""}}},
		{"-10", []struct {
			tt  token.TokenType
			lit string
		}{{token.MINUS, "-"}, {token.INT, "10"}, {token.EOF, ""}}},
	}

	for _, c := range cases {
		l := New(c.input)
		for i, exp := range c.expected {
			tok := l.NextToken()
			if tok.Type != exp.tt || tok.Literal != exp.lit {
				t.Fatalf("case %q[%d]: expected=%v/%q got=%v/%q", c.input, i, exp.tt, exp.lit, tok.Type, tok.Literal)
			}
		}
	}
}

func TestUnmatchedBraces(t *testing.T) {
	l := New("}")
	tok := l.NextToken()
	if tok.Type != token.ILLEGAL || tok.Literal != "}" {
		t.Fatalf("expected ILLEGAL '}' got %v/%q", tok.Type, tok.Literal)
	}
}

func TestNewlineColumnPositions(t *testing.T) {
	l := New("A\nB")
	// A
	tok := l.NextToken()
	if tok.Type != token.IDENT || tok.Literal != "A" || tok.Line != 1 || tok.Column != 1 {
		t.Fatalf("A token wrong: %+v", tok)
	}
	// newline
	tok = l.NextToken()
	if tok.Type != token.NEWLINE || tok.Literal != "\n" || tok.Line != 1 || tok.Column != 2 {
		t.Fatalf("NL token wrong: %+v", tok)
	}
	// B
	tok = l.NextToken()
	if tok.Type != token.IDENT || tok.Literal != "B" || tok.Line != 2 || tok.Column != 1 {
		t.Fatalf("B token wrong: %+v", tok)
	}
}
