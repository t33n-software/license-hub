# Apache License 2.0 (`Apache-2.0`)

Family: A — Open Source, permissive (A1). Canon:
[license taxonomy canon](../README.md); family reference:
[permissive README](README.md).

- Template: [`templates/permissive/apache-2.0/Apache-2.0-1.0.0.hbs`](../../../templates/permissive/apache-2.0/Apache-2.0-1.0.0.hbs)
- Steward/status: Apache Software Foundation; OSI-approved; the enterprise
  default permissive license.

## Grant, conditions, restrictions

- Grant: use, reproduction, modification, preparation of derivative works,
  public display and performance, sublicensing, and distribution, in source
  or object form, worldwide, royalty-free, non-exclusive, perpetual, and
  irrevocable.
- Conditions: give a copy of the license with every copy; cause modified
  files to carry prominent change notices; retain all copyright, patent,
  trademark, and attribution notices; preserve the contents of a NOTICE file
  in derivative distributions.
- Restrictions: no trademark rights are granted; the license names the
  Apache-2.0 text reproduction duty.

## Patent position

Express patent grant with retaliation: every contributor grants a perpetual,
worldwide, non-exclusive, royalty-free patent license covering claims
necessarily infringed by its contributions; the patent license terminates for
a licensee who institutes patent litigation alleging the work infringes a
patent.

## Compatibility

Compatible one-way into GPLv3; incompatible with GPLv2-only because of the
patent retaliation clause. Combinations are answered from the canonical
compatibility matrix in the canon.

## Adoption

Set `SPDX_LICENSE_IDENTIFIER` to `Apache-2.0` in `license.values.json`. The
template body is byte-identical to the canonical text; its appendix "How to
apply the Apache License to your work" intentionally keeps its bracketed
placeholder instructions verbatim, because those instructions describe the
per-file notice a tenant adds to source headers — they are not render
anchors.
