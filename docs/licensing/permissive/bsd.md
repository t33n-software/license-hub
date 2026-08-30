# BSD License Family and 0BSD

Family: A — Open Source, permissive (A1). Canon:
[license taxonomy canon](../README.md); family reference:
[permissive README](README.md).

## BSD 1-Clause (`BSD-1-Clause`)

- Template: [`templates/permissive/bsd-1-clause/BSD-1-Clause-1.0.0.hbs`](../../../templates/permissive/bsd-1-clause/BSD-1-Clause-1.0.0.hbs)
- Steward/status: OSI-approved; the minimal BSD form.
- Grant: redistribution and use in source and binary forms, with or without
  modification.
- Conditions: source redistributions must retain the copyright notice and the
  disclaimer.
- Patent position: no grant.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `BSD-1-Clause`.

## BSD 2-Clause "Simplified" (`BSD-2-Clause`)

- Template: [`templates/permissive/bsd-2-clause/BSD-2-Clause-1.0.0.hbs`](../../../templates/permissive/bsd-2-clause/BSD-2-Clause-1.0.0.hbs)
- Steward/status: OSI-approved; the classic permissive default.
- Grant: redistribution and use in source and binary forms, with or without
  modification.
- Conditions: source redistributions retain the notice; binary
  redistributions reproduce the notice in the documentation.
- Patent position: no grant.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `BSD-2-Clause`.

## BSD 2-Clause Plus Patent (`BSD-2-Clause-Patent`)

- Template: [`templates/permissive/bsd-2-clause-patent/BSD-2-Clause-Patent-1.0.0.hbs`](../../../templates/permissive/bsd-2-clause-patent/BSD-2-Clause-Patent-1.0.0.hbs)
- Steward/status: OSI-approved; BSD-2-Clause plus an express patent grant.
- Grant: the BSD-2-Clause copyright grant plus a perpetual, worldwide,
  royalty-free patent license limited to claims necessarily infringed by a
  contribution alone or by its combination with the work it was contributed
  to.
- Conditions: the BSD-2-Clause notice conditions.
- Patent position: express grant (no retaliation clause).
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `BSD-2-Clause-Patent`. Choose it
  when a BSD-style text with an express patent grant is required without the
  Apache-2.0 NOTICE and modification-marking machinery.

## BSD 3-Clause "New"/"Revised" (`BSD-3-Clause`)

- Template: [`templates/permissive/bsd-3-clause/BSD-3-Clause-1.0.0.hbs`](../../../templates/permissive/bsd-3-clause/BSD-3-Clause-1.0.0.hbs)
- Steward/status: OSI-approved.
- Grant: the BSD-2-Clause grant.
- Conditions: the BSD-2-Clause notice conditions plus the non-endorsement
  clause — neither the name of the copyright holder nor the names of its
  contributors may be used to endorse or promote derived products without
  specific prior written permission.
- Patent position: no grant.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `BSD-3-Clause`.

## BSD 3-Clause Clear (`BSD-3-Clause-Clear`)

- Template: [`templates/permissive/bsd-3-clause-clear/BSD-3-Clause-Clear-1.0.0.hbs`](../../../templates/permissive/bsd-3-clause-clear/BSD-3-Clause-Clear-1.0.0.hbs)
- Steward/status: OSI-approved; the BSD-3-Clause form with an explicit patent
  non-grant sentence.
- Grant: the BSD-3-Clause grant.
- Conditions: the BSD-3-Clause conditions; the template anchors both the
  copyright line and the non-endorsement owner reference.
- Patent position: explicit non-grant — no express or implied patent licenses
  are granted.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `BSD-3-Clause-Clear`. Choose it
  when patent silence must be eliminated in the restrictive direction.

## Zero-Clause BSD (`0BSD`)

- Template: [`templates/permissive/0bsd/0BSD-1.0.0.hbs`](../../../templates/permissive/0bsd/0BSD-1.0.0.hbs)
- Steward/status: OSI-approved; a zero-condition permissive license.
- Grant: use, copy, modify, and distribute for any purpose, with or without
  fee — with no conditions at all.
- Conditions: none.
- Patent position: no grant (silent).
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `0BSD`. 0BSD is a license, not a
  waiver, which makes it the safest public-domain-equivalent choice for
  EU-centric projects where copyright cannot be fully waived; see
  [../public-domain-dedication/README.md](../public-domain-dedication/README.md).
  The template anchors the copyright line and drops the steward email token.
