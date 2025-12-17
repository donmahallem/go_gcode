package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/donmahallem/go_gcode/parser"
)

func main() {
	var (
		inputPath  = flag.String("input", "", "Path to input file (omit or use '-' for stdin)")
		outputPath = flag.String("output", "", "Path to output file (omit for stdout)")
		indent     = flag.Int("indent-spaces", 2, "Number of spaces per indent level")
		envFile    = flag.String("env-file", "", "Path to JSON file with environment values (preferred)")
		envPairs   = flag.String("env", "", "Comma-separated list of key=val pairs (e.g. a=1,b=true,name=foo). Values may be JSON (use env-file if complex)")
	)
	flag.Parse()

	// Resolve input reader
	var in io.Reader
	if *inputPath == "" || *inputPath == "-" {
		in = os.Stdin
	} else {
		f, err := os.Open(*inputPath)
		if err != nil {
			log.Fatalf("failed to open input %s: %v", *inputPath, err)
		}
		defer f.Close()
		in = f
	}

	// Resolve output writer
	var out io.Writer
	if *outputPath == "" {
		out = os.Stdout
	} else {
		if err := os.MkdirAll(filepath.Dir(*outputPath), 0o755); err != nil {
			log.Fatalf("failed to create output dir: %v", err)
		}
		f, err := os.Create(*outputPath)
		if err != nil {
			log.Fatalf("failed to create output %s: %v", *outputPath, err)
		}
		defer f.Close()
		out = f
	}

	// Build environment map from env-file (JSON preferred)
	env, err := loadEnv(*envFile)
	if err != nil {
		log.Fatalf("failed to load environment: %v", err)
	}
	// Ensure indent_spaces is set or overridden by CLI flag
	env["indent_spaces"] = *indent

	// Apply CLI env pairs to override or add values from env-file
	if *envPairs != "" {
		if err := applyEnvPairs(env, *envPairs); err != nil {
			log.Fatalf("invalid env pairs: %v", err)
		}
	}

	// Run streaming parse+emit
	if err := parser.ParseAndEmit(in, out, env); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func loadEnv(envFile string) (map[string]any, error) {
	env := make(map[string]any)

	if envFile == "" {
		return env, nil
	}
	data, err := os.ReadFile(envFile)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err == nil {
		for k, v := range m {
			env[k] = v
		}
		return env, nil
	}
	// fallback: key=value per line
	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		kv := strings.SplitN(l, "=", 2)
		if len(kv) != 2 {
			continue
		}
		env[kv[0]] = parseValue(kv[1])
	}
	return env, nil
}

func applyEnvPairs(env map[string]any, pairs string) error {
	for _, p := range splitEnvPairs(pairs) {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.SplitN(p, "=", 2)
		if len(kv) != 2 {
			return fmt.Errorf("invalid pair %q, expected key=val", p)
		}
		env[kv[0]] = parseValue(kv[1])
	}
	return nil
}

// splitEnvPairs splits a comma-separated list but ignores commas inside JSON braces or quotes.
func splitEnvPairs(s string) []string {
	var parts []string
	var b strings.Builder
	depth := 0
	inSingle := false
	inDouble := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		switch ch {
		case '{', '[':
			if !inSingle && !inDouble {
				depth++
			}
		case '}', ']':
			if !inSingle && !inDouble && depth > 0 {
				depth--
			}
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case ',':
			if depth == 0 && !inSingle && !inDouble {
				parts = append(parts, b.String())
				b.Reset()
				continue
			}
		}
		b.WriteByte(ch)
	}
	if b.Len() > 0 {
		parts = append(parts, b.String())
	}
	return parts
}

func parseValue(s string) any {
	trim := strings.TrimSpace(s)
	// If it looks like JSON object/array, try to unmarshal
	if strings.HasPrefix(trim, "{") || strings.HasPrefix(trim, "[") {
		var v any
		if err := json.Unmarshal([]byte(trim), &v); err == nil {
			return v
		}
		// fall through to other parsing if JSON fails
	}
	if i, err := strconv.ParseInt(trim, 10, 64); err == nil {
		return i
	}
	if f, err := strconv.ParseFloat(trim, 64); err == nil {
		return f
	}
	if b, err := strconv.ParseBool(trim); err == nil {
		return b
	}
	// Trim surrounding quotes
	trim = strings.Trim(trim, `"'`)
	return trim
}
