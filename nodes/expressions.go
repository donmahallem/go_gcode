package nodes

import (
	"bytes"
	"fmt"

	"github.com/donmahallem/go_gcode/token"
)

type Expression interface {
	Node
	expressionNode()
}

type Identifier struct {
	Token token.Token // The token.IDENT token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }
func (i *Identifier) Evaluate(env map[string]any) (any, error) {
	if val, ok := env[i.Value]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("identifier not found: %s", i.Value)
}

type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }
func (il *IntegerLiteral) Evaluate(env map[string]any) (any, error) {
	return il.Value, nil
}

type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }
func (fl *FloatLiteral) Evaluate(env map[string]any) (any, error) {
	return fl.Value, nil
}

type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return `"` + sl.Value + `"` }
func (sl *StringLiteral) Evaluate(env map[string]any) (any, error) {
	return sl.Value, nil
}

type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. ! or -
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}
func (pe *PrefixExpression) Evaluate(env map[string]any) (any, error) {
	right, err := pe.Right.Evaluate(env)
	if err != nil {
		return nil, err
	}
	return evalPrefixExpression(pe.Operator, right)
}

type InfixExpression struct {
	Token    token.Token // The operator token, e.g. +
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}
func (ie *InfixExpression) Evaluate(env map[string]any) (any, error) {
	left, err := ie.Left.Evaluate(env)
	if err != nil {
		return nil, err
	}
	right, err := ie.Right.Evaluate(env)
	if err != nil {
		return nil, err
	}
	return evalInfixExpression(ie.Operator, left, right)
}

type IndexExpression struct {
	Token token.Token // The [ token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")
	return out.String()
}
func (ie *IndexExpression) Evaluate(env map[string]any) (any, error) {
	left, err := ie.Left.Evaluate(env)
	if err != nil {
		return nil, err
	}
	index, err := ie.Index.Evaluate(env)
	if err != nil {
		return nil, err
	}

	return evalIndexExpression(left, index)
}

type Boolean struct {
	Token token.Token
	Value bool
}

func (b *Boolean) expressionNode()      {}
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }
func (b *Boolean) String() string       { return b.Token.Literal }
func (b *Boolean) Evaluate(env map[string]any) (any, error) {
	return b.Value, nil
}

type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence Expression
	Alternative Expression
}

func (ie *IfExpression) expressionNode()      {}
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IfExpression) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())
	if ie.Alternative != nil {
		out.WriteString("else ")
		out.WriteString(ie.Alternative.String())
	}
	return out.String()
}
func (ie *IfExpression) Evaluate(env map[string]any) (any, error) {
	condition, err := ie.Condition.Evaluate(env)
	if err != nil {
		return nil, err
	}

	if isTruthy(condition) {
		return ie.Consequence.Evaluate(env)
	} else if ie.Alternative != nil {
		return ie.Alternative.Evaluate(env)
	} else {
		return nil, nil
	}
}
