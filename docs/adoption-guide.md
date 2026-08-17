# Tenant adoption guide

This guide turns a repository into a governed tenant of the license hub.

## 1. Values and lock

Create `license.values.json` with the project facts:

```json
{
  "PROJECT_NAME": "example-project",
  "LICENSE_ID": "example-project-NoRepublish-1.0",
  "COPYRIGHT_YEAR": "2026",
  "CANONICAL_SOURCE_URL": "https://github.com/<org>/example-project"
}
```

Create `license.lock.json` pinning the template release. The digest is the
SHA-256 of the template file as published in the immutable template release:

```json
{
  "template": "templates/custom/norepublish/NoRepublish-1.0.0.hbs",
  "version": "1.0.0",
  "digest": "sha256:<digest-from-release-SHA256SUMS>"
}
```

## 2. Render

```bash
license render \
  --template templates/custom/norepublish/NoRepublish-1.0.0.hbs \
  --org-defaults org-defaults.json \
  --values license.values.json \
  --out .
```

This writes `LICENSE` and `LICENSES/LicenseRef-<LICENSE_ID>.txt`. Commit both
files plus the values and lock files.

## 3. Verify (drift guard)

```bash
license verify \
  --template templates/custom/norepublish/NoRepublish-1.0.0.hbs \
  --org-defaults org-defaults.json \
  --values license.values.json \
  --lock license.lock.json \
  --dir .
```

Verification fails when the committed instance drifts from the canonical
render, when the template digest no longer matches the pin, or when a
placeholder remains unresolved.

## 4. Updates

Template updates arrive as pull requests that bump `license.lock.json` and
re-render the instance. Adoption is a tenant-controlled decision; no instance
updates automatically. Rollback means pinning the previous version and digest.
