package nodes

import (
	"fmt"
	"io"
	"strings"
)

type GroupNode struct {
	Nodes []Node
}

func (g *GroupNode) String() string {
	var b strings.Builder
	for _, node := range g.Nodes {
		b.WriteString(node.String())
	}
	return b.String()
}

func (g *GroupNode) TokenLiteral() string {
	if len(g.Nodes) > 0 {
		return g.Nodes[0].TokenLiteral()
	}
	return ""
}

// Emit streams the concatenated output of child nodes to the writer.
func (g *GroupNode) Emit(w io.Writer, env map[string]any) error {
	for _, node := range g.Nodes {
		if e, ok := node.(Emitter); ok {
			if err := e.Emit(w, env); err != nil {
				return err
			}
		} else {
			val, err := node.Evaluate(env)
			if err != nil {
				return err
			}
			if _, err := io.WriteString(w, fmt.Sprintf("%v", val)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Evaluate keeps compatibility by collecting emitted output into a string.
func (g *GroupNode) Evaluate(env map[string]any) (any, error) {
	return EmitToString(g, env)
}
