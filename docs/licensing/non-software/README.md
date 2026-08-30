# Non-Software and Adjacent Artifact Licenses — Family F

This document is the canonical family reference for licenses of artifacts that
are not software. The family taxonomy, legal foundations, compatibility
matrix, and governance gates are owned by the
[license taxonomy canon](../README.md); this document owns the
adjacent-artifact semantics and the template inventory of the family.

## 1. Family definition and artifact-type routing

Family F covers content, documentation, data, databases, hardware designs,
fonts, AI/ML model weights, and standards documents. The artifact type is a
primary decision axis: every licensing question first establishes the artifact
class, because software licenses and adjacent-artifact licenses are not
interchangeable in either direction.

**Hard rule:** Creative Commons licenses must not be used for software; CC0 is
the only CC tool acceptable for code (see
[../public-domain-dedication/cc0-1.0.md](../public-domain-dedication/cc0-1.0.md)).

## 2. Subfamilies

- **F1 Content and documentation** — the six Creative Commons 4.0 licenses,
  GFDL, Licence Art Libre, FreeBSD Documentation License.
- **F2 Data and databases** — ODbL, ODC-BY, PDDL, CDLA-Permissive,
  CDLA-Sharing, C-UDA, DL-DE, etalab.
- **F3 Open hardware** — CERN-OHL v2 family, TAPR, Solderpad.
- **F4 Fonts** — SIL OFL, IPA.
- **F5 AI/ML models** — ASWF Digital Assets; the RAIL/OpenRAIL, Llama, and
  Gemma instruments are vendor- or steward-specific custom texts (Section 4).
- **F6 Standards and specifications** — Community Specification License.

## 3. Template inventory

| Template | SPDX ID | Artifact class | Canonical documentation |
|----------|---------|----------------|-------------------------|
| `non-software/cc-by-4.0/CC-BY-4.0-1.0.0.hbs` | `CC-BY-4.0` | Content | [cc-by.md](cc-by.md) |
| `non-software/cc-by-sa-4.0/CC-BY-SA-4.0-1.0.0.hbs` | `CC-BY-SA-4.0` | Content (share-alike) | [cc-by.md](cc-by.md) |
| `non-software/cc-by-nd-4.0/CC-BY-ND-4.0-1.0.0.hbs` | `CC-BY-ND-4.0` | Content (no derivatives) | [cc-by.md](cc-by.md) |
| `non-software/cc-by-nc-4.0/CC-BY-NC-4.0-1.0.0.hbs` | `CC-BY-NC-4.0` | Content (non-commercial) | [cc-by.md](cc-by.md) |
| `non-software/cc-by-nc-sa-4.0/CC-BY-NC-SA-4.0-1.0.0.hbs` | `CC-BY-NC-SA-4.0` | Content (NC + SA) | [cc-by.md](cc-by.md) |
| `non-software/cc-by-nc-nd-4.0/CC-BY-NC-ND-4.0-1.0.0.hbs` | `CC-BY-NC-ND-4.0` | Content (most restrictive) | [cc-by.md](cc-by.md) |
| `non-software/gfdl-1.3/GFDL-1.3-1.0.0.hbs` | `GFDL-1.3-only`, `GFDL-1.3-or-later` | Documentation | [gfdl-1.3.md](gfdl-1.3.md) |
| `non-software/lal-1.3/LAL-1.3-1.0.0.hbs` | `LAL-1.3` | Free art | [lal-1.3.md](lal-1.3.md) |
| `non-software/freebsd-doc/FreeBSD-DOC-1.0.0.hbs` | `FreeBSD-DOC` | Documentation | [freebsd-doc.md](freebsd-doc.md) |
| `non-software/odbl-1.0/ODbL-1.0-1.0.0.hbs` | `ODbL-1.0` | Databases (copyleft) | [odbl-1.0.md](odbl-1.0.md) |
| `non-software/odc-by-1.0/ODC-By-1.0-1.0.0.hbs` | `ODC-By-1.0` | Databases (attribution) | [odc-by-1.0.md](odc-by-1.0.md) |
| `non-software/pddl-1.0/PDDL-1.0-1.0.0.hbs` | `PDDL-1.0` | Data (public-domain-equivalent) | [pddl-1.0.md](pddl-1.0.md) |
| `non-software/cdla-permissive-1.0/CDLA-Permissive-1.0-1.0.0.hbs` | `CDLA-Permissive-1.0` | Data | [cdla.md](cdla.md) |
| `non-software/cdla-permissive-2.0/CDLA-Permissive-2.0-1.0.0.hbs` | `CDLA-Permissive-2.0` | Data | [cdla.md](cdla.md) |
| `non-software/cdla-sharing-1.0/CDLA-Sharing-1.0-1.0.0.hbs` | `CDLA-Sharing-1.0` | Data (copyleft) | [cdla.md](cdla.md) |
| `non-software/c-uda-1.0/C-UDA-1.0-1.0.0.hbs` | `C-UDA-1.0` | AI training data | [c-uda-1.0.md](c-uda-1.0.md) |
| `non-software/dl-de-by-2.0/DL-DE-BY-2.0-1.0.0.hbs` | `DL-DE-BY-2.0` | German government data | [dl-de.md](dl-de.md) |
| `non-software/dl-de-zero-2.0/DL-DE-ZERO-2.0-1.0.0.hbs` | `DL-DE-ZERO-2.0` | German government data (zero conditions) | [dl-de.md](dl-de.md) |
| `non-software/etalab-2.0/etalab-2.0-1.0.0.hbs` | `etalab-2.0` | French government data | [etalab-2.0.md](etalab-2.0.md) |
| `non-software/cern-ohl-p-2.0/CERN-OHL-P-2.0-1.0.0.hbs` | `CERN-OHL-P-2.0` | Hardware (permissive) | [cern-ohl.md](cern-ohl.md) |
| `non-software/cern-ohl-w-2.0/CERN-OHL-W-2.0-1.0.0.hbs` | `CERN-OHL-W-2.0` | Hardware (weakly reciprocal) | [cern-ohl.md](cern-ohl.md) |
| `non-software/cern-ohl-s-2.0/CERN-OHL-S-2.0-1.0.0.hbs` | `CERN-OHL-S-2.0` | Hardware (strongly reciprocal) | [cern-ohl.md](cern-ohl.md) |
| `non-software/tapr-ohl-1.0/TAPR-OHL-1.0-1.0.0.hbs` | `TAPR-OHL-1.0` | Hardware | [tapr-ohl-1.0.md](tapr-ohl-1.0.md) |
| `non-software/shl-2.0/SHL-2.0-1.0.0.hbs` | `SHL-2.0` | Hardware | [shl.md](shl.md) |
| `non-software/shl-2.1/SHL-2.1-1.0.0.hbs` | `SHL-2.1` | Hardware | [shl.md](shl.md) |
| `non-software/ofl-1.1/OFL-1.1-1.0.0.hbs` | `OFL-1.1` | Fonts | [ofl-1.1.md](ofl-1.1.md) |
| `non-software/ipa/IPA-1.0.0.hbs` | `IPA` | Fonts | [ipa.md](ipa.md) |
| `non-software/aswf-digital-assets-1.0/ASWF-Digital-Assets-1.0-1.0.0.hbs` | `ASWF-Digital-Assets-1.0` | Media assets | [aswf-digital-assets.md](aswf-digital-assets.md) |
| `non-software/aswf-digital-assets-1.1/ASWF-Digital-Assets-1.1-1.0.0.hbs` | `ASWF-Digital-Assets-1.1` | Media assets | [aswf-digital-assets.md](aswf-digital-assets.md) |
| `non-software/community-spec-1.0/Community-Spec-1.0-1.0.0.hbs` | `Community-Spec-1.0` | Standards/specifications | [community-spec-1.0.md](community-spec-1.0.md) |

