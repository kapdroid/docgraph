package report

import (
	"strings"
	"testing"

	"github.com/kapdroid/docgraph/internal/assert"
)

func TestAllPassed(t *testing.T) {
	cases := []struct {
		name    string
		results []assert.Result
		want    bool
	}{
		{"empty is passed", nil, true},
		{"all pass", []assert.Result{{Type: "no-dangling"}}, true},
		{"one fails", []assert.Result{
			{Type: "no-dangling"},
			{Type: "no-orphan", Findings: []assert.Finding{{Message: "x"}}},
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := AllPassed(tc.results); got != tc.want {
				t.Errorf("AllPassed = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestTextReport(t *testing.T) {
	results := []assert.Result{
		{Type: "no-dangling", Findings: []assert.Finding{{Message: "a references b"}}},
		{Type: "no-orphan"},
	}
	var b strings.Builder
	n, err := Text(&b, results)
	if err != nil {
		t.Fatalf("Text returned error: %v", err)
	}
	s := b.String()
	if n != 1 {
		t.Errorf("findings written = %d, want 1", n)
	}
	for _, want := range []string{"FAIL  no-dangling (1)", "a references b", "PASS  no-orphan", "FAILED — 1 finding"} {
		if !strings.Contains(s, want) {
			t.Errorf("report missing %q:\n%s", want, s)
		}
	}
}

func TestTextReportClean(t *testing.T) {
	var b strings.Builder
	if _, err := Text(&b, []assert.Result{{Type: "no-dangling"}, {Type: "no-orphan"}}); err != nil {
		t.Fatalf("Text returned error: %v", err)
	}
	if !strings.Contains(b.String(), "ok — 2 check(s) passed") {
		t.Errorf("clean report missing ok summary:\n%s", b.String())
	}
}
