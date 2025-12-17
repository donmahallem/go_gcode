package nodes

import (
	"strings"
	"testing"

	"github.com/donmahallem/go_gcode/token"
)

func TestEvaluateErrors(t *testing.T) {
	tests := []struct {
		name          string
		input         Node
		env           map[string]any
		expectedError string
	}{
		{
			name: "Identifier not found",
			input: &Identifier{
				Token: token.Token{Type: token.IDENT, Literal: "missing"},
				Value: "missing",
			},
			env:           map[string]any{},
			expectedError: "identifier not found: missing",
		},
		{
			name: "Unknown prefix operator",
			input: &PrefixExpression{
				Token:    token.Token{Type: token.ILLEGAL, Literal: "?"},
				Operator: "?",
				Right:    &IntegerLiteral{Value: 5},
			},
			env:           nil,
			expectedError: "unknown operator: ?int64",
		},
		{
			name: "Minus prefix on string",
			input: &PrefixExpression{
				Token:    token.Token{Type: token.MINUS, Literal: "-"},
				Operator: "-",
				Right:    &StringLiteral{Value: "str"},
			},
			env:           nil,
			expectedError: "unknown operator: -string",
		},
		{
			name: "Unknown infix operator",
			input: &InfixExpression{
				Token:    token.Token{Type: token.ILLEGAL, Literal: "?"},
				Left:     &IntegerLiteral{Value: 5},
				Operator: "?",
				Right:    &IntegerLiteral{Value: 5},
			},
			env:           nil,
			expectedError: "unknown operator: int64 ? int64",
		},
		{
			name: "String infix unknown operator",
			input: &InfixExpression{
				Token:    token.Token{Type: token.MINUS, Literal: "-"},
				Left:     &StringLiteral{Value: "a"},
				Operator: "-",
				Right:    &StringLiteral{Value: "b"},
			},
			env:           nil,
			expectedError: "unknown operator: a - b",
		},
		{
			name: "Type mismatch infix",
			input: &InfixExpression{
				Token:    token.Token{Type: token.PLUS, Literal: "+"},
				Left:     &IntegerLiteral{Value: 5},
				Operator: "+",
				Right:    &StringLiteral{Value: "b"},
			},
			env:           nil,
			expectedError: "unknown operator: int64 + string",
		},
		{
			name: "Index on non-container",
			input: &IndexExpression{
				Token: token.Token{Type: token.LBRACKET, Literal: "["},
				Left:  &IntegerLiteral{Value: 5},
				Index: &IntegerLiteral{Value: 0},
			},
			env:           nil,
			expectedError: "index operator not supported: int64",
		},
		{
			name: "Map index non-string key",
			input: &IndexExpression{
				Token: token.Token{Type: token.LBRACKET, Literal: "["},
				Left:  &Identifier{Value: "m"},
				Index: &IntegerLiteral{Value: 0},
			},
			env:           map[string]any{"m": map[string]any{}},
			expectedError: "map index must be string, got int64",
		},
		{
			name: "Map key not found",
			input: &IndexExpression{
				Token: token.Token{Type: token.LBRACKET, Literal: "["},
				Left:  &Identifier{Value: "m"},
				Index: &StringLiteral{Value: "missing"},
			},
			env:           map[string]any{"m": map[string]any{}},
			expectedError: "key not found: missing",
		},
		{
			name: "Array index non-integer",
			input: &IndexExpression{
				Token: token.Token{Type: token.LBRACKET, Literal: "["},
				Left:  &Identifier{Value: "arr"},
				Index: &StringLiteral{Value: "0"},
			},
			env:           map[string]any{"arr": []any{}},
			expectedError: "array index must be integer, got string",
		},
		{
			name: "Array index out of bounds",
			input: &IndexExpression{
				Token: token.Token{Type: token.LBRACKET, Literal: "["},
				Left:  &Identifier{Value: "arr"},
				Index: &IntegerLiteral{Value: 5},
			},
			env:           map[string]any{"arr": []any{1, 2}},
			expectedError: "index out of bounds: 5",
		},
		{
			name: "Prefix expression error propagation",
			input: &PrefixExpression{
				Operator: "!",
				Right: &Identifier{
					Value: "missing",
				},
			},
			env:           map[string]any{},
			expectedError: "identifier not found: missing",
		},
		{
			name: "Infix expression left error propagation",
			input: &InfixExpression{
				Left: &Identifier{
					Value: "missing",
				},
				Operator: "+",
				Right:    &IntegerLiteral{Value: 1},
			},
			env:           map[string]any{},
			expectedError: "identifier not found: missing",
		},
		{
			name: "Infix expression right error propagation",
			input: &InfixExpression{
				Left:     &IntegerLiteral{Value: 1},
				Operator: "+",
				Right: &Identifier{
					Value: "missing",
				},
			},
			env:           map[string]any{},
			expectedError: "identifier not found: missing",
		},
		{
			name: "Index expression left error propagation",
			input: &IndexExpression{
				Left: &Identifier{
					Value: "missing",
				},
				Index: &IntegerLiteral{Value: 0},
			},
			env:           map[string]any{},
			expectedError: "identifier not found: missing",
		},
		{
			name: "Index expression index error propagation",
			input: &IndexExpression{
				Left: &Identifier{
					Value: "arr",
				},
				Index: &Identifier{
					Value: "missing",
				},
			},
			env:           map[string]any{"arr": []any{}},
			expectedError: "identifier not found: missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.input.Evaluate(tt.env)
			if err == nil {
				t.Errorf("expected error %q, got nil", tt.expectedError)
				return
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("expected error to contain %q, got %q", tt.expectedError, err.Error())
			}
		})
	}
}

