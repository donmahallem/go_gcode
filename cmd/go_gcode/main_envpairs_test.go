package main

import (
	"reflect"
	"testing"
)

func TestSplitEnvPairsBasic(t *testing.T) {
	input := `a=1,b=2,c=3`
	parts := splitEnvPairs(input)
	exp := []string{"a=1", "b=2", "c=3"}
	if !reflect.DeepEqual(parts, exp) {
		t.Fatalf("expected %v, got %v", exp, parts)
	}
}

func TestSplitEnvPairsJSONValue(t *testing.T) {
	input := `a=1,b=[1,2,3],c={"x":"y"},d=hello`
	parts := splitEnvPairs(input)
	exp := []string{"a=1", "b=[1,2,3]", "c={\"x\":\"y\"}", "d=hello"}
	if !reflect.DeepEqual(parts, exp) {
		t.Fatalf("expected %v, got %v", exp, parts)
	}
}

func TestApplyEnvPairsOverrides(t *testing.T) {
	env := map[string]any{"a": 1, "b": 2}
	if err := applyEnvPairs(env, "b=3,c=4"); err != nil {
		t.Fatalf("applyEnvPairs failed: %v", err)
	}
	if env["a"] != 1 {
		t.Fatalf("expected a==1, got %v", env["a"])
	}
	if env["b"] != int64(3) {
		t.Fatalf("expected b==3, got %v", env["b"])
	}
	if env["c"] != int64(4) {
		t.Fatalf("expected c==4, got %v", env["c"])
	}
}
