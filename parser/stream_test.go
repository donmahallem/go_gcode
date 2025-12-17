package parser

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseAndEmitBasic(t *testing.T) {
	input := `
M104 S75 ; comment
{if true}
  M106
{endif}
G1 X{10+20}
`
	r := strings.NewReader(input)
	var b bytes.Buffer
	if err := ParseAndEmit(r, &b, map[string]any{}); err != nil {
		t.Fatalf("ParseAndEmit error: %v", err)
	}
	expected := "M104 S75 ; comment\n  M106\nG1 X30\n"
	if b.String() != expected {
		t.Fatalf("expected %q, got %q", expected, b.String())
	}
}

func TestParseAndEmitConditionalElse(t *testing.T) {
	input := `
{if false}
A
{elsif true}
B
{else}
C
{endif}
`
	r := strings.NewReader(input)
	var b bytes.Buffer
	if err := ParseAndEmit(r, &b, map[string]any{}); err != nil {
		t.Fatalf("ParseAndEmit error: %v", err)
	}
	if b.String() != "B\n" {
		t.Fatalf("expected B, got %q", b.String())
	}
}

func TestParseAndEmitNestedIf(t *testing.T) {
	input := `
{if true}
X
{if false}
A
{else}
B
{endif}
Y
{endif}
`
	r := strings.NewReader(input)
	var b bytes.Buffer
	if err := ParseAndEmit(r, &b, map[string]any{}); err != nil {
		t.Fatalf("ParseAndEmit error: %v", err)
	}
	if b.String() != "X\nB\nY\n" {
		t.Fatalf("expected nested output, got %q", b.String())
	}
}
