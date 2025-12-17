package nodes

import (
	"io"

	"github.com/donmahallem/go_gcode/token"
)

type TextNode struct {
	Token token.Token
	Value string
}

func (t *TextNode) String() string {
	return t.Value
}

func (t *TextNode) TokenLiteral() string {
	return t.Token.Literal
}

func (t *TextNode) Emit(w io.Writer, env map[string]any) error {
	if _, err := io.WriteString(w, t.Value); err != nil {
		return err
	}
	return nil
}

// Evaluate kept for compatibility
func (t *TextNode) Evaluate(env map[string]any) (any, error) {
	return t.Value, nil
}
