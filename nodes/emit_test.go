package nodes

import (
	"io"
	"testing"

	"github.com/donmahallem/go_gcode/token"
)

func TestEmit_InstructionNode(t *testing.T) {
	instr := &InstructionNode{
		Command: "G1",
		Parameters: []*ParameterNode{
			{Key: "X", Value: &IntegerLiteral{Token: token.Token{Literal: "10"}, Value: 10}},
		},
	}
	res, err := EmitToString(instr, nil)
	if err != nil {
		t.Fatalf("Emit error: %v", err)
	}
	expected := "G1 X10\n"
	if res != expected {
		t.Fatalf("expected %q, got %q", expected, res)
	}
}

func TestEmit_CommentNode_Indent(t *testing.T) {
	c := &CommentNode{Value: "; comment"}
	env := map[string]any{"indent_spaces": 2, "indent_level": 1}
	res, err := EmitToString(c, env)
	if err != nil {
		t.Fatalf("Emit error: %v", err)
	}
	expected := "  ; comment\n"
	if res != expected {
		t.Fatalf("expected %q, got %q", expected, res)
	}
}

func TestEmit_GroupNode_and_Streaming(t *testing.T) {
	instr := &InstructionNode{
		Command:    "G1",
		Parameters: []*ParameterNode{{Key: "X", Value: &IntegerLiteral{Token: token.Token{Literal: "10"}, Value: 10}}},
	}
	c := &CommentNode{Value: "; comment"}
	g := &GroupNode{Nodes: []Node{instr, c}}
	env := map[string]any{"indent_spaces": 2, "indent_level": 1}
	res, err := EmitToString(g, env)
	if err != nil {
		t.Fatalf("Emit error: %v", err)
	}
	expected := "  G1 X10\n  ; comment\n"
	if res != expected {
		t.Fatalf("expected %q, got %q", expected, res)
	}

	// streaming to io.Discard should not error
	if err := instr.Emit(io.Discard, nil); err != nil {
		t.Fatalf("Emit to discard failed: %v", err)
	}
}

func TestEmit_InterpolationNode(t *testing.T) {
	i := &InterpolationNode{Expression: &IntegerLiteral{Token: token.Token{Literal: "5"}, Value: 5}}
	res, err := EmitToString(i, nil)
	if err != nil {
		t.Fatalf("Emit error: %v", err)
	}
	if res != "5" {
		t.Fatalf("expected '5', got %q", res)
	}
}

func TestEmit_ConditionalNode(t *testing.T) {
	instr := &InstructionNode{
		Command:    "G1",
		Parameters: []*ParameterNode{{Key: "X", Value: &IntegerLiteral{Token: token.Token{Literal: "10"}, Value: 10}}},
	}
	c := &CommentNode{Value: "; comment"}
	cond := &ConditionalNode{
		Conditions: []Condition{{Expression: &Boolean{Value: true}, Body: &GroupNode{Nodes: []Node{instr, c}}}},
	}
	env := map[string]any{"indent_spaces": 2, "indent_level": 0}
	res, err := EmitToString(cond, env)
	if err != nil {
		t.Fatalf("Emit error: %v", err)
	}
	expected := "  G1 X10\n  ; comment\n"
	if res != expected {
		t.Fatalf("expected %q, got %q", expected, res)
	}
}