func TestEvaluateAdditionalOps(t *testing.T) {
	tests := []struct {
		name     string
		input    Node
		env      map[string]any
		expected any
	}{
		{
			name: "Minus prefix int",
			input: &PrefixExpression{
				Operator: "-",
				Right:    &IntegerLiteral{Value: 5},
			},
			expected: int64(-5),
		},
		{
			name: "Minus prefix float",
			input: &PrefixExpression{
				Operator: "-",
				Right:    &FloatLiteral{Value: 5.5},
			},
			expected: -5.5,
		},
		{
			name: "Bang prefix nil",
			input: &PrefixExpression{
				Operator: "!",
				Right:    &Identifier{Value: "null"},
			},
			env:      map[string]any{"null": nil},
			expected: true,
		},
		{
			name: "Bang prefix false",
			input: &PrefixExpression{
				Operator: "!",
				Right:    &Identifier{Value: "f"},
			},
			env:      map[string]any{"f": false},
			expected: true,
		},
		{
			name: "String equality",
			input: &InfixExpression{
				Left:     &StringLiteral{Value: "a"},
				Operator: "==",
				Right:    &StringLiteral{Value: "a"},
			},
			expected: true,
		},
		{
			name: "String inequality",
			input: &InfixExpression{
				Left:     &StringLiteral{Value: "a"},
				Operator: "!=",
				Right:    &StringLiteral{Value: "b"},
			},
			expected: true,
		},
		{
			name: "String concatenation",
			input: &InfixExpression{
				Left:     &StringLiteral{Value: "a"},
				Operator: "+",
				Right:    &StringLiteral{Value: "b"},
			},
			expected: "ab",
		},
		{
			name: "Boolean AND",
			input: &InfixExpression{
				Left:     &Identifier{Value: "t"},
				Operator: "&&",
				Right:    &Identifier{Value: "f"},
			},
			env:      map[string]any{"t": true, "f": false},
			expected: false,
		},
		{
			name: "Boolean OR",
			input: &InfixExpression{
				Left:     &Identifier{Value: "t"},
				Operator: "||",
				Right:    &Identifier{Value: "f"},
			},
			env:      map[string]any{"t": true, "f": false},
			expected: true,
		},
		{
			name: "Number comparisons",
			input: &InfixExpression{
				Left:     &IntegerLiteral{Value: 5},
				Operator: "<",
				Right:    &IntegerLiteral{Value: 10},
			},
			expected: true,
		},
		{
			name: "Number comparisons 2",
			input: &InfixExpression{
				Left:     &IntegerLiteral{Value: 5},
				Operator: ">",
				Right:    &IntegerLiteral{Value: 10},
			},
			expected: false,
		},
		{
			name: "Number comparisons 3",
			input: &InfixExpression{
				Left:     &IntegerLiteral{Value: 5},
				Operator: "==",
				Right:    &IntegerLiteral{Value: 5},
			},
			expected: true,
		},
		{
			name: "Number comparisons 4",
			input: &InfixExpression{
				Left:     &IntegerLiteral{Value: 5},
				Operator: "!=",
				Right:    &IntegerLiteral{Value: 6},
			},
			expected: true,
		},
		{
			name: "Array index int",
			input: &IndexExpression{
				Left:  &Identifier{Value: "arr"},
				Index: &Identifier{Value: "idx"},
			},
			env:      map[string]any{"arr": []any{10, 20}, "idx": 1},
			expected: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.Evaluate(tt.env)
			if err != nil {
				t.Errorf("unexpected error: %s", err)
				return
			}
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
