package parser

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/donmahallem/go_gcode/lexer"
)

func TestParseNextAndStream(t *testing.T) {
	input := `
M104 S75 ; comment
{if true}
  M106
{endif}
G1 X{10+20}
`

	// Compare number of nodes parsed by ParseNext vs ParseProgram
	l := lexer.New(input)
	p := New(l)
	prog := p.ParseProgram()

	l2 := lexer.NewFromReader(strings.NewReader(input))
	p2 := New(l2)
	count := 0
	for {
		n, err := p2.ParseNext()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("ParseNext error: %v", err)
		}
		if n != nil {
			count++
		}
	}

	if count != len(prog.Nodes) {
		t.Fatalf("ParseNext node count %d != ParseProgram %d", count, len(prog.Nodes))
	}

	// Now test ParseStream produces same output as emitting full AST
	var buf1 bytes.Buffer
	for _, n := range prog.Nodes {
		if e, ok := n.(interface {
			Emit(io.Writer, map[string]any) error
		}); ok {
			if err := e.Emit(&buf1, map[string]any{"indent_spaces": 2, "indent_level": 1}); err != nil {
				t.Fatalf("Emit error: %v", err)
			}
		}
	}

	l3 := lexer.NewFromReader(strings.NewReader(input))
	p3 := New(l3)
	var buf2 bytes.Buffer
	if err := p3.ParseStream(&buf2, map[string]any{"indent_spaces": 2, "indent_level": 1}); err != nil {
		t.Fatalf("ParseStream error: %v", err)
	}

	if buf1.String() != buf2.String() {
		t.Fatalf("stream output differs\nfull:%q\nstream:%q", buf1.String(), buf2.String())
	}
}
