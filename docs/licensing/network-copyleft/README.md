# Network Copyleft Licenses — Family A4

This document is the canonical family reference for the network-copyleft
subfamily of Family A (Open Source). The family taxonomy, legal foundations,
compatibility matrix, and governance gates are owned by the
[license taxonomy canon](../README.md); this document owns the
network-copyleft member semantics and the template inventory of the family.

## 1. Family definition

Network copyleft closes the SaaS loophole: offering the software as a network
service triggers source obligations. Pure network use triggers nothing in
plain GPL; this family exists to close exactly that gap.

## 2. Template inventory

| Template | SPDX ID | Mechanism | Canonical documentation |
|----------|---------|-----------|-------------------------|
| `network-copyleft/agpl-3.0/AGPL-3.0-1.0.0.hbs` | `AGPL-3.0-only`, `AGPL-3.0-or-later` | §13: network interaction creates the right to receive source | [agpl-3.0.md](agpl-3.0.md) |
| `network-copyleft/cal-1.0/CAL-1.0-1.0.0.hbs` | `CAL-1.0` | Network use plus user-data portability | [cal-1.0.md](cal-1.0.md) |
| `network-copyleft/liliq-rplus-1.1/LiLiQ-Rplus-1.1-1.0.0.hbs` | `LiLiQ-Rplus-1.1` | Reciprocal forte: network use included | [liliq-rplus-1.1.md](liliq-rplus-1.1.md) |

## 3. Referenced members without a template and cross-listed members

| License | SPDX ID | Disposition |
|---------|---------|-------------|
| AGPL v1.0 | `AGPL-1.0` | Legacy (Affero); superseded by AGPL-3.0 |
| EUPL-1.2 | `EUPL-1.2` | Homed in [../weak-copyleft/eupl-1.2.md](../weak-copyleft/eupl-1.2.md) — its "communication to the public" clause covers SaaS, its primary class is weak copyleft |
| OSL-3.0 | `OSL-3.0` | Homed in [../strong-copyleft/osl-3.0.md](../strong-copyleft/osl-3.0.md) — its "external deployment" trigger is noted there |

## 4. Family conventions

1. **The network trigger is named precisely** for every member — which
   interaction creates the source duty and for which stack.
2. **Identifier discipline** — `-only` versus `-or-later` for the AGPL family.
3. **SaaS analysis is mandatory** before adopting or consuming a member:
   whether the use case is "offering the software as a service" decides the
   obligation.

## 5. Do / Don't

**Do:** choose AGPL-3.0 when community protection against SaaS capture is the
goal; publish the corresponding source to every network user; state the
`-only`/`-or-later` decision in every source header.

**Don't:** assume plain GPL covers network use; combine Family B
source-available code into network services built on these members; adopt the
legacy AGPL-1.0 text.
