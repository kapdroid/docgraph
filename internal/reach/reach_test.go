package reach

import (
	"testing"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// buildGraph: a stack that yaml-points to a reachable rule, plus an unreachable rule nothing links.
func buildGraph() *graph.Graph {
	g := graph.New()
	g.AddNode(graph.Node{ID: "stacks/go/stack.yml", Kind: "stacks"})
	g.AddNode(graph.Node{ID: "docs/rules.md", Kind: "rules"})
	g.AddNode(graph.Node{ID: "docs/lonely.md", Kind: "rules"}) // unreachable
	g.AddEdge(graph.Edge{From: "stacks/go/stack.yml", To: "docs/rules.md", Type: "yaml-pointer"})
	return g
}

func consumerCfg() *config.Config {
	return &config.Config{
		Moments:   []string{"lane", "edit"},
		Consumers: map[string]config.Consumer{"impl": {Reaches: []string{"stacks"}, At: []string{"lane"}}},
	}
}

func TestMaterialize(t *testing.T) {
	g := buildGraph()
	Materialize(consumerCfg(), g)

	if ids := ConsumerIDs(g); len(ids) != 1 || ids[0] != "consumer:impl" {
		t.Fatalf("ConsumerIDs = %v, want [consumer:impl]", ids)
	}
	if !g.Has("moment:lane") || !g.Has("moment:edit") {
		t.Error("moment nodes not materialized")
	}
	// the consumer reaches the stacks set → an edge consumer→stack
	var reachesStack, atLane bool
	for _, e := range g.Outbound("consumer:impl") {
		if e.Type == "reaches" && e.To == "stacks/go/stack.yml" {
			reachesStack = true
		}
		if e.Type == "at" && e.To == "moment:lane" {
			atLane = true
		}
	}
	if !reachesStack {
		t.Error("no reaches edge consumer→stack")
	}
	if !atLane {
		t.Error("no at edge consumer→moment:lane")
	}
}

func TestReachableFromTransitive(t *testing.T) {
	g := buildGraph()
	Materialize(consumerCfg(), g)
	r := ReachableFrom(g, ConsumerIDs(g))

	cases := []struct {
		id   string
		want bool
	}{
		{"consumer:impl", true},       // source
		{"stacks/go/stack.yml", true}, // directly reached
		{"docs/rules.md", true},       // transitively reached via stack yaml-pointer
		{"docs/lonely.md", false},     // nothing links it → unreachable (the M3 differentiator)
	}
	for _, tc := range cases {
		if got := r[tc.id]; got != tc.want {
			t.Errorf("reachable[%q] = %v, want %v", tc.id, got, tc.want)
		}
	}
}

func TestMaterializeNoConsumersIsNoop(t *testing.T) {
	g := buildGraph()
	before := len(g.Nodes())
	Materialize(&config.Config{}, g)
	if len(g.Nodes()) != before {
		t.Errorf("Materialize with no consumers changed node count %d→%d", before, len(g.Nodes()))
	}
}

func TestReachableFromEmptySources(t *testing.T) {
	if got := ReachableFrom(buildGraph(), nil); len(got) != 0 {
		t.Errorf("ReachableFrom(nil) = %v, want empty", got)
	}
}
