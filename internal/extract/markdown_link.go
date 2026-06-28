package extract

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// mdLinkRe matches a Markdown inline link [text](target). The target capture may carry an #anchor
// and/or a "title"; both are trimmed before the path is resolved.
var mdLinkRe = regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)

// markdownLink extracts intra-repo Markdown links as edges. External links (scheme://, mailto:, tel:)
// and pure in-document anchors (#section) are skipped — only references that target another file in
// the graph become edges.
type markdownLink struct{}

// Type returns the edge type this extractor handles.
func (markdownLink) Type() string { return "markdown-link" }

// Extract reads n's file and emits one edge per intra-repo Markdown link, resolving the link target
// relative to n's directory into a root-relative node ID.
func (markdownLink) Extract(_ string, n *graph.Node, _ config.EdgeRule) ([]graph.Edge, error) {
	f, err := os.Open(n.Path)
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck // read-only file, close error is not actionable

	var edges []graph.Edge
	sc := bufio.NewScanner(f)
	line := 0
	for sc.Scan() {
		line++
		for _, m := range mdLinkRe.FindAllStringSubmatch(sc.Text(), -1) {
			target, ok := linkTarget(m[1])
			if !ok {
				continue
			}
			edges = append(edges, graph.Edge{
				From: n.ID,
				To:   resolveRef(n.ID, target),
				Type: "markdown-link",
				Loc:  locOf(n.ID, line),
			})
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return edges, nil
}

// linkTarget cleans a raw link target to the bare file path and reports whether it is an intra-repo
// reference worth an edge. It strips an optional "title", any #anchor, and surrounding angle brackets,
// and rejects external schemes and pure anchors.
func linkTarget(raw string) (string, bool) {
	t := strings.TrimSpace(raw)
	if i := strings.IndexAny(t, " \t"); i >= 0 { // drop an optional title: (path "title")
		t = t[:i]
	}
	t = strings.Trim(t, "<>")
	if i := strings.IndexByte(t, '#'); i >= 0 { // drop #anchor — M1 resolves the file, not the fragment
		t = t[:i]
	}
	if t == "" { // a pure #anchor link, or empty
		return "", false
	}
	if strings.Contains(t, "://") || strings.HasPrefix(t, "mailto:") || strings.HasPrefix(t, "tel:") {
		return "", false
	}
	return t, true
}
