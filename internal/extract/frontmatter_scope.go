package extract

import (
	"fmt"
	"strings"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// frontmatterScope is the one extractor that records a node ATTRIBUTE rather than an edge: it derives
// the value a node's frontmatter field SHOULD have from the node's location (per rule.Derive) and
// stores it on the node under graph.DerivedPrefix+field. The `consistent` assertion (kap-ymj.14) then
// compares the parsed field to this derived expectation. Consistency is a property of a single node,
// not a reference between two, so forcing it through an Edge would be the wrong model (foreseen in the
// .4 seam review) — it returns no edges by design; its output is the annotation.
type frontmatterScope struct{}

// Type returns the edge type this extractor handles.
func (frontmatterScope) Type() string { return "frontmatter-scope" }

// Extract derives the expected value of rule.Field from n's location and annotates n with it. It
// returns no edges. A node whose location matches no derive rule is left unannotated (the consistent
// assertion only checks annotated nodes). Requires rule.Field.
func (frontmatterScope) Extract(n *graph.Node, rule config.EdgeRule) ([]graph.Edge, error) {
	if rule.Field == "" {
		return nil, fmt.Errorf("frontmatter-scope: edge from %q has no field", rule.From)
	}
	if !rule.MustMatchPath {
		return nil, nil // nothing to derive without the location-consistency flag
	}
	expected, ok := deriveExpected(n.ID, rule.Derive)
	if !ok {
		return nil, nil
	}
	if n.Frontmatter == nil {
		n.Frontmatter = map[string]string{}
	}
	n.Frontmatter[graph.DerivedPrefix+rule.Field] = expected
	return nil, nil
}

// deriveExpected returns the expected value for a node at id by matching id's leading path segments
// against the derive rules (first match wins). A "*" segment in a rule's Under captures the
// corresponding path segment; {1}, {2}, … in Expect are replaced by the captures in order.
func deriveExpected(id string, rules []config.DeriveRule) (string, bool) {
	segs := strings.Split(id, "/")
	for _, r := range rules {
		under := strings.Split(strings.Trim(r.Under, "/"), "/")
		if len(under) > len(segs) {
			continue
		}
		caps := make([]string, 0, len(under))
		matched := true
		for i, u := range under {
			if u == "*" {
				caps = append(caps, segs[i])
				continue
			}
			if u != segs[i] {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		out := r.Expect
		for i, c := range caps {
			out = strings.ReplaceAll(out, fmt.Sprintf("{%d}", i+1), c)
		}
		return out, true
	}
	return "", false
}
