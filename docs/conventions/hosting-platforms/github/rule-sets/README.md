# Hosting Platform: GitHub — Rule Sets
[INTENT: REFERENCE]

## Canonical source

The GitHub rule sets for the organization `t33n-software` are defined and
managed exactly once, centrally, in the repository
[`git-governance`](https://github.com/t33n-software/git-governance) under
`rulesets/github/`. That repository is the canonical source of truth for the
JSON definitions: it explains the architecture, applies the definitions, and
ships the versioned, importable artifacts.

A local copy, redefinition, or deviation in this repository is an anti-pattern
and forbidden (redundancy and drift prohibition). The only permitted
deviations are named, auditable repository exceptions that are stricter than
the organization baseline, never weaker.

## Family in use

This project (`license-hub`) uses the family **`quality-gates=full`**:

- The quality gates run on **Linux**, **Windows**, and **macOS**.
- Architectural rationale: this project ships the `license` CLI, which is
  built, attested, and verified as a native binary for all three operating
  systems; delivering for every operating system requires the full
  quality-gate matrix.

## Bound rule sets

| Rule set | Class |
|---|---|
| `push-protections: secret artifact boundary` | classless (private/internal visibility) |
| `branch-governance: ticket working branches` | classless (`~ALL`) |
| `branch-governance: develop shared line (quality-gates=full)` | full |
| `branch-governance: main shared line (quality-gates=full)` | full |
| `branch-governance: release shared lines (quality-gates=full)` | full |
| `branch-governance: support shared lines (quality-gates=full)` | full |

## Management

- Management level: the **organization** (`t33n-software`), never the
  individual repository level.
- Class membership of this repository: custom property `quality-gates=full`.
- Changes to the rule sets happen exclusively in the canonical repository and
  are then re-imported at organization level (Organization Settings →
  Repository → Rulesets).
