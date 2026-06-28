package assert

import (
	"fmt"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// cites asserts every node in the From set has at least one outbound edge into the To set — e.g.
// every reviewer rubric must cite at least one rules doc. A From node with no qualifying edge is
// flagged. Configured via rule.From and rule.To (both node-set kinds).
type cites struct{}

// Type returns the assert type this assertion handles.
func (cites) Type() string { return "cites" }

// Check flags each From-set node that does not edge to any node in the To set.
func (cites) Check(g *graph.Graph, rule config.Assertion) []Finding {
	var findings []Finding
	for _, n := range g.NodesOfKind(rule.From) {
		if edgesIntoKind(g, n.ID, rule.To) {
			continue
		}
		findings = append(findings, Finding{
			Severity: SeverityError,
			Node:     n.ID,
			Message:  fmt.Sprintf("%s (%s) does not cite any %s", n.ID, n.Kind, rule.To),
		})
	}
	return findings
}

// edgesIntoKind reports whether any outbound edge of fromID targets an existing node of the given kind.
func edgesIntoKind(g *graph.Graph, fromID, kind string) bool {
	for _, e := range g.Outbound(fromID) {
		if t, ok := g.Node(e.To); ok && t.Kind == kind {
			return true
		}
	}
	return false
}
