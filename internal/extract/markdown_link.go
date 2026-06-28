package extract

import (
	"bufio"
	"fmt"
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
//
// Scope: this is a deliberately simple inline-link regex, not a full Markdown parse. Image embeds
// ![alt](path) are treated as links (the [alt](path) substring matches) — intentional, since an image
// is also a file reference docgraph should reach. Fenced code blocks (``` / ~~~) ARE skipped so the
// examples in documentation don't register as broken references (this fix was surfaced by dogfooding
// docgraph on its own docs, kap-ymj.18). Remaining limits, acceptable for the link layer: inline-code
// `[x](y)` still matches, reference-style [text][ref] links are not followed, and a nested-bracket
// label truncates at the first ']'. Upgrade to a real parser (goldmark) only if fidelity demands it.
type markdownLink struct{}

// Type returns the edge type this extractor handles.
func (markdownLink) Type() string { return "markdown-link" }

// Extract reads n's file and emits one edge per intra-repo Markdown link, resolving the link target
// relative to n's directory into a root-relative node ID.
func (markdownLink) Extract(n *graph.Node, _ config.EdgeRule) ([]graph.Edge, error) {
	f, err := os.Open(n.Path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", n.Path, err)
	}
	defer f.Close() //nolint:errcheck // read-only file, close error is not actionable

	var edges []graph.Edge
	sc := bufio.NewScanner(f)
	line := 0
	inFence := false
	for sc.Scan() {
		line++
		text := sc.Text()
		if isFence(text) { // ``` or ~~~ — toggle fenced-code state and skip the fence line itself
			inFence = !inFence
			continue
		}
		if inFence {
			continue // links inside a code block are examples, not real references
		}
		for _, m := range mdLinkRe.FindAllStringSubmatch(text, -1) {
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
		return nil, fmt.Errorf("scanning %s: %w", n.Path, err)
	}
	return edges, nil
}

// isFence reports whether a line opens or closes a fenced code block (``` or ~~~, optionally indented
// and with an info string). Links inside such a block are documentation examples, not references.
func isFence(line string) bool {
	t := strings.TrimLeft(line, " \t")
	return strings.HasPrefix(t, "```") || strings.HasPrefix(t, "~~~")
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
