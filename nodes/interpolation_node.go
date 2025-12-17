package nodes

import (
	"fmt"
	"io"

	"github.com/donmahallem/go_gcode/token"
)

type InterpolationNode struct {
	Token      token.Token
	Expression Expression
}

func (i *InterpolationNode) String() string {
	return "{" + i.Expression.String() + "}"
}

func (i *InterpolationNode) TokenLiteral() string {
	return i.Token.Literal
}

func (i *InterpolationNode) Emit(w io.Writer, env map[string]any) error {
	val, err := i.Expression.Evaluate(env)
	if err != nil {
		return err
	}
	return writeAny(w, val)
}

func (i *InterpolationNode) Evaluate(env map[string]any) (any, error) {
	val, err := i.Expression.Evaluate(env)
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("%v", val), nil
}
