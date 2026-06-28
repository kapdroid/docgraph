// Package report renders assertion results for humans and machines. M1 ships the text report; the
// JSON/JUnit/mermaid formats are added in M3 (kap-ymj.17). Each format is a function taking the data
// it needs and an io.Writer, so the CLI stays a thin dispatcher.
package report

import (
	"fmt"
	"io"

	"github.com/kapdroid/docgraph/internal/assert"
)

// AllPassed reports whether every assertion result passed (no findings).
func AllPassed(results []assert.Result) bool {
	for _, r := range results {
		if !r.Passed() {
			return false
		}
	}
	return true
}

// errWriter records the first write error so a sequence of formatted writes can be checked once at
// the end instead of after every call (keeping the report body readable).
type errWriter struct {
	w   io.Writer
	err error
}

func (e *errWriter) printf(format string, a ...any) {
	if e.err != nil {
		return
	}
	_, e.err = fmt.Fprintf(e.w, format, a...)
}

// Text writes a human-readable pass/fail report: one line per check, then the findings of any failed
// check, and a final summary line. It returns the number of findings written and the first write
// error encountered (nil on success).
func Text(w io.Writer, results []assert.Result) (int, error) {
	ew := &errWriter{w: w}
	total := 0
	for _, r := range results {
		if r.Passed() {
			ew.printf("PASS  %s\n", r.Type)
			continue
		}
		ew.printf("FAIL  %s (%d)\n", r.Type, len(r.Findings))
		for _, f := range r.Findings {
			total++
			ew.printf("        - %s\n", f.Message)
		}
	}
	checks := len(results)
	if AllPassed(results) {
		ew.printf("\nok — %d check(s) passed\n", checks)
	} else {
		ew.printf("\nFAILED — %d finding(s) across %d check(s)\n", total, checks)
	}
	return total, ew.err
}
