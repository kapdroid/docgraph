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
