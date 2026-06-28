# Frozen acceptance checklist (ADR-0011)

QA-authored, behavior-level checks for the `docgraph` CLI — the independent specification a lane must
satisfy, verified **independently of the implementer**. This directory is **frozen** (`kapdroid
freeze`): the acceptance-guard hook + close Guard 7 refuse any edit to it once frozen, and the close
gate runs the suite **full** (never diff-scoped). The graded party never edits the test.

If a frozen case is genuinely wrong, QA re-authors it and the lane is re-frozen — the implementer
fixes **code**, never the checklist.

## The checklist (each is a runnable Go test)

| ID  | Behavior |
|-----|----------|
| AC1 | A healthy graph (`testdata/good`) exits 0 and reports `ok`. |
| AC2 | A broken graph (`testdata/broken`) exits 1 (checks failed, not a crash). |
| AC3 | A dangling link is reported, naming the missing target. |
| AC4 | An orphan node is reported, naming the unreferenced file. |
| AC5 | A missing config file exits 2 (config error, distinct from a failed check). |
| AC6 | Each check appears in the report with an explicit PASS/FAIL verdict. |
