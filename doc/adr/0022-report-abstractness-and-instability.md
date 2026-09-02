# 22. Report abstractness and instability alongside distance

Date: 2026-09-02

## Status

Accepted

## Context

ADR 0020 hid abstractness and instability as compute-only inputs: they were
not columns, flags, or policy keys. Distance is `|A + I − 1|`; readers still
need A and I to interpret D.

Policy (ADR 0021) is a first-match list of `{pattern, max-distance}`. Making
A or I selectable or gateable would reopen a metric-picker surface this
linter does not have.

## Decision

Include abstractness and instability in the reported metric set, in that
order, immediately before distance. They appear in text, JSON, CSV, and HTML.

They are not CLI-selectable (`--metrics` stays gone) and not policy-gateable.
The bound remains a maximum on distance only. The formula is unchanged:
`|A + I − 1|`. A config that names abstractness or instability as a policy
metric is rejected as an unknown setting.

Bump the report schema to version 6 because the per-package `metrics` map
gains two keys.

This ADR supersedes ADR 0020's "not reported" clause. ADR 0020's "not
selectable or gateable" clause still holds.

## Consequences

Reports show A, I, and D. Downstream consumers that assumed a single
`distance` key in `metrics` must read schema 6. Policy YAML and CLI flags
are unchanged.
