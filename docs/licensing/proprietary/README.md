# Proprietary and Closed-Source Licenses — Family C

This document is the canonical family reference for the proprietary and
closed-source license family. The family taxonomy, legal foundations,
compatibility matrix, and governance gates are owned by the
[license taxonomy canon](../README.md); this document owns the proprietary
instrument semantics and the template inventory of the family.

## 1. Family definition

Family C contains licenses that grant no source rights or only tightly
restricted commercial grants. It is the default state of software: because
copyright applies automatically, the absence of any license already means All
Rights Reserved, and every proprietary instrument is a deliberate, negotiated,
or unilateral narrowing of that default into a defined commercial grant.

## 2. Instrument types

Family C instruments divide into six classes:

1. **EULA (End User License Agreement):** a contract between licensor and end
   user defining grant, restrictions, IP reservation, warranty disclaimer,
   liability cap, maintenance, audit rights, termination, and governing law.
2. **Negotiated enterprise agreements:** MSA plus order forms, custom SLAs,
   indemnification, and source-escrow clauses.
3. **Shared Source:** source visible under restrictive terms — "look, don't
   touch".
4. **Trade secret / internal-use only:** no external grant at all.
5. **Evaluation, trial, beta, and NFR (not-for-resale) licenses:** time- or
   scope-limited grants.
6. **No license (default):** All Rights Reserved — maximum restriction by
   default law.

## 3. Commercial licensing models (metric taxonomy)

The commercial metric is independent of the legal instrument; the same EULA
can carry any of these models: per-seat / named user; concurrent / floating;
per-device; per-server / per-instance; per-core / per-CPU / per-socket;
capacity-based (PVU/RVU); site license; enterprise / unlimited (ELA); OEM /
embedded; subscription / term; perpetual plus maintenance; usage-based /
metered; feature-based / tiered editions; freemium; academic / nonprofit /
government; royalty-based.

## 4. Mandatory EULA building blocks

A complete Family C contract resolves all of the following blocks: parties and
definitions; license grant and scope (exclusive or non-exclusive,
transferable, revocable, territory, term); permitted installations and users;
restrictions (no reverse engineering except statutory rights, no rental, no
benchmarking disclosure); IP reservation; fees and audit rights;
confidentiality; warranty disclaimer; liability limitation and caps;
indemnification; support and maintenance terms; updates and upgrade policy;
termination and effect of termination; export control and sanctions
compliance; governing law and venue; severability; assignment; and the
entire-agreement clause.

## 5. Template inventory

Exactly one instrument of this family has a single fixed canonical text — the
no-license default. Every other Family C instrument is negotiated or
vendor-drafted and therefore follows the Family D drafting discipline in
[../custom/README.md](../custom/README.md); none of them can be a standard
template.

| Template | Identifier | Content | Canonical documentation |
|----------|-----------|---------|-------------------------|
| `proprietary/all-rights-reserved/AllRightsReserved-1.0.0.hbs` | `LicenseRef-<tenant LICENSE_ID>` | The organization-standard All-Rights-Reserved notice | [all-rights-reserved.md](all-rights-reserved.md) |

## 6. Family conventions

1. **Default awareness.** When no license is found, the correct
   classification is All Rights Reserved and the correct advice is that no use
   is permitted — never infer permission from public visibility of source.
2. **Instrument-metric separation.** Answer the legal question (what is
   granted) separately from the commercial question (what is metered).
3. **Boundary to Family D.** A proprietary text drafted individually for one
   vendor is a custom license and follows Family D drafting discipline; Family
   C membership describes the business model, not the drafting quality.

## 7. Do / Don't

**Do:** classify shared-source programs as proprietary with source visibility;
check evaluation and NFR grants for scope and time limits before any
production use; verify export-control and audit clauses in enterprise
agreements.

**Don't:** assume source visibility creates usage rights; treat a missing
license file as public domain; conflate the pricing metric with the legal
grant; reuse a vendor EULA as a template for another product without Family D
drafting discipline and legal review.
