package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/donmahallem/go_gcode/lexer"
	"github.com/donmahallem/go_gcode/nodes"
	"github.com/donmahallem/go_gcode/token"
)

// ParseAndEmit reads from r, parses on-the-fly, and writes rendered output to w.
// It streams top-level text and code blocks and handles conditionals without
// building AST for skipped branches.
func ParseAndEmit(r io.Reader, w io.Writer, env map[string]any) error {
	br := bufio.NewReader(r)
	// Skip initial leading newlines to match ParseProgram behavior
	for {
		b, err := br.Peek(1)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if len(b) > 0 && b[0] == '\n' {
			// consume it
			if _, err := br.ReadByte(); err != nil {
				return err
			}
			continue
		}
		break
	}
	return streamProcess(br, w, env)
}

func streamProcess(br *bufio.Reader, w io.Writer, env map[string]any) error {
	lastNewline := false
	for {
		// Read up to next '{'
		text, err := readUntilChar(br, '{')
		if len(text) > 0 {
			// Avoid emitting an extra blank line: if previous write ended with newline
			// and this text starts with newline, trim leading newlines.
			if lastNewline && strings.HasPrefix(text, "\n") {
				text = strings.TrimLeft(text, "\n")
			}
			if len(text) > 0 {
				if _, err := io.WriteString(w, text); err != nil {
					return err
				}
				lastNewline = strings.HasSuffix(text, "\n")
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		// We are at a '{'. Read code until '}'
		code, err := readUntilChar(br, '}')
		if err != nil {
			return fmt.Errorf("unterminated '{' block: %w", err)
		}
		trimmed := strings.TrimSpace(code)
		// Control blocks: if, elsif, else, endif
		if strings.HasPrefix(trimmed, "if ") {
			// handle conditional streaming; trimmed includes the 'if ...'
			if err := handleConditionalStream(br, w, env, trimmed[len("if "):], &lastNewline); err != nil {
				return err
			}
			continue
		}
		if trimmed == "endif" || strings.HasPrefix(trimmed, "elsif ") || trimmed == "else" {
			// Unexpected top-level control; treat as literal
			if _, err := io.WriteString(w, "{"+code+"}"); err != nil {
				return err
			}
			lastNewline = strings.HasSuffix("{"+code+"}", "\n")
			continue
		}

		// Otherwise treat as interpolation/expression; parse and evaluate and format
		s, err := evalExpressionFormatted(trimmed, env)
		if err != nil {
			return err
		}
		if _, err := io.WriteString(w, s); err != nil {
			return err
		}
		lastNewline = strings.HasSuffix(s, "\n")
	}
}

// readUntilChar reads from br until the delimiter char is consumed. It returns the string
// read before the delimiter (excluding the delimiter). If EOF is reached before the delimiter,
// io.EOF is returned as error.
func readUntilChar(br *bufio.Reader, delim byte) (string, error) {
	var sb strings.Builder
	for {
		b, err := br.ReadByte()
		if err != nil {
			return sb.String(), err
		}
		if b == delim {
			return sb.String(), nil
		}
		sb.WriteByte(b)
	}
}

// evalExpressionRaw parses the expression in `code` and returns the raw evaluated value
func evalExpressionRaw(code string, env map[string]any) (any, error) {
	l := lexer.New("{" + code + "}")
	p := New(l)
	node := p.parseInterpolation(token.RBRACE)
	if node == nil {
		return nil, fmt.Errorf("failed to parse expression: %s", code)
	}
	interp := node.(*nodes.InterpolationNode)
	// Evaluate the underlying expression to get raw typed value
	return interp.Expression.Evaluate(env)
}

func evalExpressionFormatted(code string, env map[string]any) (string, error) {
	v, err := evalExpressionRaw(code, env)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", v), nil
}

func handleConditionalStream(br *bufio.Reader, w io.Writer, env map[string]any, firstConditionExpr string, lastNewline *bool) error {
	// Evaluate first condition
	val, err := evalExpressionRaw(firstConditionExpr, env)
	if err != nil {
		return err
	}
	chosen := isTruthy(val)
	// If chosen, we should stream nodes until we hit an elsif/else/endif at this level
	// If not chosen, skip until we find an elsif/else/endif and evaluate those.

	for {
		// Read text and braces until we reach a branch marker or endif at this level
		text, err := readUntilChar(br, '{')
		if err == io.EOF {
			return fmt.Errorf("unterminated conditional")
		}
		if err != nil {
			return err
		}

		// If chosen, write text
		if chosen {
			// Avoid extra leading blank line: if text starts with newlines, remove one or all
			if strings.HasPrefix(text, "\n") {
				if *lastNewline {
					text = strings.TrimLeft(text, "\n")
				} else {
					text = strings.TrimPrefix(text, "\n")
				}
			}
			if len(text) > 0 {
				if _, err := io.WriteString(w, text); err != nil {
					return err
				}
				*lastNewline = strings.HasSuffix(text, "\n")
			}
		}

		// Read the code inside braces
		code, err := readUntilChar(br, '}')
		if err != nil {
			return fmt.Errorf("unterminated '{' in conditional body: %w", err)
		}
		trimmed := strings.TrimSpace(code)

		// Handle nested conditional
		if strings.HasPrefix(trimmed, "if ") {
			if chosen {
				// Emit nested conditional result
				if err := handleConditionalStream(br, w, env, trimmed[len("if "):], lastNewline); err != nil {
					return err
				}
			} else {
				// Skip entire nested conditional
				if err := skipConditional(br); err != nil {
					return err
				}
			}
			continue
		}

		// Branch markers at current level
		if strings.HasPrefix(trimmed, "elsif ") {
			if chosen {
				// We have already emitted chosen branch; skip rest until endif
				if err := skipToEndOfConditional(br); err != nil {
					return err
				}
				return nil
			}
			// Not chosen yet: evaluate this elsif
			val, err := evalExpressionRaw(strings.TrimSpace(trimmed[len("elsif "):]), env)
			if err != nil {
				return err
			}
			chosen = isTruthy(val)
			continue
		}

		if trimmed == "else" {
			if chosen {
				// Already emitted branch; skip rest until endif
				if err := skipToEndOfConditional(br); err != nil {
					return err
				}
				return nil
			}
			// This else is our chosen branch now
			chosen = true
			continue
		}

		if trimmed == "endif" {
			// Done with conditional
			return nil
		}

		// Regular interpolation or other code inside the body
		if chosen {
			// Evaluate and format
			s, err := evalExpressionFormatted(trimmed, env)
			if err != nil {
				return err
			}
			if _, err := io.WriteString(w, s); err != nil {
				return err
			}
			*lastNewline = strings.HasSuffix(s, "\n")
		} else {
			// not chosen: ignore
		}
	}
}

// skipConditional consumes tokens for a nested conditional without emitting anything.
func skipConditional(br *bufio.Reader) error {
	nested := 1
	for nested > 0 {
		// read until next '{'
		_, err := readUntilChar(br, '{')
		if err == io.EOF {
			return fmt.Errorf("unterminated nested conditional")
		}
		if err != nil {
			return err
		}
		code, err := readUntilChar(br, '}')
		if err != nil {
			return err
		}
		trimmed := strings.TrimSpace(code)
		if strings.HasPrefix(trimmed, "if ") {
			nested++
			continue
		}
		if trimmed == "endif" {
			nested--
			continue
		}
		// other tokens ignored
	}
	return nil
}

// skipToEndOfConditional skips the rest of the current conditional until the matching endif.
func skipToEndOfConditional(br *bufio.Reader) error {
	return skipConditional(br)
}

func isTruthy(val any) bool {
	switch v := val.(type) {
	case bool:
		return v
	case nil:
		return false
	default:
		return true
	}
}
