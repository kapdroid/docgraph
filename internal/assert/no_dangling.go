package assert

import (
	"fmt"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// noDangling flags every edge whose target node does not exist in the graph — a reference to a
// missing file/id. This is the first thing a link layer must catch.
type noDangling struct{}

// Type returns the assert type this assertion handles.
func (noDangling) Type() string { return "no-dangling" }

// Check returns a finding for each edge pointing at an absent node.
func (noDangling) Check(g *graph.Graph, _ config.Assertion) []Finding {
	var findings []Finding
	for _, e := range g.Edges() {
		if g.Has(e.To) {
			continue
		}
		edge := e
		findings = append(findings, Finding{
			Node:    e.From,
			Edge:    &edge,
			Message: fmt.Sprintf("%s references %q which does not exist (at %s)", e.From, e.To, e.Loc),
		})
	}
	return findings
}
