# License Hub Policy

This document is the binding policy for template families, versioning, and
tenant adoption of the license hub.

## 1. Family assignment

Every template belongs to exactly one family directory (see
`templates/README.md`). Source-available and custom licenses are never
labeled as OSI Open Source. A modified standard license always receives a new
name and a `LicenseRef-` identifier; it never keeps the standard name.

## 2. Legal SemVer for template releases

Template families are versioned independently with semantic versioning:

| Bump | Meaning | Tenant impact |
|------|---------|---------------|
| Patch (1.0.0 → 1.0.1) | Editorial only; no meaning change | Adoption recommended, risk-free |
| Minor (1.0.0 → 1.1.0) | Clarification or extended permission without tightened restriction | Adoption via pull request; changelog note required |
| Major (1.x → 2.0.0) | Grant or restriction changes materially | Active per-tenant adoption decision; `-only` binding prevents silent upgrades |

Every template release requires legal-counsel review before publication.

## 3. Tenant adoption

1. The tenant carries `license.values.json` (project facts) and
   `license.lock.json` (template path, version, SHA-256 digest).
2. Instances are produced only by rendering the pinned template with the
   `license` CLI (`render`) — never by hand-editing rendered files.
3. The tenant CI runs `license verify` as a merge-blocking drift guard.
4. Template updates propagate as tenant-controlled pull requests; no tenant
   updates automatically.

## 4. Placeholder contract

Templates use Handlebars-style anchors (`{{KEY}}`, uppercase snake case).
The closed anchor set is:

| Anchor | Source of truth |
|--------|-----------------|
| `PROJECT_NAME`, `LICENSE_ID`, `COPYRIGHT_YEAR`, `CANONICAL_SOURCE_URL` | Tenant `license.values.json` |
| `COPYRIGHT_HOLDER`, `GOVERNING_LAW`, `VENUE`, `PERMISSION_CONTACT` | Hub `org-defaults.json` |

A rendered instance with an unresolved anchor is invalid; the render gate
rejects it (`grep -RE '\{\{[A-Z0-9_]+\}\}'` must return zero matches).

## 5. Ownership

The `templates/` tree is legal-governance content. Changes require the review
boundary recorded in `.github/CODEOWNERS` and the legal-counsel gate.
