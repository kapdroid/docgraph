// Package extract turns discovered nodes into graph edges. Each reference kind (markdown links, YAML
// pointers, regex citations, …) is an Extractor; a Registry wires the built-ins and runs every
// configured edge rule over the node-set it reads from. Adding a kind is one file implementing
// Extractor plus one line in NewRegistry — no switch sprawl (the pluggable-registry invariant).
package extract

import (
	"fmt"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// Extractor produces the edges out of a single node for one edge rule. Implementations read the
// node's file (resolved against root) and emit edges whose From is the node and whose To is a
// root-relative, slash-cleaned node ID — the same identity scheme discovery assigns, so a target
// that does not resolve to a real node is a dangling edge the no-dangling assertion catches.
type Extractor interface {
	// Type is the docgraph.yml edge `type` this extractor handles.
	Type() string
	// Extract emits the edges for one node under rule. The node carries everything needed: n.Path
	// (root-inclusive, for reading the file) and n.ID (root-relative, the resolution base) — so no
	// separate root is threaded through. A read/parse failure on the node's own file is returned; an
	// unresolved reference is NOT an error — it becomes an edge to a missing node (reported by
	// no-dangling, not here).
	Extract(n *graph.Node, rule config.EdgeRule) ([]graph.Edge, error)
}

// Registry holds the extractors keyed by edge type. Construct it with NewRegistry; it carries no
// global state, so tests build their own and the CLI builds one (no package-level mutable registry).
type Registry struct {
	byType map[string]Extractor
}

// NewRegistry returns a registry with every built-in extractor registered. New extractor kinds are
// added here, one line each.
func NewRegistry() *Registry {
	r := &Registry{byType: map[string]Extractor{}}
	r.Register(markdownLink{})
	r.Register(yamlPointer{})
	r.Register(regexCite{})
	r.Register(jsonPath{})
	return r
}

// Register adds or replaces the extractor for its Type.
func (r *Registry) Register(e Extractor) {
	r.byType[e.Type()] = e
}

// Get returns the extractor for an edge type and whether one is registered.
func (r *Registry) Get(t string) (Extractor, bool) {
	e, ok := r.byType[t]
	return e, ok
}

// Extract runs every edge rule over the nodes of its From set, adding the produced edges to g. It
// fails fast if a rule names an edge type with no registered extractor (a config the loader allowed
// but this build cannot execute) or if reading a node's file fails.
func (r *Registry) Extract(g *graph.Graph, edges []config.EdgeRule) error {
	for i, rule := range edges {
		ex, ok := r.Get(rule.Type)
		if !ok {
			return fmt.Errorf("extract: edge[%d]: no extractor registered for type %q", i, rule.Type)
		}
		for _, n := range g.NodesOfKind(rule.From) {
			es, err := ex.Extract(n, rule)
			if err != nil {
				return fmt.Errorf("extract: %s on %s: %w", rule.Type, n.ID, err)
			}
			for _, e := range es {
				g.AddEdge(e)
			}
		}
	}
	return nil
}
