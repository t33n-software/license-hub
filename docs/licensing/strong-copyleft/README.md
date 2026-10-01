# Strong Copyleft Licenses — Family A3

This document is the canonical family reference for the strong-copyleft
subfamily of Family A (Open Source). The family taxonomy, legal foundations,
compatibility matrix, and governance gates are owned by the
[license taxonomy canon](../README.md); this document owns the strong-copyleft
member semantics and the template inventory of the family.

## 1. Family definition

Strong copyleft requires derivative and combined works to remain under the
same license upon distribution, with full corresponding source disclosure.
The distribution trigger fires when copies reach third parties; pure private
and internal use and modification stay free.

## 2. Template inventory

| Template | SPDX ID | Notes | Canonical documentation |
|----------|---------|-------|-------------------------|
| `strong-copyleft/gpl-2.0/GPL-2.0-1.0.0.hbs` | `GPL-2.0-only`, `GPL-2.0-or-later` | Distribution-triggered; the "or later" choice decides v3 compatibility | [gpl.md](gpl.md) |
| `strong-copyleft/gpl-3.0/GPL-3.0-1.0.0.hbs` | `GPL-3.0-only`, `GPL-3.0-or-later` | Patent grant and retaliation (§11), anti-tivoization (§6), cure period (§8) | [gpl.md](gpl.md) |
| `strong-copyleft/osl-3.0/OSL-3.0-1.0.0.hbs` | `OSL-3.0` | Contract-style strong copyleft including network use | [osl-3.0.md](osl-3.0.md) |
| `strong-copyleft/esa-pl-strong-copyleft-2.4/ESA-PL-strong-copyleft-2.4-1.0.0.hbs` | `ESA-PL-strong-copyleft-2.4` | EU-aligned strongly reciprocal variant | [esa-pl-strong-copyleft-2.4.md](esa-pl-strong-copyleft-2.4.md) |
| `strong-copyleft/liliq-r-1.1/LiLiQ-R-1.1-1.0.0.hbs` | `LiLiQ-R-1.1` | Québec reciprocal variant | [liliq-r-1.1.md](liliq-r-1.1.md) |

## 3. Referenced members without a template

| License | SPDX ID | Exclusion rationale |
|---------|---------|---------------------|
| Common Public License 1.0 | `CPL-1.0` | Superseded by the EPL line |
| CeCILL v2.0 | `CECILL-2.0` | Superseded by CECILL-2.1 (homed in [../weak-copyleft/cecill-2.1.md](../weak-copyleft/cecill-2.1.md)) |
| Copyleft-next 0.3.x | `copyleft-next-0.3.0`, `copyleft-next-0.3.1` | Experimental; no stable steward release |
| Sleepycat License | `Sleepycat` | Vendor-bound (Berkeley DB dual-licensing context) |
| NASA Open Source Agreement 1.3 | `NASA-1.3` | Special-purpose government license; not for general reuse |
| OSET Public License 2.1 | `OSET-PL-2.1` | Special-purpose election-technology license |
| Open Software License 1.0 / 2.1 | `OSL-1.0`, `OSL-2.1` | Legacy; superseded by OSL-3.0 |
| Cryptographic Autonomy License 1.0 | `CAL-1.0` | Homed in [../network-copyleft/cal-1.0.md](../network-copyleft/cal-1.0.md) — its operative distinction is the network trigger |

## 4. Family conventions

1. **The "or-later" decision is explicit.** For the GPL family the tenant
   declares `-only` or `-or-later` in its notice headers; the template text is
   shared and never carries the suffix decision.
2. **Source disclosure is the core duty.** Distribution of derivatives
   requires the full corresponding source under the same license.
3. **Patent position is always stated.**
4. **Compatibility is answered from the canonical matrix** — GPLv2-only with
   GPLv3-only is incompatible unless "or-later" is declared.

## 5. Do / Don't

**Do:** publish corresponding source with every distribution; mark
modifications where required; choose GPL-3.0 when patent retaliation and a
cure period are wanted; state the `-only`/`-or-later` decision in every source
header.

**Don't:** combine GPLv2-only code with Apache-2.0; assume network use
triggers anything in plain GPL (that is the network-copyleft family);
relicense a strong-copyleft work without 100 percent copyright control.
