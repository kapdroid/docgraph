package assert

import (
	"fmt"
	"sort"
	"strings"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// consistent asserts that every node carrying a location-DERIVED expectation (written by the
// frontmatter-scope extractor under graph.DerivedPrefix+field) has a parsed frontmatter field equal
// to it. A mismatch — including a missing/empty parsed field — is flagged, which is how a typo'd ADR
// scope is caught. Optional rule.In scopes the check to specific node kinds. The assertion is generic:
// it validates whatever fields frontmatter-scope derived, with no hardcoded field name.
type consistent struct{}

// Type returns the assert type this assertion handles.
func (consistent) Type() string { return "consistent" }

// Check flags each derived field whose parsed value differs from the location-derived expectation.
func (consistent) Check(g *graph.Graph, rule config.Assertion) []Finding {
	inScope := kindSet(rule.In)
	var findings []Finding
	for _, n := range g.Nodes() {
		if len(inScope) > 0 && !inScope[n.Kind] {
			continue
		}
		for _, field := range derivedFields(n) {
			expected := n.Frontmatter[graph.DerivedPrefix+field]
			actual := n.Frontmatter[field]
			if actual == expected {
				continue
			}
			findings = append(findings, Finding{
				Severity: SeverityError,
				Node:     n.ID,
				Message: fmt.Sprintf("%s: %s = %q but its location implies %q",
					n.ID, field, actual, expected),
			})
		}
	}
	return findings
}

// derivedFields returns, sorted for deterministic output, the field names a node carries a derived
// expectation for (the keys prefixed with graph.DerivedPrefix, with the prefix stripped).
func derivedFields(n *graph.Node) []string {
	var fields []string
	for k := range n.Frontmatter {
		if strings.HasPrefix(k, graph.DerivedPrefix) {
			fields = append(fields, strings.TrimPrefix(k, graph.DerivedPrefix))
		}
	}
	sort.Strings(fields)
	return fields
}
