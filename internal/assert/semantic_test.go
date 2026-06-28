package assert

import (
	"testing"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

func TestCites(t *testing.T) {
	g := graph.New()
	g.AddNode(graph.Node{ID: "r1", Kind: "reviewers"})
	g.AddNode(graph.Node{ID: "r2", Kind: "reviewers"}) // cites nothing → flagged
	g.AddNode(graph.Node{ID: "d1", Kind: "docs"})
	g.AddEdge(graph.Edge{From: "r1", To: "d1", Type: "regex-cite"})
	fs := cites{}.Check(g, config.Assertion{Type: "cites", From: "reviewers", To: "docs"})
	if len(fs) != 1 || fs[0].Node != "r2" {
		t.Fatalf("cites findings = %+v, want only r2", fs)
	}
}

func TestRegistered(t *testing.T) {
	g := graph.New()
	g.AddNode(graph.Node{ID: "index", Kind: "docs"})
	g.AddNode(graph.Node{ID: "adr-1", Kind: "adrs"})
	g.AddNode(graph.Node{ID: "adr-2", Kind: "adrs"}) // not linked from index → flagged
	g.AddEdge(graph.Edge{From: "index", To: "adr-1", Type: "markdown-link"})

	t.Run("flags unregistered", func(t *testing.T) {
		fs := registered{}.Check(g, config.Assertion{Type: "registered", From: "adrs", Registry: "index"})
		if len(fs) != 1 || fs[0].Node != "adr-2" {
			t.Fatalf("registered findings = %+v, want only adr-2", fs)
		}
	})
	t.Run("missing registry node", func(t *testing.T) {
		fs := registered{}.Check(g, config.Assertion{Type: "registered", From: "adrs", Registry: "nope"})
		if len(fs) != 1 {
			t.Fatalf("expected 1 finding for absent registry, got %+v", fs)
		}
	})
	t.Run("no registry configured", func(t *testing.T) {
		fs := registered{}.Check(g, config.Assertion{Type: "registered", From: "adrs"})
		if len(fs) != 1 {
			t.Fatalf("expected 1 finding for unconfigured registry, got %+v", fs)
		}
	})
}

func TestConsistent(t *testing.T) {
	g := graph.New()
	// correct: parsed scope equals the derived expectation
	g.AddNode(graph.Node{ID: "ok.md", Kind: "adrs", Frontmatter: map[string]string{
		"scope": "stack:go", graph.DerivedPrefix + "scope": "stack:go",
	}})
	// typo'd scope: parsed differs from derived → flagged (the M2 DoD)
	g.AddNode(graph.Node{ID: "typo.md", Kind: "adrs", Frontmatter: map[string]string{
		"scope": "engine", graph.DerivedPrefix + "scope": "stack:go",
	}})
	// missing scope: parsed empty, derived set → flagged
	g.AddNode(graph.Node{ID: "missing.md", Kind: "adrs", Frontmatter: map[string]string{
		graph.DerivedPrefix + "scope": "engine",
	}})
	// not annotated (no derivation) → never flagged
	g.AddNode(graph.Node{ID: "plain.md", Kind: "docs", Frontmatter: map[string]string{"scope": "whatever"}})

	fs := consistent{}.Check(g, config.Assertion{Type: "consistent"})
	flagged := map[string]bool{}
	for _, f := range fs {
		flagged[f.Node] = true
	}
	if len(fs) != 2 || !flagged["typo.md"] || !flagged["missing.md"] {
		t.Fatalf("consistent findings = %+v, want exactly {typo.md, missing.md}", fs)
	}
}

func TestSemanticAssertsRegistered(t *testing.T) {
	r := NewRegistry()
	for _, typ := range []string{"cites", "registered", "consistent"} {
		if _, ok := r.Get(typ); !ok {
			t.Errorf("assert %q not registered", typ)
		}
	}
}
