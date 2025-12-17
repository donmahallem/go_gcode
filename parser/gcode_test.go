package parser

import (
	"testing"

	"github.com/donmahallem/go_gcode/lexer"
	"github.com/donmahallem/go_gcode/nodes"
)

func TestParseGCodeInstructions(t *testing.T) {
	input := `
G1 X10 Y20.5
M104 S200
M1002 judge_flag g29_before_print_flag
G1 X-10
G1 X{val}
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	// Expected nodes:
	// 1. G1 X10 Y20.5
	// 2. M104 S200
	// 3. M1002 judge_flag g29_before_print_flag
	// 4. G1 X-10
	// 5. G1 X{val}

	if len(program.Nodes) != 5 {
		t.Fatalf("program.Nodes has wrong length. got=%d", len(program.Nodes))
	}

	// 1. G1 X10 Y20.5
	stmt1, ok := program.Nodes[0].(*nodes.InstructionNode)
	if !ok {
		t.Fatalf("stmt1 is not InstructionNode. got=%T", program.Nodes[0])
	}
	if stmt1.Command != "G1" {
		t.Errorf("stmt1.Command wrong. got=%q", stmt1.Command)
	}
	if len(stmt1.Parameters) != 2 {
		t.Fatalf("stmt1.Parameters wrong length. got=%d", len(stmt1.Parameters))
	}
	if stmt1.Parameters[0].Key != "X10" {
		t.Errorf("stmt1 param 0 key wrong. got=%q", stmt1.Parameters[0].Key)
	}
	if stmt1.Parameters[0].Value != nil {
		t.Errorf("stmt1 param 0 value wrong. got=%v", stmt1.Parameters[0].Value)
	}
	if stmt1.Parameters[1].Key != "Y20.5" {
		t.Errorf("stmt1 param 1 key wrong. got=%q", stmt1.Parameters[1].Key)
	}
	if stmt1.Parameters[1].Value != nil {
		t.Errorf("stmt1 param 1 value wrong. got=%v", stmt1.Parameters[1].Value)
	}

	// 2. M104 S200
	stmt2 := program.Nodes[1].(*nodes.InstructionNode)
	if stmt2.Command != "M104" {
		t.Errorf("stmt2.Command wrong. got=%q", stmt2.Command)
	}
	if len(stmt2.Parameters) != 1 {
		t.Fatalf("stmt2.Parameters wrong length. got=%d", len(stmt2.Parameters))
	}
	if stmt2.Parameters[0].Key != "S200" {
		t.Errorf("stmt2 param 0 key wrong. got=%q", stmt2.Parameters[0].Key)
	}

	// 3. M1002 judge_flag g29_before_print_flag
	stmt3 := program.Nodes[2].(*nodes.InstructionNode)
	if stmt3.Command != "M1002" {
		t.Errorf("stmt3.Command wrong. got=%q", stmt3.Command)
	}
	if len(stmt3.Parameters) != 2 {
		t.Fatalf("stmt3.Parameters wrong length. got=%d", len(stmt3.Parameters))
	}
	if stmt3.Parameters[0].Key != "judge_flag" {
		t.Errorf("stmt3 param 0 key wrong. got=%q", stmt3.Parameters[0].Key)
	}
	if stmt3.Parameters[0].Value != nil {
		t.Errorf("stmt3 param 0 value should be nil. got=%v", stmt3.Parameters[0].Value)
	}
	if stmt3.Parameters[1].Key != "g29_before_print_flag" {
		t.Errorf("stmt3 param 1 key wrong. got=%q", stmt3.Parameters[1].Key)
	}

	// 4. G1 X-10
	stmt4 := program.Nodes[3].(*nodes.InstructionNode)
	if stmt4.Command != "G1" {
		t.Errorf("stmt4.Command wrong. got=%q", stmt4.Command)
	}
	if stmt4.Parameters[0].Key != "X-10" {
		t.Errorf("stmt4 param 0 key wrong. got=%q", stmt4.Parameters[0].Key)
	}
	if stmt4.Parameters[0].Value != nil {
		t.Errorf("stmt4 param 0 value wrong. got=%v", stmt4.Parameters[0].Value)
	}

	// 5. G1 X{val}
	stmt5 := program.Nodes[4].(*nodes.InstructionNode)
	if stmt5.Command != "G1" {
		t.Errorf("stmt5.Command wrong. got=%q", stmt5.Command)
	}
	if stmt5.Parameters[0].Key != "X" {
		t.Errorf("stmt5 param 0 key wrong. got=%q", stmt5.Parameters[0].Key)
	}
	_, ok = stmt5.Parameters[0].Value.(*nodes.InterpolationNode)
	if !ok {
		t.Errorf("stmt5 param 0 value is not InterpolationNode. got=%T", stmt5.Parameters[0].Value)
	}
}
