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
  "COPYRIGHT_HOLDER": "t33n Software",
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
| Ecosystem manifest license fields | Nowhere — aligned only by the render | Drift guard rejects hand edits |

## 5. Ecosystem manifest license fields

The declaration seam (`git-governance.quality.json`, field
`toolchain.language`) selects the ecosystem row whose manifest license
fields the render aligns to the lock projection:

| Language | Manifest | Aligned fields |
|----------|----------|----------------|
| `node-typescript` | `package.json` | `license` → `"SEE LICENSE IN LICENSE"` (or the declared SPDX identifier) |
| `python` | `pyproject.toml` | `[project]` `license` → `LicenseRef-<LICENSE_ID>` (or the declared SPDX identifier), `license-files` → `["LICENSE"]` |
| `rust` | `Cargo.toml` | `[package]` `license-file` → `"LICENSE"` (or `[package]` `license` → the declared SPDX identifier) |
| `maven` | `pom.xml` | `licenses`/`license` `name` → the LICENSE_ID form (or the declared SPDX identifier), `url` → the canonical source URL |
| `nuget` | `*.nuspec` / `*.csproj` | `<license type="expression">` → the declared SPDX identifier, `<license type="file">` → `LICENSE` (the custom family); MSBuild: `PackageLicenseExpression` / `PackageLicenseFile` (discovered in the render target directory) |
| `composer` | `composer.json` | `license` → `LicenseRef-<LICENSE_ID>` (or the declared SPDX identifier) |
| `ruby` | `*.gemspec` | `spec.license` → `LicenseRef-<LICENSE_ID>` (or the declared SPDX identifier); the plural `spec.licenses = [<entry>]` spelling is the value-equal alternative (discovered in the render target directory) |
| `elixir` | `mix.exs` | `package` keyword list `licenses` → `["LicenseRef-<LICENSE_ID>"]` (or the declared SPDX identifier); the insertion anchors at the package keyword list |
| `haskell` | `*.cabal` | top-level `license` → `LicenseRef-<LICENSE_ID>` (or the declared SPDX identifier), `license-file` → `LICENSE`; the `license-files` list spelling with exactly `LICENSE` is the value-equal alternative (discovered in the render target directory) |

Manifest license fields are never hand-edited: the verify lane fails closed
on diverging fields, and the render is the sanctioned writer. Deprecated
declaration forms — the PEP 621 `license` table subkeys, the `License ::`
classifier entries, the nuspec `licenseUrl` element, and the MSBuild
`PackageLicenseUrl` property — are reported as findings whose remediation
is an explicit tenant decision. The cargo license keys are mutually
exclusive (`license` in lieu of `license-file`); a manifest that declares
both keys is a fail-closed finding whose resolution is an explicit tenant
decision. A pom that declares multiple `license` elements is the same
class of finding, a composer license array that declares multiple
expressions is the same class of finding, and so is a gemspec license
array that declares multiple entries or a gemspec that declares the
license assignment more than once. A mix.exs whose `licenses` entry
declares multiple identifiers, whose package configuration or licenses
entry is declared more than once, or whose licenses entry lives outside
the package keyword list is the same class of finding. A .cabal that
declares multiple license file entries, that declares the license field or
the file surface more than once, or that carries a license surface inside
a component section instead of the top-level package description is the
same class of finding.
