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

// Text writes a human-readable pass/fail report: one line per check, then the findings of any failed
// check, and a final summary line. It returns the number of findings written.
func Text(w io.Writer, results []assert.Result) int {
	total := 0
	for _, r := range results {
		if r.Passed() {
			fmt.Fprintf(w, "PASS  %s\n", r.Type)
			continue
		}
		fmt.Fprintf(w, "FAIL  %s (%d)\n", r.Type, len(r.Findings))
		for _, f := range r.Findings {
			total++
			fmt.Fprintf(w, "        - %s\n", f.Message)
		}
	}
	checks := len(results)
	if AllPassed(results) {
		fmt.Fprintf(w, "\nok — %d check(s) passed\n", checks)
	} else {
		fmt.Fprintf(w, "\nFAILED — %d finding(s) across %d check(s)\n", total, checks)
	}
	return total
}
