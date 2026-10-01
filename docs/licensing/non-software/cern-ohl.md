# CERN Open Hardware Licence v2 Family

Family: F — Non-Software, F3 (Open Hardware). Canon:
[license taxonomy canon](../README.md); family reference:
[non-software README](README.md).

The CERN Open Hardware Licence version 2 family covers hardware designs in
three reciprocity strengths; all v2 variants are OSI-approved. The legacy
CERN-OHL-1.1 and 1.2 texts carry no template.

## CERN-OHL-P-2.0 (`CERN-OHL-P-2.0`) — permissive

- Template: [`templates/non-software/cern-ohl-p-2.0/CERN-OHL-P-2.0-1.0.0.hbs`](../../../templates/non-software/cern-ohl-p-2.0/CERN-OHL-P-2.0-1.0.0.hbs)
- Grant: use, copy, modify, and manufacture the hardware design, including
  commercially, without reciprocity.
- Conditions: notices and the license text accompany distributions.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `CERN-OHL-P-2.0`.

## CERN-OHL-W-2.0 (`CERN-OHL-W-2.0`) — weakly reciprocal

- Template: [`templates/non-software/cern-ohl-w-2.0/CERN-OHL-W-2.0-1.0.0.hbs`](../../../templates/non-software/cern-ohl-w-2.0/CERN-OHL-W-2.0-1.0.0.hbs)
- Grant: the same grant, with weak reciprocity — modifications to the licensed
  design files remain under the CERN-OHL-W, while separately developed
  components keep their own terms.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `CERN-OHL-W-2.0`.

## CERN-OHL-S-2.0 (`CERN-OHL-S-2.0`) — strongly reciprocal

- Template: [`templates/non-software/cern-ohl-s-2.0/CERN-OHL-S-2.0-1.0.0.hbs`](../../../templates/non-software/cern-ohl-s-2.0/CERN-OHL-S-2.0-1.0.0.hbs)
- Grant: the same grant, with strong reciprocity — products built on the
  design must make the complete corresponding source available under the
  CERN-OHL-S.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `CERN-OHL-S-2.0`. Treat it with
  the same combination analysis duty as software strong copyleft.

## Shared notes

- Patent position: the v2 family includes an express patent license from
  contributors for the licensed designs.
- All three template bodies are byte-identical to their canonical texts.