## 4. Referenced members without a template

| License | SPDX ID | Exclusion rationale |
|---------|---------|---------------------|
| Creative Commons 1.0–3.0 (incl. ported variants) | `CC-BY-*-1.0/2.0/3.0` … | Legacy versions; the 4.0 suite is the current canonical form |
| GFDL 1.1 / 1.2 | `GFDL-1.1-*`, `GFDL-1.2-*` | Legacy; superseded by GFDL-1.3 |
| Licence Art Libre 1.2 | `LAL-1.2` | Legacy; superseded by LAL-1.3 |
| CERN-OHL 1.1 / 1.2 | `CERN-OHL-1.1`, `CERN-OHL-1.2` | Legacy; the v2 family is the current canonical form |
| SIL OFL 1.0 | `OFL-1.0` | Legacy; superseded by OFL-1.1 |
| Bitstream Vera / Charter / mplus / Lucida font licenses | various | Ecosystem-bound font licenses for specific historical fonts |
| BigScience BLOOM RAIL 1.0 / OpenRAIL-M | custom family | AI-model use-restriction texts with variant selection and a mandatory legal gate; adopting one is a per-project legal decision, not a standard template |
| Llama Community License | custom (Meta) | Vendor-owned custom text with an acceptable-use policy and a 700M-MAU restriction clause |
| Gemma Terms of Use | custom (Google) | Vendor-owned custom use policy |
| W3C document licenses | — | Organization-bound document notices, not general templates |
| IETF Trust Legal Provisions | — | Organization-bound document notices |
| ISO permission notices | — | Organization-bound document notices |

## 5. Family conventions

1. **Artifact-first routing.** The artifact class is fixed before any license
   comparison, and software is never licensed with Family F instruments (CC0
   excepted as a dedication).
2. **Copyleft parity awareness.** Share-alike semantics exist inside Family F
   (CC-BY-SA for content, ODbL for databases, CERN-OHL-S for hardware,
   CDLA-Sharing for data) and create the same combination analysis duty as
   software copyleft.
3. **NC/ND caution.** NC and ND elements remove the artifact from the
   open-definition space; they are legitimate choices but are always declared
   as restrictions, never as open licensing.

## 6. Do / Don't

**Do:** use CC-BY-4.0 or CC-BY-SA-4.0 for documentation and content; use
ODbL-1.0 for copyleft databases; use CERN-OHL v2 for open hardware; use
OFL-1.1 for fonts; check model-specific use policies before redistributing AI
model weights.

**Don't:** apply CC licenses to software; treat model weights as covered by
the code license of the training framework; mix ODbL data into proprietary
databases without share-alike analysis; ignore reserved font names when
distributing OFL fonts under a new name.
