# CLI Convention: Value Domains
[INTENT: SPECIFICATION]

This document is the binding convention for the value-domain model of every
CLI tool of the organization: every flag and every positional argument is
assigned to exactly one value class, and the class determines the binding
help duty. How the value domains are sourced and projected into every channel
is owned by [single-source-of-truth.md](single-source-of-truth.md) and is not
restated here.

## 1. The eight-class model

Every flag and every positional argument of every CLI tool is assigned to
**exactly one** of the following eight value classes. The class bindingly
determines what the help must show (Sections 2 to 6) and what the remaining
channels project. A flag without a class assignment is an architecture
defect, because its help duty is then undefined and the drift control cannot
verify it.

| # | Class | Definition | Help duty (short form) |
|---|-------|-----------|------------------------|
| 1 | `closed-enum` | Finite, fixed value set | Complete list of the values accepted by this endpoint (Section 2) |
| 2 | `shaped` | Grammar template with a fixed skeleton or prefix | Grammar + canonical example + subset rules (Section 3) |
| 3 | `free-constrained` | Free text with validation rules | Compact rule set including the forbidden and an example (Section 4) |
| 4 | `structural-reference` | Path, reference, or identifier with runtime-dependent validity | Form + resolution rule, no full-prevention promise (Section 5) |
| 5 | `scalar-bounded` | Number, duration, size | Type, unit, value range, default (Section 6) |
| 6 | `boolean-switch` | Switch without a value | Effect + default (+ negation form, if present) (Section 6) |
| 7 | `composite-token` | Repeatable `TOKEN=VALUE` form | Transport form + token grammar + example (Section 6) |
| 8 | `secret-reference` | Reference to a secret | Reference forms only (environment, file, broker); never value acceptance via argument (Section 6) |

Positive example — the class assignment of a command definition
(illustrative):

```text
--strategy  -> closed-enum          (values: check, auto, rebase, merge)
--version   -> shaped               (grammar: <major>.<minor>.<patch>)
--message   -> free-constrained     (rules: length, character class, forbidden)
--config    -> structural-reference (form: path; resolution: existence at runtime)
--timeout   -> scalar-bounded       (positive duration, default 30s)
--dry-run   -> boolean-switch       (effect: plan without mutation, default false)
```

Negative example: a new flag is registered without a class assignment; as a
consequence its help text carries neither values nor rules. This violates the
class duty: without a class it is undecidable which help duty applies, and
the drift control cannot verify the flag.

## 2. Class `closed-enum` — the complete endpoint-specific value list

When a flag owns a finite, fixed value set, the help **must** show **100 % of
the values accepted by this endpoint**. Two precisions are binding:

1. **Endpoint reference over total domain.** What is shown is the set that
   *this* endpoint accepts. If an endpoint accepts only a subset of a larger
   domain, the help shows the subset (or names list plus restriction). A help
   that lists values the endpoint rejects produces misbehavior itself; a help
   that omits accepted values does so as well.
2. **Derivation over literal.** The list is rendered at registration time
   from the canonical value source, never duplicated as hand-maintained text.

Positive example:

```text
--strategy string   check, auto, rebase, or merge (default "check")
--kind string       blocker, docs, or release-prep
```

Both texts name the complete accepted set of their endpoint; two endpoints
with the same flag name but different domains each show their own subset.

Negative example:

```text
--type string   change family
```

The value set is fixed, but the help names not a single value: the consumer
must guess and learns the domain only through an error message — exactly the
state the core principle forbids. Equally forbidden: a list containing values
this endpoint rejects (superset), or a list that withholds accepted values
(subset).

## 3. Class `shaped` — grammar, example, subsets

When a flag owns a grammar form with a fixed skeleton (fixed prefix,
placeholder segments, versioned form), the help **must** show the grammar
itself, a canonical example, and any subset rules of this endpoint (which
prefixes or forms this endpoint accepts or excludes).

Positive example:

```text
--affected-line string   main, release/<semver>, or support/<major.minor>;
                         example: release/2.8.0
--target-line string     declared develop, release/<semver>, or
                         support/<major.minor> target; example: support/2.7
```

