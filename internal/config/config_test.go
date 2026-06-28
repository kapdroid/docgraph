package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadValid(t *testing.T) {
	cfg, err := Load(filepath.Join("testdata", "valid.yml"))
	if err != nil {
		t.Fatalf("Load(valid.yml) returned error: %v", err)
	}
	if cfg.Root != "." {
		t.Errorf("Root = %q, want %q", cfg.Root, ".")
	}
	if len(cfg.Nodes) != 3 {
		t.Errorf("len(Nodes) = %d, want 3", len(cfg.Nodes))
	}
	if got := cfg.Nodes["adrs"].Frontmatter; len(got) != 3 {
		t.Errorf("adrs.Frontmatter = %v, want 3 fields", got)
	}
	if len(cfg.Edges) != 3 {
		t.Errorf("len(Edges) = %d, want 3", len(cfg.Edges))
	}
	if cfg.Edges[1].Key != "rules" || cfg.Edges[1].To != "docs" {
		t.Errorf("Edges[1] = %+v, want yaml-pointer key=rules to=docs", cfg.Edges[1])
	}
	if !cfg.Edges[2].MustMatchPath {
		t.Errorf("Edges[2].MustMatchPath = false, want true")
	}
	if len(cfg.Moments) != 4 {
		t.Errorf("len(Moments) = %d, want 4", len(cfg.Moments))
	}
	if c := cfg.Consumers["implementer"]; len(c.Reaches) != 1 || len(c.At) != 2 {
		t.Errorf("Consumers[implementer] = %+v, want reaches=1 at=2", c)
	}
	if len(cfg.Assert) != 3 {
		t.Errorf("len(Assert) = %d, want 3", len(cfg.Assert))
	}
}

func TestRootDefaults(t *testing.T) {
	cfg, err := Parse([]byte("nodes:\n  docs: { glob: \"*.md\" }\n"))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if cfg.Root != "." {
		t.Errorf("Root = %q, want default %q", cfg.Root, ".")
	}
}

func TestLoadRejectsMalformed(t *testing.T) {
	cases := []struct {
		name      string
		file      string
		wantInErr string // substring pinning WHICH rule rejected it (so a fixture can't pass for the wrong reason)
	}{
		{"yaml syntax error", "bad_syntax.yml", "parsing yaml"},
		{"empty nodes", "bad_empty_nodes.yml", "no node-sets defined"},
		{"node-set without glob", "bad_no_glob.yml", "no glob"},
		{"unknown edge type", "bad_unknown_edge.yml", "unknown type"},
		{"edge missing from", "bad_missing_from.yml", "missing from"},
		{"edge from undefined node-set", "bad_undefined_from.yml", "not a defined node-set"},
		{"edge to undefined node-set", "bad_to_undefined.yml", "not a defined node-set"},
		{"unknown assert type", "bad_unknown_assert.yml", "unknown type"},
		{"assert references undefined node-set", "bad_assert_undefined_set.yml", "not a defined node-set"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Load(filepath.Join("testdata", tc.file))
			if err == nil {
				t.Fatalf("Load(%s) = nil error, want a rejection", tc.file)
			}
			if !strings.Contains(err.Error(), tc.wantInErr) {
				t.Errorf("Load(%s) error = %q, want it to contain %q", tc.file, err, tc.wantInErr)
			}
		})
	}
}

func TestLoadReachableConsumersKeyword(t *testing.T) {
	// reachable's `from: consumers` is a keyword, not a node-set, and must NOT be rejected by the
	// assert node-set validation.
	if _, err := Load(filepath.Join("testdata", "valid_reachable.yml")); err != nil {
		t.Fatalf("Load(valid_reachable) rejected the consumers keyword: %v", err)
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join("testdata", "does-not-exist.yml")); err == nil {
		t.Fatal("Load(missing) = nil error, want an error")
	}
}
