# GNU General Public License (GPL)

Family: A — Open Source, strong copyleft (A3). Canon:
[license taxonomy canon](../README.md); family reference:
[strong-copyleft README](README.md).

## GPL v2.0 (`GPL-2.0-only`, `GPL-2.0-or-later`)

- Template: [`templates/strong-copyleft/gpl-2.0/GPL-2.0-1.0.0.hbs`](../../../templates/strong-copyleft/gpl-2.0/GPL-2.0-1.0.0.hbs)
- Steward/status: Free Software Foundation; OSI-approved; the reference
  strong copyleft (the Linux kernel is v2-only).

### Grant, conditions, restrictions

- Grant: use, copy, modify, and distribute the program.
- Conditions: distribution of the program or a work based on it requires the
  full corresponding source under the GPL, with notices and a copy of the
  license.
- Restrictions: no further restrictions may be imposed; derivative works stay
  under the GPL on distribution.

### Patent position

No express grant (patent-silent).

### Compatibility

GPLv2-only is incompatible with GPLv3-only and with Apache-2.0; the
"or-later" declaration decides v3 compatibility. Answered from the canonical
compatibility matrix.

### Adoption

Set `SPDX_LICENSE_IDENTIFIER` to `GPL-2.0-only` or `GPL-2.0-or-later` — the
text is shared and the suffix is the tenant's notice decision. The template
body is byte-identical to the canonical text; its how-to-apply section keeps
its placeholder instructions verbatim.

## GPL v3.0 (`GPL-3.0-only`, `GPL-3.0-or-later`)

- Template: [`templates/strong-copyleft/gpl-3.0/GPL-3.0-1.0.0.hbs`](../../../templates/strong-copyleft/gpl-3.0/GPL-3.0-1.0.0.hbs)
- Steward/status: Free Software Foundation; OSI-approved; the modern strong
  copyleft.

### Grant, conditions, restrictions

- Grant: the GPL grant plus the v3 clarifications.
- Conditions: the strong-copyleft source duties on distribution.
- Restrictions: anti-tivoization (§6 — installation information for user
  products), no additional restrictions, and a cure period for violations
  (§8).

### Patent position

Express grant with retaliation (§11).

### Compatibility

One-way compatible with Apache-2.0 (Apache-2.0 code into a GPLv3 project);
linkable with AGPL-3.0 both ways via the §13 clauses. Answered from the
canonical compatibility matrix.

### Adoption

Set `SPDX_LICENSE_IDENTIFIER` to `GPL-3.0-only` or `GPL-3.0-or-later`. The
template body is byte-identical to the canonical text; its how-to-apply
section keeps its placeholder instructions verbatim.
