# Render and Verify

This document defines the `license` CLI: its commands, its determinism
contract, and the proofs its verification executes. The organization-wide CLI
conventions this tool follows are owned by
[../conventions/cli/README.md](../conventions/cli/README.md); this document
binds them to the concrete command surface and never restates them.

The CLI is a zero-dependency Go command-line tool built entirely on the
standard library. It runs on Linux, macOS, and Windows without any runtime
assumption and produces deterministically identical bytes on every host.

## 1. Commands

Every command owns a complete help area: `license <command> --help` renders
the full input contract — usage, every flag with its value domain, canonical
examples, and the exit behavior — generated from the command registry. The
root answers `--help` (or `-h`, or the `help` command) with the command
families and the global flags.

### `license version`

Prints the binary version, commit, and build date. The root additionally
answers `--version`; with `--output json` the version record is a
machine-readable JSON document.

### `license digest --template <path>`

Prints the canonical `sha256:<hex>` digest of a template file. Use it to
compute or cross-check the digest pinned in `license.lock.json`.

### `license render`

```bash
license render \
  --template templates/<family>/<template>/<Name>-<semver>.hbs \
  --org-defaults org-defaults.json \
  --values license.values.json \
  --out . \
  --yes
```

Renders the pinned template with the organization defaults and the tenant
values and writes the instance: `LICENSE` plus the REUSE text file
(`LICENSES/<SPDX_LICENSE_IDENTIFIER>.txt` when the values declare a listed
identifier, otherwise `LICENSES/LicenseRef-<LICENSE_ID>.txt`). The render
fails closed when a required value is missing or a placeholder anchor remains
unresolved. The command prints the written paths and the template digest.

`render` is the mutating command of this tool. It requires the explicit
confirmation flag `--yes` in non-interactive contexts and asks for
confirmation on an interactive terminal; without confirmation nothing is
written. `--dry-run` prints the plan — the target paths and the template
digest — without writing any file. A repeated render is idempotent: it
rewrites the same canonical bytes.

### `license verify`

```bash
license verify \
  --template templates/<family>/<template>/<Name>-<semver>.hbs \
  --org-defaults org-defaults.json \
  --values license.values.json \
  --lock license.lock.json \
  --dir .
```

Re-renders the canonical template with the tenant values and the organization
defaults and proves three properties fail closed:

1. **Pin integrity** — the template bytes still match the digest pinned in
   `license.lock.json` (when `--lock` is given).
2. **Drift freedom** — the committed `LICENSE` and `LICENSES/` instance equal
   the canonical render byte for byte; hand edits are detected.
3. **Completeness** — no unresolved `{{PLACEHOLDER}}` anchor remains.

Every violation is reported; the command exits with the governance-rejection
code when any violation exists.

## 2. Output contract

Every command accepts the global flags:

- `--output human|json` (default `human`): the machine-readable form is one
  JSON document per call, versioned as `schemaVersion: 1` and evolving
  additively only.
- `--quiet`: suppresses the successful human output; errors are never
  suppressed.

In the human form, successes print to stdout and failures to stderr. In the
JSON form every document prints to stdout. Errors are structured records with
a stable code, the affected field, the actual and the expected value, the
violated rule, a valid example, and the remediation; the human and the
machine form carry identical information.

## 3. Exit codes

| Code | Class |
|------|-------|
| 0 | success |
| 1 | execution failure (for example an unreadable input file) |
| 2 | usage error (invalid or missing input, missing confirmation) |
| 3 | governance rejection (`verify` found violations) |

## 4. Environment mapping

Every flag owns an environment variable; the precedence order is
`flag > environment variable > default`:

| Flag | Environment |
|------|-------------|
| `--template` | `LICENSE_TEMPLATE` |
| `--org-defaults` | `LICENSE_ORG_DEFAULTS` |
| `--values` | `LICENSE_VALUES` |
| `--out` | `LICENSE_OUT` |
| `--lock` | `LICENSE_LOCK` |
| `--dir` | `LICENSE_DIR` |
| `--output` | `LICENSE_OUTPUT` |
| `--quiet` | `LICENSE_QUIET` |
| `--dry-run` | `LICENSE_DRY_RUN` |
| `--yes` | `LICENSE_YES` |

The CLI owns no configuration file: every input is already an explicit file,
so the organization-wide precedence order reduces to flag, environment,
default.

## 5. Offline capability and timeouts

Every command works offline. The tool starts no external processes and
performs no network access, so bounded configurable timeouts are not
applicable.

## 6. Determinism and instance discipline

A license instance is a generated, committed artifact. Generation happens at
maintenance time (initial adoption or template update), never at tenant build
time, because the license text must exist in the source tree itself: hosting
detection, release archives, and clones read the tree, not build outputs.

Instances are re-rendered, never hand-edited. The committed instance must
equal the render output of the pinned template plus values; corrections belong
into the template (a new template version) or into the values. Editing a
rendered instance by hand to silence the guard is a policy violation.

## 7. Consumption contract

In CI the CLI is consumed exclusively as a digest-pinned, signature- and
attestation-verified release binary of this hub, resolved through the
consuming repository's pinned tooling module. Source-based execution in tenant
CI — a hub checkout or `go run` — is not permitted. Local authoring with the
CLI (render, verify, digest) is the development path and does not replace the
CI gate.
