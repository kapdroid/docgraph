// Package app wires the docgraph pipeline: load config → discover nodes → extract edges → run
// assertions → render the report. It is the testable core behind cmd/docgraph (the binary is a thin
// flag wrapper), so the pipeline can be exercised without exec'ing a process.
package app

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kapdroid/docgraph/internal/assert"
	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/discover"
	"github.com/kapdroid/docgraph/internal/extract"
	"github.com/kapdroid/docgraph/internal/graph"
	"github.com/kapdroid/docgraph/internal/reach"
	"github.com/kapdroid/docgraph/internal/report"
)

// build loads and resolves the config, then runs discover → extract → materialize, returning the
// fully built graph (nodes, edges, and the consumers×moments layer) ready for assertions or explain.
// It is the shared front half of every command so the pipeline lives in exactly one place.
func build(cfgPath string) (*config.Config, *graph.Graph, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}
	// Resolve the config's root relative to the config file's location (not the process CWD), so
	// `docgraph -config path/to/docgraph.yml` works from anywhere. config.Load stays a pure parser;
	// filesystem resolution is the app layer's job. An absolute root is honored as-is.
	if !filepath.IsAbs(cfg.Root) {
		cfg.Root = filepath.Join(filepath.Dir(cfgPath), cfg.Root)
	}
	g, err := discover.Discover(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("discover: %w", err)
	}
	if err := extract.NewRegistry().Extract(g, cfg.Edges); err != nil {
		return nil, nil, fmt.Errorf("extract: %w", err)
	}
	// Inject the consumers×moments layer (M3) so reachability sees consumers as entry nodes.
	// No-op when no consumers are configured.
	reach.Materialize(cfg, g)
	return cfg, g, nil
}

// Run builds the graph, runs the configured assertions, writes the text report to w, and returns
// whether every check passed. A configuration/discovery/extraction failure (as opposed to an
// assertion finding) is returned as an error — the caller maps that to a distinct exit code.
func Run(cfgPath string, w io.Writer) (bool, error) {
	cfg, g, err := build(cfgPath)
	if err != nil {
		return false, err
	}
	results, err := assert.NewRegistry().Check(g, cfg.Assert)
	if err != nil {
		return false, fmt.Errorf("assert: %w", err)
	}
	if err := report.Text(w, results); err != nil {
		return false, fmt.Errorf("report: %w", err)
	}
	return report.AllPassed(results), nil
}

// errWriter records the first write error so a sequence of formatted writes can be checked once.
type errWriter struct {
	w   io.Writer
	err error
}

func (e *errWriter) printf(format string, a ...any) {
	if e.err == nil {
		_, e.err = fmt.Fprintf(e.w, format, a...)
	}
}

// Explain builds the graph and writes how the given node is reached from the consumer entry nodes: a
// path (consumer → … → node) if reachable, or a clear "unreachable" line plus the node's direct
// referrers (inbound edges) to help debug why. Returns an error if the pipeline could not run or a
// write failed.
func Explain(cfgPath, node string, w io.Writer) error {
	_, g, err := build(cfgPath)
	if err != nil {
		return err
	}
	ew := &errWriter{w: w}
	switch {
	case !g.Has(node):
		ew.printf("%s: no such node in the graph\n", node)
	default:
		sources := reach.ConsumerIDs(g)
		if path, ok := reach.PathTo(g, sources, node); ok {
			ew.printf("%s is reachable:\n  %s\n", node, strings.Join(path, " → "))
			break
		}
		noConsumers := ""
		if len(sources) == 0 {
			noConsumers = " (no consumers configured)"
		}
		ew.printf("%s is UNREACHABLE from consumers%s\n", node, noConsumers)
		if inbound := g.Inbound(node); len(inbound) == 0 {
			ew.printf("  nothing references it (it is an orphan)\n")
		} else {
			ew.printf("  referenced by (but none on a path from a consumer):\n")
			for _, e := range inbound {
				ew.printf("    %s (%s)\n", e.From, e.Type)
			}
		}
	}
	return ew.err
}
