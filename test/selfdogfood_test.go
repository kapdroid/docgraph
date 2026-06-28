package test

import (
	"io"
	"path/filepath"
	"testing"

	"github.com/kapdroid/docgraph/internal/app"
)

// TestSelfDogfood runs docgraph over its OWN docgraph.yml: the repo's documentation must stay
// link-clean. This is docgraph being its own first customer (kap-ymj.18).
func TestSelfDogfood(t *testing.T) {
	passed, err := app.Run(filepath.Join("..", "docgraph.yml"), io.Discard)
	if err != nil {
		t.Fatalf("self-dogfood could not run: %v", err)
	}
	if !passed {
		t.Error("docgraph's own docs failed docgraph.yml (run: docgraph -config docgraph.yml)")
	}
}
