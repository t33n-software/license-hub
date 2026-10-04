# Ecosystem License Metadata Specification

The license hub owns the render and verify machinery of canonical license
instances. Beyond the LICENSE file family, package ecosystems declare
license metadata on their own manifest surfaces, and this specification
defines how the hub derives, aligns, and proves those surfaces. It binds the
derivation law, the npm adapter contract, the fail-closed verification
classes, and the surgical alignment discipline.

## 1. Derivation law

The declaration is the truth; the manifest field is its deterministic
projection; the license lock is the single source of truth of the license
type.

1. **Declaration.** The project's ecosystem is contractually declared by the
   governed seam document (`git-governance.quality.json`, field
   `toolchain.language`); the organization fleet inventory mirrors the same
   declaration. The declaration is consumed read-only — never copied and
   never re-declared.
2. **Matrix row.** The declared language selects exactly one ecosystem row.
   The implemented row set: `node-typescript` selects the npm surface
   (`package.json`, field `license`); `python` selects the pyproject.toml
   surface (PEP 621/639: field `license` as the SPDX expression string and
   field `license-files` as the glob list); Go selects no manifest field —
   the `LICENSE` file at the module root is the truth surface, and the
   adapter aligns nothing. Additional rows are grown content-driven when a
   consuming surface exists, with the field semantics verified against the
   official ecosystem documentation before the row is implemented; unverified
   semantics fail closed instead of guessing.
3. **Projection.** The expected field value is derived from the license lock
   (template, version, digest) together with the tenant values: a declared
   SPDX identifier projects itself; the file-based custom family projects
   the bound target form. The manifest field is never an independent source,
   and no process derives the lock from the manifest.

## 2. The npm alignment contract

The legal npm `license` forms are a valid SPDX license expression, the
custom-file reference form `SEE LICENSE IN <file>`, and `UNLICENSED`. The
bound organization target form for the file-based custom family is the
string `"SEE LICENSE IN LICENSE"` — the npm-canonical reference that points
the registry and every scanner at the committed `LICENSE` file at the
repository root. Declaring an SPDX `LicenseRef-...` identifier directly in
the field is rejected for the private custom family: the family is declared
by the referenced file, and the file reference keeps the registry
declaration and the committed text provably identical.

The npm manifest carries two other identity surfaces beside the license
field: `name` and `author`. Both belong to the package identity lane and are
never read or written by the license render.

## 3. The Python alignment contract

The Python row targets `pyproject.toml` under PEP 621/639. The bound
organization target form for the file-based custom family is the SPDX
sideload expression `LicenseRef-<LICENSE_ID>` (the tenant `LICENSE_ID`
value), paired with the license text glob list `["LICENSE"]` in the
`license-files` field; a declared `SPDX_LICENSE_IDENTIFIER` projects itself
into both surfaces.

The adapter's scanner reads the `[project]` table strictly and lexes the
remaining document structurally. A structural break anywhere — an
unterminated string, a malformed or duplicated header, a duplicate key, a
value the scanner cannot interpret — refuses the whole surface fail-closed.
The dotted-key forms that would create or shadow the license surface
outside the canonical table (a dotted key in `[project]` whose first
segment names a license-relevant key, and any root-level declaration of the
project name) are refused for the same reason.

The deprecated PEP 621 `license` table subkeys (`text`/`file`) are the same
license field in its deprecated value form: the render aligns them to the
string expression form. The `License ::` classifier entries are the
deprecated license declaration form: the verify lane reports them as a
finding whose remediation is an explicit tenant decision, and the alignment
never rewrites the classifiers array.

## 4. Render alignment discipline

The render is the sanctioned writer of tenant license surfaces, and the same
governed act aligns the ecosystem license fields. Four disciplines bind
every alignment:

- **Declaration-anchored:** the target value is derived from the lock and
  the matrix row, never from a scan of the manifest.
- **Surgical:** only the license projection surface of the manifest is
  written; every other byte — including `name`, `author`, ordering,
  formatting, and unknown fields — is preserved exactly.
- **Idempotent:** a re-run over an aligned manifest is a no-op and reports
  zero deltas; a missing member is inserted deterministically; drift reports
  name exactly the diverging surfaces.
- **Single-owner:** the alignment lives exclusively in the license hub CLI
  (`cmd/license`, commands `render` and `verify`); no provisioning tool, CI
  step, or script may reimplement it.

The render skips the alignment with an explicit report when the declared
manifest is absent, and the dry-run plan previews the alignment without
writing. A non-string license member or an unscannable manifest is refused
fail-closed instead of being rewritten.

## 5. Verification finding classes

Verification is fail-closed in both directions and yields exactly three
finding classes:

1. **Declared, manifest missing or diverging** — the declared ecosystem's
   manifest lacks the license field or carries a diverging value: the
   finding names the observed value, the projected target from the lock, and
   the remediation (run the render).
2. **Manifest without declaration (shadow surface)** — a manifest license
   surface exists where no covering declaration exists: the finding names
   the declaration state; the shadow surface is either removed or the
   declaration is completed, by explicit decision.
3. **Verify proof failure** — the verify lane cannot prove the surface
   (unscannable manifest, unreadable declaration): error; the surface stays
   unproven and merge-blocking controls fail closed.

A diverging field is never adopted from the manifest into the lock; the
correction direction is always lock → manifest through the render.

## 6. Adapter set and growth rule

The npm and Python adapters are implemented. Every additional ecosystem
adapter is born content-driven when a consuming project surface exists, and
its matrix row records either the canonical field semantics (verified
against the official ecosystem documentation at specification time) or the
explicit verdict that the ecosystem carries no manifest field — for those
rows the file family is the declared truth and the adapter aligns nothing.

## 7. Do / Don't

**Do:** derive every expected field value from the lock through the matrix
row; prove declaration, manifest, and field fail-closed in both directions;
align only through the sanctioned render, surgically and idempotently;
report every skip and finding with the observed and target values.

**Don't:** hand-edit a manifest license field outside the render; write the
lock from the manifest or adopt a diverging field value as truth; run a
verify-only process without the sanctioned writer; touch `name`, `author`,
or any non-license field during an alignment; provision license and language
families in one mega-tool; invent field semantics without verification.