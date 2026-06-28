# docgraph.yml reference

The config describes your graph and the checks over it. Paths are relative to `root` (which is itself
relative to the config file's location). See the [README](../README.md) for usage.

```yaml
root: .                              # base dir for globs (default ".")

nodes:                               # named node-sets, discovered by glob
  docs:  { glob: "**/*.md" }
  adrs:  { glob: "**/decisions/adr-*.md", frontmatter: [id, scope, status] }
  stacks: { glob: "stacks/*/stack.yml" }

edges:                               # reference extractors run over a From node-set
  - { type: markdown-link, from: docs }                          # [text](path#anchor)
  - { type: yaml-pointer,  from: stacks, key: rules, to: docs }  # a YAML key whose value is a path
  - { type: regex-cite,    from: agents, pattern: 'stacks/\S+\.md' }  # capture group 1 = the path
  - { type: json-path,     from: config, path: "hooks[].command" }   # [] iterates arrays
  - { type: frontmatter-scope, from: adrs, field: scope, mustMatchPath: true,
      derive:                                                    # path prefix -> expected value
        - { under: "docs/decisions",     expect: "engine" }
        - { under: "stacks/*/decisions", expect: "stack:{1}" }  # {1} = the captured * segment
        - { under: "tenants/*/decisions", expect: "tenant:{1}" } }

moments: [session, lane, edit, review]   # lifecycle points consumers read at (M3)

consumers:                               # who reads the graph, what they reach, when (M3)
  implementer: { reaches: [stacks], at: [lane, edit] }

assert:
  - { type: no-dangling }                          # every edge target exists
  - { type: no-orphan, in: [docs, adrs] }          # no unreferenced node in these sets
  - { type: cites, from: reviewers, to: docs }     # every reviewer cites a doc
  - { type: registered, from: adrs, registry: docs/decisions/README.md }  # listed in the index
  - { type: consistent }                           # frontmatter-scope field == derived value
  - { type: reachable, set: stacks, from: consumers }  # every stack reachable from a consumer
  - { type: acyclic }                              # optional: no reference cycles
```

## Extractors

| type | reads | edge target |
|---|---|---|
| `markdown-link` | inline markdown links | file-relative path (anchors/titles stripped; external/mailto skipped; fenced code skipped) |
| `yaml-pointer` | the value at `key` (dotted for nested) | root-relative path |
| `regex-cite` | each match of `pattern` (capture group 1, else whole match) | root-relative path |
| `json-path` | string values at `path` (`[]` iterates arrays) | root-relative path |
| `frontmatter-scope` | derives `field`'s expected value from the file location | annotation (no edge), consumed by `consistent` |

## Assertions

| type | fails when |
|---|---|
| `no-dangling` | an edge points at a node that does not exist |
| `no-orphan` | a node (optionally scoped by `in`) has no inbound edge (abstract consumer/moment nodes exempt) |
| `cites` | a `from`-set node has no edge into the `to` set |
| `registered` | a `from`-set node is not referenced from the `registry` node |
| `consistent` | a parsed frontmatter field differs from its location-derived value |
| `reachable` | a `set` node is not reachable from the entry sources (`from: consumers`, or a node-set) |
| `acyclic` | the reference graph contains a directed cycle |
