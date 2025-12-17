package lexer

import (
	"strings"
	"testing"

	"github.com/donmahallem/go_gcode/token"
)

func TestStreamLexerMatchesInMemory(t *testing.T) {
	input := `
M104 S75 ; comment
{if temp > 50}
  M106
{endif}
G1 X{10+20}
`

	l1 := New(input)
	l2 := NewFromReader(strings.NewReader(input))

	for {
		t1 := l1.NextToken()
		t2 := l2.NextToken()

		if t1.Type != t2.Type || t1.Literal != t2.Literal {
			t.Fatalf("tokens differ: in-memory=%v/%q stream=%v/%q", t1.Type, t1.Literal, t2.Type, t2.Literal)
		}

		if t1.Type == token.EOF {
			break
		}
	}
}
