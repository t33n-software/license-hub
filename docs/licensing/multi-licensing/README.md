# Combination and Multi-Licensing Mechanisms — Family G

This document is the canonical family reference for the combination and
multi-licensing meta-layer. The family taxonomy, legal foundations,
compatibility matrix, and governance gates are owned by the
[license taxonomy canon](../README.md); this document owns the combination
mechanisms.

## 1. Family definition — why this family carries no text templates

Family G is the meta-layer of the licensing taxonomy: it contains **no license
texts of its own** but the mechanisms that combine, offer, or extend licenses
from families A–F. Every Family G analysis preserves the underlying family
rules of each combined license; a combination mechanism never relaxes the
strictest member's obligations.

Because the mechanisms own no texts, `templates/multi-licensing/` carries no
license-text templates: a composition is served by the member templates of the
other families plus the SPDX expression that binds them. This document is the
canonical reference for forming those compositions.

## 2. G1 — Dual-licensing (commercial plus open)

Dual-licensing offers the same code under two parallel offers: a strong
copyleft license for the community (GPL/AGPL) and a paid commercial license
for proprietary embedding. Canonical examples are MySQL (GPL/commercial), Qt
(GPL/LGPL/commercial), Ghostscript, and Berkeley DB (Sleepycat).

**Hard prerequisite:** 100 percent copyright control over every contribution,
achieved through a CLA or copyright assignment; without it the rights holder
cannot lawfully issue the commercial offer. Projects with external
contributions under inbound-equals-outbound cannot switch to a dual-licensing
model retroactively without relicensing every contribution.

The commercial side of a dual-licensing offer is always a custom instrument
under Family D discipline — see [../custom/README.md](../custom/README.md).

## 3. G2 — Multi-licensing (recipient's choice)

Multi-licensing offers the recipient a choice between licenses through an
SPDX `OR` expression: `MIT OR Apache-2.0` (the Rust ecosystem standard) or
`Artistic-1.0 OR GPL-1.0-or-later` (Perl). The recipient picks one license;
obligations of the unchosen licenses do not apply.

To form a recipient-choice composition from this hub, adopt each member
template into the tenant's `LICENSES/` area and declare the expression in the
per-file headers, for example `SPDX-License-Identifier: MIT OR Apache-2.0`
with `LICENSES/MIT.txt` and `LICENSES/Apache-2.0.txt` rendered from
[../permissive/mit.md](../permissive/mit.md) and
[../permissive/apache-2.0.md](../permissive/apache-2.0.md).

## 4. G3 — License stacking (cumulative)

Stacking applies multiple licenses simultaneously through an SPDX `AND`
expression: `BSD-3-Clause AND MIT` means both licenses apply at once — a dual
obligation, not a choice. Every stacked license's conditions must be satisfied
concurrently.

## 5. G4 — SPDX license expressions (syntax standard)

SPDX expressions are the canonical machine-readable combination syntax.
Operators are `AND`, `OR`, and `WITH` (exception), plus parentheses for
grouping; identifiers are case-sensitive. The `+` operator is deprecated;
versioned licenses use the `-only` and `-or-later` identifier suffixes
instead. Reference examples: `Apache-2.0 OR MIT`, `GPL-2.0-only WITH
Classpath-exception-2.0`, `(MIT AND CC-BY-4.0)`.

Every declared expression must parse under the SPDX syntax rules, use current
identifiers, and reference only real licenses and registered exceptions.

## 6. G5 — License exceptions (WITH)

Exceptions modify a base license through the `WITH` operator. The canonical
inventory includes `Classpath-exception-2.0` (linking),
`Autoconf-exception-2.0/3.0`, `GCC-exception-3.1`, `LLVM-exception`,
`Font-exception-2.0`, `OpenSSL-exception`, `Bison-exception-2.2`,
`Qt-GPL-exception-1.0`, and the Solderpad hardware line `SHL-2.0/2.1`.

Exceptions are never standalone licenses: they live in the expression that
binds them to their base license (for example `GPL-2.0-only WITH
Classpath-exception-2.0`), so they carry no text templates in this hub. Never
invent exception identifiers not present on the SPDX exceptions list.

## 7. G6 — Inbound contribution licensing

Inbound mechanisms determine the license under which contributions arrive:

| Mechanism | Function |
|-----------|----------|
| CLA (Contributor License Agreement) | Contributors grant broad rights including relicensing (Apache CLA, Google CLA); enables dual-licensing |
| DCO (Developer Certificate of Origin) | A `Signed-off-by` line certifies the right to contribute (Linux kernel model); lightweight, no rights transfer |
| Copyright assignment | FSF-style full assignment to the project steward |
| Inbound = outbound rule | Default norm: contributions arrive under the project's existing license |

CLA texts are contract instruments with per-entity schedules and follow the
Family D drafting discipline with the legal-counsel gate; a DCO needs no
template — it is the `Signed-off-by` line in every commit.

## 8. Family conventions

1. **Choice-versus-obligation discipline.** `OR` grants a recipient choice,
   `AND` accumulates obligations, and confusing them in either direction is a
   compliance defect.
2. **Control prerequisite.** Dual-licensing advice is only valid with proven
   100 percent copyright control.
3. **Expression validity.** Every declared expression parses under the SPDX
   syntax rules, uses current identifiers (no `+`), and references only real
   licenses and registered exceptions.

## 9. Do / Don't

**Do:** write `MIT OR Apache-2.0` when recipients should choose; write
`BSD-3-Clause AND MIT` when both apply; use `GPL-2.0-only WITH
Classpath-exception-2.0` for linking exceptions; adopt a DCO for lightweight
inbound governance and a CLA when relicensing freedom is required.

**Don't:** use the deprecated `+` operator; treat an `AND` stack as a menu;
attempt dual-licensing without complete inbound rights control; invent
exception identifiers; treat a CLA as a copyright assignment or a DCO as a
rights transfer.
