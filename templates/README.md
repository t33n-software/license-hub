# License template families

This directory is the canonical home of every license template family of the
organization. Each family directory holds one subdirectory per template,
versioned by file name (`<Name>-<semver>.hbs`) and released as an immutable,
digest-pinned template release.

## Family taxonomy

| Directory | Family | Content rule |
|-----------|--------|--------------|
| `permissive/` | A — Open Source, permissive | OSI-approved permissive templates |
| `weak-copyleft/` | A — Open Source, weak copyleft | OSI-approved weak-copyleft templates |
| `strong-copyleft/` | A — Open Source, strong copyleft | OSI-approved strong-copyleft templates |
| `network-copyleft/` | A — Open Source, network copyleft | OSI-approved network-copyleft templates |
| `source-available/` | B — Source-Available | Non-OSI source-available standard templates |
| `proprietary/` | C — Proprietary | Closed-source agreement templates |
| `custom/` | D — Custom | Organization-drafted `LicenseRef` templates |
| `public-domain-dedication/` | E — Public-Domain dedication | Dedication templates with jurisdiction caveats |
| `non-software/` | F — Non-software artifacts | Documentation, data, font, hardware templates |
| `multi-licensing/` | G — Combination mechanisms | Dual-/multi-licensing arrangement templates |

Family G carries no license-text templates: the combination mechanisms own no
texts of their own and are served by composing member templates through SPDX
expressions — see
[`../docs/licensing/multi-licensing/README.md`](../docs/licensing/multi-licensing/README.md)
and the family note in [`multi-licensing/README.md`](multi-licensing/README.md).

Every template directory carries the template file, a `CHANGELOG.md` with the
legal SemVer history, and a `README.md` documentation seam that re-references
the canonical template documentation under `../docs/licensing/<family>/`.
Template bodies never carry comments: rendering is pure placeholder
substitution, so any comment would be emitted into the rendered legal text of
every tenant.

New templates are added only through the governed ticket workflow and require
legal review before release (see [`../POLICY.md`](../POLICY.md)). The complete
template contract — placement, naming, anchors, render layout, and SemVer —
lives in
[`../docs/infrastructure/template-contract.md`](../docs/infrastructure/template-contract.md).

## Active templates

