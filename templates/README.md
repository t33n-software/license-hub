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

Empty families carry a `.gitkeep` marker and are reserved for later governed
additions. New templates are added only through the governed ticket workflow
and require legal review before release (see `POLICY.md`).

## Active templates

| Template | Family | Latest version | Release tag |
|----------|--------|----------------|-------------|
| `custom/norepublish/NoRepublish-1.0.0.hbs` | D — Custom | 1.0.0 | `norepublish/v1.0.0` |
