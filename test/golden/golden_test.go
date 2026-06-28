// Package golden is the integration layer of the test pyramid: it builds the real docgraph binary and
// runs it over the canonical testdata/ fixtures, comparing stdout and exit code against committed
// golden files. It is the Go equivalent of an end-to-end CLI test. Regenerate goldens with:
//
//	go test ./test/golden/ -update
package golden

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var update = flag.Bool("update", false, "update golden files instead of comparing")

// buildBinary compiles cmd/docgraph into a temp dir and returns its path. Building by the module-
// qualified package path works regardless of the test's working directory.
func buildBinary(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "docgraph")
	out, err := exec.Command("go", "build", "-o", bin, "github.com/kapdroid/docgraph/cmd/docgraph").CombinedOutput()
	if err != nil {
		t.Fatalf("build docgraph: %v\n%s", err, out)
	}
	return bin
}

// exitCode returns the process exit code from a *exec.Cmd run error (0 when err is nil).
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode()
	}
	return -1
}

func TestCLIGolden(t *testing.T) {
	bin := buildBinary(t)
	cases := []struct {
		name     string
		config   string
		golden   string
		wantExit int
	}{
		{"good", "../../testdata/good/docgraph.yml", "testdata/good.golden", 0},
		{"broken", "../../testdata/broken/docgraph.yml", "testdata/broken.golden", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := exec.Command(bin, "-config", tc.config).Output()
			if code := exitCode(err); code != tc.wantExit {
				t.Fatalf("exit = %d, want %d (err=%v)", code, tc.wantExit, err)
			}
			if *update {
				if werr := os.WriteFile(tc.golden, out, 0o600); werr != nil {
					t.Fatalf("update golden: %v", werr)
				}
				return
			}
			want, rerr := os.ReadFile(tc.golden)
			if rerr != nil {
				t.Fatalf("read golden (run with -update first): %v", rerr)
			}
			if string(out) != string(want) {
				t.Errorf("stdout mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", tc.name, out, want)
			}
		})
	}
}
