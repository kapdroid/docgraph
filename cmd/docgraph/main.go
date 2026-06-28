// Command docgraph is a semantic doc/rule dependency-graph linter: it models a project's docs, rules,
// and ADRs as a graph and asserts reachability and consistency, not merely that links are alive.
//
// Usage:
//
//	docgraph -config docgraph.yml
//
// Exit codes: 0 = all checks passed, 1 = one or more checks failed, 2 = a configuration or runtime
// error (the linter could not run).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kapdroid/docgraph/internal/app"
)

func main() {
	cfgPath := flag.String("config", "docgraph.yml", "path to the docgraph.yml config")
	flag.Parse()

	passed, err := app.Run(*cfgPath, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "docgraph:", err)
		os.Exit(2)
	}
	if !passed {
		os.Exit(1)
	}
}
