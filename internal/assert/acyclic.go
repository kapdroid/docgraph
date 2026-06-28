package assert

import (
	"strings"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// acyclic asserts the reference graph has no directed cycle, reporting each back-edge it finds with
// the cycle path. Optional (a doc graph may legitimately have mutual links) — enable it only where a
// DAG is required. Traversal follows edges to existing nodes; dangling edges are skipped.
type acyclic struct{}

// Type returns the assert type this assertion handles.
func (acyclic) Type() string { return "acyclic" }

// node colors for the depth-first cycle search.
const (
	white = iota // unvisited
	gray         // on the current DFS stack
	black        // fully explored
)

// Check returns a finding for each back-edge (a cycle) found by a depth-first search over the graph.
func (acyclic) Check(g *graph.Graph, _ config.Assertion) []Finding {
	color := map[string]int{}
	var findings []Finding
	var path []string

	var visit func(id string)
	visit = func(id string) {
		color[id] = gray
		path = append(path, id)
		for _, e := range g.Outbound(id) {
			if !g.Has(e.To) {
				continue
			}
			switch color[e.To] {
			case white:
				visit(e.To)
			case gray: // back-edge closing a cycle
				findings = append(findings, Finding{
					Severity: SeverityError,
					Node:     e.To,
					Message:  "cycle: " + cyclePath(path, e.To),
				})
			}
		}
		path = path[:len(path)-1]
		color[id] = black
	}

	for _, n := range g.Nodes() { // sorted → deterministic
		if color[n.ID] == white {
			visit(n.ID)
		}
	}
	return findings
}

// cyclePath renders the cycle closing at `to`: the slice of path from where `to` first appears,
// followed by `to` again to show the closure (e.g. "a → b → a"). `to` is always on `path` here (it is
// the gray node a back-edge points at), so the start-at-0 default is never actually used.
func cyclePath(path []string, to string) string {
	start := 0
	for i, id := range path {
		if id == to {
			start = i
			break
		}
	}
	return strings.Join(append(append([]string{}, path[start:]...), to), " → ")
}
