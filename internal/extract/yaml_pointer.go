package extract

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// yamlPointer extracts an edge from a YAML key whose value is a repo path — e.g. a stack.yml
// `rules: stacks/go/rules.md` becomes an edge from the stack node to the rules doc. The key may be
// dotted to reach a nested value (`overrides.gate_prepend`). Configured via rule.Key (required).
type yamlPointer struct{}

// Type returns the edge type this extractor handles.
func (yamlPointer) Type() string { return "yaml-pointer" }

// Extract reads n's YAML, navigates rule.Key, and emits one edge to the path value (root-relative).
// An absent key or a non-scalar value yields no edge (not an error — only this node's own read/parse
// failing is). A missing-but-required pointer is a `registered`/`cites` assertion's concern, not the
// extractor's.
func (yamlPointer) Extract(n *graph.Node, rule config.EdgeRule) ([]graph.Edge, error) {
	if rule.Key == "" {
		return nil, fmt.Errorf("yaml-pointer: edge from %q has no key", rule.From)
	}
	raw, err := os.ReadFile(n.Path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", n.Path, err)
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", n.Path, err)
	}
	val, ok := navigateYAML(doc, strings.Split(rule.Key, "."))
	if !ok {
		return nil, nil
	}
	s, ok := val.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return nil, nil
	}
	return []graph.Edge{{
		From: n.ID,
		To:   rootRef(s),
		Type: "yaml-pointer",
		Loc:  fmt.Sprintf("%s:%s", n.ID, rule.Key),
	}}, nil
}

// navigateYAML walks a dotted key path through nested maps, returning the leaf value and whether the
// full path resolved to a present value.
func navigateYAML(doc map[string]any, keys []string) (any, bool) {
	var cur any = doc
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		cur, ok = m[k]
		if !ok {
			return nil, false
		}
	}
	return cur, true
}