| Template | Family | Latest version | Release tag |
|----------|--------|----------------|-------------|
| `templates/permissive/0bsd/0BSD-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/afl-3.0/AFL-3.0-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/apache-2.0/Apache-2.0-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/artistic-2.0/Artistic-2.0-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/bsd-1-clause/BSD-1-Clause-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/bsd-2-clause/BSD-2-Clause-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/bsd-2-clause-patent/BSD-2-Clause-Patent-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/bsd-3-clause/BSD-3-Clause-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/bsd-3-clause-clear/BSD-3-Clause-Clear-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/blueoak-1.0.0/BlueOak-1.0.0-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/bsl-1.0/BSL-1.0-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/ecl-2.0/ECL-2.0-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/esa-pl-permissive-2.4/ESA-PL-permissive-2.4-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/fair/Fair-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/fsfap/FSFAP-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/isc/ISC-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/mit/MIT-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/mit-0/MIT-0-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/mit-modern-variant/MIT-Modern-Variant-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/miros/MirOS-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/ms-pl/MS-PL-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/mulanpsl-2.0/MulanPSL-2.0-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/permissive/zlib/Zlib-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/weak-copyleft/apsl-2.0/APSL-2.0-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/weak-copyleft/cddl-1.0/CDDL-1.0-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/weak-copyleft/cddl-1.1/CDDL-1.1-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/weak-copyleft/cecill-2.1/CECILL-2.1-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/weak-copyleft/cpal-1.0/CPAL-1.0-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/weak-copyleft/epl-2.0/EPL-2.0-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/weak-copyleft/erlpl-1.1/ErlPL-1.1-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/weak-copyleft/esa-pl-weak-copyleft-2.4/ESA-PL-weak-copyleft-2.4-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/weak-copyleft/eupl-1.2/EUPL-1.2-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/weak-copyleft/lgpl-2.1/LGPL-2.1-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/weak-copyleft/lgpl-3.0/LGPL-3.0-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/permissive/liliq-p-1.1/LiLiQ-P-1.1-1.0.0.hbs` | A — Open Source, permissive | 1.0.0 | — |
| `templates/strong-copyleft/liliq-r-1.1/LiLiQ-R-1.1-1.0.0.hbs` | A — Open Source, strong copyleft | 1.0.0 | — |
| `templates/network-copyleft/liliq-rplus-1.1/LiLiQ-Rplus-1.1-1.0.0.hbs` | A — Open Source, network copyleft | 1.0.0 | — |
| `templates/weak-copyleft/mpl-2.0/MPL-2.0-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/weak-copyleft/ms-rl/MS-RL-1.0.0.hbs` | A — Open Source, weak copyleft | 1.0.0 | — |
| `templates/strong-copyleft/gpl-2.0/GPL-2.0-1.0.0.hbs` | A — Open Source, strong copyleft | 1.0.0 | — |
| `templates/strong-copyleft/gpl-3.0/GPL-3.0-1.0.0.hbs` | A — Open Source, strong copyleft | 1.0.0 | — |
| `templates/strong-copyleft/osl-3.0/OSL-3.0-1.0.0.hbs` | A — Open Source, strong copyleft | 1.0.0 | — |
| `templates/strong-copyleft/esa-pl-strong-copyleft-2.4/ESA-PL-strong-copyleft-2.4-1.0.0.hbs` | A — Open Source, strong copyleft | 1.0.0 | — |
| `templates/network-copyleft/agpl-3.0/AGPL-3.0-1.0.0.hbs` | A — Open Source, network copyleft | 1.0.0 | — |
| `templates/network-copyleft/cal-1.0/CAL-1.0-1.0.0.hbs` | A — Open Source, network copyleft | 1.0.0 | — |
| `templates/source-available/sspl-1.0/SSPL-1.0-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/elastic-2.0/Elastic-2.0-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/fsl-1.1-alv2/FSL-1.1-ALv2-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/fsl-1.1-mit/FSL-1.1-MIT-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/hippocratic-2.1/Hippocratic-2.1-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/json/JSON-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/polyform-noncommercial-1.0.0/PolyForm-Noncommercial-1.0.0-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/polyform-shield-1.0.0/PolyForm-Shield-1.0.0-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/polyform-perimeter-1.0.1/PolyForm-Perimeter-1.0.1-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/polyform-strict-1.0.0/PolyForm-Strict-1.0.0-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/polyform-small-business-1.0.0/PolyForm-Small-Business-1.0.0-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/polyform-free-trial-1.0.0/PolyForm-Free-Trial-1.0.0-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/source-available/polyform-internal-use-1.0.0/PolyForm-Internal-Use-1.0.0-1.0.0.hbs` | B — Source-Available | 1.0.0 | — |
| `templates/proprietary/all-rights-reserved/AllRightsReserved-1.0.0.hbs` | C — Proprietary | 1.0.0 | — |
| `templates/public-domain-dedication/cc0-1.0/CC0-1.0-1.0.0.hbs` | E — Public-Domain dedication | 1.0.0 | — |
| `templates/public-domain-dedication/unlicense/Unlicense-1.0.0.hbs` | E — Public-Domain dedication | 1.0.0 | — |
| `templates/public-domain-dedication/wtfpl/WTFPL-1.0.0.hbs` | E — Public-Domain dedication | 1.0.0 | — |
| `templates/public-domain-dedication/blessing/blessing-1.0.0.hbs` | E — Public-Domain dedication | 1.0.0 | — |
| `templates/non-software/cc-by-4.0/CC-BY-4.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/cc-by-sa-4.0/CC-BY-SA-4.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/cc-by-nd-4.0/CC-BY-ND-4.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/cc-by-nc-4.0/CC-BY-NC-4.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/cc-by-nc-sa-4.0/CC-BY-NC-SA-4.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/cc-by-nc-nd-4.0/CC-BY-NC-ND-4.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/gfdl-1.3/GFDL-1.3-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/lal-1.3/LAL-1.3-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/freebsd-doc/FreeBSD-DOC-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/odbl-1.0/ODbL-1.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/odc-by-1.0/ODC-By-1.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/pddl-1.0/PDDL-1.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/cdla-permissive-1.0/CDLA-Permissive-1.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/cdla-permissive-2.0/CDLA-Permissive-2.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/cdla-sharing-1.0/CDLA-Sharing-1.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/c-uda-1.0/C-UDA-1.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/dl-de-by-2.0/DL-DE-BY-2.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/dl-de-zero-2.0/DL-DE-ZERO-2.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/etalab-2.0/etalab-2.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/cern-ohl-p-2.0/CERN-OHL-P-2.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/cern-ohl-w-2.0/CERN-OHL-W-2.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/cern-ohl-s-2.0/CERN-OHL-S-2.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/tapr-ohl-1.0/TAPR-OHL-1.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/shl-2.0/SHL-2.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/shl-2.1/SHL-2.1-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/ofl-1.1/OFL-1.1-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/ipa/IPA-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/aswf-digital-assets-1.0/ASWF-Digital-Assets-1.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/aswf-digital-assets-1.1/ASWF-Digital-Assets-1.1-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/non-software/community-spec-1.0/Community-Spec-1.0-1.0.0.hbs` | F — Non-software artifacts | 1.0.0 | — |
| `templates/custom/norepublish/NoRepublish-1.0.0.hbs` | D — Custom | 1.0.0 | `norepublish/v1.0.0` |

A `—` release tag marks a template whose first legal-gated release is pending;
the tag will be `<template-dir>/v<semver>` once the immutable template release
lane publishes it after legal-counsel review.