package extract

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// jsonPath extracts edges from JSON string values at a configured path. The path is dot-separated
// keys; a segment ending in "[]" iterates an array (e.g. "hooks.PreToolUse[].command"). Every string
// leaf reached becomes a root-relative edge. Configured via rule.Path (required). Non-string leaves
// are ignored. Point it at a path you know holds repo paths; a value that isn't one becomes a
// dangling edge the no-dangling assertion reports.
type jsonPath struct{}

// Type returns the edge type this extractor handles.
func (jsonPath) Type() string { return "json-path" }

// Extract reads n's JSON and emits an edge for each string value at rule.Path.
func (jsonPath) Extract(n *graph.Node, rule config.EdgeRule) ([]graph.Edge, error) {
	if rule.Path == "" {
		return nil, fmt.Errorf("json-path: edge from %q has no path", rule.From)
	}
	raw, err := os.ReadFile(n.Path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", n.Path, err)
	}
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", n.Path, err)
	}
	var edges []graph.Edge
	for _, v := range collectJSON(doc, strings.Split(rule.Path, ".")) {
		if strings.TrimSpace(v) == "" {
			continue
		}
		edges = append(edges, graph.Edge{
			From: n.ID,
			To:   rootRef(v),
			Type: "json-path",
			Loc:  fmt.Sprintf("%s:%s", n.ID, rule.Path),
		})
	}
	return edges, nil
}

// collectJSON walks a dotted path (a "[]" suffix iterates an array) and returns every string leaf
// reached. An unmatched key or type mismatch yields no values (not an error).
func collectJSON(cur any, segs []string) []string {
	if len(segs) == 0 {
		if s, ok := cur.(string); ok {
			return []string{s}
		}
		return nil
	}
	seg, rest := segs[0], segs[1:]
	iterArray := strings.HasSuffix(seg, "[]")
	seg = strings.TrimSuffix(seg, "[]")

	next := cur
	if seg != "" {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		if next, ok = m[seg]; !ok {
			return nil
		}
	}
	if !iterArray {
		return collectJSON(next, rest)
	}
	list, ok := next.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, e := range list {
		out = append(out, collectJSON(e, rest)...)
	}
	return out
}
