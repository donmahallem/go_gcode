# go_gcode 📄

A small Go library for parsing and emitting G-code-like files with embedded control blocks and expressions.

This repository provides two main modes of operation:

- **Full AST mode** — parse the entire document into an in-memory AST and manipulate or emit from it.
- **Streaming mode** — parse and emit on-the-fly for low memory consumption (suitable for very large inputs).

Both modes share a common rendering API (`Emitter`) so you can switch between them without changing the rest of your code.

---

## Quick Start ✨

Install (from your module-aware project):

```bash
# From the repo root (module already in this workspace)
go test ./...
```

Run unit tests and benchmarks:

```bash
# Tests
go test ./...

# Benchmarks (example for nodes)
go test ./nodes -bench . -benchmem
```

---

## Key Concepts 🔑

- Emitter: streaming write API implemented by statement nodes.

```go
// Emitter writes the rendered output directly into an io.Writer.
type Emitter interface {
    Emit(w io.Writer, env map[string]any) error
}

// Helper to collect emitted output into a string
func EmitToString(e Emitter, env map[string]any) (string, error)
```

- Evaluate: expression nodes (numbers, booleans, identifiers) produce values via `Evaluate(env) (any, error)`.

- Streaming parser: use `lexer.NewFromReader(io.Reader)` + `parser.New(tokenSource)` + `p.ParseStream(w, env)` or the convenience `ParseAndEmit(r io.Reader, w io.Writer, env)` helper.

- Incremental parse: use `p.ParseNext()` to read single top-level statements if you need custom streaming logic.

---

## Examples 🧾

1. Full AST parse + emit (useful if you need the AST or transforming it):

```go
l := lexer.New(inputString)
p := parser.New(l)
program := p.ParseProgram()
// Emit whole AST to a writer
_ = nodes.EmitToString(program, env)
```

2. Streaming parse + emit (low memory, suited for 10+ MB files):

```go
r := os.Open("large.gcode")
defer r.Close()

// ParseAndEmit reads, parses and emits streaming to writer w
err := parser.ParseAndEmit(r, os.Stdout, map[string]any{"indent_spaces": 2})
if err != nil {
    log.Fatalf("streaming failed: %v", err)
}
```

3. Parser-level streaming (fine-grained control):

```go
r := os.Open("large.gcode")
lex := lexer.NewFromReader(r)
p := parser.New(lex)
for {
    node, err := p.ParseNext()
    if err == io.EOF { break }
    if err != nil { /* handle */ }
    if e, ok := node.(nodes.Emitter); ok {
        _ = e.Emit(out, env)
    } else {
        v, _ := node.Evaluate(env)
        io.WriteString(out, fmt.Sprintf("%v", v))
    }
}
```

---

## Notes & Design Rationale 💡

- The streaming workflow evaluates control blocks (interpolations and conditionals) inline; for conditionals it evaluates branch conditions and streams only the chosen branch — avoiding building AST for skipped branches.
- Expression evaluation still uses `Evaluate(...)` for typed results; `InterpolationNode` implements `Emit` by evaluating and formatting its expression.
- The streaming lexer (`lexer.NewFromReader`) is compatible with the parser by implementing a `TokenSource`-like interface.

---

## Environment & CLI

- Environment variables:
  - Provide environment via `-env-file path` with a **JSON** object (preferred). Example: `{"arr": [1, 2], "name": "PLA"}`.
  - Values support arrays and maps (JSON), and scalar types (number, boolean, string).

## Developer commands 🔧

- Run all tests: `go test ./...`
- Run benchmarks: `go test ./nodes -bench . -benchmem`
- Run a specific test: `go test ./parser -run TestParseAndEmitBasic -v`

---

If you want, I can add a longer examples page, include migration guidance for existing callers, or push this README into a new CHANGELOG/CONTRIBUTING file.
