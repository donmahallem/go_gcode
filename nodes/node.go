package nodes

import (
	"bytes"
	"fmt"
	"io"
)

type Emitter interface {
	Emit(w io.Writer, env map[string]any) error
}

func EmitToString(e Emitter, env map[string]any) (string, error) {
	var buf bytes.Buffer
	if err := e.Emit(&buf, env); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func writeAny(w io.Writer, v any) error {
	s := fmt.Sprintf("%v", v)
	if sw, ok := w.(interface{ WriteString(string) (int, error) }); ok {
		_, err := sw.WriteString(s)
		return err
	}
	_, err := io.WriteString(w, s)
	return err
}

type Node interface {
	String() string
	TokenLiteral() string
	Evaluate(env map[string]any) (any, error)
}
