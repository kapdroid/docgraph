package extract

import (
	"bufio"
	"fmt"
	"os"
	"regexp"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// regexCite extracts edges from a regex that captures a repo path/id — e.g. an agent file citing
// `stacks/[a-z_]+/reviewers/[a-z-]+\.md`. If the pattern has a capture group, group 1 is the path;
// otherwise the whole match is. Each unique match per line becomes a root-relative edge. Configured
// via rule.Pattern (required). The same citation appearing on multiple lines yields one edge per line
// (each with its own Loc), which a `cites` assertion can dedupe if it cares.
type regexCite struct{}

// Type returns the edge type this extractor handles.
func (regexCite) Type() string { return "regex-cite" }

// Extract compiles rule.Pattern and emits an edge for each match in n's file.
func (regexCite) Extract(n *graph.Node, rule config.EdgeRule) ([]graph.Edge, error) {
	if rule.Pattern == "" {
		return nil, fmt.Errorf("regex-cite: edge from %q has no pattern", rule.From)
	}
	re, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return nil, fmt.Errorf("regex-cite: compiling pattern %q: %w", rule.Pattern, err)
	}
	f, err := os.Open(n.Path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", n.Path, err)
	}
	defer f.Close() //nolint:errcheck // read-only file, close error is not actionable

	var edges []graph.Edge
	sc := bufio.NewScanner(f)
	line := 0
	for sc.Scan() {
		line++
		for _, m := range re.FindAllStringSubmatch(sc.Text(), -1) {
			cited := m[0]
			if len(m) > 1 {
				cited = m[1] // first capture group is the path when present
			}
			if cited == "" {
				continue
			}
			edges = append(edges, graph.Edge{
				From: n.ID,
				To:   rootRef(cited),
				Type: "regex-cite",
				Loc:  locOf(n.ID, line),
			})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scanning %s: %w", n.Path, err)
	}
	return edges, nil
}
