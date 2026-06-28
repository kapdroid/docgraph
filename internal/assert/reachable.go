package assert

import (
	"fmt"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
	"github.com/kapdroid/docgraph/internal/reach"
)

// reachable is the M3 differentiator: it asserts every node in a target set is reachable from the
// consumer entry nodes by following edges of ANY type. A rule/doc that exists but no consumer can
// reach — through links, yaml pointers, citations, anything — is flagged. This is the question generic
// link-checkers cannot answer. Configured via rule.Set (the target node-set) and rule.From: the
// literal keyword "consumers" (default — entries are the materialized consumer nodes, the .5/.15
// seam), or any node-set name to use that set's members as the entry points.
type reachable struct{}

// Type returns the assert type this assertion handles.
func (reachable) Type() string { return "reachable" }

// Check flags each node of rule.Set not reachable from the entry nodes. With rule.From == "consumers"
// (or empty), the entries are the injected consumer nodes; the reachable set is the BFS closure.
func (reachable) Check(g *graph.Graph, rule config.Assertion) []Finding {
	sources := entrySources(g, rule.From)
	if len(sources) == 0 {
		return []Finding{{
			Severity: SeverityError,
			Message:  fmt.Sprintf("reachable: no entry sources for from=%q (are consumers configured?)", rule.From),
		}}
	}
	reached := reach.ReachableFrom(g, sources)
	var findings []Finding
	for _, n := range g.NodesOfKind(rule.Set) {
		if reached[n.ID] {
			continue
		}
		findings = append(findings, Finding{
			Severity: SeverityError,
			Node:     n.ID,
			Message:  fmt.Sprintf("%s (%s) is unreachable from %s", n.ID, n.Kind, rule.From),
		})
	}
	return findings
}

// entrySources resolves the reachability entry node IDs for a reachable rule. "consumers" (the
// default) means the injected consumer nodes; any other value names a node-set whose members are
// entries.
func entrySources(g *graph.Graph, from string) []string {
	if from == "" || from == "consumers" {
		return reach.ConsumerIDs(g)
	}
	var ids []string
	for _, n := range g.NodesOfKind(from) {
		ids = append(ids, n.ID)
	}
	return ids
}
