// Package config defines the docgraph.yml schema and loads it into validated structs.
//
// docgraph.yml is the declarative description of a project's doc/rule graph: which files are nodes,
// which references between them are edges, and which reachability/consistency assertions must hold.
// The schema is intentionally small — only the fields milestones M1–M3 consume exist (YAGNI).
package config

import (
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is a parsed, validated docgraph.yml.
type Config struct {
	// Root is the directory globs are resolved against; defaults to "." when omitted.
	Root string `yaml:"root"`
	// Nodes maps a node-set name (e.g. "docs") to how its members are discovered.
	Nodes map[string]NodeSet `yaml:"nodes"`
	// Edges are the reference-extraction rules run over the discovered nodes.
	Edges []EdgeRule `yaml:"edges"`
	// Moments are the named lifecycle points a consumer reaches the graph at (M3).
	Moments []string `yaml:"moments"`
	// Consumers maps a consumer name to the node-sets it must reach and when (M3).
	Consumers map[string]Consumer `yaml:"consumers"`
	// Assert is the ordered list of checks the graph must satisfy.
	Assert []Assertion `yaml:"assert"`
}

// NodeSet describes how the members of one node-set are discovered.
type NodeSet struct {
	// Glob selects the files in this set, relative to Root (doublestar syntax).
	Glob string `yaml:"glob"`
	// Frontmatter lists YAML frontmatter fields to parse off each member, if any.
	Frontmatter []string `yaml:"frontmatter"`
}

// EdgeRule is one reference-extraction rule. Which fields are meaningful depends on Type; the
// per-extractor packages validate their own required fields (M1 markdown-link uses only From).
type EdgeRule struct {
	// Type names the extractor (markdown-link, yaml-pointer, regex-cite, json-path, frontmatter-scope).
	Type string `yaml:"type"`
	// From is the node-set the rule reads edges out of; it must name a defined node-set.
	From string `yaml:"from"`
	// To optionally constrains the node-set edges may point into (yaml-pointer, cites).
	To string `yaml:"to"`
	// Key is the YAML key whose value is a path (yaml-pointer).
	Key string `yaml:"key"`
	// Pattern is the regex capturing a path/id (regex-cite).
	Pattern string `yaml:"pattern"`
	// Path is the JSON path whose value is a node path (json-path).
	Path string `yaml:"path"`
	// Field is the frontmatter field that must match a location-derived value (frontmatter-scope).
	Field string `yaml:"field"`
	// MustMatchPath enables the location-consistency check for frontmatter-scope.
	MustMatchPath bool `yaml:"mustMatchPath"`
	// Derive maps a node's location to its expected Field value (frontmatter-scope). The expected
	// value convention is repo-specific, so it is configured here, not hardcoded (portability).
	Derive []DeriveRule `yaml:"derive"`
}

// DeriveRule maps a path prefix to an expected value, for frontmatter-scope. Under is a "/"-segmented
// path prefix where a "*" segment is a wildcard capture; Expect is a template where {1}, {2}, … are
// substituted with the captured segments. First matching rule wins. Example:
//
//	{ under: "stacks/*/decisions", expect: "stack:{1}" }
//
// derives "stack:go" for stacks/go/decisions/adr-0015-x.md.
type DeriveRule struct {
	// Under is the "/"-segmented path prefix to match, with "*" capturing one segment.
	Under string `yaml:"under"`
	// Expect is the expected value template; {n} is replaced by the n-th captured "*" segment.
	Expect string `yaml:"expect"`
}

// Consumer is a reader of the graph: the node-sets it must reach, at which moments (M3 reachable).
type Consumer struct {
	// Reaches lists the node-set names this consumer must be able to reach.
	Reaches []string `yaml:"reaches"`
	// At lists the moments the consumer reads the graph at.
	At []string `yaml:"at"`
}

// Assertion is one check over the built graph. Fields beyond Type are assertion-specific.
type Assertion struct {
	// Type names the check (no-dangling, no-orphan, cites, registered, consistent, reachable, acyclic).
	Type string `yaml:"type"`
	// In limits a check to specific node-sets (no-orphan).
	In []string `yaml:"in"`
	// From is the source node-set (cites) or the reach source (reachable: the literal "consumers").
	From string `yaml:"from"`
	// To is the required target node-set (cites).
	To string `yaml:"to"`
	// Set is the node-set whose members must all be reachable (reachable).
	Set string `yaml:"set"`
	// Registry is the node a set's members must appear in (registered).
	Registry string `yaml:"registry"`
}

// knownEdgeTypes and knownAssertTypes are the recognized schema vocabularies; loading rejects any
// type outside them so a typo'd config fails fast rather than silently doing nothing.
var (
	knownEdgeTypes = map[string]bool{
		"markdown-link": true, "yaml-pointer": true, "regex-cite": true,
		"json-path": true, "frontmatter-scope": true,
	}
	knownAssertTypes = map[string]bool{
		"no-dangling": true, "no-orphan": true, "cites": true, "registered": true,
		"consistent": true, "reachable": true, "acyclic": true,
	}
)

// Load reads, parses, and validates a docgraph.yml at path. It returns a descriptive error for an
// unreadable file, malformed YAML, or a config that is syntactically valid but semantically wrong
// (an edge/assert that names an undefined node-set or an unknown type).
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: reading %s: %w", path, err)
	}
	cfg, err := Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("config: %s: %w", path, err)
	}
	return cfg, nil
}

// Parse parses and validates docgraph.yml content already in memory. It is the testable core of Load.
func Parse(raw []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parsing yaml: %w", err)
	}
	if cfg.Root == "" {
		cfg.Root = "."
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	return &cfg, nil
}

// validate enforces the cross-field invariants the type system cannot.
func (c *Config) validate() error {
	if len(c.Nodes) == 0 {
		return errors.New("no node-sets defined (nodes: is empty)")
	}
	for name, ns := range c.Nodes {
		if ns.Glob == "" {
			return fmt.Errorf("node-set %q has no glob", name)
		}
	}
	for i, e := range c.Edges {
		if !knownEdgeTypes[e.Type] {
			return fmt.Errorf("edge[%d]: unknown type %q", i, e.Type)
		}
		if e.From == "" {
			return fmt.Errorf("edge[%d] (%s): missing from", i, e.Type)
		}
		if _, ok := c.Nodes[e.From]; !ok {
			return fmt.Errorf("edge[%d] (%s): from %q is not a defined node-set", i, e.Type, e.From)
		}
		if e.To != "" {
			if _, ok := c.Nodes[e.To]; !ok {
				return fmt.Errorf("edge[%d] (%s): to %q is not a defined node-set", i, e.Type, e.To)
			}
		}
	}
	// Assert node-set references (In/From/To/Set) are validated by the per-assert packages that
	// consume them (M2/M3 beads) — mirroring the per-extractor field validation above (KISS); here we
	// only pin the assert vocabulary itself.
	for i, a := range c.Assert {
		if !knownAssertTypes[a.Type] {
			return fmt.Errorf("assert[%d]: unknown type %q", i, a.Type)
		}
	}
	return nil
}
