# License Taxonomy Canon

This document is the canonical classification and routing surface for every
license artifact governed by this project. It defines the complete license
family taxonomy, the legal foundations every licensing decision presupposes,
the cross-family compatibility and supply-chain rules, and the routing
contract that delegates per-family depth to the family documents under
`docs/licensing/<family>/`.

This document is a knowledge artifact, not legal advice. Jurisdiction-specific
binding decisions always require qualified legal counsel.

## 1. Legal foundations

Every licensing decision rests on five fixed legal foundations that no license
choice can bypass.

**Copyright default.** Software is protected by copyright automatically upon
creation; no registration is required. No license file means "All Rights
Reserved": nobody may lawfully use, copy, modify, or distribute the code, and
publishing source on a public host grants no usage rights beyond limited
platform terms of service. A license is a permission grant from the rights
holder; without it every use is an infringement.

**License instrument types.** Three instrument classes exist: bare licenses
(unilateral permission grants with conditions; most FOSS licenses such as MIT,
BSD, GPL), contract-style licenses (EULAs with offer/acceptance mechanics,
consideration, and negotiated terms), and public-domain dedications (rights
waivers such as CC0 and the Unlicense, which are not licenses and are legally
fragile where copyright cannot be fully waived, for example under German
Urheberrecht — robust dedications therefore include a fallback permissive
license clause).

**Enforceability.** FOSS licenses are enforceable copyright conditions, not
mere contractual covenants. GPL enforcement has an established history,
especially in Germany. Source-available licenses such as SSPL and BUSL have
limited courtroom track record, which leaves residual legal uncertainty.

**Rights dimensions.** Copyright is the core right every software license
addresses. Patents are separate rights granted explicitly only by some
licenses (Apache-2.0, GPLv3); MIT and BSD contain no explicit patent grant.
Trademarks are never covered by copyright licenses and require a separate
trademark policy. Export control and sanctions apply independently of the
chosen license.

**Core clause canon.** Every complete license resolves eight clause groups:
grant of rights, conditions, restrictions, warranty disclaimer, limitation of
liability, termination, governing law and venue, and versioning semantics
("only" versus "or-later").

## 2. The seven license families

The complete licensing space partitions into exactly seven families. Every
license artifact classifies into exactly one family before any further
analysis; family G is a meta-layer that combines licenses from the other
families.

| ID | Family | Core property | OSI-approved | Owning document |
|----|--------|---------------|--------------|-----------------|
| A | Open-Source Licenses | Comply with the Open Source Definition | Yes (per license) | `open-source` areas: [permissive](permissive/README.md), [weak-copyleft](weak-copyleft/README.md), [strong-copyleft](strong-copyleft/README.md), [network-copyleft](network-copyleft/README.md) |
| B | Source-Available Licenses | Source visible, but OSD-violating restrictions (field-of-use, SaaS, competition) | No | [source-available](source-available/README.md) |
| C | Proprietary / Closed-Source Licenses | No source rights or tightly restricted commercial grants | No | [proprietary](proprietary/README.md) |
| D | Custom / Self-Written Licenses | Individually drafted license texts | No (unless submitted and approved) | [custom](custom/README.md) |
| E | Public Domain Dedications | Rights waivers, not licenses | Mixed (Unlicense and 0BSD yes; CC0 no) | [public-domain-dedication](public-domain-dedication/README.md) |
| F | Non-Software / Adjacent Artifact Licenses | Content, documentation, data, hardware, fonts, AI models | Mixed | [non-software](non-software/README.md) |
| G | Combination and Multi-Licensing Mechanisms | Dual/multi-licensing, exceptions, SPDX expressions | N/A (meta-layer) | [multi-licensing](multi-licensing/README.md) |

### 2.1 Family A subfamilies

Family A splits by copyleft strength. **A1 Permissive**: minimal conditions,
sublicensable (MIT, BSD variants, Apache-2.0, ISC, 0BSD, zlib, and ecosystem
variants). **A2 Weak Copyleft**: component- or file-scoped copyleft
(LGPL-2.1/3.0, MPL-2.0, EPL-1.0/2.0, CDDL, EUPL-1.2, MS-RL, CPAL-1.0,
OSL-3.0, CeCILL). **A3 Strong Copyleft**: derivative works remain under the
same license on distribution (GPL-2.0, GPL-3.0, CAL-1.0, OSL, Sleepycat).
**A4 Network Copyleft**: network service triggers source obligations
(AGPL-3.0, EUPL-1.2, OSL-3.0, CAL-1.0).

### 2.2 Family B subfamilies

