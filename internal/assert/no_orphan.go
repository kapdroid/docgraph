package assert

import (
	"fmt"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// noOrphan flags nodes that nothing references (zero inbound edges). When rule.In is set, only nodes
// of those kinds are checked — so a config can scope the check to the sets where being unreferenced
// is a defect (e.g. docs, adrs) and ignore sets that are legitimately only sources (e.g. agents).
//
// M1 scope: "orphan" here means strictly zero inbound edges. Entry-point exemption (a node that is a
// legitimate root reached by a consumer, not by another doc) arrives with the M3 consumers×moments
// reachability model; until then, scope the check via rule.In.
type noOrphan struct{}

// Type returns the assert type this assertion handles.
func (noOrphan) Type() string { return "no-orphan" }

// Check returns a finding for each in-scope node with no inbound edges.
func (noOrphan) Check(g *graph.Graph, rule config.Assertion) []Finding {
	inScope := kindSet(rule.In)
	var findings []Finding
	for _, n := range g.Nodes() {
		if len(inScope) > 0 && !inScope[n.Kind] {
			continue
		}
		if len(g.Inbound(n.ID)) > 0 {
			continue
		}
		findings = append(findings, Finding{
			Severity: SeverityError,
			Node:     n.ID,
			Message:  fmt.Sprintf("%s (%s) is an orphan — nothing references it", n.ID, n.Kind),
		})
	}
	return findings
}

// kindSet turns a list of node-set kinds into a set for O(1) membership; empty list → empty set
// (meaning "all kinds in scope").
func kindSet(kinds []string) map[string]bool {
	if len(kinds) == 0 {
		return nil
	}
	s := make(map[string]bool, len(kinds))
	for _, k := range kinds {
		s[k] = true
	}
	return s
}
