# Creative Commons 4.0 Suite

Family: F — Non-Software, F1 (Content and Documentation). Canon:
[license taxonomy canon](../README.md); family reference:
[non-software README](README.md).

The Creative Commons 4.0 International suite provides six licenses composed
from the four elements BY (attribution), SA (share-alike), NC
(non-commercial), and ND (no derivatives). **None of them may be used for
software** — CC0 excepted, which lives in
[../public-domain-dedication/cc0-1.0.md](../public-domain-dedication/cc0-1.0.md).
Legacy versions 1.0–3.0 including ported jurisdiction variants exist on the
SPDX list and carry no template.

## CC-BY-4.0 (`CC-BY-4.0`)

- Template: [`templates/non-software/cc-by-4.0/CC-BY-4.0-1.0.0.hbs`](../../../templates/non-software/cc-by-4.0/CC-BY-4.0-1.0.0.hbs)
- The most permissive member: attribution only.
- Conditions: attribution with link, license indication, and change marking.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `CC-BY-4.0`.

## CC-BY-SA-4.0 (`CC-BY-SA-4.0`)

- Template: [`templates/non-software/cc-by-sa-4.0/CC-BY-SA-4.0-1.0.0.hbs`](../../../templates/non-software/cc-by-sa-4.0/CC-BY-SA-4.0-1.0.0.hbs)
- Attribution plus share-alike — the CC copyleft; adaptations carry the same
  license. One-way compatible into GPLv3 per the canonical compatibility
  matrix.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `CC-BY-SA-4.0`.

## CC-BY-ND-4.0 (`CC-BY-ND-4.0`)

- Template: [`templates/non-software/cc-by-nd-4.0/CC-BY-ND-4.0-1.0.0.hbs`](../../../templates/non-software/cc-by-nd-4.0/CC-BY-ND-4.0-1.0.0.hbs)
- Attribution plus no derivatives: the work may be shared but adaptations may
  not be shared. This is a restriction, not open licensing — declare it as
  such.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `CC-BY-ND-4.0`.

## CC-BY-NC-4.0 (`CC-BY-NC-4.0`)

- Template: [`templates/non-software/cc-by-nc-4.0/CC-BY-NC-4.0-1.0.0.hbs`](../../../templates/non-software/cc-by-nc-4.0/CC-BY-NC-4.0-1.0.0.hbs)
- Attribution plus non-commercial: commercial use requires a separate
  license. This is a restriction, not open licensing.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `CC-BY-NC-4.0`.

## CC-BY-NC-SA-4.0 (`CC-BY-NC-SA-4.0`)

- Template: [`templates/non-software/cc-by-nc-sa-4.0/CC-BY-NC-SA-4.0-1.0.0.hbs`](../../../templates/non-software/cc-by-nc-sa-4.0/CC-BY-NC-SA-4.0-1.0.0.hbs)
- Non-commercial plus share-alike.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `CC-BY-NC-SA-4.0`.

## CC-BY-NC-ND-4.0 (`CC-BY-NC-ND-4.0`)

- Template: [`templates/non-software/cc-by-nc-nd-4.0/CC-BY-NC-ND-4.0-1.0.0.hbs`](../../../templates/non-software/cc-by-nc-nd-4.0/CC-BY-NC-ND-4.0-1.0.0.hbs)
- The most restrictive member: attribution, non-commercial, no derivatives.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `CC-BY-NC-ND-4.0`.

## Shared notes (all six)

- Patent position: no grant (the CC 4.0 suite is patent-silent).
- All six template bodies are byte-identical to their canonical texts.
- Attribution form: retain the copyright notice, a link to the license, an
  indication of the license, and a change marker for modified material.
