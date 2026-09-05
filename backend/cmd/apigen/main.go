// Command apigen derives the API contract from the Go types and writes the
// OpenAPI document plus the generated clients. Run it with `make apigen`;
// a test fails if the checked-in output is stale.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/tks/backend/internal/apispec"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	files, err := apispec.Generate()
	if err != nil {
		fmt.Fprintln(os.Stderr, "apigen:", err)
		os.Exit(1)
	}
	for name, body := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "apigen:", err)
			os.Exit(1)
		}
		if err := os.WriteFile(path, body, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "apigen:", err)
			os.Exit(1)
		}
		fmt.Printf("  wrote %s (%d bytes)\n", name, len(body))
	}
}