The second text also shows the endpoint-specific subset: it lists the forms
this endpoint accepts — and omits the form it rejects.

Negative example:

```text
--release string   the release line
```

The grammar (`release/<semver>`, SemVer without a leading `v`) stays
invisible; inputs such as `v2.8.0`, `2.8`, or `release/2.8` fail only at
validation, although the form is completely statically known.

## 4. Class `free-constrained` — the compact rule set with allowed/forbidden

When a flag accepts free text whose validation can fail, the help **must**
proactively show the compact rule set so that misbehavior is prevented before
the failed validation. The rule set comprises, as far as applicable to the
flag:

- allowed characters / character class;
- disallowed characters and forbidden content;
- length or size limits;
- the governing naming convention;
- the applied grammar or regex rule;
- one canonical valid example;
- the primary conditions under which validation fails.

The exhaustive grammar remains the property of the referenced specification
(help level 3, see [../help/README.md](../help/README.md)); the help carries
the operationally decisive rule set.

**Labeling law:** the help bindingly distinguishes between "**rejected by
validation**" (machine rule, hard rejection) and "**convention-violating**"
(content contract that is not machine-enforced). Without this distinction the
consumer learns false expectations about the rejection behavior.

Positive example:

```text
--slug string   1-100 lowercase ASCII letters or digits, words joined by
                single hyphens (rejected unless matching
                ^[a-z0-9]+(?:-[a-z0-9]+)*$); example: add-export-button
```

Negative example:

```text
--slug string   branch description
```

The flag validates against a character class, a length, and a naming
convention — the help names none of them; every rule violation is discovered
only by error. Equally forbidden: a rule text that promises a hard rejection
the validation does not execute (untrue contract), or one that labels a
convention rule as a validation rejection.

## 5. Class `structural-reference` — form and resolution rule without a full-prevention promise

When a flag references paths, refs, or identifiers whose validity is decidable
only at runtime (existence, registration, state), the help **must** name the
**form** and the **resolution rule** — and **must not** pretend it could
fully pre-empt validity. A help that promises an exhaustive list of valid
instances it cannot guarantee is an untrue contract and forbidden.

Positive example:

```text
--base string   canonical branch name or <remote>/<branch> on the selected
                remote; existence is resolved at runtime; example:
                origin/develop
--record string repository-relative record path; defaults to the ticket
                record path
```

Negative example:

```text
--base string   one of: develop, main, release/2.8.0, ...
```

An instance-based list in the help is unprovable at write time and possibly
stale at read time — it promises a full prevention that only the runtime can
deliver. The duty of this class is form plus resolution, not an instance
list.

## 6. Classes `scalar-bounded`, `boolean-switch`, `composite-token`, `secret-reference`

Four compact classes with their own help duties:

1. **`scalar-bounded`:** the help shows type, unit, value range, and default.
2. **`boolean-switch`:** the help shows the effect of the set switch and the
   default; if a negation form exists, it is named.
3. **`composite-token`:** repeatable key-value flags show the transport form
   (for example `TOKEN=VALUE`), the token grammar (allowed token characters,
   reserved tokens), the repeatability, and an example. If the transport form
   of the CLI deviates from the grammar of the target artifact (for example
   `=` on the CLI versus `: ` in the artifact), exactly this mapping is shown.
4. **`secret-reference`:** secrets are **never** accepted as a value via
   argument (process-list and history exposure). The help shows only the
   allowed reference forms (environment variable, file, broker) — a
   value-taking secret argument is per se a contract breach.

Positive example:

```text
--timeout duration   positive duration for external processes (default 30s)
--push               validate and push after committing (default false)
--footer strings     footer as TOKEN=VALUE; token: letters, digits, hyphens;
                     repeatable; example: --footer Refs=#123
--token-env string   environment variable that carries the access token
                     (default TOOL_TOKEN)
```

Negative example:

```text
--timeout duration   timeout
--footer strings     footers
--token string       the access token
```

The first text withholds range and default; the second withholds transport
form and token grammar; the third accepts a secret as a plaintext argument
and thereby violates the secret law independent of any help text.
