// Command docgraph is a semantic doc/rule dependency-graph linter: it models a project's docs, rules,
// and ADRs as a graph and asserts reachability and consistency, not merely that links are alive.
//
// Usage:
//
//	docgraph -config docgraph.yml              # lint: run the configured checks
//	docgraph -config docgraph.yml -explain N   # trace how node N is reached from consumers
//
// Exit codes: 0 = all checks passed, 1 = one or more checks failed, 2 = a configuration or runtime
// error (the linter could not run). -explain always exits 0/2 (it reports, it does not gate).
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kapdroid/docgraph/internal/app"
)

func main() {
	cfgPath := flag.String("config", "docgraph.yml", "path to the docgraph.yml config")
	explain := flag.String("explain", "", "trace how the given node ID is reached from consumers, then exit")
	flag.Parse()

	if *explain != "" {
		if err := app.Explain(*cfgPath, *explain, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "docgraph:", err)
			os.Exit(2)
		}
		return
	}

	passed, err := app.Run(*cfgPath, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "docgraph:", err)
		os.Exit(2)
	}
	if !passed {
		os.Exit(1)
	}
}
