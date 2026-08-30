# AllRightsReserved — Proprietary No-License-Default Notice

Family: C — Proprietary / Closed-Source. Canon:
[license taxonomy canon](../README.md); family reference:
[proprietary README](README.md).

- Template: [`templates/proprietary/all-rights-reserved/AllRightsReserved-1.0.0.hbs`](../../../templates/proprietary/all-rights-reserved/AllRightsReserved-1.0.0.hbs)
- Identifier: `LicenseRef-<LICENSE_ID>` per tenant — this notice is not on the
  SPDX License List and follows the `LicenseRef-` identification discipline.
- Status: the organization-standard no-license-default instrument, drafted
  under the Family D drafting canon and subject to the legal-counsel gate
  before any release.

## Content summary

- **No license grant:** all rights in the software are reserved; no right to
  use, copy, modify, merge, publish, distribute, sublicense, sell, host as a
  service, or create derivative works is granted; platform terms of service
  grant at most limited viewing and forking on that platform.
- **Permissions:** any use beyond mandatory law requires prior written
  permission via the organization permission contact.
- **No warranty / liability:** as-is, with the mandatory EU/DE liability
  floor preserved.
- **Governing law, venue, language:** the organization defaults
  (`GOVERNING_LAW`, `VENUE`); English is the controlling language.
- **Versioning:** fixed version per release ("-only" semantics); no automatic
  upgrade.
- **Machine-readability:** `LicenseRef-<LICENSE_ID>` identification, REUSE
  3.2 placement, and per-file SPDX headers.

## Adoption

Leave `SPDX_LICENSE_IDENTIFIER` unset in `license.values.json` — the instance
renders as `LICENSE` plus `LICENSES/LicenseRef-<LICENSE_ID>.txt`. Choose this
instrument when the project is closed-source and no usage rights are granted
to anyone outside the organization.
