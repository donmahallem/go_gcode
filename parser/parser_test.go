package parser

import (
	"testing"

	"github.com/donmahallem/go_gcode/lexer"
	"github.com/donmahallem/go_gcode/nodes"
)

func BenchmarkParseProgram(b *testing.B) {
	input := `
M104 S75
{if temp > 50}
  M106
{endif}
G1 X{10+20}
`
	for b.Loop() {
		l := lexer.New(input)
		p := New(l)
		_ = p.ParseProgram()
	}
}

func TestParseProgram(t *testing.T) {
	input := `
M104 S75
{if temp > 50}
  M106
{endif}
G1 X{10+20}
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser has %d errors: %v", len(p.Errors()), p.Errors())
	}

	if len(program.Nodes) != 3 {
		t.Fatalf("program.Nodes does not contain 3 statements. got=%d",
			len(program.Nodes))
	}
	stmt, ok := program.Nodes[0].(*nodes.InstructionNode)
	if !ok {
		t.Fatalf("program.Nodes[0] is not nodes.InstructionNode. got=%T", program.Nodes[0])
	}
	if stmt.Command != "M104" {
		t.Errorf("stmt.Command not 'M104'. got=%q", stmt.Command)
	}

	cond, ok := program.Nodes[1].(*nodes.ConditionalNode)
	if !ok {
		t.Fatalf("program.Nodes[1] is not nodes.ConditionalNode. got=%T", program.Nodes[1])
	}
	if len(cond.Conditions) != 1 {
		t.Fatalf("cond.Conditions does not contain 1 condition. got=%d", len(cond.Conditions))
	}
}

func TestParseExpressions(t *testing.T) {
	testArgs := make(map[string]any)
	testArgs["arr"] = []any{10, 20, 30}
	testArgs["i"] = 2
	testArgs["a"] = 10
	testArgs["b"] = 20
	tests := []struct {
		input       string
		expectedStr string
		expectedVal string
	}{
		{"{5}", "{5}", "5"},
		{"{5 + 5}", "{(5 + 5)}", "10"},
		{"{5 - 5}", "{(5 - 5)}", "0"},
		{"{5 * 5}", "{(5 * 5)}", "25"},
		{"{5 / 5}", "{(5 / 5)}", "1"},
		{"{-5}", "{(-5)}", "-5"},
		{"{!true}", "{(!true)}", "false"},
		{"{5 + 5 * 2}", "{(5 + (5 * 2))}", "15"},
		{"{5 + 5 + 2}", "{((5 + 5) + 2)}", "12"},
		{"{a + b}", "{(a + b)}", "30"},
		{"{arr[1]}", "{(arr[1])}", "20"},
		{"{arr[i]}", "{(arr[i])}", "30"},
	}
	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()

		if len(p.Errors()) > 0 {
			t.Errorf("parser has errors for input %q: %v", tt.input, p.Errors())
			continue
		}

		if len(program.Nodes) != 1 {
			t.Errorf("program.Nodes has wrong length for input %q. got=%d", tt.input, len(program.Nodes))
			continue
		}

		interp, ok := program.Nodes[0].(*nodes.InterpolationNode)
		if !ok {
			t.Errorf("program.Nodes[0] is not InterpolationNode. got=%T", program.Nodes[0])
			continue
		}

		if interp.String() != tt.expectedStr {
			t.Errorf("interp.String() wrong. expected=%q, got=%q", tt.expectedStr, interp.String())
		}
		evaluated, err := interp.Evaluate(testArgs)
		if err != nil {
			t.Errorf("interp.Evaluate() returned error: %s", err)
			continue
		}
		if evaluated != tt.expectedVal {
			t.Errorf("interp.Evaluate() wrong. expected=%q, got=%q", tt.expectedVal, evaluated)
		}
	}
}

func TestParseErrorRecovery(t *testing.T) {
	input := `
{if 1 == }
  M1
{endif}
G1 X10
`
	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Errorf("Expected errors, got none")
	}

	foundG1 := false
	for _, node := range program.Nodes {
		if instr, ok := node.(*nodes.InstructionNode); ok {
			if instr.Command == "G1" {
				foundG1 = true
				break
			}
		}
	}

	if !foundG1 {
		t.Errorf("Did not find 'G1 X10' after error recovery")
	}
}
