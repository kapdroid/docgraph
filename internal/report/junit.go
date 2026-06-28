package report

import (
	"encoding/xml"
	"io"

	"github.com/kapdroid/docgraph/internal/assert"
)

// JUnit shapes — one testsuite, one testcase per check, one failure element per finding. This is the
// format CI systems ingest to show docgraph results alongside unit tests.
type junitSuite struct {
	XMLName  xml.Name    `xml:"testsuite"`
	Name     string      `xml:"name,attr"`
	Tests    int         `xml:"tests,attr"`
	Failures int         `xml:"failures,attr"`
	Cases    []junitCase `xml:"testcase"`
}

type junitCase struct {
	Name     string         `xml:"name,attr"`
	Failures []junitFailure `xml:"failure"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
}

// JUnit writes the check results as a JUnit XML test suite: each assertion is a test case, each
// finding a failure. Exit-gating stays the caller's job (via AllPassed).
func JUnit(w io.Writer, results []assert.Result) error {
	suite := junitSuite{Name: "docgraph", Tests: len(results)}
	for _, r := range results {
		c := junitCase{Name: r.Type}
		for _, f := range r.Findings {
			c.Failures = append(c.Failures, junitFailure{Message: f.Message})
		}
		if !r.Passed() {
			suite.Failures++
		}
		suite.Cases = append(suite.Cases, c)
	}
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(suite); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\n")
	return err
}
