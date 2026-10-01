# CLI Convention: Interaction Model
[INTENT: SPECIFICATION]

This document is the binding convention for the interaction model of every
CLI tool of the organization: non-interactive completeness, same-source
prompts, the mutation confirmation, the dry-run duty, terminal detection, and
accessible output.

## 1. Non-interactive completeness

Every capability of the tool is reachable without a terminal — through flags
or explicit configuration. The interactive mode is a guided comfort layer,
never the only path; if a required value is missing in a non-interactive
call, the call fails closed with a named remediation instead of hanging.

## 2. Prompts from the same source

Interactive prompts and selects carry the same rule text and the same options
as the static help — rendered from the same source (see
[../values/single-source-of-truth.md](../values/single-source-of-truth.md)),
with live validation. Two truths between prompt and help are forbidden.

## 3. Mutation confirmation and dry-run

Mutating operations require an explicit confirmation or an explicit
confirmation flag (for example `--yes`); every mutating command owns a
`--dry-run` that shows the plan without mutation. The confirmation request
shows the plan beforehand (dry-run equivalent).

## 4. Documented terminal detection

The interaction mode (for example `auto|always|never`) is documented; in
non-terminal contexts (CI, agents) the tool never hangs in a prompt loop. The
detection rule is part of the tool's contract documentation.

## 5. Accessible output

Where form- or color-based rendering is used, an accessible, line-oriented
output form is available.

Positive example:

```text
tool release cut --version 2.8.0 --dry-run     # shows the plan, mutates nothing
tool release cut --version 2.8.0 --yes         # non-interactive, complete
tool release cut                               # interactive: the select shows the
                                               # same values/rules as the static help
```

Negative example: a required value is reachable only interactively (no flag
exists); or the tool opens a prompt loop in a CI environment and blocks the
run; or the interactive select shows different values than the static help
(two truths, forbidden by the single-source-of-truth law).
