package app

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGoodPasses(t *testing.T) {
	var out strings.Builder
	passed, err := Run(filepath.Join("testdata", "good", "docgraph.yml"), &out)
	if err != nil {
		t.Fatalf("Run(good) error: %v", err)
	}
	if !passed {
		t.Errorf("Run(good) passed = false, want true\n%s", out.String())
	}
	if !strings.Contains(out.String(), "PASS  no-dangling") {
		t.Errorf("output missing PASS line:\n%s", out.String())
	}
}

func TestRunBrokenFailsWithDanglingAndOrphan(t *testing.T) {
	var out strings.Builder
	passed, err := Run(filepath.Join("testdata", "broken", "docgraph.yml"), &out)
	if err != nil {
		t.Fatalf("Run(broken) error: %v", err)
	}
	if passed {
		t.Fatalf("Run(broken) passed = true, want false\n%s", out.String())
	}
	s := out.String()
	// the M1 Definition of Done: flags a broken link AND an orphan
	if !strings.Contains(s, "ghost.md") {
		t.Errorf("output missing the dangling link (ghost.md):\n%s", s)
	}
	if !strings.Contains(s, "orphan.md") {
		t.Errorf("output missing the orphan (orphan.md):\n%s", s)
	}
	if !strings.Contains(s, "FAIL  no-dangling") || !strings.Contains(s, "FAIL  no-orphan") {
		t.Errorf("output missing FAIL lines:\n%s", s)
	}
}

func TestRunConfigErrorIsError(t *testing.T) {
	var out strings.Builder
	if _, err := Run(filepath.Join("testdata", "nope.yml"), &out); err == nil {
		t.Fatal("Run(missing config) = nil error, want an error (distinct from a failed check)")
	}
}

func TestExplainReachable(t *testing.T) {
	var out strings.Builder
	if err := Explain(filepath.Join("testdata", "reach", "docgraph.yml"), "docs/reachable.md", &out); err != nil {
		t.Fatalf("Explain error: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "is reachable") || !strings.Contains(s, "consumer:impl") || !strings.Contains(s, "docs/reachable.md") {
		t.Errorf("explain output missing reachable path:\n%s", s)
	}
}

func TestExplainUnreachable(t *testing.T) {
	var out strings.Builder
	if err := Explain(filepath.Join("testdata", "reach", "docgraph.yml"), "docs/lonely.md", &out); err != nil {
		t.Fatalf("Explain error: %v", err)
	}
	if s := out.String(); !strings.Contains(s, "UNREACHABLE") {
		t.Errorf("explain output should mark lonely.md unreachable:\n%s", s)
	}
}

func TestExplainNoSuchNode(t *testing.T) {
	var out strings.Builder
	if err := Explain(filepath.Join("testdata", "reach", "docgraph.yml"), "docs/ghost.md", &out); err != nil {
		t.Fatalf("Explain error: %v", err)
	}
	if s := out.String(); !strings.Contains(s, "no such node") {
		t.Errorf("explain output should report no such node:\n%s", s)
	}
}

func TestRenderFormats(t *testing.T) {
	cfg := filepath.Join("testdata", "broken", "docgraph.yml")
	for _, format := range []string{"text", "junit", "json", "mermaid"} {
		t.Run(format, func(t *testing.T) {
			var out strings.Builder
			if _, err := Render(cfg, format, &out); err != nil {
				t.Fatalf("Render(%s) error: %v", format, err)
			}
			if out.Len() == 0 {
				t.Errorf("Render(%s) produced no output", format)
			}
		})
	}
	if _, err := Render(cfg, "bogus", &strings.Builder{}); err == nil {
		t.Error("Render(bogus) = nil error, want unknown-format failure")
	}
}
