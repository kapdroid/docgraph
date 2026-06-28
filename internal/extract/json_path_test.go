package extract

import (
	"sort"
	"testing"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

func jsonGraph() *graph.Graph {
	g := graph.New()
	g.AddNode(graph.Node{ID: ".config.json", Kind: "config", Path: "testdata/jp/.config.json"})
	return g
}

func extractJSONTargets(t *testing.T, path string) []string {
	t.Helper()
	g := jsonGraph()
	r := NewRegistry()
	if err := r.Extract(g, []config.EdgeRule{{Type: "json-path", From: "config", Path: path}}); err != nil {
		t.Fatalf("Extract(%q): %v", path, err)
	}
	var targets []string
	for _, e := range g.Outbound(".config.json") {
		if e.Type != "json-path" {
			t.Errorf("edge type = %q, want json-path", e.Type)
		}
		targets = append(targets, e.To)
	}
	sort.Strings(targets)
	return targets
}

func TestJSONPath(t *testing.T) {
	cases := []struct {
		name string
		path string
		want []string
	}{
		{"array of strings", "docs[]", []string{"docs/a.md", "docs/b.md"}},
		{"nested array of objects", "hooks.PreToolUse[].command", []string{"scripts/beat.sh", "scripts/guard.sh"}},
		{"absent key yields nothing", "missing[].x", nil},
		{"non-array where [] expected yields nothing", "docs.x", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := extractJSONTargets(t, tc.path)
			if len(got) != len(tc.want) {
				t.Fatalf("targets = %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("targets = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestJSONPathErrors(t *testing.T) {
	g := jsonGraph()
	r := NewRegistry()
	if err := r.Extract(g, []config.EdgeRule{{Type: "json-path", From: "config"}}); err == nil {
		t.Error("json-path with no path = nil error, want failure")
	}
}
