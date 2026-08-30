# NoRepublish — Organization Custom License

Family: D — Custom / Self-Written (source-available in effect). **Not
OSI-approved open source.** Canon: [license taxonomy canon](../README.md);
family reference: [custom README](README.md).

- Template: [`templates/custom/norepublish/NoRepublish-1.0.0.hbs`](../../../templates/custom/norepublish/NoRepublish-1.0.0.hbs)
- Identifier: `LicenseRef-<LICENSE_ID>` per tenant, for example
  `LicenseRef-license-hub-NoRepublish-1.0` for this repository.
- Status: the canonical organization custom text, version 1.0; drafted under
  the 15-block canon and released under the legal-counsel gate.

## Content summary

- **Grant:** use, copy, clone, build, and modify for own use — commercially
  and non-commercially — including the official release artifacts published at
  the canonical source.
- **Restriction — no republication:** the software and any substantially
  similar work must not be published or redistributed, under any name, in any
  commerciality form; renaming or repackaging never removes substantial
  similarity.
- **Permitted distribution:** patches against the canonical source, and build
  scripts or tooling that obtain the software exclusively from the canonical
  source without bundling it.
- **Conditions:** retention of all copyright, license, and attribution
  notices; patches identify their modifications.
- **Patent:** grant for permitted use with retaliation on patent litigation.
- **Trademark:** no brand rights; renaming creates no distribution right.
- **Warranty and liability:** as-is with the mandatory EU/DE liability floor.
- **Termination:** automatic on republication; 30-day cure for other
  breaches.
- **Governing law, venue, language:** the organization defaults; English is
  the controlling language.
- **Versioning:** "-only" semantics per release; the licensor may relicense
  future releases.
- **Machine-readability:** `LicenseRef-<LICENSE_ID>` identification, REUSE
  3.2 placement, and per-file SPDX headers.

## Adoption

Leave `SPDX_LICENSE_IDENTIFIER` unset in `license.values.json` — the instance
renders as `LICENSE` plus `LICENSES/LicenseRef-<LICENSE_ID>.txt`. The full
adoption contract is owned by [../../adoption-guide.md](../../adoption-guide.md).
