package assert

import (
	"sort"
	"testing"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// testGraph: docs root→mid→leaf→ghost(dangling); docs "orphan" and adrs "lonely" both unreferenced.
//
//	root → mid → leaf → ghost(missing)
//	orphan (docs, no inbound)
//	lonely (adrs, no inbound)
func testGraph() *graph.Graph {
	g := graph.New()
	for _, id := range []string{"root", "mid", "leaf", "orphan"} {
		g.AddNode(graph.Node{ID: id, Kind: "docs", Path: id + ".md"})
	}
	g.AddNode(graph.Node{ID: "lonely", Kind: "adrs", Path: "lonely.md"})
	g.AddEdge(graph.Edge{From: "root", To: "mid", Type: "markdown-link", Loc: "root.md:1"})
	g.AddEdge(graph.Edge{From: "mid", To: "leaf", Type: "markdown-link", Loc: "mid.md:1"})
	g.AddEdge(graph.Edge{From: "leaf", To: "ghost", Type: "markdown-link", Loc: "leaf.md:1"})
	return g
}

func findingNodes(fs []Finding) []string {
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = f.Node
	}
	sort.Strings(out)
	return out
}

func TestNoDangling(t *testing.T) {
	fs := noDangling{}.Check(testGraph(), config.Assertion{Type: "no-dangling"})
	if len(fs) != 1 {
		t.Fatalf("no-dangling findings = %d, want 1", len(fs))
	}
	if fs[0].Edge == nil || fs[0].Edge.To != "ghost" {
		t.Errorf("finding = %+v, want the leaf→ghost edge", fs[0])
	}
	if fs[0].Severity != SeverityError {
		t.Errorf("finding severity = %q, want %q", fs[0].Severity, SeverityError)
	}
}

func TestNoOrphan(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{"scoped to docs", []string{"docs"}, []string{"orphan", "root"}},
		{"all kinds", nil, []string{"lonely", "orphan", "root"}},
		{"scoped to adrs", []string{"adrs"}, []string{"lonely"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := noOrphan{}.Check(testGraph(), config.Assertion{Type: "no-orphan", In: tc.in})
			got := findingNodes(fs)
			if len(got) != len(tc.want) {
				t.Fatalf("orphans = %v, want %v", got, tc.want)
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Fatalf("orphans = %v, want %v", got, tc.want)
				}
			}
		})
	}
}

func TestRegistryCheck(t *testing.T) {
	r := NewRegistry()
	results, err := r.Check(testGraph(), []config.Assertion{
		{Type: "no-dangling"},
		{Type: "no-orphan", In: []string{"docs"}},
	})
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}
	if results[0].Type != "no-dangling" || results[0].Passed() {
		t.Errorf("result[0] = %+v, want no-dangling failed", results[0])
	}
	if results[1].Passed() {
		t.Errorf("result[1] no-orphan should have failed (root, orphan)")
	}
}

func TestRegistryUnknownType(t *testing.T) {
	r := NewRegistry()
	if _, err := r.Check(testGraph(), []config.Assertion{{Type: "telepathy"}}); err == nil {
		t.Fatal("Check with unknown assert type = nil error, want failure")
	}
}

func TestPassedOnCleanGraph(t *testing.T) {
	g := graph.New()
	g.AddNode(graph.Node{ID: "a", Kind: "docs"})
	g.AddNode(graph.Node{ID: "b", Kind: "docs"})
	g.AddEdge(graph.Edge{From: "a", To: "b", Type: "markdown-link"})
	// b is referenced; a is a root (zero inbound) → orphan only if we don't scope it out. Check that
	// no-dangling passes cleanly here.
	fs := noDangling{}.Check(g, config.Assertion{Type: "no-dangling"})
	if len(fs) != 0 {
		t.Errorf("no-dangling on clean graph = %v, want none", fs)
	}
}
