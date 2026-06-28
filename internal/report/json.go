package report

import (
	"encoding/json"
	"io"

	"github.com/kapdroid/docgraph/internal/assert"
	"github.com/kapdroid/docgraph/internal/graph"
)

// jsonGraph is the JSON wire shape: the discovered graph plus the check results, for machine consumers
// (dashboards, other tools). Stable field names; deterministic order (graph.Nodes/Edges are ordered).
type jsonGraph struct {
	Nodes  []jsonNode  `json:"nodes"`
	Edges  []jsonEdge  `json:"edges"`
	Checks []jsonCheck `json:"checks"`
}

type jsonNode struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Abstract bool   `json:"abstract,omitempty"`
}

type jsonEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"`
	Loc  string `json:"loc,omitempty"`
}

type jsonCheck struct {
	Type     string        `json:"type"`
	Passed   bool          `json:"passed"`
	Findings []jsonFinding `json:"findings,omitempty"`
}

type jsonFinding struct {
	Severity string `json:"severity"`
	Node     string `json:"node,omitempty"`
	Message  string `json:"message"`
}

// JSON writes the graph and check results as indented JSON.
func JSON(w io.Writer, g *graph.Graph, results []assert.Result) error {
	out := jsonGraph{}
	for _, n := range g.Nodes() {
		out.Nodes = append(out.Nodes, jsonNode{ID: n.ID, Kind: n.Kind, Abstract: n.Abstract})
	}
	for _, e := range g.Edges() {
		out.Edges = append(out.Edges, jsonEdge{From: e.From, To: e.To, Type: e.Type, Loc: e.Loc})
	}
	for _, r := range results {
		c := jsonCheck{Type: r.Type, Passed: r.Passed()}
		for _, f := range r.Findings {
			c.Findings = append(c.Findings, jsonFinding{Severity: f.Severity, Node: f.Node, Message: f.Message})
		}
		out.Checks = append(out.Checks, c)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
