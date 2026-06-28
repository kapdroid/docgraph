// Package reach models the consumers×moments layer (M3) and computes reachability over the graph.
//
// A consumer (e.g. "implementer") reads certain node-sets ("reaches") at certain lifecycle moments
// ("at"). Materialize injects each consumer and moment as an abstract node and links each consumer to
// the nodes of the sets it reaches — turning "is every rule delivered to its reader?" into a pure
// graph-reachability question. ReachableFrom then answers it with a BFS. This is the M3 differentiator:
// not "is this link alive" but "does every rule/doc reach its intended consumer through any path".
package reach

import (
	"sort"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// Kinds and ID prefixes for the abstract nodes this layer injects.
const (
	// ConsumerKind is the Kind of an injected consumer node.
	ConsumerKind = "consumer"
	// MomentKind is the Kind of an injected moment node.
	MomentKind = "moment"
	// ConsumerPrefix prefixes a consumer node's ID (ConsumerPrefix+name).
	ConsumerPrefix = "consumer:"
	// MomentPrefix prefixes a moment node's ID (MomentPrefix+name).
	MomentPrefix = "moment:"

	// reachesEdge links a consumer to a node-set it directly reaches.
	reachesEdge = "reaches"
	// atEdge links a consumer to a moment it reads the graph at.
	atEdge = "at"
)

// Materialize injects the config's consumers and moments into g as abstract nodes and adds, for each
// consumer, a "reaches" edge to every node of each set it reaches and an "at" edge to each of its
// moments. It is a no-op when no consumers are configured. Call it after extraction and before the
// reachable assertion so the assert sees a graph where consumers are first-class entry nodes.
func Materialize(cfg *config.Config, g *graph.Graph) {
	if len(cfg.Consumers) == 0 {
		return // moments are only meaningful as consumer `at` targets — nothing to inject
	}
	for _, m := range cfg.Moments {
		g.AddNode(graph.Node{ID: MomentPrefix + m, Kind: MomentKind, Abstract: true})
	}
	names := make([]string, 0, len(cfg.Consumers))
	for name := range cfg.Consumers {
		names = append(names, name)
	}
	sort.Strings(names) // deterministic edge insertion order (cfg.Consumers is a map)
	for _, name := range names {
		c := cfg.Consumers[name]
		cid := ConsumerPrefix + name
		g.AddNode(graph.Node{ID: cid, Kind: ConsumerKind, Abstract: true})
		for _, set := range c.Reaches {
			for _, n := range g.NodesOfKind(set) {
				g.AddEdge(graph.Edge{From: cid, To: n.ID, Type: reachesEdge, Loc: cid})
			}
		}
		for _, at := range c.At {
			g.AddEdge(graph.Edge{From: cid, To: MomentPrefix + at, Type: atEdge, Loc: cid})
		}
	}
}

// ConsumerIDs returns the IDs of the injected consumer nodes, sorted (Nodes() is sorted).
func ConsumerIDs(g *graph.Graph) []string {
	var ids []string
	for _, n := range g.NodesOfKind(ConsumerKind) {
		ids = append(ids, n.ID)
	}
	return ids
}

// ReachableFrom returns the set of node IDs reachable from any source by following outbound edges
// (BFS). Sources themselves are included. Edges to non-existent nodes (dangling) are not traversable
// (there is no node to expand) but are recorded as reached IDs only if they are real nodes.
func ReachableFrom(g *graph.Graph, sources []string) map[string]bool {
	seen := map[string]bool{}
	queue := make([]string, 0, len(sources))
	for _, s := range sources {
		if !seen[s] {
			seen[s] = true
			queue = append(queue, s)
		}
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		for _, e := range g.Outbound(id) {
			if seen[e.To] || !g.Has(e.To) {
				continue
			}
			seen[e.To] = true
			queue = append(queue, e.To)
		}
	}
	return seen
}
