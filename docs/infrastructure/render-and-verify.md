# Render and Verify

This document defines the `license` CLI: its commands, its determinism
contract, and the proofs its verification executes.

The CLI is a zero-dependency Go command-line tool built entirely on the
standard library. It runs on Linux, macOS, and Windows without any runtime
assumption and produces deterministically identical bytes on every host.

## 1. Commands

### `license version`

Prints the binary version, commit, and build date.

### `license digest --template <path>`

Prints the canonical `sha256:<hex>` digest of a template file. Use it to
compute or cross-check the digest pinned in `license.lock.json`.

### `license render`

```bash
license render \
  --template templates/<family>/<template>/<Name>-<semver>.hbs \
  --org-defaults org-defaults.json \
  --values license.values.json \
  --out .
```

Renders the pinned template with the organization defaults and the tenant
values and writes the instance: `LICENSE` plus the REUSE text file
(`LICENSES/<SPDX_LICENSE_IDENTIFIER>.txt` when the values declare a listed
identifier, otherwise `LICENSES/LicenseRef-<LICENSE_ID>.txt`). The render
fails closed when a required value is missing or a placeholder anchor remains
unresolved. The command prints the written paths and the template digest.

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

Every violation is reported; the command exits non-zero when any violation
exists.

## 2. Determinism and instance discipline

A license instance is a generated, committed artifact. Generation happens at
maintenance time (initial adoption or template update), never at tenant build
time, because the license text must exist in the source tree itself: hosting
detection, release archives, and clones read the tree, not build outputs.

Instances are re-rendered, never hand-edited. The committed instance must
equal the render output of the pinned template plus values; corrections belong
into the template (a new template version) or into the values. Editing a
rendered instance by hand to silence the guard is a policy violation.

## 3. Consumption contract

In CI the CLI is consumed exclusively as a digest-pinned, signature- and
attestation-verified release binary of this hub, resolved through the
consuming repository's pinned tooling module. Source-based execution in tenant
CI — a hub checkout or `go run` — is not permitted. Local authoring with the
CLI (render, verify, digest) is the development path and does not replace the
CI gate.
