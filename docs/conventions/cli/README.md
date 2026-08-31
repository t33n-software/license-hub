# CLI Conventions
[INTENT: REFERENCE]

This directory is the canonical home of the organization-wide infrastructure
conventions for command-line tools (CLI tools). These conventions apply to
**every** CLI tool of the organization — present and future — independent of
programming language, subject matter, and project. A deviation in an
individual case is an explicitly documented, reasoned exception, never a
silent one.

## The core principle: discoverable-closed

The help of a CLI tool is the only discovery channel that by construction has
three properties at once: **zero side effects** (pure reading, no mutation),
**zero network dependency** (fully offline), and **zero version drift** (the
help is produced by the same installed binary that executes the call). No
other channel — documentation website, wiki, chat, error message — guarantees
all three, so the help is the authoritative contract surface of the tool.

The architectural law follows from this: **everything the validation knows
must be discoverable before the call.** A tool that reveals its value domains
only in error messages forces every consumer — human and LLM agent alike —
into a costly guess-and-error cycle. Fail-closed (the early, hard rejection of
invalid input with named remediation) remains mandatory, but it is only half
of the architecture; the target architecture is **discoverable-closed**: the
complete input domain is knowable before the first call.

Consumer equality: the help is read equally by humans in terminals and by
LLM agents working under a help-first contract (read the current help before
every invocation and derive inputs from it). Both consumer classes must be
able to derive a valid call from the help alone.

## Convention domains

Every subdomain owns its convention in exactly one document:

| Domain | Convention content | Owning document |
|---|---|---|
| Help | The universal help contract on every command node, the help level model, and the generation mandate | [help/README.md](help/README.md) |
| Value domains | The eight-class value-domain model and the binding help duty per class | [values/README.md](values/README.md) |
| Value-domain sourcing | The single-source-of-truth register and its five rendering channels; drift blocks the release | [values/single-source-of-truth.md](values/single-source-of-truth.md) |
| Output | The stable output contract, structured coded error records, semantic exit codes, `--quiet`, and the secret-free output rule | [output/README.md](output/README.md) |
| Interaction | Non-interactive completeness, same-source prompts, mutation confirmation, `--dry-run`, terminal detection, accessible output | [interaction/README.md](interaction/README.md) |
| Errors | The error philosophy: fail-closed at the earliest boundary, truthful remediation, no partial success as success, missing evidence is never a pass | [errors/README.md](errors/README.md) |
| Configuration | The organization-wide precedence order, environment mapping, secret handling, bounded timeouts, and the offline declaration | [configuration/README.md](configuration/README.md) |
| Security | Read-only default, documented idempotency, audit records, least privilege, and signed, attested distribution | [security/README.md](security/README.md) |
| Lifecycle | Semantic versioning on the CLI contract, the deprecation pipeline, stability levels, and additive schema evolution | [lifecycle/README.md](lifecycle/README.md) |
| Testing | The test and verification law: help contract tests, property-based domain tests, the help-first consumer test, and drift blocking CI | [testing/README.md](testing/README.md) |
| Distribution | Dependency-free artifact form, cross-platform parity, documented lifecycle, and telemetry rules | [distribution/README.md](distribution/README.md) |
| Identity | One stable binary name, the machine-readable root version, the grouping law of the command tree, and the visibility rule | [identity/README.md](identity/README.md) |

## Overarching rules

1. **Scope.** These conventions bind every CLI tool of the organization, in
   every programming language and every subject domain.
2. **Agnosticism mandate.** No rule in these documents is phrased or reasoned
   bound to a programming language, a framework, a subject domain, or a single
   project.
3. **Single source of truth.** No value domain, rule text, or value list is
   maintained independently in two places; reference replaces copy across
   document, code, and channel boundaries.
4. **Drift law.** Every machine-readable projection of a canonical rule is
   contract-tested against its source; drift is a release-blocking defect, not
   a warning.
5. **Truth law of help.** Help texts, examples, and remediation references
   must be and stay true: they show only values that are accepted, promise
   only rejections that occur, and reference only surfaces that carry the
   answer.
6. **Secret law.** Secrets are never accepted as CLI arguments and never shown
   in outputs, errors, logs, or audit records.
7. **Domain anchoring.** Every leaf command is its own responsibility boundary
   (a use-case boundary) and therefore owns a complete help area; value
   domains are domain knowledge and live in the domain or policy layer, not in
   the presentation layer.
8. **Consumer equality.** Every rule applies equally to human and agentic
   consumers; a surface servable only by humans (for example a required value
   reachable only interactively) violates the interaction convention.
