package extract

import (
	"sort"
	"testing"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

func agentGraph() *graph.Graph {
	g := graph.New()
	g.AddNode(graph.Node{ID: ".claude/agents/go-reviewer.md", Kind: "agents", Path: "testdata/rc/.claude/agents/go-reviewer.md"})
	g.AddNode(graph.Node{ID: "stacks/go/reviewers/go-reviewer.md", Kind: "reviewers", Path: "testdata/rc/stacks/go/reviewers/go-reviewer.md"})
	return g
}

func TestRegexCite(t *testing.T) {
	g := agentGraph()
	r := NewRegistry()
	rule := config.EdgeRule{Type: "regex-cite", From: "agents", Pattern: `stacks/[a-z_]+/reviewers/[a-z-]+\.md`}
	if err := r.Extract(g, []config.EdgeRule{rule}); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	var targets []string
	for _, e := range g.Outbound(".claude/agents/go-reviewer.md") {
		if e.Type != "regex-cite" {
			t.Errorf("edge type = %q, want regex-cite", e.Type)
		}
		targets = append(targets, e.To)
	}
	sort.Strings(targets)
	want := []string{"stacks/go/reviewers/extra-reviewer.md", "stacks/go/reviewers/go-reviewer.md"}
	if len(targets) != len(want) || targets[0] != want[0] || targets[1] != want[1] {
		t.Errorf("targets = %v, want %v", targets, want)
	}
}

func TestRegexCiteCaptureGroup(t *testing.T) {
	g := agentGraph()
	r := NewRegistry()
	// a capture group selects the path out of a larger match
	rule := config.EdgeRule{Type: "regex-cite", From: "agents", Pattern: `rubric at (stacks/\S+\.md)`}
	if err := r.Extract(g, []config.EdgeRule{rule}); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	edges := g.Outbound(".claude/agents/go-reviewer.md")
	if len(edges) != 1 || edges[0].To != "stacks/go/reviewers/go-reviewer.md" {
		t.Errorf("edges = %+v, want one edge to the captured path", edges)
	}
}

func TestRegexCiteErrors(t *testing.T) {
	g := agentGraph()
	r := NewRegistry()
	if err := r.Extract(g, []config.EdgeRule{{Type: "regex-cite", From: "agents"}}); err == nil {
		t.Error("regex-cite with no pattern = nil error, want failure")
	}
	if err := r.Extract(g, []config.EdgeRule{{Type: "regex-cite", From: "agents", Pattern: "([unclosed"}}); err == nil {
		t.Error("regex-cite with invalid pattern = nil error, want failure")
	}
}
