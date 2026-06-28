// Package discover turns a config's node-set globs into graph nodes: it walks the filesystem under
// the config root, matches each node-set's doublestar glob, and parses any requested YAML frontmatter
// off the matched files. It adds nodes only — extractors add edges later. Node IDs are slash-cleaned
// paths relative to the root, the single identity scheme every extractor resolves references to.
package discover

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"gopkg.in/yaml.v3"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// Discover builds a graph of nodes from cfg by expanding each node-set glob under cfg.Root and
// parsing declared frontmatter fields. The returned graph has no edges. A glob that matches nothing
// is not an error (an empty set is legitimate); an unreadable file or malformed frontmatter is.
func Discover(cfg *config.Config) (*graph.Graph, error) {
	g := graph.New()
	fsys := os.DirFS(cfg.Root)
	for name, ns := range cfg.Nodes {
		matches, err := doublestar.Glob(fsys, ns.Glob, doublestar.WithFilesOnly())
		if err != nil {
			return nil, fmt.Errorf("discover: node-set %q glob %q: %w", name, ns.Glob, err)
		}
		for _, rel := range matches {
			rel = filepath.ToSlash(rel)
			node := graph.Node{ID: rel, Kind: name, Path: filepath.Join(cfg.Root, rel)}
			if len(ns.Frontmatter) > 0 {
				fm, err := frontmatter(fsys, rel, ns.Frontmatter)
				if err != nil {
					return nil, fmt.Errorf("discover: %s: %w", rel, err)
				}
				node.Frontmatter = fm
			}
			g.AddNode(node)
		}
	}
	return g, nil
}

// fenceDelim is the line that opens and closes a YAML frontmatter block at the top of a file.
var fenceDelim = []byte("---")

// frontmatter reads rel through fsys and returns the requested fields parsed from a leading
// "---"-fenced YAML block. A file with no frontmatter yields an empty (non-nil) map — not an error,
// since not every member of a set carries frontmatter. Requested fields absent from the block are
// omitted. List/scalar values are flattened to their YAML string form.
func frontmatter(fsys fs.FS, rel string, fields []string) (map[string]string, error) {
	raw, err := fs.ReadFile(fsys, rel)
	if err != nil {
		return nil, fmt.Errorf("reading: %w", err)
	}
	block, ok := frontmatterBlock(raw)
	out := map[string]string{}
	if !ok {
		return out, nil
	}
	var all map[string]any
	if err := yaml.Unmarshal(block, &all); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}
	for _, f := range fields {
		if v, present := all[f]; present {
			out[f] = scalarString(v)
		}
	}
	return out, nil
}

// frontmatterBlock returns the bytes between the opening and closing "---" fences and whether a
// well-formed block was found. The opening fence must be the first line of the file.
func frontmatterBlock(raw []byte) ([]byte, bool) {
	lines := bytes.SplitAfter(raw, []byte("\n"))
	if len(lines) == 0 || !bytes.Equal(bytes.TrimRight(lines[0], "\r\n"), fenceDelim) {
		return nil, false
	}
	var block bytes.Buffer
	for _, line := range lines[1:] {
		if bytes.Equal(bytes.TrimRight(line, "\r\n"), fenceDelim) {
			return block.Bytes(), true
		}
		block.Write(line)
	}
	return nil, false // no closing fence
}

// scalarString renders a frontmatter value as a stable string: scalars verbatim, sequences joined
// with ", " (so a list field like covers/affected_paths is still queryable as text).
func scalarString(v any) string {
	if seq, ok := v.([]any); ok {
		parts := make([]string, len(seq))
		for i, e := range seq {
			parts[i] = fmt.Sprintf("%v", e)
		}
		return strings.Join(parts, ", ")
	}
	return fmt.Sprintf("%v", v)
}
