package assert

import (
	"testing"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
	"github.com/kapdroid/docgraph/internal/reach"
)

// reachGraph: a consumer reaches stacks; the stack links one rule; a second rule is unreachable.
func reachGraph() *graph.Graph {
	g := graph.New()
	g.AddNode(graph.Node{ID: "stacks/go/stack.yml", Kind: "stacks"})
	g.AddNode(graph.Node{ID: "docs/rules.md", Kind: "rules"})
	g.AddNode(graph.Node{ID: "docs/lonely.md", Kind: "rules"}) // unreachable
	g.AddEdge(graph.Edge{From: "stacks/go/stack.yml", To: "docs/rules.md", Type: "yaml-pointer"})
	cfg := &config.Config{
		Moments:   []string{"lane"},
		Consumers: map[string]config.Consumer{"impl": {Reaches: []string{"stacks"}, At: []string{"lane"}}},
	}
	reach.Materialize(cfg, g)
	return g
}

func TestReachableFlagsUnreachable(t *testing.T) {
	fs := reachable{}.Check(reachGraph(), config.Assertion{Type: "reachable", Set: "rules", From: "consumers"})
	if len(fs) != 1 || fs[0].Node != "docs/lonely.md" {
		t.Fatalf("reachable findings = %+v, want only docs/lonely.md", fs)
	}
}

func TestReachableAllReachablePasses(t *testing.T) {
	g := graph.New()
	g.AddNode(graph.Node{ID: "stacks/go/stack.yml", Kind: "stacks"})
	g.AddNode(graph.Node{ID: "docs/rules.md", Kind: "rules"})
	g.AddEdge(graph.Edge{From: "stacks/go/stack.yml", To: "docs/rules.md", Type: "yaml-pointer"})
	reach.Materialize(&config.Config{Moments: []string{"lane"},
		Consumers: map[string]config.Consumer{"impl": {Reaches: []string{"stacks"}, At: []string{"lane"}}}}, g)
	fs := reachable{}.Check(g, config.Assertion{Type: "reachable", Set: "rules", From: "consumers"})
	if len(fs) != 0 {
		t.Errorf("reachable findings = %+v, want none (all rules reachable)", fs)
	}
}

func TestReachableNoConsumers(t *testing.T) {
	g := graph.New()
	g.AddNode(graph.Node{ID: "docs/rules.md", Kind: "rules"})
	fs := reachable{}.Check(g, config.Assertion{Type: "reachable", Set: "rules", From: "consumers"})
	if len(fs) != 1 {
		t.Fatalf("expected 1 config-level finding for no consumers, got %+v", fs)
	}
}

func TestAcyclic(t *testing.T) {
	t.Run("flags a cycle", func(t *testing.T) {
		g := graph.New()
		g.AddNode(graph.Node{ID: "a", Kind: "docs"})
		g.AddNode(graph.Node{ID: "b", Kind: "docs"})
		g.AddEdge(graph.Edge{From: "a", To: "b", Type: "markdown-link"})
		g.AddEdge(graph.Edge{From: "b", To: "a", Type: "markdown-link"})
		fs := acyclic{}.Check(g, config.Assertion{Type: "acyclic"})
		if len(fs) == 0 {
			t.Fatal("acyclic found no cycle in a↔b graph")
		}
	})
	t.Run("passes a DAG", func(t *testing.T) {
		g := graph.New()
		g.AddNode(graph.Node{ID: "a", Kind: "docs"})
		g.AddNode(graph.Node{ID: "b", Kind: "docs"})
		g.AddNode(graph.Node{ID: "c", Kind: "docs"})
		g.AddEdge(graph.Edge{From: "a", To: "b", Type: "markdown-link"})
		g.AddEdge(graph.Edge{From: "a", To: "c", Type: "markdown-link"})
		g.AddEdge(graph.Edge{From: "b", To: "c", Type: "markdown-link"})
		fs := acyclic{}.Check(g, config.Assertion{Type: "acyclic"})
		if len(fs) != 0 {
			t.Errorf("acyclic flagged a DAG: %+v", fs)
		}
	})
}
