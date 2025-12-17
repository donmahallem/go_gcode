package nodes

import (
	"testing"

	"github.com/donmahallem/go_gcode/token"
)

func TestEvaluate(t *testing.T) {
	tests := []struct {
		input    Node
		env      map[string]any
		expected any
	}{
		{
			&IntegerLiteral{Token: token.Token{Type: token.INT, Literal: "5"}, Value: 5},
			nil,
			int64(5),
		},
		{
			&FloatLiteral{Token: token.Token{Type: token.FLOAT, Literal: "5.5"}, Value: 5.5},
			nil,
			5.5,
		},
		{
			&StringLiteral{Token: token.Token{Type: token.STRING, Literal: "hello"}, Value: "hello"},
			nil,
			"hello",
		},
		{
			&Identifier{Token: token.Token{Type: token.IDENT, Literal: "x"}, Value: "x"},
			map[string]any{"x": 10},
			10,
		},
		{
			&InfixExpression{
				Token:    token.Token{Type: token.PLUS, Literal: "+"},
				Left:     &IntegerLiteral{Value: 5},
				Operator: "+",
				Right:    &IntegerLiteral{Value: 10},
			},
			nil,
			15.0, // 5 + 10 -> float64
		},
		{
			&InfixExpression{
				Token:    token.Token{Type: token.PLUS, Literal: "+"},
				Left:     &FloatLiteral{Value: 5.5},
				Operator: "+",
				Right:    &FloatLiteral{Value: 4.5},
			},
			nil,
			10.0,
		},
		{
			&PrefixExpression{
				Token:    token.Token{Type: token.BANG, Literal: "!"},
				Operator: "!",
				Right:    &Identifier{Value: "trueVal"},
			},
			map[string]any{"trueVal": true},
			false,
		},
		{
			&IndexExpression{
				Token: token.Token{Type: token.LBRACKET, Literal: "["},
				Left:  &Identifier{Value: "arr"},
				Index: &IntegerLiteral{Value: 1},
			},
			map[string]any{"arr": []any{10, 20, 30}},
			20,
		},
		{
			&IndexExpression{
				Token: token.Token{Type: token.LBRACKET, Literal: "["},
				Left:  &Identifier{Value: "m"},
				Index: &StringLiteral{Value: "key"},
			},
			map[string]any{"m": map[string]any{"key": "value"}},
			"value",
		},
		{
			&Boolean{Token: token.Token{Type: token.TRUE, Literal: "true"}, Value: true},
			nil,
			true,
		},
		{
			&Boolean{Token: token.Token{Type: token.FALSE, Literal: "false"}, Value: false},
			nil,
			false,
		},
		{
			&PrefixExpression{
				Token:    token.Token{Type: token.MINUS, Literal: "-"},
				Operator: "-",
				Right:    &IntegerLiteral{Value: 5},
			},
			nil,
			int64(-5),
		},
		{
			&InfixExpression{
				Token:    token.Token{Type: token.LT, Literal: "<"},
				Left:     &IntegerLiteral{Value: 5},
				Operator: "<",
				Right:    &IntegerLiteral{Value: 10},
			},
			nil,
			true,
		},
		{
			&InfixExpression{
				Token:    token.Token{Type: token.GT, Literal: ">"},
				Left:     &IntegerLiteral{Value: 5},
				Operator: ">",
				Right:    &IntegerLiteral{Value: 10},
			},
			nil,
			false,
		},
		{
			&InfixExpression{
				Token:    token.Token{Type: token.EQ, Literal: "=="},
				Left:     &IntegerLiteral{Value: 5},
				Operator: "==",
				Right:    &IntegerLiteral{Value: 5},
			},
			nil,
			true,
		},
		{
			&InfixExpression{
				Token:    token.Token{Type: token.NOT_EQ, Literal: "!="},
				Left:     &IntegerLiteral{Value: 5},
				Operator: "!=",
				Right:    &IntegerLiteral{Value: 10},
			},
			nil,
			true,
		},
		{
			&IfExpression{
				Token:       token.Token{Type: token.IF, Literal: "if"},
				Condition:   &Boolean{Value: true},
				Consequence: &IntegerLiteral{Value: 10},
			},
			nil,
			int64(10),
		},
		{
			&IfExpression{
				Token:       token.Token{Type: token.IF, Literal: "if"},
				Condition:   &Boolean{Value: false},
				Consequence: &IntegerLiteral{Value: 10},
				Alternative: &IntegerLiteral{Value: 20},
			},
			nil,
			int64(20),
		},
	}

	for i, tt := range tests {
		result, err := tt.input.Evaluate(tt.env)
		if err != nil {
			t.Errorf("tests[%d] - evaluation error: %s", i, err)
			continue
		}

		if result != tt.expected {
			t.Errorf("tests[%d] - wrong result. expected=%v, got=%v", i, tt.expected, result)
		}
	}
}

