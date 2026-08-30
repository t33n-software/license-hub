# PolyForm Modular License Family

Family: B — Source-Available, group B3 (Add-On and Modular Restriction).
**Not OSI-approved open source.** Canon:
[license taxonomy canon](../README.md); family reference:
[source-available README](README.md).

The PolyForm project publishes standardized modular source-available texts.
All seven variants are templated; the variant choice is the tenant's
restriction-axis decision. Two variants are on the SPDX License List; the
other five carry the steward text from the PolyForm project (marked below).

## PolyForm-Noncommercial-1.0.0 (`PolyForm-Noncommercial-1.0.0`)

- Template: [`templates/source-available/polyform-noncommercial-1.0.0/PolyForm-Noncommercial-1.0.0-1.0.0.hbs`](../../../templates/source-available/polyform-noncommercial-1.0.0/PolyForm-Noncommercial-1.0.0-1.0.0.hbs)
- Restriction axis: commercial use — any purpose that is not commercial is
  free; commercial use requires a separate license.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `PolyForm-Noncommercial-1.0.0`.

## PolyForm-Shield-1.0.0 (steward text)

- Template: [`templates/source-available/polyform-shield-1.0.0/PolyForm-Shield-1.0.0-1.0.0.hbs`](../../../templates/source-available/polyform-shield-1.0.0/PolyForm-Shield-1.0.0-1.0.0.hbs)
- Restriction axis: competition — no offering of a competing product or
  service.
- Adoption: the identifier is not on the SPDX License List; leave
  `SPDX_LICENSE_IDENTIFIER` unset so the instance renders as
  `LICENSES/LicenseRef-<LICENSE_ID>.txt`, and record the steward identity
  `PolyForm-Shield-1.0.0` in the tenant documentation.

## PolyForm-Perimeter-1.0.1 (steward text)

- Template: [`templates/source-available/polyform-perimeter-1.0.1/PolyForm-Perimeter-1.0.1-1.0.0.hbs`](../../../templates/source-available/polyform-perimeter-1.0.1/PolyForm-Perimeter-1.0.1-1.0.0.hbs)
- Restriction axis: competition — no competing product at the same
  functionality perimeter. The template carries the steward's current 1.0.1
  patch text.
- Adoption: not SPDX-listed; leave `SPDX_LICENSE_IDENTIFIER` unset and record
  the steward identity `PolyForm-Perimeter-1.0.1`.

## PolyForm-Strict-1.0.0 (steward text)

- Template: [`templates/source-available/polyform-strict-1.0.0/PolyForm-Strict-1.0.0-1.0.0.hbs`](../../../templates/source-available/polyform-strict-1.0.0/PolyForm-Strict-1.0.0-1.0.0.hbs)
- Restriction axis: strict — no distribution or modification beyond the
  narrowest reading; evaluation-grade restriction.
- Adoption: not SPDX-listed; leave `SPDX_LICENSE_IDENTIFIER` unset and record
  the steward identity `PolyForm-Strict-1.0.0`.

## PolyForm-Small-Business-1.0.0 (`PolyForm-Small-Business-1.0.0`)

- Template: [`templates/source-available/polyform-small-business-1.0.0/PolyForm-Small-Business-1.0.0-1.0.0.hbs`](../../../templates/source-available/polyform-small-business-1.0.0/PolyForm-Small-Business-1.0.0-1.0.0.hbs)
- Restriction axis: commercial use with a size threshold — free for small
  businesses below the defined headcount and revenue limits.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `PolyForm-Small-Business-1.0.0`.

## PolyForm-Free-Trial-1.0.0 (steward text)

- Template: [`templates/source-available/polyform-free-trial-1.0.0/PolyForm-Free-Trial-1.0.0-1.0.0.hbs`](../../../templates/source-available/polyform-free-trial-1.0.0/PolyForm-Free-Trial-1.0.0-1.0.0.hbs)
- Restriction axis: time — free evaluation for a defined trial period.
- Adoption: not SPDX-listed; leave `SPDX_LICENSE_IDENTIFIER` unset and record
  the steward identity `PolyForm-Free-Trial-1.0.0`.

## PolyForm-Internal-Use-1.0.0 (steward text)

- Template: [`templates/source-available/polyform-internal-use-1.0.0/PolyForm-Internal-Use-1.0.0-1.0.0.hbs`](../../../templates/source-available/polyform-internal-use-1.0.0/PolyForm-Internal-Use-1.0.0-1.0.0.hbs)
- Restriction axis: internal use — free inside the organization, no external
  distribution.
- Adoption: not SPDX-listed; leave `SPDX_LICENSE_IDENTIFIER` unset and record
  the steward identity `PolyForm-Internal-Use-1.0.0`.

## Compliance classification (all variants)

Proprietary-grade: legal review before adoption, explicit scanner policy
entry, and no free mixing with open-source works in either direction.
