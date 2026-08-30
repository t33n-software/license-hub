# Source-Available Licenses — Family B

This document is the canonical family reference for the source-available
license family. The family taxonomy, legal foundations, compatibility matrix,
and governance gates are owned by the [license taxonomy canon](../README.md);
this document owns the source-available member semantics and the template
inventory of the family.

## 1. Family definition and honest declaration

Family B contains licenses whose source code is visible but whose restrictions
violate the Open Source Definition, typically through field-of-use, SaaS, or
anti-competition clauses. This family is the modern commercial
open-source-protection layer.

**No Family B license is OSI-approved open source.** Presenting one as open
source is a category error forbidden in this project. For compliance purposes
every Family B license is handled like proprietary software: legal review
required, no assumption of redistribution rights, and mandatory SaaS/offering
analysis.

## 2. Restriction axes

Before comparing members, the restriction axis is identified, because members
on different axes are not substitutes:

- **Commercial use** — Commons Clause, Sustainable Use
- **SaaS offering** — SSPL-1.0, Elastic-2.0, RSALv2
- **Competition** — PolyForm-Shield
- **Time** — BUSL-1.1, FSL-1.1, Fair Source (delayed open source)
- **Ethics** — Hippocratic-2.1, RAIL

## 3. Template inventory

| Template | SPDX ID | Group | Canonical documentation |
|----------|---------|-------|-------------------------|
| `source-available/sspl-1.0/SSPL-1.0-1.0.0.hbs` | `SSPL-1.0` | B1 SaaS-Protection | [sspl-1.0.md](sspl-1.0.md) |
| `source-available/elastic-2.0/Elastic-2.0-1.0.0.hbs` | `Elastic-2.0` | B1 SaaS-Protection | [elastic-2.0.md](elastic-2.0.md) |
| `source-available/fsl-1.1-alv2/FSL-1.1-ALv2-1.0.0.hbs` | `FSL-1.1-ALv2` | B2 Delayed Open Source | [fsl.md](fsl.md) |
| `source-available/fsl-1.1-mit/FSL-1.1-MIT-1.0.0.hbs` | `FSL-1.1-MIT` | B2 Delayed Open Source | [fsl.md](fsl.md) |
| `source-available/hippocratic-2.1/Hippocratic-2.1-1.0.0.hbs` | `Hippocratic-2.1` | B4 Ethical | [hippocratic-2.1.md](hippocratic-2.1.md) |
| `source-available/json/JSON-1.0.0.hbs` | `JSON` | B4 Values-Restricted | [json.md](json.md) |
| `source-available/polyform-noncommercial-1.0.0/PolyForm-Noncommercial-1.0.0-1.0.0.hbs` | `PolyForm-Noncommercial-1.0.0` | B3 Modular | [polyform.md](polyform.md) |
| `source-available/polyform-shield-1.0.0/PolyForm-Shield-1.0.0-1.0.0.hbs` | `PolyForm-Shield-1.0.0` (steward text) | B3 Modular | [polyform.md](polyform.md) |
| `source-available/polyform-perimeter-1.0.1/PolyForm-Perimeter-1.0.1-1.0.0.hbs` | `PolyForm-Perimeter-1.0.1` (steward text) | B3 Modular | [polyform.md](polyform.md) |
| `source-available/polyform-strict-1.0.0/PolyForm-Strict-1.0.0-1.0.0.hbs` | `PolyForm-Strict-1.0.0` (steward text) | B3 Modular | [polyform.md](polyform.md) |
| `source-available/polyform-small-business-1.0.0/PolyForm-Small-Business-1.0.0-1.0.0.hbs` | `PolyForm-Small-Business-1.0.0` | B3 Modular | [polyform.md](polyform.md) |
| `source-available/polyform-free-trial-1.0.0/PolyForm-Free-Trial-1.0.0-1.0.0.hbs` | `PolyForm-Free-Trial-1.0.0` (steward text) | B3 Modular | [polyform.md](polyform.md) |
| `source-available/polyform-internal-use-1.0.0/PolyForm-Internal-Use-1.0.0-1.0.0.hbs` | `PolyForm-Internal-Use-1.0.0` (steward text) | B3 Modular | [polyform.md](polyform.md) |

## 4. Referenced members without a template

| License | Identifier | Exclusion rationale |
|---------|-----------|---------------------|
| Business Source License 1.1 | `BUSL-1.1` | Parametrized text: the Additional Use Grant, the Change Date, and the Change License are free legal parameters embedded in the text, which the closed anchor set of the hub cannot resolve; adoption is a per-tenant legal decision — see [../custom/README.md](../custom/README.md) |
| Commons Clause | none (add-on) | Not a standalone license: it layers a no-sale condition onto an existing license and the combination must not keep the original license name — applying it is custom drafting under Family D discipline |
| Confluent Community License | `LicenseRef-...` | Vendor-owned custom text (Confluent); adopting it for other products is Family D drafting |
| Redis Source Available License 2.0 | `LicenseRef-...` | Vendor-owned custom text (Redis Ltd.); Family D discipline |
| Fair Source License | `LicenseRef-...` (fair.io) | Not on the SPDX License List; evolving steward terms — re-evaluate on steward stabilization |
| Sustainable Use License | custom (n8n) | Vendor-owned custom text; Family D discipline |
| BigScience RAIL / OpenRAIL | custom family | AI-model use-restriction family with variant selection and a mandatory legal gate — see [../non-software/README.md](../non-software/README.md) |

## 5. Family conventions

1. **Honest declaration.** Every reference names the family as
   source-available, never as open source.
2. **Restriction-axis separation** before any comparison (Section 2).
3. **Identifier discipline.** Standardized members carry their SPDX IDs;
   non-standardized members are identified as custom `LicenseRef-...` texts
   and treated with Family D drafting discipline.
4. **Enforceability caution.** Source-available licenses have limited
   courtroom track record; residual legal uncertainty is stated whenever a
   Family B license is recommended.
5. **Proprietary-grade compliance.** A Family B dependency in a build is a
   compliance event, not a routine import: legal review, explicit scanner
   policy entry, and no free mixing with open-source works in either
   direction.

## 6. Do / Don't

**Do:** state the exact restriction axis and conversion semantics (Change
Date, Change License) when evaluating delayed-open-source members; check
whether a use case is "offering the software as a service" before approving
SSPL or Elastic-2.0 dependencies; treat `LicenseRef`-identified Family B texts
with Family D rigor.

**Don't:** call any Family B license open source; assume a non-competing
mirror or internal use is automatically permitted without reading the specific
text; combine Family B code into open-source projects; treat the Commons
Clause as a standalone license or keep the original license name after
applying it.
