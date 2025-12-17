package parser

import (
	"testing"

	"github.com/donmahallem/go_gcode/lexer"
	"github.com/donmahallem/go_gcode/nodes"
)

func TestParseInstructionWithComment(t *testing.T) {
	input := "G1 x10 ; move 10 on x"
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	if len(program.Nodes) != 2 {
		t.Fatalf("program.Nodes does not contain 2 statements. got=%d",
			len(program.Nodes))
	}

	stmt1, ok := program.Nodes[0].(*nodes.InstructionNode)
	if !ok {
		t.Fatalf("program.Nodes[0] is not nodes.InstructionNode. got=%T", program.Nodes[0])
	}
	if stmt1.Command != "G1" {
		t.Errorf("stmt1.Command wrong. expected='G1', got=%q", stmt1.Command)
	}

	stmt2, ok := program.Nodes[1].(*nodes.CommentNode)
	if !ok {
		t.Fatalf("program.Nodes[1] is not nodes.CommentNode. got=%T", program.Nodes[1])
	}
	if stmt2.Value != "; move 10 on x" {
		t.Errorf("stmt2.Value wrong. expected='; move 10 on x', got=%q", stmt2.Value)
	}
}
