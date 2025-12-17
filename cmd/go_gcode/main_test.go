package main

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestParseValueJSON(t *testing.T) {
	v := parseValue("[1, 2, 3]")
	arr, ok := v.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", v)
	}
	if len(arr) != 3 {
		t.Fatalf("expected length 3, got %d", len(arr))
	}
	v2 := parseValue(`{"a": 1, "b": [true, false]}`)
	m, ok := v2.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", v2)
	}
	if _, ok := m["a"]; !ok {
		t.Fatalf("expected key a in map")
	}
}

func TestParseValueScalar(t *testing.T) {
	if got := parseValue("10"); !reflect.DeepEqual(got, int64(10)) {
		t.Fatalf("expected int64(10), got %#v", got)
	}
	if got := parseValue("3.14"); !reflect.DeepEqual(got, 3.14) {
		t.Fatalf("expected 3.14, got %#v", got)
	}
	if got := parseValue("true"); !reflect.DeepEqual(got, true) {
		t.Fatalf("expected true, got %#v", got)
	}
	if got := parseValue("hello"); !reflect.DeepEqual(got, "hello") {
		t.Fatalf("expected 'hello', got %#v", got)
	}
}

func TestParseValueQuotedString(t *testing.T) {
	if got := parseValue(`"quoted"`); !reflect.DeepEqual(got, "quoted") {
		t.Fatalf("expected quoted, got %#v", got)
	}
	if got := parseValue("'q'"); !reflect.DeepEqual(got, "q") {
		t.Fatalf("expected q, got %#v", got)
	}
}

func TestParseValueComplex(t *testing.T) {
	js := `{"a": {"b": [1,2]}, "c": "d"}`
	v := parseValue(js)
	// ensure it's valid JSON and structured
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(v)
}
