package nodes

import (
	"testing"
)

func TestPrettyPrint(t *testing.T) {

	program := &GroupNode{
		Nodes: []Node{
			&ConditionalNode{
				Conditions: []Condition{
					{
						Expression: &Boolean{Value: true},
						Body: &GroupNode{
							Nodes: []Node{
								&InstructionNode{
									Command: "G1",
									Parameters: []*ParameterNode{
										{Key: "X10", Value: nil},
									},
								},
								&CommentNode{Value: "; comment"},
							},
						},
					},
				},
			},
		},
	}

	env := map[string]any{
		"indent_spaces": 2,
	}

	res, err := program.Evaluate(env)
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}

	expected := "  G1 X10\n  ; comment\n"
	if res != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, res)
	}
}

func TestDoubleSpaceReduction(t *testing.T) {
	// G1  X10  Y20 -> G1 X10 Y20
	instr := &InstructionNode{
		Command: "G1",
		Parameters: []*ParameterNode{
			{Key: "X10", Value: nil},
			{Key: "Y20", Value: nil},
		},
	}

	res, err := instr.Evaluate(nil)
	if err != nil {
		t.Fatalf("Evaluate error: %v", err)
	}

	expected := "G1 X10 Y20\n"
	if res != expected {
		t.Errorf("Expected %q, got %q", expected, res)
	}
}
