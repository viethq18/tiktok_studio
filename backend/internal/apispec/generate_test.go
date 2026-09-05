package apispec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tks/backend/internal/apispec"
)

// The generated OpenAPI document and clients are checked in so the mobile and
// web apps build without running Go. This test is what stops them going stale:
// change a Go struct without running `make apigen` and it fails here rather
// than at runtime on someone's phone.
func TestGeneratedClientsAreUpToDate(t *testing.T) {
	files, err := apispec.Generate()
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	root := filepath.Join("..", "..", "..")

	for name, want := range files {
		path := filepath.Join(root, name)
		got, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("%s is missing — run `make apigen`", name)
			continue
		}
		if string(got) != string(want) {
			t.Errorf("%s is stale — run `make apigen` and commit the result", name)
		}
	}
}

// Every operation must be reachable and uniquely named, since the operation id
// becomes a method name in the generated clients.
func TestOperationIDsAreUniqueAndComplete(t *testing.T) {
	seen := map[string]bool{}
	for _, op := range apispec.Operations() {
		if op.OperationID == "" {
			t.Errorf("%s %s has no operation id", op.Method, op.Path)
		}
		if seen[op.OperationID] {
			t.Errorf("duplicate operation id %q — it would collide as a client method", op.OperationID)
		}
		seen[op.OperationID] = true
		if op.Summary == "" {
			t.Errorf("%s has no summary; it would generate an undocumented method", op.OperationID)
		}
	}
}
