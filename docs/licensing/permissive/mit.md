# MIT License Family

Family: A — Open Source, permissive (A1). Canon:
[license taxonomy canon](../README.md); family reference:
[permissive README](README.md).

## MIT (`MIT`)

- Template: [`templates/permissive/mit/MIT-1.0.0.hbs`](../../../templates/permissive/mit/MIT-1.0.0.hbs)
- Steward/status: OSI-approved; the most widely used open-source license.
- Grant: use, copy, modify, merge, publish, distribute, sublicense, and sell,
  free of charge, for any purpose.
- Conditions: the copyright notice and the permission notice must be included
  in all copies or substantial portions.
- Restrictions: none beyond the notice condition; the software is provided
  as-is with a full warranty disclaimer and liability exclusion.
- Patent position: no grant.
- Compatibility: compatible everywhere; notices must be retained.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `MIT` in `license.values.json`;
  the template anchors the copyright line to `COPYRIGHT_YEAR` and
  `COPYRIGHT_HOLDER`.

## MIT No Attribution (`MIT-0`)

- Template: [`templates/permissive/mit-0/MIT-0-1.0.0.hbs`](../../../templates/permissive/mit-0/MIT-0-1.0.0.hbs)
- Steward/status: OSI-approved MIT variant without the attribution
  requirement.
- Grant: same scope as MIT.
- Conditions: none; even the notice retention of MIT is dropped.
- Patent position: no grant.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `MIT-0`; the template anchors the
  copyright line. Choose MIT-0 when attribution-free reuse is the goal while
  remaining inside a reviewed license text.

## MIT Modern Variant (`MIT-Modern-Variant`)

- Template: [`templates/permissive/mit-modern-variant/MIT-Modern-Variant-1.0.0.hbs`](../../../templates/permissive/mit-modern-variant/MIT-Modern-Variant-1.0.0.hbs)
- Steward/status: OSI-approved MIT variant drafted without the written-
  agreement phrasing of the original.
- Grant: use, copy, modify, and distribute for any purpose without royalty
  fees.
- Conditions: the copyright notice and the two disclaimer paragraphs must
  appear in all copies.
- Patent position: no grant.
- Adoption: set `SPDX_LICENSE_IDENTIFIER` to `MIT-Modern-Variant`; the
  canonical text references a copyright notice above the text without carrying
  one, so the template prepends the anchored copyright line ahead of the
  verbatim text.
