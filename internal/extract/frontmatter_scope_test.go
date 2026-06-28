package extract

import (
	"testing"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

func TestDeriveExpected(t *testing.T) {
	rules := []config.DeriveRule{
		{Under: "docs/decisions", Expect: "engine"},
		{Under: "stacks/*/decisions", Expect: "stack:{1}"},
		{Under: "tenants/*/decisions", Expect: "tenant:{1}"},
	}
	cases := []struct {
		name string
		id   string
		want string
		ok   bool
	}{
		{"engine adr", "docs/decisions/adr-0001-x.md", "engine", true},
		{"stack adr", "stacks/go/decisions/adr-0015-x.md", "stack:go", true},
		{"tenant adr", "tenants/docgraph/decisions/adr-9.md", "tenant:docgraph", true},
		{"no rule matches", "somewhere/else/adr.md", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := deriveExpected(tc.id, rules)
			if ok != tc.ok || got != tc.want {
				t.Errorf("deriveExpected(%q) = (%q,%v), want (%q,%v)", tc.id, got, ok, tc.want, tc.ok)
			}
		})
	}
}

func TestFrontmatterScopeAnnotates(t *testing.T) {
	g := graph.New()
	g.AddNode(graph.Node{ID: "stacks/go/decisions/adr-0015-x.md", Kind: "adrs", Frontmatter: map[string]string{"scope": "stack:go"}})
	g.AddNode(graph.Node{ID: "docs/decisions/adr-0002-y.md", Kind: "adrs", Frontmatter: map[string]string{"scope": "WRONG"}})
	r := NewRegistry()
	rule := config.EdgeRule{
		Type: "frontmatter-scope", From: "adrs", Field: "scope", MustMatchPath: true,
		Derive: []config.DeriveRule{
			{Under: "docs/decisions", Expect: "engine"},
			{Under: "stacks/*/decisions", Expect: "stack:{1}"},
		},
	}
	if err := r.Extract(g, []config.EdgeRule{rule}); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	// frontmatter-scope emits no edges — it annotates nodes
	if len(g.Edges()) != 0 {
		t.Errorf("frontmatter-scope emitted %d edges, want 0 (it annotates)", len(g.Edges()))
	}
	n1, _ := g.Node("stacks/go/decisions/adr-0015-x.md")
	if got := n1.Frontmatter[graph.DerivedPrefix+"scope"]; got != "stack:go" {
		t.Errorf("derived scope = %q, want stack:go", got)
	}
	n2, _ := g.Node("docs/decisions/adr-0002-y.md")
	if got := n2.Frontmatter[graph.DerivedPrefix+"scope"]; got != "engine" {
		t.Errorf("derived scope = %q, want engine (actual is WRONG — consistent assert will catch it)", got)
	}
}

func TestFrontmatterScopeRequiresField(t *testing.T) {
	g := graph.New()
	g.AddNode(graph.Node{ID: "a.md", Kind: "adrs"})
	r := NewRegistry()
	if err := r.Extract(g, []config.EdgeRule{{Type: "frontmatter-scope", From: "adrs", MustMatchPath: true}}); err == nil {
		t.Fatal("frontmatter-scope with no field = nil error, want failure")
	}
}