Family B is the modern commercial open-source-protection layer and is treated
as proprietary-with-source-access for compliance. **B1 SaaS-Protection**
(SSPL-1.0, Elastic-2.0, Confluent Community, Redis RSALv2). **B2 Delayed Open
Source** (BUSL-1.1 with Additional Use Grant, Change Date, Change License;
FSL-1.1; Fair Source). **B3 Add-On and Modular Restriction** (Commons Clause;
PolyForm Noncommercial, Shield, Perimeter, Strict, Small-Business, Free-Trial,
Internal-Use). **B4 Fair-Code / Ethical / Values-Restricted** (Sustainable
Use, Hippocratic-2.1, JSON, BigScience RAIL / OpenRAIL).

### 2.3 Family C instruments

Family C covers EULAs, negotiated enterprise agreements, shared-source
programs ("look, don't touch"), trade-secret internal use,
evaluation/trial/beta/NFR grants, and the no-license default (All Rights
Reserved). It also owns the commercial licensing metric taxonomy (per-seat,
concurrent, per-device, per-server, per-core, capacity, site, enterprise, OEM,
subscription, perpetual-plus-maintenance, usage-based, feature-tiered,
freemium, academic, royalty-based) and the mandatory EULA building blocks.

### 2.4 Family D scope

Family D covers individually drafted license texts. Custom licenses are lawful
but the highest-risk family: they must use SPDX `LicenseRef-<idstring>`
identification, REUSE 3.2 file placement, the 15-block drafting canon, and
naming discipline (a modified standard license never keeps the standard name).
Every custom license with commercial or compliance impact passes the
legal-counsel gate before publication.

### 2.5 Family E instruments

Family E covers CC0-1.0 (waiver plus fallback license; not OSI-approved), the
Unlicense (OSI-approved dedication), 0BSD (a zero-condition license, not a
waiver), WTFPL (informal, not OSI-approved), the Public Domain Mark (a label,
not a license), and blessing-style dedications. In jurisdictions where authors
cannot fully waive copyright (Germany, Austria), pure dedications degrade to
their fallback license, which makes 0BSD the safest public-domain-equivalent
choice for EU-centric projects.

### 2.6 Family F subfamilies

Family F covers non-code artifacts. **F1 Content and documentation** (the six
Creative Commons 4.0 licenses CC-BY, CC-BY-SA, CC-BY-ND, CC-BY-NC,
CC-BY-NC-SA, CC-BY-NC-ND; GFDL; Licence Art Libre; FreeBSD Documentation
License — Creative Commons licenses must not be used for software, CC0
excepted). **F2 Data and databases** (ODbL-1.0, ODC-BY-1.0, PDDL-1.0,
CDLA-Permissive, CDLA-Sharing, C-UDA-1.0, DL-DE-BY-2.0, DL-DE-ZERO-2.0,
etalab-2.0). **F3 Open hardware** (CERN-OHL v1/v2 permissive, weakly and
strongly reciprocal variants; TAPR; Solderpad). **F4 Fonts** (SIL OFL-1.0/1.1,
IPA). **F5 AI/ML models** (BigScience RAIL, OpenRAIL-M, Llama Community
License, Gemma Terms of Use, ASWF Digital Assets). **F6 Standards and
specifications** (Community Specification License, W3C, IETF, ISO permission
notices).

### 2.7 Family G mechanisms

Family G is the combination meta-layer: dual-licensing (same code under
copyleft plus paid commercial license, requiring 100 percent copyright control
via CLA or assignment), multi-licensing with recipient choice (`MIT OR
Apache-2.0`), license stacking (`BSD-3-Clause AND MIT`), SPDX expression
syntax (`AND`, `OR`, `WITH`, parentheses, `-only`/`-or-later` instead of the
deprecated `+`), license exceptions (`Classpath-exception-2.0`,
`GCC-exception-3.1`, `LLVM-exception`, `OpenSSL-exception`, and others), and
inbound contribution mechanisms (CLA, DCO `Signed-off-by`, copyright
assignment, inbound-equals-outbound default).

## 3. Compatibility and distribution triggers

License combination outcomes follow this canonical compatibility matrix; a
combination question is answered from this matrix before any case-specific
analysis.

| Combination | Result |
|-------------|--------|
| MIT / BSD / ISC / 0BSD into any work | Compatible everywhere (retain notices) |
| Apache-2.0 into GPLv3 project | Compatible (one-way into GPLv3) |
| Apache-2.0 into GPLv2-only project | Incompatible (patent retaliation conflict) |
| GPLv2-only with GPLv3-only | Incompatible (unless "or-later") |
| CDDL with GPL | Incompatible |
| MPL-2.0 with GPL | Compatible unless Exhibit B notice present |
| LGPL-2.1/3.0 in proprietary application | Allowed via dynamic linking plus relink mechanism; static linking needs object files or an exception |
| AGPL-3.0 with GPL-3.0 | Linkable both ways via the §13 clauses; combined network work stays AGPL |
| EPL-2.0 with GPL | Only with the Secondary Licenses notice |
| EUPL-1.2 with GPL/AGPL/MPL/LGPL/EPL/OSL/CeCILL | Explicitly compatible per appendix |
| CC BY-SA 4.0 into GPL | One-way compatible into GPLv3 |
| CC0 into any | Compatible |
| Family B (SSPL/BUSL/Elastic-2.0) into any open combination | Treat as proprietary: no free mixing; legal review required |

