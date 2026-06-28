// Package assert runs the checks a docgraph must satisfy over a built graph. Each check kind
// (no-dangling, no-orphan, and the M2/M3 semantic/reachability checks) is an Assertion; a Registry
// wires the built-ins and runs every configured assertion, returning one Result per check so the
// report can show per-check pass/fail. Adding a check is one file implementing Assertion plus one
// line in NewRegistry — the same pluggable pattern as the extractors.
package assert

import (
	"fmt"

	"github.com/kapdroid/docgraph/internal/config"
	"github.com/kapdroid/docgraph/internal/graph"
)

// Severity classifies a finding. Every M1 finding is an error (a violation that fails the check);
// SeverityWarn exists for M2/M3 checks that want to report without failing.
const (
	// SeverityError marks a finding that fails its assertion.
	SeverityError = "error"
	// SeverityWarn marks an advisory finding that does not fail its assertion.
	SeverityWarn = "warn"
)

// Finding is a single violation of one assertion. Node and/or Edge locate it; both may be empty when
// the violation is graph-wide.
type Finding struct {
	// Severity is SeverityError (fails the check) or SeverityWarn (advisory). M1 checks emit errors.
	Severity string
	// Node is the ID of the offending node, if the violation is about a node.
	Node string
	// Edge is the offending edge, if the violation is about an edge (nil otherwise).
	Edge *graph.Edge
	// Message is the human-readable explanation. For path-shaped findings (M3 reachable --explain)
	// the path is rendered into Message; structured path output is a later concern if needed.
	Message string
}

// Result is the outcome of running one assertion: its type and any findings (empty Findings = pass).
type Result struct {
	// Type is the assertion type that produced this result.
	Type string
	// Findings are the violations; an empty slice means the assertion passed.
	Findings []Finding
}

// Passed reports whether the assertion found no violations.
func (r Result) Passed() bool { return len(r.Findings) == 0 }

// Assertion checks the built graph under one configured assert rule and returns its findings.
//
// Seam invariant (decided kap-ymj.5, for M2/M3): Check receives only the graph and its own rule — NOT
// the whole Config. An assertion that needs consumers/moments (the M3 reachable assert) consumes them
// as nodes/edges already materialized in g: the M3 consumers×moments build (kap-ymj.15) injects
// Config.Consumers/Moments as consumer/moment-kind nodes, so reachable is pure BFS over g. Keep it
// this way; only a genuine need for whole-Config access justifies widening this signature (which would
// ripple to every assertion).
type Assertion interface {
	// Type is the docgraph.yml assert `type` this assertion handles.
	Type() string
	// Check returns the violations of this assertion over g under rule (empty/nil = pass).
	Check(g *graph.Graph, rule config.Assertion) []Finding
}

// Registry holds assertions keyed by type. Construct it with NewRegistry; no global state.
type Registry struct {
	byType map[string]Assertion
}

// NewRegistry returns a registry with every built-in assertion registered. New assertion kinds are
// added here, one line each.
func NewRegistry() *Registry {
	r := &Registry{byType: map[string]Assertion{}}
	r.Register(noDangling{})
	r.Register(noOrphan{})
	r.Register(cites{})
	r.Register(registered{})
	r.Register(consistent{})
	return r
}

// Register adds or replaces the assertion for its Type.
func (r *Registry) Register(a Assertion) {
	r.byType[a.Type()] = a
}

// Get returns the assertion for a type and whether one is registered.
func (r *Registry) Get(t string) (Assertion, bool) {
	a, ok := r.byType[t]
	return a, ok
}

// Check runs every configured assertion over g, in config order, returning one Result each. It fails
// fast if a rule names an assert type with no registered assertion (a config the loader allowed but
// this build cannot execute).
func (r *Registry) Check(g *graph.Graph, asserts []config.Assertion) ([]Result, error) {
	results := make([]Result, 0, len(asserts))
	for i, rule := range asserts {
		a, ok := r.Get(rule.Type)
		if !ok {
			return nil, fmt.Errorf("assert: assert[%d]: no assertion registered for type %q", i, rule.Type)
		}
		results = append(results, Result{Type: rule.Type, Findings: a.Check(g, rule)})
	}
	return results, nil
}
