// Package app wires the docgraph pipeline: load config → discover nodes → extract edges → run
// assertions → render the report. It is the testable core behind cmd/docgraph (the binary is a thin
// flag wrapper), so the pipeline can be exercised without exec'ing a process.
package app

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/kapdroid/docgraph/internal/assert"
	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/discover"
	"github.com/kapdroid/docgraph/internal/extract"
	"github.com/kapdroid/docgraph/internal/report"
)

// Run loads the docgraph.yml at cfgPath, builds the graph, runs the configured assertions, writes the
// text report to w, and returns whether every check passed. A configuration, discovery, or extraction
// failure (as opposed to an assertion finding) is returned as an error — the caller maps that to a
// distinct exit code from "checks failed".
func Run(cfgPath string, w io.Writer) (bool, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return false, fmt.Errorf("load config: %w", err)
	}
	// Resolve the config's root relative to the config file's location (not the process CWD), so
	// `docgraph -config path/to/docgraph.yml` works from anywhere. config.Load stays a pure parser;
	// filesystem resolution is the app layer's job. An absolute root is honored as-is.
	if !filepath.IsAbs(cfg.Root) {
		cfg.Root = filepath.Join(filepath.Dir(cfgPath), cfg.Root)
	}
	g, err := discover.Discover(cfg)
	if err != nil {
		return false, fmt.Errorf("discover: %w", err)
	}
	if err := extract.NewRegistry().Extract(g, cfg.Edges); err != nil {
		return false, fmt.Errorf("extract: %w", err)
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
