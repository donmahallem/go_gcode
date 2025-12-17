package nodes

import (
	"io"
	"strings"

	"github.com/donmahallem/go_gcode/token"
)

type InstructionNode struct {
	Token      token.Token // The command token (e.g. G1, M104)
	Command    string
	Parameters []*ParameterNode
}

func (in *InstructionNode) expressionNode()      {}
func (in *InstructionNode) TokenLiteral() string { return in.Token.Literal }

func (in *InstructionNode) String() string {
	var out strings.Builder
	out.WriteString(in.Command)
	for _, p := range in.Parameters {
		out.WriteString(" ")
		out.WriteString(p.String())
	}
	return out.String()
}

func (in *InstructionNode) Emit(w io.Writer, env map[string]any) error {
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

	if _, err := io.WriteString(w, in.Command); err != nil {
		return err
	}

	for _, p := range in.Parameters {
		if _, err := io.WriteString(w, " "); err != nil {
			return err
		}
		if _, err := io.WriteString(w, p.Key); err != nil {
			return err
		}
		if p.Value != nil {
			if e, ok := p.Value.(Emitter); ok {
				if err := e.Emit(w, env); err != nil {
					return err
				}
			} else {
				val, err := p.Value.Evaluate(env)
				if err != nil {
					return err
				}
				if err := writeAny(w, val); err != nil {
					return err
				}
			}
		}
	}
	if _, err := io.WriteString(w, "\n"); err != nil {
		return err
	}
	return nil
}

// Evaluate kept for compatibility
func (in *InstructionNode) Evaluate(env map[string]any) (any, error) {
	return EmitToString(in, env)
}

type ParameterNode struct {
	Key   string // The key (e.g. X, Y, or the flag name)
	Value Node   // The value (e.g. 10, {val}), or nil if it's a flag
}

func (pn *ParameterNode) String() string {
	var out strings.Builder
	out.WriteString(pn.Key)
	if pn.Value != nil {
		// GCode usually doesn't have space between key and value (X10), but flags do (judge_flag)
		// If it's a KV pair, we might want to check if we should add space?
		// Usually X10.
		out.WriteString(pn.Value.String())
	}
	return out.String()
}
