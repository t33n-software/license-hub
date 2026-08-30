# Public Domain Dedications — Family E

This document is the canonical family reference for the public-domain
dedication family. The family taxonomy, legal foundations, compatibility
matrix, and governance gates are owned by the
[license taxonomy canon](../README.md); this document owns the dedication
semantics and the template inventory of the family.

## 1. Family definition

Family E contains public-domain dedications and public-domain-equivalent
instruments. A dedication is a rights waiver, not a license: instead of
granting permission under conditions, the author attempts to give up the
rights themselves. This distinction matters because some jurisdictions do not
allow a full copyright waiver, which is why robust dedications carry a
fallback permissive license clause and why zero-condition licenses exist as
the legally safer instrument.

## 2. Jurisdiction caveat

In jurisdictions where authors cannot fully waive copyright — notably Germany
and Austria — a pure dedication degrades to its fallback license. This is why
CC0 and the Unlicense include one, and why 0BSD, which is a license rather
than a waiver, is the safest public-domain-equivalent choice for EU-centric
projects. When the rights holder is subject to such a jurisdiction,
recommending a bare dedication without a fallback clause is an error.

## 3. Template inventory

| Template | SPDX ID | Status | Canonical documentation |
|----------|---------|--------|-------------------------|
| `public-domain-dedication/cc0-1.0/CC0-1.0-1.0.0.hbs` | `CC0-1.0` | FSF-free, GPL-compatible; not OSI-approved | [cc0-1.0.md](cc0-1.0.md) |
| `public-domain-dedication/unlicense/Unlicense-1.0.0.hbs` | `Unlicense` | OSI-approved dedication | [unlicense.md](unlicense.md) |
| `public-domain-dedication/wtfpl/WTFPL-1.0.0.hbs` | `WTFPL` | FSF-free, not OSI-approved | [wtfpl.md](wtfpl.md) |
| `public-domain-dedication/blessing/blessing-1.0.0.hbs` | `blessing` | Blessing-style dedication | [blessing.md](blessing.md) |

## 4. Referenced members without a template and cross-references

| Instrument | SPDX ID | Disposition |
|------------|---------|-------------|
| Public Domain Mark | `CC-PDM-1.0` | Not a license — a label for works already in the public domain; never applied to new works |
| 0BSD | `0BSD` | Homed in [../permissive/bsd.md](../permissive/bsd.md) — it is a zero-condition license, not a waiver, and the recommended EU-safe public-domain equivalent |

## 5. Family conventions

1. **Instrument honesty.** A dedication is never called a license, and a
   zero-condition license is never called a waiver; the legal mechanics
   differ.
2. **Approval status is always stated.** 0BSD and Unlicense are OSI-approved,
   CC0 is not (while remaining FSF-free and GPL-compatible), WTFPL is neither
   OSI-approved nor corporate-policy-friendly.
3. **Patent caution.** CC0 expressly does not license patents, which is one
   reason it failed OSI review; where patent silence is unacceptable, 0BSD or
   MIT-0 is the correct public-domain-equivalent.

## 6. Do / Don't

**Do:** choose 0BSD when maximum permissiveness with OSI approval and EU
robustness is required; choose CC0 for content, data, and documentation
dedications; verify corporate policy acceptance before adopting WTFPL or
blessing-style texts.

**Don't:** apply the Public Domain Mark to new works; assume a dedication is
effective worldwide without the fallback analysis; treat "no license" as
public domain — absent a dedication or license, the work remains All Rights
Reserved.
