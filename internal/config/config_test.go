package config

import (
	"path/filepath"
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
		name string
		file string
	}{
		{"yaml syntax error", "bad_syntax.yml"},
		{"unknown edge type", "bad_unknown_edge.yml"},
		{"edge from undefined node-set", "bad_undefined_from.yml"},
		{"node-set without glob", "bad_no_glob.yml"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Load(filepath.Join("testdata", tc.file)); err == nil {
				t.Fatalf("Load(%s) = nil error, want a rejection", tc.file)
			}
		})
	}
}

func TestLoadMissingFile(t *testing.T) {
	if _, err := Load(filepath.Join("testdata", "does-not-exist.yml")); err == nil {
		t.Fatal("Load(missing) = nil error, want an error")
	}
}
