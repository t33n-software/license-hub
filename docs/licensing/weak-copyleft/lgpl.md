# GNU Lesser General Public License (LGPL)

Family: A — Open Source, weak copyleft (A2). Canon:
[license taxonomy canon](../README.md); family reference:
[weak-copyleft README](README.md).

## LGPL v2.1 (`LGPL-2.1-only`, `LGPL-2.1-or-later`)

- Template: [`templates/weak-copyleft/lgpl-2.1/LGPL-2.1-1.0.0.hbs`](../../../templates/weak-copyleft/lgpl-2.1/LGPL-2.1-1.0.0.hbs)
- Steward/status: Free Software Foundation; OSI-approved; the classic
  library-level copyleft.

### Grant, conditions, restrictions

- Grant: use, copy, modify, and distribute the library.
- Conditions: distribution triggers source obligations for the library;
  modifications to the library must be published under the same terms;
  combined works must allow relinking — dynamic linking, or static linking
  with object files or an equivalent relink mechanism.
- Restrictions: the copyleft stays scoped to the library; the larger work
  remains under its own terms when the linking conditions are met.

### Patent position

No grant (patent-silent, like GPL-2.0).

### Adoption

Set `SPDX_LICENSE_IDENTIFIER` to `LGPL-2.1-only` or `LGPL-2.1-or-later` — the
text is shared and the suffix is the tenant's notice decision, declared in the
per-file source headers. The template body is byte-identical to the canonical
text; its how-to-apply section keeps its placeholder instructions verbatim.

## LGPL v3.0 (`LGPL-3.0-only`, `LGPL-3.0-or-later`)

- Template: [`templates/weak-copyleft/lgpl-3.0/LGPL-3.0-1.0.0.hbs`](../../../templates/weak-copyleft/lgpl-3.0/LGPL-3.0-1.0.0.hbs)
- Steward/status: Free Software Foundation; OSI-approved; GPLv3-based
  library-level copyleft.

### Grant, conditions, restrictions

- Grant: the GPLv3 grant plus the additional permissions of the LGPL.
- Conditions: the LGPL-2.1-style library conditions restated on the GPLv3
  base, including the relink mechanism and the combined-work notices.
- Restrictions: GPLv3 terms apply to the library itself, including
  anti-tivoization.

### Patent position

Express grant with retaliation through the GPLv3 base (§11).

### Compatibility

Compatible with GPLv3. Combinations are answered from the canonical
compatibility matrix in the canon.

### Adoption

Set `SPDX_LICENSE_IDENTIFIER` to `LGPL-3.0-only` or `LGPL-3.0-or-later`. The
template body is byte-identical to the canonical text; its how-to-apply
section keeps its placeholder instructions verbatim.
