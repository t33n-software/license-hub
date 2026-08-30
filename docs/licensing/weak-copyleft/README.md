# Weak Copyleft Licenses — Family A2

This document is the canonical family reference for the weak-copyleft
subfamily of Family A (Open Source). The family taxonomy, legal foundations,
compatibility matrix, and governance gates are owned by the
[license taxonomy canon](../README.md); this document owns the weak-copyleft
member semantics and the template inventory of the family.

## 1. Family definition

Weak copyleft scopes the copyleft obligation to the licensed component or
files; linking or larger-work combination with proprietary code remains
allowed under conditions. The defining question for every member is the exact
copyleft scope — file-level, module-level, or library-level — and the linking
or relinking duty that comes with it.

## 2. Template inventory

| Template | SPDX ID | Copyleft scope | Canonical documentation |
|----------|---------|----------------|-------------------------|
| `weak-copyleft/lgpl-2.1/LGPL-2.1-1.0.0.hbs` | `LGPL-2.1-only`, `LGPL-2.1-or-later` | Library-level | [lgpl.md](lgpl.md) |
| `weak-copyleft/lgpl-3.0/LGPL-3.0-1.0.0.hbs` | `LGPL-3.0-only`, `LGPL-3.0-or-later` | Library-level | [lgpl.md](lgpl.md) |
| `weak-copyleft/mpl-2.0/MPL-2.0-1.0.0.hbs` | `MPL-2.0` | File-level | [mpl-2.0.md](mpl-2.0.md) |
| `weak-copyleft/epl-2.0/EPL-2.0-1.0.0.hbs` | `EPL-2.0` | Module-level | [epl-2.0.md](epl-2.0.md) |
| `weak-copyleft/cddl-1.0/CDDL-1.0-1.0.0.hbs` | `CDDL-1.0` | File-level | [cddl.md](cddl.md) |
| `weak-copyleft/cddl-1.1/CDDL-1.1-1.0.0.hbs` | `CDDL-1.1` | File-level | [cddl.md](cddl.md) |
| `weak-copyleft/eupl-1.2/EUPL-1.2-1.0.0.hbs` | `EUPL-1.2` | Copyleft plus network clause | [eupl-1.2.md](eupl-1.2.md) |
| `weak-copyleft/ms-rl/MS-RL-1.0.0.hbs` | `MS-RL` | File-level reciprocal | [ms-rl.md](ms-rl.md) |
| `weak-copyleft/cpal-1.0/CPAL-1.0-1.0.0.hbs` | `CPAL-1.0` | Copyleft plus attribution badge | [cpal-1.0.md](cpal-1.0.md) |
| `weak-copyleft/apsl-2.0/APSL-2.0-1.0.0.hbs` | `APSL-2.0` | File-level-ish | [apsl-2.0.md](apsl-2.0.md) |
| `weak-copyleft/cecill-2.1/CECILL-2.1-1.0.0.hbs` | `CECILL-2.1` | GPL-compatible copyleft | [cecill-2.1.md](cecill-2.1.md) |
| `weak-copyleft/esa-pl-weak-copyleft-2.4/ESA-PL-weak-copyleft-2.4-1.0.0.hbs` | `ESA-PL-weak-copyleft-2.4` | Module-level | [esa-pl-weak-copyleft-2.4.md](esa-pl-weak-copyleft-2.4.md) |
| `weak-copyleft/erlpl-1.1/ErlPL-1.1-1.0.0.hbs` | `ErlPL-1.1` | File-level | [erlpl-1.1.md](erlpl-1.1.md) |

## 3. Referenced members without a template and cross-listed members

| License | SPDX ID | Disposition |
|---------|---------|-------------|
| Mozilla Public License 1.1 | `MPL-1.1` | Legacy; superseded by MPL-2.0 |
| Eclipse Public License 1.0 | `EPL-1.0` | Legacy and GPL-incompatible; superseded by EPL-2.0 |
| EUPL 1.0 / 1.1 | `EUPL-1.0`, `EUPL-1.1` | Legacy; superseded by EUPL-1.2 |
| Open Software License 3.0 | `OSL-3.0` | Homed in [../strong-copyleft/osl-3.0.md](../strong-copyleft/osl-3.0.md) — its dominant class is strong copyleft |
| Cryptographic Autonomy License 1.0 | `CAL-1.0` | Homed in [../network-copyleft/cal-1.0.md](../network-copyleft/cal-1.0.md) — its operative distinction is the network trigger |
| LiLiQ-P 1.1 | `LiLiQ-P-1.1` | Homed in [../permissive/liliq-p-1.1.md](../permissive/liliq-p-1.1.md) — permissive variant |
| LiLiQ-R 1.1 | `LiLiQ-R-1.1` | Homed in [../strong-copyleft/liliq-r-1.1.md](../strong-copyleft/liliq-r-1.1.md) — reciprocal variant |
| LiLiQ-Rplus 1.1 | `LiLiQ-Rplus-1.1` | Homed in [../network-copyleft/liliq-rplus-1.1.md](../network-copyleft/liliq-rplus-1.1.md) — reciprocal-forte variant with network scope |

## 4. Family conventions

1. **Scope statement.** Every member doc names the exact copyleft scope
   (file, module, or library) and the linking or relinking duty.
2. **Identifier discipline.** Exact SPDX identifiers including `-only` versus
   `-or-later` for the LGPL family; the deprecated `+` operator is never used.
3. **GPL compatibility is stated explicitly.** MPL-2.0 Exhibit B removes GPL
   compatibility; EPL-2.0 gains it only through the Secondary Licenses notice;
   CDDL stays GPL-incompatible; EUPL-1.2 carries an explicit compatibility
   appendix.
4. **Patent position is always stated.**

## 5. Do / Don't

**Do:** keep modifications to the licensed files or library under the same
license; provide the relink mechanism for LGPL; state GPL compatibility from
the canonical matrix; mark modifications where the member requires it.

**Don't:** assume MPL-2.0 code with an Exhibit B notice is GPL-compatible;
treat CDDL as GPL-compatible; adopt the legacy MPL-1.1, EPL-1.0, or EUPL-1.0/1.1
texts for new projects; extend the copyleft scope beyond the licensed
component without a strong-copyleft decision.
