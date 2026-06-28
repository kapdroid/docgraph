package assert

import (
	"fmt"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// registered asserts every node in the From set is referenced by a registry node — e.g. every ADR
// must be linked from the ADR index. A From node with no inbound edge from the registry node is
// flagged (it exists but is not registered). Configured via rule.From (node-set kind) and
// rule.Registry (the registry node's ID).
type registered struct{}

// Type returns the assert type this assertion handles.
func (registered) Type() string { return "registered" }

// Check flags each From-set node not referenced by the registry node. If the registry node itself is
// absent from the graph, that is a single configuration-level finding (nothing can be registered).
func (registered) Check(g *graph.Graph, rule config.Assertion) []Finding {
	if rule.Registry == "" {
		return []Finding{{Severity: SeverityError, Message: "registered: no registry node configured"}}
	}
	if !g.Has(rule.Registry) {
		return []Finding{{
			Severity: SeverityError,
			Node:     rule.Registry,
			Message:  fmt.Sprintf("registered: registry node %q does not exist", rule.Registry),
		}}
	}
	var findings []Finding
	for _, n := range g.NodesOfKind(rule.From) {
		if n.ID == rule.Registry {
			continue // the registry need not register itself
		}
		if referencedBy(g, n.ID, rule.Registry) {
			continue
		}
		findings = append(findings, Finding{
			Severity: SeverityError,
			Node:     n.ID,
			Message:  fmt.Sprintf("%s (%s) is not registered in %s", n.ID, n.Kind, rule.Registry),
		})
	}
	return findings
}

// referencedBy reports whether targetID has an inbound edge from the node fromID.
func referencedBy(g *graph.Graph, targetID, fromID string) bool {
	for _, e := range g.Inbound(targetID) {
		if e.From == fromID {
			return true
		}
	}
	return false
}
