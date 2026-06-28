package discover

import (
	"testing"

	"github.com/kapdroid/docgraph/internal/config"
)

func projConfig() *config.Config {
	return &config.Config{
		Root: "testdata/proj",
		Nodes: map[string]config.NodeSet{
			"docs":   {Glob: "docs/*.md"},
			"adrs":   {Glob: "**/decisions/adr-*.md", Frontmatter: []string{"id", "scope", "status", "covers"}},
			"stacks": {Glob: "stacks/*/stack.yml"},
		},
	}
}

func TestDiscoverNodes(t *testing.T) {
	g, err := Discover(projConfig())
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	cases := []struct {
		name      string
		id        string
		wantKind  string
		wantExist bool
	}{
		{"doc a", "docs/a.md", "docs", true},
		{"doc b", "docs/b.md", "docs", true},
		{"doc nofm", "docs/nofm.md", "docs", true},
		{"adr", "decisions/adr-0001-foo.md", "adrs", true},
		{"stack yml", "stacks/go/stack.yml", "stacks", true},
		{"not discovered", "does/not/exist.md", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n, ok := g.Node(tc.id)
			if ok != tc.wantExist {
				t.Fatalf("Node(%q) exists=%v, want %v", tc.id, ok, tc.wantExist)
			}
			if ok && n.Kind != tc.wantKind {
				t.Errorf("Node(%q).Kind = %q, want %q", tc.id, n.Kind, tc.wantKind)
			}
		})
	}
}

func TestDiscoverFrontmatter(t *testing.T) {
	g, err := Discover(projConfig())
	if err != nil {
		t.Fatalf("Discover returned error: %v", err)
	}
	adr, ok := g.Node("decisions/adr-0001-foo.md")
	if !ok {
		t.Fatal("adr node not discovered")
	}
	want := map[string]string{
		"id":     "ADR-0001",
		"scope":  "engine",
		"status": "decided",
		"covers": "graphs, edges", // a list flattened to text
	}
	for k, v := range want {
		if got := adr.Frontmatter[k]; got != v {
			t.Errorf("Frontmatter[%q] = %q, want %q", k, got, v)
		}
	}
}

func TestFrontmatterAbsentIsEmpty(t *testing.T) {
	// a doc set without a Frontmatter request leaves Frontmatter nil; a member with no fenced block
	// (nofm.md) would yield an empty map if requested. Here docs requests none → nil.
	g, _ := Discover(projConfig())
	n, _ := g.Node("docs/nofm.md")
	if n.Frontmatter != nil {
		t.Errorf("docs node Frontmatter = %v, want nil (no frontmatter requested)", n.Frontmatter)
	}
}

func TestEmptyGlobIsNotError(t *testing.T) {
	cfg := &config.Config{
		Root:  "testdata/proj",
		Nodes: map[string]config.NodeSet{"none": {Glob: "no/such/*.txt"}},
	}
	g, err := Discover(cfg)
	if err != nil {
		t.Fatalf("Discover with empty-matching glob errored: %v", err)
	}
	if len(g.Nodes()) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(g.Nodes()))
	}
}

func TestDiscoverTolerantFrontmatter(t *testing.T) {
	// a frontmatter whose NON-requested fields are YAML-invalid (regression_signal: >.* , coverage[..])
	// must not break extraction of the requested fields (the kapdroid-dogfood case, kap-ymj.21).
	cfg := &config.Config{
		Root:  "testdata/quirk",
		Nodes: map[string]config.NodeSet{"adrs": {Glob: "adr-*.md", Frontmatter: []string{"id", "scope"}}},
	}
	g, err := Discover(cfg)
	if err != nil {
		t.Fatalf("Discover errored on quirky frontmatter: %v", err)
	}
	n, ok := g.Node("adr-bad.md")
	if !ok {
		t.Fatal("adr-bad.md not discovered")
	}
	if n.Frontmatter["id"] != "ADR-0099" || n.Frontmatter["scope"] != "engine" {
		t.Errorf("requested fields = %v, want id=ADR-0099 scope=engine", n.Frontmatter)
	}
}

func TestDiscoverExclude(t *testing.T) {
	cfg := &config.Config{
		Root: "testdata/proj",
		Nodes: map[string]config.NodeSet{
			"all": {Glob: "**/*.md", Exclude: []string{"decisions/**"}},
		},
	}
	g, err := Discover(cfg)
	if err != nil {
		t.Fatalf("Discover: %v", err)
	}
	if _, ok := g.Node("docs/a.md"); !ok {
		t.Error("docs/a.md should be discovered")
	}
	if _, ok := g.Node("decisions/adr-0001-foo.md"); ok {
		t.Error("decisions/adr-0001-foo.md should be EXCLUDed")
	}
}

func TestDiscoverBadExcludePattern(t *testing.T) {
	cfg := &config.Config{
		Root:  "testdata/proj",
		Nodes: map[string]config.NodeSet{"all": {Glob: "**/*.md", Exclude: []string{"[bad"}}},
	}
	if _, err := Discover(cfg); err == nil {
		t.Fatal("Discover with a malformed exclude pattern = nil error, want failure")
	}
}