func TestEvaluateConditional(t *testing.T) {
	// {if x > 5} "big" {else} "small" {endif}
	cond := &ConditionalNode{
		Conditions: []Condition{
			{
				Expression: &InfixExpression{
					Left:     &Identifier{Value: "x"},
					Operator: ">",
					Right:    &IntegerLiteral{Value: 5},
				},
				Body: &TextNode{Value: "big"},
			},
		},
		Else: &TextNode{Value: "small"},
	}

	// Case 1: x = 10 -> "big"
	res, err := cond.Evaluate(map[string]any{"x": 10})
	if err != nil {
		t.Fatalf("eval error: %s", err)
	}
	if res != "big" {
		t.Errorf("expected 'big', got %v", res)
	}

	// Case 2: x = 2 -> "small"
	res, err = cond.Evaluate(map[string]any{"x": 2})
	if err != nil {
		t.Fatalf("eval error: %s", err)
	}
	if res != "small" {
		t.Errorf("expected 'small', got %v", res)
	}
}

func TestEvaluateInterpolation(t *testing.T) {
	// {x + 10}
	interp := &InterpolationNode{
		Expression: &InfixExpression{
			Left:     &Identifier{Value: "x"},
			Operator: "+",
			Right:    &IntegerLiteral{Value: 10},
		},
	}

	res, err := interp.Evaluate(map[string]any{"x": 5})
	if err != nil {
		t.Fatalf("eval error: %s", err)
	}

	// Result should be string "15" (or "15.0" depending on float conversion)
	// In our evaluator, 5 + 10 becomes 15.0 (float64) because of toFloat conversion in evalNumberInfixExpression
	// And InterpolationNode uses fmt.Sprintf("%v", val)
	if res != "15" {
		t.Errorf("expected '15', got %q", res)
	}
}

func TestEvaluateConditionalElsif(t *testing.T) {
	// {if x > 10} "big" {elsif x > 5} "medium" {else} "small" {endif}
	cond := &ConditionalNode{
		Conditions: []Condition{
			{
				Expression: &InfixExpression{
					Left:     &Identifier{Value: "x"},
					Operator: ">",
					Right:    &IntegerLiteral{Value: 10},
				},
				Body: &TextNode{Value: "big"},
			},
			{
				Expression: &InfixExpression{
					Left:     &Identifier{Value: "x"},
					Operator: ">",
					Right:    &IntegerLiteral{Value: 5},
				},
				Body: &TextNode{Value: "medium"},
			},
		},
		Else: &TextNode{Value: "small"},
	}

	tests := []struct {
		x        int64
		expected string
	}{
		{15, "big"},
		{8, "medium"},
		{2, "small"},
	}

	for _, tt := range tests {
		res, err := cond.Evaluate(map[string]any{"x": tt.x})
		if err != nil {
			t.Fatalf("x=%d eval error: %s", tt.x, err)
		}
		if res != tt.expected {
			t.Errorf("x=%d expected %q, got %q", tt.x, tt.expected, res)
		}
	}
}
