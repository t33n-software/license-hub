# Tenant and Hub Control Files

This document defines every control file of the licensing architecture: what
it carries, where it lives, and who edits it.

## 1. `license.values.json` — tenant project facts

Lives in the tenant repository root and is edited in the tenant. It carries
the project facts of the adopting repository:

```json
{
  "PROJECT_NAME": "example-project",
  "LICENSE_ID": "example-project-NoRepublish-1.0",
  "COPYRIGHT_YEAR": "2026",
  "CANONICAL_SOURCE_URL": "https://github.com/<org>/example-project"
}
```

Required keys: `PROJECT_NAME`, `LICENSE_ID`, `COPYRIGHT_YEAR`,
`CANONICAL_SOURCE_URL`.

Optional key: `SPDX_LICENSE_IDENTIFIER` — set it to the exact SPDX identifier
of a listed standard license (for example `"MIT"` or `"Apache-2.0"`) so the
rendered REUSE text file is emitted as `LICENSES/<ID>.txt`. Leave it unset for
custom and unlisted instruments, which keep the
`LICENSES/LicenseRef-<LICENSE_ID>.txt` form.

## 2. `license.lock.json` — tenant template pin

Lives in the tenant repository root and is never hand-edited; it changes only
through a tenant pull request that adopts a template update. It pins the
template reference by path, version, and SHA-256 digest:

```json
{
  "template": "templates/custom/norepublish/NoRepublish-1.0.0.hbs",
  "version": "1.0.0",
  "digest": "sha256:<digest-from-release-SHA256SUMS>"
}
```

The digest is the SHA-256 of the template file as published in the immutable
template release. The verify lane fails closed when the template bytes no
longer match the pinned digest.

## 3. `org-defaults.json` — hub organization constants

Lives in this repository and is edited only here, under the legal review
boundary. It carries the organization constants injected into every render:

```json
{
  "COPYRIGHT_HOLDER": "CyberT33N",
  "GOVERNING_LAW": "the Federal Republic of Germany",
  "VENUE": "Germany",
  "PERMISSION_CONTACT": "https://github.com/t33n-software"
}
```

These constants are public by nature — they appear in every rendered license —
and therefore never belong in a secret store.

## 4. Edit-location discipline

| File | Edit location | Gate |
|------|---------------|------|
| Templates | Hub only | Governed ticket workflow plus legal review |
| `org-defaults.json` | Hub only | Legal review |
| `license.values.json` | Tenant | Tenant pull request |
| `license.lock.json` | Tenant, via adoption pull request | Verify lane digest proof |
| Rendered instances (`LICENSE`, `LICENSES/`) | Nowhere — regenerated only | Drift guard rejects hand edits |
