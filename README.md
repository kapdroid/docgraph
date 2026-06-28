# docgraph

A **semantic doc/rule dependency-graph linter**. Generic link-checkers verify "is this hyperlink
alive." docgraph verifies the thing that actually matters in a docs-as-code / rules / ADR system:
**does every rule/doc reach its intended reader, through any link type, and is anything orphaned or
unreachable?** That is a graph-reachability problem no off-the-shelf link-checker models.

You describe your project's graph declaratively in `docgraph.yml` — what files are **nodes**, what
references between them are **edges** (markdown links, YAML pointers, regex citations, JSON paths,
frontmatter scopes), and what **assertions** must hold (no dangling links, no orphans, citations,
registration, location-consistency, and **reachability** from a reader to a rule). docgraph builds the
graph and checks it.

## Install

```sh
go install github.com/kapdroid/docgraph/cmd/docgraph@latest
```

## Usage

```sh
docgraph -config docgraph.yml                 # lint: run the configured checks (text report)
docgraph -config docgraph.yml -format junit   # JUnit XML for CI  (also: json, mermaid)
docgraph -config docgraph.yml -explain NODE   # trace how NODE is reached from its consumers
```

Exit codes: `0` all checks passed · `1` one or more checks failed · `2` the linter could not run
(bad config / unreadable file). `-explain` reports and never gates.

See **[docs/CONFIG.md](docs/CONFIG.md)** for the full `docgraph.yml` reference.

## What it checks

- **Edges (extractors):** `markdown-link`, `yaml-pointer`, `regex-cite`, `json-path`,
  `frontmatter-scope`.
- **Assertions:** `no-dangling`, `no-orphan`, `cites`, `registered`, `consistent`, `reachable`,
  optional `acyclic`.

The differentiator is `reachable`: model your *consumers* (who reads the docs) and the *moments* they
read at, and docgraph proves every rule in a set is reachable from a consumer through any chain of
references — catching a rule that exists but never reaches the agent/human who needs it.

## Milestones

- **M1 — link layer:** config + discovery + `markdown-link` + `no-dangling`/`no-orphan` + CLI.
- **M2 — semantic edges:** `yaml-pointer` · `regex-cite` · `json-path` · `frontmatter-scope` +
  `cites`/`registered`/`consistent`.
- **M3 — reachability:** consumers×moments model + `reachable` + `--explain` + mermaid/JSON/JUnit emit.

## Dogfooding

docgraph lints its own docs (`docgraph.yml` in this repo) and is the first customer of itself inside
the Kapdroid engine, where it replaces a bash `verify-context.sh` graph check.

## Development

```sh
go test ./...                      # unit + integration + golden + acceptance
go test ./test/golden/ -update     # refresh CLI golden files
```
