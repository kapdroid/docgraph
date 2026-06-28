package extract

import (
	"sort"
	"testing"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// docsGraph builds a graph of the three testdata docs (root = testdata), as discovery would.
func docsGraph() *graph.Graph {
	g := graph.New()
	for _, id := range []string{"docs/a.md", "docs/b.md", "docs/sub/c.md"} {
		g.AddNode(graph.Node{ID: id, Kind: "docs", Path: "testdata/" + id})
	}
	return g
}

func TestMarkdownLinkExtract(t *testing.T) {
	g := docsGraph()
	r := NewRegistry()
	if err := r.Extract(g, []config.EdgeRule{{Type: "markdown-link", From: "docs"}}); err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}

	var targets []string
	for _, e := range g.Outbound("docs/a.md") {
		if e.Type != "markdown-link" {
			t.Errorf("edge type = %q, want markdown-link", e.Type)
		}
		targets = append(targets, e.To)
	}
	sort.Strings(targets)

	// b.md appears twice ([b] and [titled]); external/mailto/anchor are skipped; ghost.md is a
	// dangling edge but still emitted.
	want := []string{"docs/b.md", "docs/b.md", "docs/ghost.md", "docs/sub/c.md"}
	if len(targets) != len(want) {
		t.Fatalf("targets = %v, want %v", targets, want)
	}
	for i := range want {
		if targets[i] != want[i] {
			t.Fatalf("targets = %v, want %v", targets, want)
		}
	}
}

func TestLinkTarget(t *testing.T) {
	cases := []struct {
		name   string
		raw    string
		want   string
		wantOK bool
	}{
		{"plain", "b.md", "b.md", true},
		{"anchor stripped", "sub/c.md#head", "sub/c.md", true},
		{"title stripped", `b.md "the title"`, "b.md", true},
		{"angle brackets", "<b.md>", "b.md", true},
		{"external https", "https://example.com", "", false},
		{"mailto", "mailto:x@y.com", "", false},
		{"tel", "tel:+1555", "", false},
		{"pure anchor", "#section", "", false},
		{"empty", "", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := linkTarget(tc.raw)
			if ok != tc.wantOK || got != tc.want {
				t.Errorf("linkTarget(%q) = (%q,%v), want (%q,%v)", tc.raw, got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestResolveRef(t *testing.T) {
	cases := []struct {
		name   string
		from   string
		target string
		want   string
	}{
		{"sibling", "docs/a.md", "b.md", "docs/b.md"},
		{"subdir", "docs/a.md", "sub/c.md", "docs/sub/c.md"},
		{"parent", "docs/sub/c.md", "../a.md", "docs/a.md"},
		{"root-absolute", "docs/a.md", "/README.md", "README.md"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resolveRef(tc.from, tc.target); got != tc.want {
				t.Errorf("resolveRef(%q,%q) = %q, want %q", tc.from, tc.target, got, tc.want)
			}
		})
	}
}

func TestExtractUnknownType(t *testing.T) {
	g := docsGraph()
	r := NewRegistry()
	err := r.Extract(g, []config.EdgeRule{{Type: "telepathy", From: "docs"}})
	if err == nil {
		t.Fatal("Extract with unknown type = nil error, want failure")
	}
}

func TestExtractOwnFileReadFails(t *testing.T) {
	// the seam's error contract: a read failure on the node's OWN file is an error (vs an unresolved
	// reference, which is a dangling edge, not an error). M2 extractors must honor the same contract.
	g := graph.New()
	g.AddNode(graph.Node{ID: "docs/gone.md", Kind: "docs", Path: "testdata/docs/gone.md"})
	r := NewRegistry()
	if err := r.Extract(g, []config.EdgeRule{{Type: "markdown-link", From: "docs"}}); err == nil {
		t.Fatal("Extract over a node with a missing file = nil error, want failure")
	}
}

func TestMarkdownLinkSkipsFencedCode(t *testing.T) {
	g := graph.New()
	g.AddNode(graph.Node{ID: "docs/fenced.md", Kind: "docs", Path: "testdata/docs/fenced.md"})
	r := NewRegistry()
	if err := r.Extract(g, []config.EdgeRule{{Type: "markdown-link", From: "docs"}}); err != nil {
		t.Fatalf("Extract: %v", err)
	}
	for _, e := range g.Outbound("docs/fenced.md") {
		if e.To == "docs/should-not-resolve.md" {
			t.Errorf("link inside a fenced code block was extracted: %+v", e)
		}
	}
	if n := len(g.Outbound("docs/fenced.md")); n != 2 { // [b] and [c], not the fenced one
		t.Errorf("got %d edges, want 2 (fenced example excluded)", n)
	}
}
