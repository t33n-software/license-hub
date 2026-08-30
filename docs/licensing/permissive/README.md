# Permissive Licenses — Family A1

This document is the canonical family reference for the permissive subfamily
of Family A (Open Source). The family taxonomy, legal foundations,
compatibility matrix, and governance gates are owned by the
[license taxonomy canon](../README.md); this document owns the permissive
member semantics and the template inventory of the family.

## 1. Family definition and admission gate

Permissive licenses impose minimal conditions (attribution and notice
retention), carry no copyleft, allow sublicensing, and permit embedding into
proprietary works. Every member of this family complies with the Open Source
Definition and has passed the OSI license-review process. A license that fails
any OSD criterion or never passed review is not a member of this family,
regardless of marketing language.

## 2. Template inventory

Every template listed here is versioned independently with legal SemVer and
released through the immutable template release lane after legal-counsel
review.

| Template | SPDX ID | Patent position | Canonical documentation |
|----------|---------|-----------------|-------------------------|
| `permissive/mit/MIT-1.0.0.hbs` | `MIT` | No grant | [mit.md](mit.md) |
| `permissive/mit-0/MIT-0-1.0.0.hbs` | `MIT-0` | No grant | [mit.md](mit.md) |
| `permissive/mit-modern-variant/MIT-Modern-Variant-1.0.0.hbs` | `MIT-Modern-Variant` | No grant | [mit.md](mit.md) |
| `permissive/bsd-1-clause/BSD-1-Clause-1.0.0.hbs` | `BSD-1-Clause` | No grant | [bsd.md](bsd.md) |
| `permissive/bsd-2-clause/BSD-2-Clause-1.0.0.hbs` | `BSD-2-Clause` | No grant | [bsd.md](bsd.md) |
| `permissive/bsd-2-clause-patent/BSD-2-Clause-Patent-1.0.0.hbs` | `BSD-2-Clause-Patent` | Express grant | [bsd.md](bsd.md) |
| `permissive/bsd-3-clause/BSD-3-Clause-1.0.0.hbs` | `BSD-3-Clause` | No grant | [bsd.md](bsd.md) |
| `permissive/bsd-3-clause-clear/BSD-3-Clause-Clear-1.0.0.hbs` | `BSD-3-Clause-Clear` | Explicit non-grant | [bsd.md](bsd.md) |
| `permissive/0bsd/0BSD-1.0.0.hbs` | `0BSD` | No grant | [bsd.md](bsd.md) |
| `permissive/apache-2.0/Apache-2.0-1.0.0.hbs` | `Apache-2.0` | Express grant plus retaliation | [apache-2.0.md](apache-2.0.md) |
| `permissive/isc/ISC-1.0.0.hbs` | `ISC` | No grant | [isc.md](isc.md) |
| `permissive/zlib/Zlib-1.0.0.hbs` | `Zlib` | No grant | [zlib.md](zlib.md) |
| `permissive/bsl-1.0/BSL-1.0-1.0.0.hbs` | `BSL-1.0` | No grant | [bsl-1.0.md](bsl-1.0.md) |
| `permissive/afl-3.0/AFL-3.0-1.0.0.hbs` | `AFL-3.0` | Express grant | [afl-3.0.md](afl-3.0.md) |
| `permissive/artistic-2.0/Artistic-2.0-1.0.0.hbs` | `Artistic-2.0` | Express grant plus retaliation | [artistic-2.0.md](artistic-2.0.md) |
| `permissive/ecl-2.0/ECL-2.0-1.0.0.hbs` | `ECL-2.0` | Express grant plus retaliation | [ecl-2.0.md](ecl-2.0.md) |
| `permissive/ms-pl/MS-PL-1.0.0.hbs` | `MS-PL` | Express grant plus retaliation | [ms-pl.md](ms-pl.md) |
| `permissive/mulanpsl-2.0/MulanPSL-2.0-1.0.0.hbs` | `MulanPSL-2.0` | Express grant plus retaliation | [mulanpsl-2.0.md](mulanpsl-2.0.md) |
| `permissive/liliq-p-1.1/LiLiQ-P-1.1-1.0.0.hbs` | `LiLiQ-P-1.1` | Express grant plus retaliation | [liliq-p-1.1.md](liliq-p-1.1.md) |
| `permissive/blueoak-1.0.0/BlueOak-1.0.0-1.0.0.hbs` | `BlueOak-1.0.0` | Express grant | [blueoak-1.0.0.md](blueoak-1.0.0.md) |
| `permissive/fair/Fair-1.0.0.hbs` | `Fair` | No grant | [fair.md](fair.md) |
| `permissive/miros/MirOS-1.0.0.hbs` | `MirOS` | No grant | [miros.md](miros.md) |
| `permissive/fsfap/FSFAP-1.0.0.hbs` | `FSFAP` | No grant | [fsfap.md](fsfap.md) |
| `permissive/esa-pl-permissive-2.4/ESA-PL-permissive-2.4-1.0.0.hbs` | `ESA-PL-permissive-2.4` | Express grant plus retaliation | [esa-pl-permissive-2.4.md](esa-pl-permissive-2.4.md) |

