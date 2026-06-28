// Package test holds cross-package integration tests over the canonical testdata/ fixtures. This
// file guards the fixtures themselves: it pins that the "good" graph passes every check and the
// "broken" graph contains exactly the one dangling link and one orphan the rest of the suite (golden,
// acceptance) relies on. If someone edits a fixture and shifts those counts, this fails loudly.
package test

import (
	"path/filepath"
	"testing"

	"github.com/kapdroid/docgraph/internal/assert"
	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/discover"
	"github.com/kapdroid/docgraph/internal/extract"
)

// checkFixture runs the config→discover→extract→assert pipeline over a fixture directory and returns
// the assertion results keyed by type.
func checkFixture(t *testing.T, dir string) map[string]assert.Result {
	t.Helper()
	cfg, err := config.Load(filepath.Join(dir, "docgraph.yml"))
	if err != nil {
		t.Fatalf("load %s: %v", dir, err)
	}
	cfg.Root = filepath.Join(dir, cfg.Root) // resolve root relative to the config file (as app.Run does)
	g, err := discover.Discover(cfg)
	if err != nil {
		t.Fatalf("discover %s: %v", dir, err)
	}
	if err := extract.NewRegistry().Extract(g, cfg.Edges); err != nil {
		t.Fatalf("extract %s: %v", dir, err)
	}
	results, err := assert.NewRegistry().Check(g, cfg.Assert)
	if err != nil {
		t.Fatalf("assert %s: %v", dir, err)
	}
	byType := map[string]assert.Result{}
	for _, r := range results {
		byType[r.Type] = r
	}
	return byType
}

func TestGoodFixturePasses(t *testing.T) {
	for typ, r := range checkFixture(t, filepath.Join("..", "testdata", "good")) {
		if !r.Passed() {
			t.Errorf("good fixture: %s failed with %d finding(s): %+v", typ, len(r.Findings), r.Findings)
		}
	}
}

func TestBrokenFixtureHasExactlyOneDanglingAndOneOrphan(t *testing.T) {
	byType := checkFixture(t, filepath.Join("..", "testdata", "broken"))
	if got := len(byType["no-dangling"].Findings); got != 1 {
		t.Errorf("broken fixture: no-dangling findings = %d, want exactly 1", got)
	}
	if got := len(byType["no-orphan"].Findings); got != 1 {
		t.Errorf("broken fixture: no-orphan findings = %d, want exactly 1: %+v", got, byType["no-orphan"].Findings)
	}
}
