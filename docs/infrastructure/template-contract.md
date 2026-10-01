# Template Contract

This document defines the binding contract of every license template in
`templates/`: placement, naming, the placeholder system, the render output
layout, versioning, and the release and review gates.

## 1. Placement and naming

1. Every template belongs to exactly one family directory under `templates/`;
   the family taxonomy and its content rules are owned by
   [../../templates/README.md](../../templates/README.md).
2. Each template occupies exactly one versioned directory per license:
   `templates/<family>/<template-kebab>/`.
3. The template file is versioned by file name: `<Name>-<semver>.hbs`, where
   `<Name>` is the exact SPDX identifier for listed licenses (for example
   `Apache-2.0-1.0.0.hbs`) or the PascalCase instrument name for unlisted
   instruments (for example `NoRepublish-1.0.0.hbs`).
4. Every template directory carries a `CHANGELOG.md` recording the legal
   SemVer history and a `README.md` re-referencing the canonical template
   documentation under `../licensing/`.
5. Template bodies never carry comments: rendering is pure placeholder
   substitution, so any comment would be emitted into the rendered legal text
   of every tenant. The sibling `README.md` is the documentation seam.

## 2. Placeholder system

Templates use Handlebars-style anchors (`{{KEY}}`, uppercase snake case).
Rendering is pure substitution; there is no template logic, no conditionals,
and no loops.

The closed required anchor set is:

| Anchor | Source of truth |
|--------|-----------------|
| `PROJECT_NAME`, `LICENSE_ID`, `COPYRIGHT_YEAR`, `CANONICAL_SOURCE_URL` | Tenant `license.values.json` |
| `COPYRIGHT_HOLDER`, `GOVERNING_LAW`, `VENUE`, `PERMISSION_CONTACT` | Hub `org-defaults.json` |

One optional anchor exists:

| Anchor | Source of truth | Effect |
|--------|-----------------|--------|
| `SPDX_LICENSE_IDENTIFIER` | Tenant `license.values.json` | Selects the listed-license output stem (Section 3) |

A rendered instance with an unresolved anchor is invalid; the render gate
rejects it (the placeholder pattern `\{\{[A-Z0-9_]+\}\}` must return zero
matches on the rendered output).

## 3. Render output layout

A render produces exactly two files with identical content:

1. `LICENSE` at the repository root — the detection anchor.
2. The full text in the REUSE license area:
   - with `SPDX_LICENSE_IDENTIFIER` set (listed standard licenses):
     `LICENSES/<SPDX_LICENSE_IDENTIFIER>.txt`;
   - without it (custom and unlisted instruments):
     `LICENSES/LicenseRef-<LICENSE_ID>.txt`.

Custom licenses are never on the SPDX License List and must be identified with
the `LicenseRef-<idstring>` syntax; a modified standard license always
receives a new name and a `LicenseRef-` identifier and never keeps the
standard name. Listed standard licenses keep their exact SPDX identifier,
including `-only` versus `-or-later` discipline for the GPL family; the
deprecated `+` operator is never used.

Per-file source headers (`SPDX-License-Identifier` and
`SPDX-FileCopyrightText`) follow the machine-readability section of the
template and the REUSE 3.2 specification.

## 4. Legal SemVer

Template families are versioned independently with semantic versioning:

| Bump | Meaning | Tenant impact |
|------|---------|---------------|
| Patch (1.0.0 → 1.0.1) | Editorial only; no meaning change | Adoption recommended, risk-free |
| Minor (1.0.0 → 1.1.0) | Clarification or extended permission without tightened restriction | Adoption via pull request; changelog note required |
| Major (1.x → 2.0.0) | Grant or restriction changes materially | Active per-tenant adoption decision; `-only` binding prevents silent upgrades |

Every template release requires legal-counsel review before publication.

## 5. Immutability and pinning

Template families are published as versioned, immutable release artifacts —
not as movable tags alone. Every release contains the template files, a
checksum file, and a provenance attestation. Tenants pin version plus SHA-256
digest in their lock file; the verify lane checks the reference against the
pinned digest. Rollback is trivial: pin the previous version and digest.

## 6. Change discipline

Templates are edited only in this repository, through the governed ticket
workflow, with the legal review boundary recorded in
[../../.github/CODEOWNERS](../../.github/CODEOWNERS). Hand edits to rendered
instances in tenants are rejected by the drift guard; corrections belong into
the template (a new template version) or into the tenant values.
