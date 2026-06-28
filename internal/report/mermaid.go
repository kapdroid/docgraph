package report

import (
	"fmt"
	"io"

	"github.com/kapdroid/docgraph/internal/graph"
)

// Mermaid writes the graph as a Mermaid flowchart (`flowchart LR`), one line per node and per edge.
// Node IDs are mapped to safe sequential tokens (n0, n1, …) because Mermaid identifiers cannot contain
// path characters; the real ID is the node label. Edges are labelled with the extractor type. This is
// a visualization, not a check — it never fails.
func Mermaid(w io.Writer, g *graph.Graph) error {
	ew := &errWriter{w: w}
	ew.printf("flowchart LR\n")

	nodes := g.Nodes()
	token := make(map[string]string, len(nodes))
	for i, n := range nodes {
		token[n.ID] = fmt.Sprintf("n%d", i)
	}
	for _, n := range nodes {
		ew.printf("  %s[%q]\n", token[n.ID], n.ID)
	}
	for _, e := range g.Edges() {
		from, ok := token[e.From]
		if !ok {
			continue // edge from a node not in the graph (shouldn't happen) — skip
		}
		to, ok := token[e.To]
		if !ok {
			// dangling target: render a distinct missing node so the dead edge is visible
			to = "missing_" + sanitize(e.To)
			ew.printf("  %s[%q]:::missing\n", to, e.To+" (missing)")
		}
		ew.printf("  %s -->|%s| %s\n", from, e.Type, to)
	}
	ew.printf("  classDef missing stroke-dasharray: 5 5;\n")
	return ew.err
}

// sanitize makes a string usable as a Mermaid identifier (alphanumerics and underscore only).
func sanitize(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
			b = append(b, c)
			continue
		}
		b = append(b, '_')
	}
	return string(b)
}
