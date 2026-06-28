// Package acceptance is the FROZEN, QA-authored behavioral spec for the docgraph CLI (ADR-0011). It
// is verified independently of the implementer: once frozen (kapdroid freeze), the acceptance-guard
// hook and close Guard 7 refuse edits to this directory, and the close gate runs it full. Each test
// is one row of test/acceptance/README.md. These drive the real binary over the canonical fixtures.
package acceptance

import (
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// build compiles the docgraph binary into a temp dir (module-qualified path, CWD-independent).
func build(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "docgraph")
	if out, err := exec.Command("go", "build", "-o", bin, "github.com/kapdroid/docgraph/cmd/docgraph").CombinedOutput(); err != nil {
		t.Fatalf("build docgraph: %v\n%s", err, out)
	}
	return bin
}

// run executes docgraph with the given config and returns combined stdout and the exit code.
func run(t *testing.T, bin, config string) (string, int) {
	t.Helper()
	out, err := exec.Command(bin, "-config", config).Output()
	code := 0
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		} else {
			t.Fatalf("run docgraph: %v", err)
		}
	}
	return string(out), code
}

func fixture(name string) string {
	return filepath.Join("..", "..", "testdata", name, "docgraph.yml")
}

func TestAC1_HealthyGraphPassesExitZero(t *testing.T) {
	out, code := run(t, build(t), fixture("good"))
	if code != 0 {
		t.Fatalf("AC1: exit = %d, want 0\n%s", code, out)
	}
	if !strings.Contains(out, "ok") {
		t.Errorf("AC1: report missing 'ok':\n%s", out)
	}
}

func TestAC2_BrokenGraphExitsOne(t *testing.T) {
	out, code := run(t, build(t), fixture("broken"))
	if code != 1 {
		t.Fatalf("AC2: exit = %d, want 1 (failed checks, not a crash)\n%s", code, out)
	}
}

func TestAC3_DanglingLinkReported(t *testing.T) {
	out, _ := run(t, build(t), fixture("broken"))
	if !strings.Contains(out, "ghost.md") || !strings.Contains(out, "does not exist") {
		t.Errorf("AC3: dangling link not reported with its target:\n%s", out)
	}
}

func TestAC4_OrphanReported(t *testing.T) {
	out, _ := run(t, build(t), fixture("broken"))
	if !strings.Contains(out, "orphan.md") || !strings.Contains(out, "orphan") {
		t.Errorf("AC4: orphan not reported with its file:\n%s", out)
	}
}

func TestAC5_MissingConfigExitsTwo(t *testing.T) {
	out, code := run(t, build(t), fixture("does-not-exist"))
	if code != 2 {
		t.Fatalf("AC5: exit = %d, want 2 (config error)\n%s", code, out)
	}
}

func TestAC6_EachCheckHasAVerdict(t *testing.T) {
	out, _ := run(t, build(t), fixture("broken"))
	for _, check := range []string{"no-dangling", "no-orphan"} {
		if !strings.Contains(out, "PASS  "+check) && !strings.Contains(out, "FAIL  "+check) {
			t.Errorf("AC6: check %q has no PASS/FAIL verdict:\n%s", check, out)
		}
	}
}
