# CLI Convention: Value-Domain Single Source of Truth
[INTENT: SPECIFICATION]

This document is the binding convention for the sourcing of value domains of
every CLI tool of the organization: every value domain is defined exactly
once, and every consumption channel renders from that one source. The value
classes themselves are owned by [README.md](README.md) and are not restated
here.

## 1. One source, five channels

Every value domain — value lists, grammars, rule sets, limits, examples — is
defined **exactly once** in the domain or policy layer of the tool. Five
consumption channels render from this one source, never from their own
copies:

```text
K1  static --help texts
K2  interactive prompts and selects
K3  error contracts (Expected/Rule/Example/Remediation)
K4  shell completion on the value level
K5  machine-readable discovery endpoint (for example policy/describe with
    JSON output)
```

Binding rules:

1. **No literal duplicates — anywhere.** Error texts, too, must not carry
   value lists as handwritten strings; they render the same source. A
   duplicated list silently drifts apart at the next domain extension.
2. **Endpoint subsets via declared filters.** If an endpoint accepts a
   subset, the subset is produced through a declared filter on the source,
   never as a separate hand copy.
3. **Contract tests pin every projection.** Tests prove that help, errors,
   prompts, completion, and the discovery output match the source (see
   [../testing/README.md](../testing/README.md)).
4. **Drift is release-blocking.** A deviation between a projection and the
   canonical source is a defect that blocks the release — not a warning.

Positive example:

```text
Domain registry (one source) -> help renderer, prompt renderer, error
renderer, completion renderer, discovery endpoint; one contract test per
channel pins the projection against the registry.
```

Negative example: the value list exists four times independently — as a help
string, as an error string, as a prompt text, and as a documentation table.
At the next domain extension at least two surfaces show outdated values —
parallel truths that lose their right to exist.

## 2. The multi-decision matrix: value class × channel

The binding overall view fixes, per value class, what every channel shows.
Row rule of the matrix: **no cell may contain a divergent answer — only a
compressed one** (same source, different level of detail).

| Value class | K1 static help | K2 interactive prompt | K3 error contract | K4 completion | K5 discovery endpoint |
|---|---|---|---|---|---|
| `closed-enum` | full endpoint list | select with labels + descriptions | list in `Expected` | complete values | list + metadata |
| `shaped` | grammar + example + subset | rule text + live validation | rule + example | prefix form | pattern/grammar |
| `free-constrained` | rule set + example + rejection/convention label | rule text + live validation | concrete violation + rule | — | limits/regex |
| `structural-reference` | form + resolution rule | form + resolution rule | resolution error + form | path/ref suggestions | — |
| `scalar-bounded` | type + range + default | range + default | range | — | range |
| `boolean-switch` | effect + default | — | — | — | — |
| `composite-token` | transport form + token grammar + example | — | form error + example | token prefixes | token grammar |
| `secret-reference` | reference forms | reference forms | reference error | — | — |

A dash marks a channel that carries no projection for the class; it never
marks permission to maintain a divergent answer on that channel.
