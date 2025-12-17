package nodes

import (
	"io"
	"testing"

	"github.com/donmahallem/go_gcode/token"
)

func BenchmarkInstructionNode(b *testing.B) {
	instr := &InstructionNode{
		Command:    "G1",
		Parameters: []*ParameterNode{{Key: "X", Value: &IntegerLiteral{Token: token.Token{Literal: "10"}, Value: 10}}},
	}

	b.Run("Emit", func(b *testing.B) {
		for b.Loop() {
			_ = instr.Emit(io.Discard, nil)
		}
	})

	b.Run("Evaluate", func(b *testing.B) {
		for b.Loop() {
			_, _ = instr.Evaluate(nil)
		}
	})
}

func BenchmarkGroupNode(b *testing.B) {
	instr := &InstructionNode{
		Command:    "G1",
		Parameters: []*ParameterNode{{Key: "X", Value: &IntegerLiteral{Token: token.Token{Literal: "10"}, Value: 10}}},
	}
	c := &CommentNode{Value: "; comment"}
	g := &GroupNode{Nodes: []Node{instr, c}}
	env := map[string]any{"indent_spaces": 2, "indent_level": 1}

	b.Run("Emit", func(b *testing.B) {
		for b.Loop() {
			_ = g.Emit(io.Discard, env)
		}
	})

	b.Run("Evaluate", func(b *testing.B) {
		for b.Loop() {
			_, _ = g.Evaluate(env)
		}
	})
}

func BenchmarkGroupNode_Large(b *testing.B) {
	// build a big group of instructions
	nodes := make([]Node, 0, 1000)
	for i := 0; i < 1000; i++ {
		nodes = append(nodes, &InstructionNode{Command: "G1", Parameters: []*ParameterNode{{Key: "X", Value: &IntegerLiteral{Token: token.Token{Literal: "10"}, Value: 10}}}})
	}
	g := &GroupNode{Nodes: nodes}

	b.Run("Emit", func(b *testing.B) {
		for b.Loop() {
			_ = g.Emit(io.Discard, nil)
		}
	})

	b.Run("Evaluate", func(b *testing.B) {
		for b.Loop() {
			_, _ = g.Evaluate(nil)
		}
	})
}
