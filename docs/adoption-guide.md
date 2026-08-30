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

For a listed standard license, additionally set `SPDX_LICENSE_IDENTIFIER` to
its exact SPDX identifier so the REUSE text file renders as
`LICENSES/<ID>.txt`:

```json
{
  "PROJECT_NAME": "example-project",
  "LICENSE_ID": "example-project-MIT",
  "COPYRIGHT_YEAR": "2026",
  "CANONICAL_SOURCE_URL": "https://github.com/<org>/example-project",
  "SPDX_LICENSE_IDENTIFIER": "MIT"
}
```

Leave the key unset for custom and unlisted instruments, which keep the
`LICENSES/LicenseRef-<LICENSE_ID>.txt` form. For license texts shared by an
`-only`/`-or-later` pair (GPL, LGPL, AGPL, GFDL), declare the chosen suffix
through this key (for example `GPL-3.0-only`); the template text is shared.

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

This writes `LICENSE` and the REUSE text file (`LICENSES/<ID>.txt` for listed
standard licenses, `LICENSES/LicenseRef-<LICENSE_ID>.txt` for custom and
unlisted instruments). Commit both files plus the values and lock files.

The canonical per-template documentation — family, grant, conditions,
restrictions, patent position, and adoption notes — lives in
[licensing/](licensing/README.md); the operational contracts live in
[infrastructure/](infrastructure/README.md). Multi-license compositions
(recipient choice `OR`, cumulative stacking `AND`, `WITH` exceptions) are
formed from member templates per
[licensing/multi-licensing/README.md](licensing/multi-licensing/README.md).

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

Running this step in tenant CI/CD is not optional: it is a merge-blocking
required check whose binding convention lives in
[docs/tenant-verify-convention.md](tenant-verify-convention.md).

## 4. Updates

Template updates arrive as pull requests that bump `license.lock.json` and
re-render the instance. Adoption is a tenant-controlled decision; no instance
updates automatically. Rollback means pinning the previous version and digest.
