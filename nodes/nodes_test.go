package nodes

import (
	"testing"

	"github.com/donmahallem/go_gcode/token"
)

func TestConditionalNode(t *testing.T) {
	// {if x} body {elsif y} body2 {else} elseBody {endif}
	cond := &ConditionalNode{
		Token: token.Token{Type: token.IF, Literal: "if"},
		Conditions: []Condition{
			{
				Token:      token.Token{Type: token.IF, Literal: "if"},
				Expression: &Identifier{Value: "x"},
				Body:       &TextNode{Value: "body"},
			},
			{
				Token:      token.Token{Type: token.ELSIF, Literal: "elsif"},
				Expression: &Identifier{Value: "y"},
				Body:       &TextNode{Value: "body2"},
			},
		},
		Else: &TextNode{Value: "elseBody"},
	}

	if cond.TokenLiteral() != "if" {
		t.Errorf("TokenLiteral wrong. expected='if', got=%q", cond.TokenLiteral())
	}

	expectedString := "{if x}\nbody{elsif y}\nbody2{else}\nelseBody{endif}\n"
	if cond.String() != expectedString {
		t.Errorf("String wrong. expected=%q, got=%q", expectedString, cond.String())
	}
}

func TestInterpolationNode(t *testing.T) {
	// {x + 1}
	interp := &InterpolationNode{
		Token: token.Token{Type: token.LBRACE, Literal: "{"},
		Expression: &InfixExpression{
			Left:     &Identifier{Value: "x"},
			Operator: "+",
			Right:    &IntegerLiteral{Token: token.Token{Literal: "1"}, Value: 1},
		},
	}

	if interp.TokenLiteral() != "{" {
		t.Errorf("TokenLiteral wrong. expected='{', got=%q", interp.TokenLiteral())
	}

	expectedString := "{(x + 1)}"
	if interp.String() != expectedString {
		t.Errorf("String wrong. expected=%q, got=%q", expectedString, interp.String())
	}
}
