package extract

import (
	"testing"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

func stackGraph() *graph.Graph {
	g := graph.New()
	g.AddNode(graph.Node{ID: "stacks/go/stack.yml", Kind: "stacks", Path: "testdata/yp/stacks/go/stack.yml"})
	g.AddNode(graph.Node{ID: "docs/rules.md", Kind: "docs", Path: "testdata/yp/docs/rules.md"})
	g.AddNode(graph.Node{ID: "docs/notes.md", Kind: "docs", Path: "testdata/yp/docs/notes.md"})
	return g
}

func TestYAMLPointer(t *testing.T) {
	cases := []struct {
		name   string
		key    string
		wantTo string // "" = expect no edge
	}{
		{"top-level key", "rules", "docs/rules.md"},
		{"nested dotted key", "overrides.notes", "docs/notes.md"},
		{"absent key yields no edge", "missing", ""},
		{"non-scalar (map) yields no edge", "overrides", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := stackGraph()
			r := NewRegistry()
			err := r.Extract(g, []config.EdgeRule{{Type: "yaml-pointer", From: "stacks", Key: tc.key, To: "docs"}})
			if err != nil {
				t.Fatalf("Extract: %v", err)
			}
			edges := g.Outbound("stacks/go/stack.yml")
			if tc.wantTo == "" {
				if len(edges) != 0 {
					t.Fatalf("expected no edge, got %+v", edges)
				}
				return
			}
			if len(edges) != 1 {
				t.Fatalf("expected 1 edge, got %+v", edges)
			}
			if edges[0].To != tc.wantTo || edges[0].Type != "yaml-pointer" {
				t.Errorf("edge = %+v, want To=%q type=yaml-pointer", edges[0], tc.wantTo)
			}
		})
	}
}

func TestYAMLPointerMissingKeyConfig(t *testing.T) {
	g := stackGraph()
	r := NewRegistry()
	if err := r.Extract(g, []config.EdgeRule{{Type: "yaml-pointer", From: "stacks"}}); err == nil {
		t.Fatal("yaml-pointer with no key = nil error, want failure")
	}
}
