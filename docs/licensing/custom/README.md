# Custom and Self-Written Licenses — Family D

This document is the canonical family reference for the custom license family.
The family taxonomy, legal foundations, compatibility matrix, and governance
gates are owned by the [license taxonomy canon](../README.md); this document
owns the custom-drafting architecture and the template inventory of the
family.

## 1. Family definition and lawfulness

Family D contains individually drafted license texts — proprietary custom
licenses, custom EULAs, and custom public licenses. Custom licensing is
lawful: any rights holder may draft their own license text, and the result is
legally valid if drafted correctly. Family D is simultaneously the
highest-risk family, because untested text is interpreted against its drafter
(contra proferentem) and carries ecosystem friction that standardized licenses
do not.

## 2. Established standards and conventions

| Standard | Rule |
|----------|------|
| SPDX License List | Custom licenses are not on the list and must be identified with the `LicenseRef-[idstring]` syntax, for example `SPDX-License-Identifier: LicenseRef-MyCompany-Community-1.0` |
| SPDX `AdditionRef-` | Additional terms on top of a listed license use the form `MIT WITH AdditionRef-MyExtraRequirement` |
| Linux Foundation guidance | `LicenseRef-` and `AdditionRef-` policy and best practices |
| REUSE Specification 3.2 | The custom license text lives in `LICENSES/LicenseRef-MyLicense.txt`; each source file carries `SPDX-License-Identifier` and `SPDX-FileCopyrightText` headers |
| OSI license-proliferation principle | Do not create or submit new open-source licenses without compelling reason; prefer established licenses |
| Naming discipline | A modified standard license must not keep the original license's name; "MIT with extra clause" is a new license, not MIT |

## 3. The 15 mandatory drafting blocks

Every custom license text resolves all fifteen blocks; a text missing any
block is legally incomplete:

1. Parties and definitions (Licensor, Licensee, Software, Derivative Work,
   Distribution, Network/SaaS Use, Affiliate, Internal Use)
2. Grant of rights (exact verbs; exclusivity; revocability; territory; term;
   transferability)
3. Scope limits (user counts, devices, cores, environments, field-of-use,
   internal-only)
4. Conditions (attribution form and placement, notice retention,
   source-disclosure triggers, reciprocity, marking of modifications)
5. Restrictions (no sublicensing, no resale, no competing product,
   non-commercial only, no derivatives, no removal of notices, no license-key
   circumvention)
6. Patent clause (grant or explicit non-grant; retaliation)
7. Trademark clause (no brand rights; naming rules for forks)
8. Warranty disclaimer (AS IS; jurisdiction-aware — EU consumer and mandatory
   liability limits)
9. Limitation of liability (caps, exclusion of indirect damages, carve-outs
   for willful misconduct and gross negligence — mandatory in DE/EU)
10. Termination (automatic on breach, cure period, effect on sublicenses,
    survival)
11. Compliance and audit (record-keeping, audit rights, certification on
    demand)
12. Governing law, venue, language
13. Boilerplate (severability, waiver, assignment, entire agreement,
    amendments, notices)
14. Versioning and change mechanism (fixed version versus "or later";
    conversion clauses if applicable)
15. Machine-readability (`LicenseRef-` identifier, REUSE-conformant
    placement, verbatim distribution rule)

## 4. Risk profile

Family D carries four characteristic risks. **Enforceability risk:** untested
text, with ambiguity interpreted against the drafter. **Ecosystem friction:**
no automatic detection (hosting-platform detectors and scanners report
"Unknown"), blocking by corporate license policies, and only partial registry
support for `LicenseRef-`. **License proliferation:** every custom text
multiplies legal review cost for all downstream users. **Mandatory legal
counsel:** any custom license with commercial or compliance impact requires
qualified lawyer review; this gate is absolute and no architectural analysis
replaces it.

## 5. Modified standard licenses

Changing only copyright-holder placeholders (year and name) keeps a license
standard. Changing any substantive term creates a new custom license: rename
it, mark it `LicenseRef-`, and treat it as Family D. A modified standard
license under the original standard name is a naming-discipline violation and
a supply-chain false declaration.

## 6. Template inventory

| Template | Identifier | Content | Canonical documentation |
|----------|-----------|---------|-------------------------|
| `custom/norepublish/NoRepublish-1.0.0.hbs` | `LicenseRef-<tenant LICENSE_ID>` | The organization No-Republish source-available text: permissive use grant paired with a hard republication ban | [norepublish.md](norepublish.md) |

## 7. Do / Don't

**Do:** assign a unique `LicenseRef-<idstring>` per license text variant;
place texts REUSE-3.2-conformant under `LICENSES/`; run `reuse lint` as a
quality gate; state the source-available family honestly in README and
metadata; pass every publication through the legal-counsel gate.

**Don't:** publish a custom license under a standard license's name; ship a
text missing any of the 15 drafting blocks; assume scanner detection will work
for `LicenseRef-` texts without a compliance-policy entry; treat a custom
license as OSI-approved without a successful OSI review; adapt a canonical
boilerplate by editing its invariant text body instead of its declared
placeholder set.