## 3. Referenced members without a template

The following licenses are part of the permissive knowledge space but carry no
hub template, for the stated reason. This list is part of the family
completeness contract: every referenced license is either templated or
explicitly excluded here.

| License | SPDX ID | Exclusion rationale |
|---------|---------|---------------------|
| BSD 4-Clause "Original" | `BSD-4-Clause` | Obsolete advertising clause; must not be adopted for new code |
| Apache 1.1 / 1.0 | `Apache-1.1`, `Apache-1.0` | Legacy; superseded by Apache-2.0 |
| Historical Permission Notice | `HPND` | Legacy permissive template form |
| Mulan Permissive v1 | `MulanPSL-1.0` | Superseded by MulanPSL-2.0 |
| PHP License 3.01 | `PHP-3.01` | Ecosystem-bound (PHP runtime stack) |
| Ruby License | `Ruby` | Ecosystem-bound (Ruby runtime stack) |
| Python Software Foundation | `PSF-2.0`, `Python-2.0` | Ecosystem-bound (Python runtime stack) |
| curl License | `curl` | Ecosystem-bound (curl project) |
| ICU License | `ICU` | Ecosystem-bound (Unicode/ICU project) |
| PostgreSQL License | `PostgreSQL` | Semantically redundant with the MIT/BSD templates |
| NCSA / U. Illinois | `NCSA` | Semantically redundant with the MIT/BSD templates |
| Beerware | `Beerware` | Informal; no single steward-canonical text |
| WTFPL | `WTFPL` | Routed to family E — see [../public-domain-dedication/README.md](../public-domain-dedication/README.md) |
| The Unlicense | `Unlicense` | Routed to family E — see [../public-domain-dedication/README.md](../public-domain-dedication/README.md) |
| SQLite Blessing | `blessing` | Routed to family E — see [../public-domain-dedication/README.md](../public-domain-dedication/README.md) |
| Artistic License 1.0 | `Artistic-1.0` | Legacy and ambiguous; superseded by Artistic-2.0 |

## 4. Family conventions

1. **Patent position is always stated.** Every member doc and the inventory
   table declare the patent position: express grant with retaliation, express
   grant, no grant, or explicit non-grant.
2. **Attribution discipline.** Copyright and permission notices accompany
   every copy; the templates carry the copyright line in anchored form where
   the canonical text provides one.
3. **Identifier discipline.** The exact SPDX identifier is used everywhere,
   and the deprecated `+` operator is never used.
4. **Copyleft axis.** This family is the zero point of the copyleft axis
   (permissive → weak → strong → network); combination questions are answered
   from the canonical compatibility matrix in the canon.

## 5. Do / Don't

**Do:** retain all notices in copies and substantial portions; state the
patent position when advising; prefer Apache-2.0 when an express patent grant
is required; prefer 0BSD for maximum permissiveness with EU robustness.

**Don't:** remove or alter copyright notices; imply patent rights for
no-grant members; adopt excluded legacy or ecosystem-bound texts for new
projects; call any family B license permissive or open source.
