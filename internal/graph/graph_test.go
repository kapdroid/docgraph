package graph

import "testing"

// fixtureGraph builds a small graph:
//
//	a → b, a → c, b → c, c → missing (dangling), and an isolated orphan node "orphan".
func fixtureGraph() *Graph {
	g := New()
	for _, id := range []string{"a", "b", "c", "orphan"} {
		g.AddNode(Node{ID: id, Kind: "docs", Path: id + ".md"})
	}
	g.AddEdge(Edge{From: "a", To: "b", Type: "markdown-link", Loc: "a.md:1"})
	g.AddEdge(Edge{From: "a", To: "c", Type: "markdown-link", Loc: "a.md:2"})
	g.AddEdge(Edge{From: "b", To: "c", Type: "markdown-link", Loc: "b.md:1"})
	g.AddEdge(Edge{From: "c", To: "missing", Type: "markdown-link", Loc: "c.md:1"})
	return g
}

func TestAdjacency(t *testing.T) {
	g := fixtureGraph()
	cases := []struct {
		name    string
		id      string
		wantOut int
		wantIn  int
	}{
		{"a has two out, zero in", "a", 2, 0},
		{"b one out one in", "b", 1, 1},
		{"c one out two in", "c", 1, 2},
		{"orphan zero both", "orphan", 0, 0},
		{"missing target has inbound but no node", "missing", 0, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := len(g.Outbound(tc.id)); got != tc.wantOut {
				t.Errorf("Outbound(%q) = %d, want %d", tc.id, got, tc.wantOut)
			}
			if got := len(g.Inbound(tc.id)); got != tc.wantIn {
				t.Errorf("Inbound(%q) = %d, want %d", tc.id, got, tc.wantIn)
			}
		})
	}
}

func TestHasAndNode(t *testing.T) {
	g := fixtureGraph()
	if !g.Has("a") {
		t.Error("Has(a) = false, want true")
	}
	if g.Has("missing") {
		t.Error("Has(missing) = true, want false (it is only an edge target)")
	}
	n, ok := g.Node("b")
	if !ok || n.Path != "b.md" {
		t.Errorf("Node(b) = %+v, %v; want a node with Path b.md", n, ok)
	}
}

func TestAddNodeOverwrites(t *testing.T) {
	g := New()
	g.AddNode(Node{ID: "x", Kind: "docs"})
	g.AddNode(Node{ID: "x", Kind: "adrs", Frontmatter: map[string]string{"scope": "engine"}})
	n, _ := g.Node("x")
	if n.Kind != "adrs" || n.Frontmatter["scope"] != "engine" {
		t.Errorf("re-AddNode did not overwrite: got %+v", n)
	}
}

func TestNodesSortedAndOfKind(t *testing.T) {
	g := New()
	g.AddNode(Node{ID: "z", Kind: "docs"})
	g.AddNode(Node{ID: "a", Kind: "adrs"})
	g.AddNode(Node{ID: "m", Kind: "docs"})
	nodes := g.Nodes()
	if len(nodes) != 3 || nodes[0].ID != "a" || nodes[2].ID != "z" {
		t.Errorf("Nodes() not sorted by ID: %v", ids(nodes))
	}
	docs := g.NodesOfKind("docs")
	if len(docs) != 2 || docs[0].ID != "m" || docs[1].ID != "z" {
		t.Errorf("NodesOfKind(docs) = %v, want [m z]", ids(docs))
	}
}

func ids(ns []*Node) []string {
	out := make([]string, len(ns))
	for i, n := range ns {
		out[i] = n.ID
	}
	return out
}
