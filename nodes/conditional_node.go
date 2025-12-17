package nodes

import (
	"io"

	"github.com/donmahallem/go_gcode/token"
)

/*
 * Condition represents a single condition and its associated body in a ConditionalNode.
 * For example, in {if x} body1 {elsif y} body2 {else} elseBody {endif},
 * there are two Conditions: one for "if x" and one for "elsif y".
 */
type Condition struct {
	Token      token.Token
	Expression Expression
	Body       Node
}

type ConditionalNode struct {
	Token      token.Token
	Conditions []Condition
	Else       Node
}

func (c *ConditionalNode) String() string {
	res := ""
	for i, cond := range c.Conditions {
		if i == 0 {
			res += "{if " + cond.Expression.String() + "}\n"
		} else {
			res += "{elsif " + cond.Expression.String() + "}\n"
		}
		res += cond.Body.String()
	}
	if c.Else != nil {
		res += "{else}\n"
		res += c.Else.String()
	}
	res += "{endif}\n"
	return res
}

func (c *ConditionalNode) TokenLiteral() string {
	return c.Token.Literal
}

func (c *ConditionalNode) Emit(w io.Writer, env map[string]any) error {
	newEnv := make(map[string]any)
	for k, v := range env {
		newEnv[k] = v
	}

	currentLevel := 0
	if val, ok := env["indent_level"]; ok {
		if i, ok := val.(int); ok {
			currentLevel = i
		}
	}
	newEnv["indent_level"] = currentLevel + 1

	for _, cond := range c.Conditions {
		val, err := cond.Expression.Evaluate(env)
		if err != nil {
			return err
		}
		if isTruthy(val) {
			if e, ok := cond.Body.(Emitter); ok {
				return e.Emit(w, newEnv)
			}
			val, err := cond.Body.Evaluate(newEnv)
			if err != nil {
				return err
			}
			return writeAny(w, val)
		}
	}
	if c.Else != nil {
		if e, ok := c.Else.(Emitter); ok {
			return e.Emit(w, newEnv)
		}
		val, err := c.Else.Evaluate(newEnv)
		if err != nil {
			return err
		}
		return writeAny(w, val)
	}
	return nil
}

func (c *ConditionalNode) Evaluate(env map[string]any) (any, error) {
	return EmitToString(c, env)
}
