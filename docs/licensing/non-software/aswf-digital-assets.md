# ASWF Digital Assets License (`ASWF-Digital-Assets-1.0`, `ASWF-Digital-Assets-1.1`)

Family: F — Non-Software, F5 (AI/ML and media assets). Canon:
[license taxonomy canon](../README.md); family reference:
[non-software README](README.md).

The Academy Software Foundation license for digital assets (media, test
footage, reference content). Both referenced versions carry templates.

## ASWF-Digital-Assets-1.0 (`ASWF-Digital-Assets-1.0`)

- Template: [`templates/non-software/aswf-digital-assets-1.0/ASWF-Digital-Assets-1.0-1.0.0.hbs`](../../../templates/non-software/aswf-digital-assets-1.0/ASWF-Digital-Assets-1.0-1.0.0.hbs)
- Grant: use, copy, modify, and distribute the digital assets, including
  commercially.
- Conditions: the asset name identifies the assets; outputs may refer to the
  asset name only to identify the assets.
- Restrictions: no use of the copyright holder's names or trademarks to
  promote derived products without prior written permission.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `ASWF-Digital-Assets-1.0`.

## ASWF-Digital-Assets-1.1 (`ASWF-Digital-Assets-1.1`)

- Template: [`templates/non-software/aswf-digital-assets-1.1/ASWF-Digital-Assets-1.1-1.0.0.hbs`](../../../templates/non-software/aswf-digital-assets-1.1/ASWF-Digital-Assets-1.1-1.0.0.hbs)
- The 1.1 update: clarified trademark and attribution wording on the same
  grant structure.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `ASWF-Digital-Assets-1.1`.

## Shared notes

- Patent position: no grant (patent-silent) for both.
- Both templates anchor the asset name and owner notice lines to the
  `PROJECT_NAME`, `COPYRIGHT_YEAR`, and `COPYRIGHT_HOLDER` anchors; the rest
  of each text is byte-identical to the canonical text.
- For AI model weights specifically, the vendor- and steward-specific
  instruments (RAIL/OpenRAIL, Llama, Gemma) are custom texts with a mandatory
  legal gate — see [README.md](README.md) Section 4.