Obligation triggers split into two classes. Distribution or conveyance
triggers fire when copies reach third parties (GPL, LGPL, MPL, EPL, CDDL,
Apache-2.0). Network triggers close the SaaS loophole: pure network use
triggers nothing in GPL, closed by AGPL-3.0 §13, OSL-3.0 external deployment,
EUPL-1.2 communication to the public, CAL-1.0, and radically by SSPL-1.0
(entire service stack). All FOSS licenses permit unlimited private and
internal use and modification without disclosure.

Attribution and notice obligations apply across all families: license text and
copyright notices accompany copies; Apache-2.0 additionally preserves NOTICE
file contents; BSD-3-Clause forbids endorsement by name; CC BY requires
attribution with link, license indication, and change marking; CPAL-1.0
requires a UI attribution badge; Apache-2.0, GPLv3, and Elastic-2.0 require
marking modifications.

Patent strategy distinguishes three positions: express grant with retaliation
(Apache-2.0, GPLv3, MPL-2.0, EPL-2.0, Elastic-2.0, MS-PL/MS-RL), no grant
(MIT, BSD-2/3-Clause, GPL-2.0), and explicit non-grant (BSD-3-Clause-Clear,
CC0).

## 4. Supply-chain compliance

Enterprise consumption of third-party licenses runs through a fixed toolchain
layer. SBOM standards are SPDX (ISO/IEC 5962) and CycloneDX (ECMA-424), with
the NTIA minimum elements as the content baseline. Scanning and detection use
tools such as ScanCode, FOSSology, Black Duck, FOSSA, Snyk, ORT,
ClearlyDefined, and hosting-platform detection such as GitHub Licensee or
GitLab License Compliance. Policy gates enforce allow/deny license lists per
project, copyleft quarantine, Family B block rules, and a `LicenseRef` review
queue. REUSE 3.2 provides per-file `SPDX-License-Identifier` and
`SPDX-FileCopyrightText` headers plus a `LICENSES/` directory,
machine-verifiable through `reuse lint`. Inbound governance enforces CLA/DCO
rules and re-checks dependency licenses on every version upgrade, because
license drift on version bumps is a real supply-chain event.

## 5. Hosting-platform and registry boundary

License texts are platform-agnostic; hosting platforms differ only in
detection, templates, and compliance tooling. The legal layer of licensing
lives in this documentation plane, while platform-specific operational
mechanics (detection behavior, template pickers, compliance features) belong
to the hosting platform's own documentation. No platform-specific legal
semantics are defined here.

Package registries form the distribution layer with their own license
mechanisms: npm uses a `package.json` SPDX expression with a
`SEE LICENSE IN <file>` escape hatch; PyPI uses PEP 639 `License-Expression`
plus `License-File`; Maven Central uses POM `<licenses>`; NuGet uses
`<license type="expression">` or `type="file"`; crates.io uses `license` or
`license-file`; Go modules have no manifest field and rely on `LICENSE` file
detection by pkg.go.dev; OCI images use the
`org.opencontainers.image.licenses` annotation; RubyGems uses `spec.license`.

## 6. Decision axes for licensing questions

Every licensing question is normalized onto ten axes before options are
formed:

1. Rights-holder goal (adoption versus control versus monetization versus
   compliance hygiene)
2. Distribution mode (source, binary, SaaS/network, embedded/OEM,
   internal-only)
3. Copyleft tolerance (none, weak, strong, network)
4. Patent strategy (express grant needed versus no-grant acceptable)
5. Compatibility constraints (inbound licenses of the existing stack and
   target ecosystem norms)
6. Commercial model fit (open-core, dual-licensing, SaaS protection, support,
   marketplace)
7. Enforcement and jurisdiction (governing law, EU/DE specifics, export
   control)
8. Artifact type (code versus documentation, data, model weights, hardware,
   fonts — the family F routing axis)
9. Platform and registry mechanics (detection, templates, registry SPDX
   fields)
10. Supply-chain posture (SBOM, scanner policy, CLA/DCO inbound, Family B
    quarantine)

## 7. Governance and legal-counsel gates

Three gates are absolute in this documentation plane.

**Family correctness.** Every artifact is declared in its true family; calling
a source-available, proprietary, or custom license "open source" in the OSI
sense is forbidden.

**Legal-counsel gate.** Custom license drafting and any jurisdiction-binding
recommendation require qualified legal counsel before publication;
architectural analysis never replaces legal advice.

**Single source of truth.** Each licensing truth lives exactly once at its
narrowest owner: the taxonomy in this document, family semantics in the family
documents, concrete license texts in `templates/`, and the multi-tenant
operation contract in `../infrastructure/` and `../adoption-guide.md`.
