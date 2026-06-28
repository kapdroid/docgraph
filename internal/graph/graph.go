// Package graph is the in-memory model docgraph builds and asserts over: typed nodes (artifacts and
// abstract reader/moment nodes) connected by directed reference edges, with inbound/outbound
// adjacency. It is pure data — no I/O, no globbing — so it is cheap to construct in tests and reused
// unchanged by discovery (fills nodes), the extractors (add edges), and every assertion (M1–M3).
package graph

import "sort"

// DerivedPrefix namespaces a node Frontmatter key holding a value DERIVED by an extractor (not parsed
// from the file) — e.g. frontmatter-scope stores the location-derived expected value under
// DerivedPrefix+field, which the consistent assertion compares against the parsed field. Keeping it a
// reserved prefix lets producer (extract) and consumer (assert) agree without a side channel.
const DerivedPrefix = "@derived:"

// Node is one vertex: a file artifact (Path set, Abstract false) or an abstract node such as a
// consumer/role or a lifecycle moment (Abstract true, Path empty).
type Node struct {
	// ID is the unique identity used by edges; for file nodes it is the cleaned relative path.
	ID string
	// Kind is the node-set the node belongs to (e.g. "docs", "adrs") or an abstract kind ("consumer").
	Kind string
	// Path is the file path for a file node; empty for an abstract node.
	Path string
	// Abstract is true for non-file nodes (consumers, moments) that have no on-disk artifact.
	Abstract bool
	// Frontmatter holds parsed scalar frontmatter fields, populated by discovery (nil if none).
	Frontmatter map[string]string
}

// Edge is a directed reference "From references To", produced by an extractor. To may name a node
// that does not exist in the graph — that is a dangling edge, which the no-dangling assertion reports.
type Edge struct {
	// From is the ID of the referencing node (always a real node in the graph).
	From string
	// To is the ID the reference points at; it may be absent from the graph (dangling).
	To string
	// Type is the extractor that produced the edge (e.g. "markdown-link").
	Type string
	// Loc is the source location of the reference, "path:line", for reporting.
	Loc string
}

// Graph holds nodes keyed by ID and the edges between them, with both adjacency directions
// precomputed for O(1) neighbor lookup.
type Graph struct {
	nodes map[string]*Node
	edges []Edge
	out   map[string][]Edge
	in    map[string][]Edge
}

// New returns an empty graph ready for AddNode/AddEdge.
func New() *Graph {
	return &Graph{
		nodes: map[string]*Node{},
		out:   map[string][]Edge{},
		in:    map[string][]Edge{},
	}
}

// AddNode inserts or replaces the node with n.ID. Re-adding an existing ID overwrites it (discovery
// is the single writer of nodes, so last-write-wins is intentional and keeps callers simple).
func (g *Graph) AddNode(n Node) {
	node := n
	g.nodes[n.ID] = &node
}

// AddEdge records a directed edge and updates both adjacency indexes. The To node need not exist yet.
func (g *Graph) AddEdge(e Edge) {
	g.edges = append(g.edges, e)
	g.out[e.From] = append(g.out[e.From], e)
	g.in[e.To] = append(g.in[e.To], e)
}

// Node returns the node with the given ID and whether it exists.
func (g *Graph) Node(id string) (*Node, bool) {
	n, ok := g.nodes[id]
	return n, ok
}

// Has reports whether a node with the given ID exists.
func (g *Graph) Has(id string) bool {
	_, ok := g.nodes[id]
	return ok
}

// Nodes returns all nodes sorted by ID, for deterministic iteration and stable report output.
func (g *Graph) Nodes() []*Node {
	out := make([]*Node, 0, len(g.nodes))
	for _, n := range g.nodes {
		out = append(out, n)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// Edges returns all edges in insertion order. The slice is read-only — do not append to it (it
// shares backing storage with the graph; docgraph builds-then-reads single-threaded).
func (g *Graph) Edges() []Edge {
	return g.edges
}

// Outbound returns the edges leaving the node with the given ID (nil if none). Read-only (see Edges).
func (g *Graph) Outbound(id string) []Edge {
	return g.out[id]
}

// Inbound returns the edges arriving at the given ID, including edges from nodes that point at an ID
// with no node (so a dangling target still has discoverable referrers). Read-only (see Edges).
func (g *Graph) Inbound(id string) []Edge {
	return g.in[id]
}

// NodesOfKind returns the nodes whose Kind matches, sorted by ID.
func (g *Graph) NodesOfKind(kind string) []*Node {
	var out []*Node
	for _, n := range g.Nodes() {
		if n.Kind == kind {
			out = append(out, n)
		}
	}
	return out
}
