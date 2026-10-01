# Server Side Public License v1 (`SSPL-1.0`)

Family: B — Source-Available, group B1 (SaaS-Protection). **Not OSI-approved
open source.** Canon: [license taxonomy canon](../README.md); family
reference: [source-available README](README.md).

- Template: [`templates/source-available/sspl-1.0/SSPL-1.0-1.0.0.hbs`](../../../templates/source-available/sspl-1.0/SSPL-1.0-1.0.0.hbs)
- Steward/status: MongoDB; the OSI review was withdrawn in 2019; also used by
  Redis in a dual arrangement since 2024.

## Grant, conditions, restrictions

- Grant: the GPLv3-style grant for non-service use.
- Conditions: the GPLv3-style copyleft duties.
- Restrictions: §13 — offering the software as a service requires publishing
  the entire service stack source (management, backup, hosting, and monitoring
  software), a radically wider trigger than AGPL.

## Patent position

Express grant with retaliation through the GPLv3 base.

## Compliance classification

Proprietary-grade: legal review before adoption, no assumption of
redistribution or SaaS rights, explicit scanner policy entry, and no free
mixing with open-source works.

## Adoption

Set `SPDX_LICENSE_IDENTIFIER` to `SSPL-1.0`. Choose it only when the entire
service-stack source duty is the intended protection. The template body is
byte-identical to the canonical text.
