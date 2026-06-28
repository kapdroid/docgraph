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

func TestAddNodeMergesMultiKind(t *testing.T) {
	g := New()
	g.AddNode(Node{ID: "x", Kind: "docs"})
	g.AddNode(Node{ID: "x", Kind: "adrs", Frontmatter: map[string]string{"scope": "engine"}})
	n, _ := g.Node("x")
	// merged: belongs to BOTH sets, frontmatter kept, primary Kind is lexicographically-first ("adrs")
	if !g.IsKind("x", "docs") || !g.IsKind("x", "adrs") {
		t.Errorf("x should be a member of both docs and adrs")
	}
	if n.Kind != "adrs" {
		t.Errorf("primary Kind = %q, want lexicographically-first 'adrs'", n.Kind)
	}
	if n.Frontmatter["scope"] != "engine" {
		t.Errorf("merged frontmatter lost: got %+v", n.Frontmatter)
	}
	if len(g.NodesOfKind("docs")) != 1 || len(g.NodesOfKind("adrs")) != 1 {
		t.Errorf("NodesOfKind should return x for both sets")
	}
}

func TestAddNodePrimaryKindDeterministic(t *testing.T) {
	// order of AddNode must not change the primary Kind (lexicographically-first wins either way)
	g1 := New()
	g1.AddNode(Node{ID: "x", Kind: "zeta"})
	g1.AddNode(Node{ID: "x", Kind: "alpha"})
	g2 := New()
	g2.AddNode(Node{ID: "x", Kind: "alpha"})
	g2.AddNode(Node{ID: "x", Kind: "zeta"})
	n1, _ := g1.Node("x")
	n2, _ := g2.Node("x")
	if n1.Kind != "alpha" || n2.Kind != "alpha" {
		t.Errorf("primary Kind not deterministic: g1=%q g2=%q, want both 'alpha'", n1.Kind, n2.Kind)
	}
}

func TestAddNodeMergeFillsPathAndSkipsEmptyKind(t *testing.T) {
	g := New()
	g.AddNode(Node{ID: "x", Kind: "docs"})                // no Path yet
	g.AddNode(Node{ID: "x", Kind: "", Path: "real/x.md"}) // empty Kind must not register a "" set
	n, _ := g.Node("x")
	if n.Path != "real/x.md" {
		t.Errorf("merge did not fill empty Path: got %q", n.Path)
	}
	if g.IsKind("x", "") {
		t.Error("empty Kind should not create a membership entry")
	}
	if !g.IsKind("x", "docs") {
		t.Error("original docs membership lost after empty-Kind re-add")
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
