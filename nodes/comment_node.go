package nodes

import (
	"io"
	"strings"

	"github.com/donmahallem/go_gcode/token"
)

/*
 * CommentNode represents a comment in the G-code.
 * For example: "; This is a comment"
 */
type CommentNode struct {
	Token token.Token
	Value string
}

func (c *CommentNode) String() string {
	return c.Value
}

func (c *CommentNode) TokenLiteral() string {
	return c.Token.Literal
}

func (c *CommentNode) Emit(w io.Writer, env map[string]any) error {
	// Indentation
	indentSpaces := 0
	if val, ok := env["indent_spaces"]; ok {
		if i, ok := val.(int); ok {
			indentSpaces = i
		}
	}
	indentLevel := 0
	if val, ok := env["indent_level"]; ok {
		if i, ok := val.(int); ok {
			indentLevel = i
		}
	}

	if indentSpaces > 0 && indentLevel > 0 {
		if _, err := io.WriteString(w, strings.Repeat(" ", indentSpaces*indentLevel)); err != nil {
			return err
		}
	}

	if _, err := io.WriteString(w, c.Value); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}
	return nil
}

// Evaluate kept for compatibility
func (c *CommentNode) Evaluate(env map[string]any) (any, error) {
	return EmitToString(c, env)
}
