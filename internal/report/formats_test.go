package report

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/kapdroid/docgraph/internal/assert"
	"github.com/kapdroid/docgraph/internal/graph"
)

func sampleGraph() *graph.Graph {
	g := graph.New()
	g.AddNode(graph.Node{ID: "docs/a.md", Kind: "docs"})
	g.AddNode(graph.Node{ID: "docs/b.md", Kind: "docs"})
	g.AddEdge(graph.Edge{From: "docs/a.md", To: "docs/b.md", Type: "markdown-link", Loc: "docs/a.md:1"})
	g.AddEdge(graph.Edge{From: "docs/a.md", To: "docs/ghost.md", Type: "markdown-link", Loc: "docs/a.md:2"})
	return g
}

func sampleResults() []assert.Result {
	return []assert.Result{
		{Type: "no-dangling", Findings: []assert.Finding{{Severity: assert.SeverityError, Node: "docs/a.md", Message: "dangling to docs/ghost.md"}}},
		{Type: "no-orphan"},
	}
}

func TestMermaid(t *testing.T) {
	var b strings.Builder
	if err := Mermaid(&b, sampleGraph()); err != nil {
		t.Fatalf("Mermaid error: %v", err)
	}
	s := b.String()
	for _, want := range []string{"flowchart LR", `"docs/a.md"`, "-->|markdown-link|", "missing"} {
		if !strings.Contains(s, want) {
			t.Errorf("mermaid output missing %q:\n%s", want, s)
		}
	}
}

func TestJSONValidAndComplete(t *testing.T) {
	var b strings.Builder
	if err := JSON(&b, sampleGraph(), sampleResults()); err != nil {
		t.Fatalf("JSON error: %v", err)
	}
	var out struct {
		Nodes  []struct{ ID, Kind string }
		Edges  []struct{ From, To, Type string }
		Checks []struct {
			Type   string
			Passed bool
		}
	}
	if err := json.Unmarshal([]byte(b.String()), &out); err != nil {
		t.Fatalf("emitted JSON is invalid: %v\n%s", err, b.String())
	}
	if len(out.Nodes) != 2 || len(out.Edges) != 2 || len(out.Checks) != 2 {
		t.Errorf("json counts: nodes=%d edges=%d checks=%d, want 2/2/2", len(out.Nodes), len(out.Edges), len(out.Checks))
	}
	if out.Checks[0].Passed {
		t.Error("no-dangling check should be Passed=false in JSON")
	}
}

func TestJUnitValidAndComplete(t *testing.T) {
	var b strings.Builder
	if err := JUnit(&b, sampleResults()); err != nil {
		t.Fatalf("JUnit error: %v", err)
	}
	var suite struct {
		Tests    int `xml:"tests,attr"`
		Failures int `xml:"failures,attr"`
		Cases    []struct {
			Name     string `xml:"name,attr"`
			Failures []struct {
				Message string `xml:"message,attr"`
			} `xml:"failure"`
		} `xml:"testcase"`
	}
	body := strings.TrimPrefix(b.String(), xml.Header)
	if err := xml.Unmarshal([]byte(body), &suite); err != nil {
		t.Fatalf("emitted JUnit is invalid XML: %v\n%s", err, b.String())
	}
	if suite.Tests != 2 || suite.Failures != 1 {
		t.Errorf("junit tests=%d failures=%d, want 2/1", suite.Tests, suite.Failures)
	}
	if len(suite.Cases) != 2 || len(suite.Cases[0].Failures) != 1 {
		t.Errorf("junit cases shape wrong: %+v", suite.Cases)
	}
}
