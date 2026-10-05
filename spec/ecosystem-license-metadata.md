# Ecosystem License Metadata Specification

The license hub owns the render and verify machinery of canonical license
instances. Beyond the LICENSE file family, package ecosystems declare
license metadata on their own manifest surfaces, and this specification
defines how the hub derives, aligns, and proves those surfaces. It binds the
derivation law, the ecosystem adapter contracts, the fail-closed verification
classes, and the surgical alignment discipline.

## 1. Derivation law

The declaration is the truth; the manifest field is its deterministic
projection; the license lock is the single source of truth of the license
type.

1. **Declaration.** The project's ecosystem is contractually declared by the
   governed seam document (`git-governance.quality.json`, field
   `toolchain.language`); the organization fleet inventory mirrors the same
   declaration. The declaration is consumed read-only — never copied and
   never re-declared.
2. **Matrix row.** The declared language selects exactly one ecosystem row.
   The implemented row set: `node-typescript` selects the npm surface
   (`package.json`, field `license`); `python` selects the pyproject.toml
   surface (PEP 621/639: field `license` as the SPDX expression string and
   field `license-files` as the glob list); `rust` selects the Cargo.toml
   surface (the mutually exclusive `[package]` keys `license`, an SPDX 2.3
   expression, and `license-file`, the path to the license text inside the
   package);    `maven` selects the pom.xml surface (the project-level
   `licenses` element: every `license` element carries `name` and `url`);
   `nuget` selects the NuGet packaging surfaces (the `*.nuspec` `<license>`
   element and the `*.csproj` MSBuild pack properties, discovered in the
   render target directory); `composer` selects the composer.json surface
   (field `license`: an SPDX expression string or an array of SPDX
expressions; the `proprietary` form is the documented non-SPDX value);
    `ruby` selects the `*.gemspec` surface (the `spec.license` string
    assignment or the value-equal `spec.licenses` array assignment; every
    entry is a single SPDX identifier, `LicenseRef-<idstring>` is the
    documented non-SPDX entry form, and entries are limited to 64
characters); `elixir` selects the mix.exs surface (the `licenses` entry
   inside the `package` keyword list of the project configuration: a bracket
   list of SPDX identifiers or `LicenseRef-<idstring>` entries for custom
   licenses included in the package); `haskell` selects the `*.cabal` surface
   (the top-level fields of the package description: `license`, an SPDX
   expression case-sensitive since cabal-version: 2.2 where
   `LicenseRef-<idstring>` is the documented expression form for custom
   licenses, and `license-file`, the single filename of the license text —
   the `license-files` list form with exactly the projected file is the
   value-equal alternative spelling; discovered in the render target
   directory); Go selects no manifest field —
   the `LICENSE` file at the module root is the truth surface, and the
   adapter aligns nothing. Additional rows are grown content-driven when a
   consuming surface exists, with the field semantics verified against the
   official ecosystem documentation before the row is implemented; unverified
   semantics fail closed instead of guessing.
3. **Projection.** The expected field value is derived from the license lock
   (template, version, digest) together with the tenant values: a declared
   SPDX identifier projects itself; the file-based custom family projects
   the bound target form. The manifest field is never an independent source,
   and no process derives the lock from the manifest.

## 2. The npm alignment contract

The legal npm `license` forms are a valid SPDX license expression, the
custom-file reference form `SEE LICENSE IN <file>`, and `UNLICENSED`. The
bound organization target form for the file-based custom family is the
string `"SEE LICENSE IN LICENSE"` — the npm-canonical reference that points
the registry and every scanner at the committed `LICENSE` file at the
repository root. Declaring an SPDX `LicenseRef-...` identifier directly in
the field is rejected for the private custom family: the family is declared
by the referenced file, and the file reference keeps the registry
declaration and the committed text provably identical.

The npm manifest carries two other identity surfaces beside the license
field: `name` and `author`. Both belong to the package identity lane and are
never read or written by the license render.

## 3. The Python alignment contract

The Python row targets `pyproject.toml` under PEP 621/639. The bound
organization target form for the file-based custom family is the SPDX
sideload expression `LicenseRef-<LICENSE_ID>` (the tenant `LICENSE_ID`
value), paired with the license text glob list `["LICENSE"]` in the
`license-files` field; a declared `SPDX_LICENSE_IDENTIFIER` projects itself
into both surfaces.

The adapter's scanner reads the `[project]` table strictly and lexes the
remaining document structurally. A structural break anywhere — an
unterminated string, a malformed or duplicated header, a duplicate key, a
value the scanner cannot interpret — refuses the whole surface fail-closed.
The dotted-key forms that would create or shadow the license surface
outside the canonical table (a dotted key in `[project]` whose first
segment names a license-relevant key, and any root-level declaration of the
project name) are refused for the same reason.

The deprecated PEP 621 `license` table subkeys (`text`/`file`) are the same
license field in its deprecated value form: the render aligns them to the
string expression form. The `License ::` classifier entries are the
deprecated license declaration form: the verify lane reports them as a
finding whose remediation is an explicit tenant decision, and the alignment
never rewrites the classifiers array.

## 4. The Rust alignment contract

The Rust row targets `Cargo.toml`. The cargo license keys of the
`[package]` table are mutually exclusive: `license` carries an SPDX 2.3
license expression, and `license-file` carries the path to the license text
inside the package — each key is declared in lieu of the other. The bound
organization target form for the file-based custom family is
`license-file = "LICENSE"` — the canonical license text at the repository
root; a declared `SPDX_LICENSE_IDENTIFIER` projects itself into the
`license` expression key. Because the keys are exclusive, a form switch
replaces the declared key line with the projected assignment, and a
manifest that declares both keys is refused fail-closed: the surface cannot
be aligned or proven deterministically, and the resolution is an explicit
tenant decision.

The adapter's scanner reads the `[package]` table strictly and lexes the
remaining document structurally through the shared TOML skeleton lexer of
the package. A structural break anywhere — an unterminated string, a
malformed or duplicated header, a duplicate key, a value the scanner cannot
interpret — refuses the whole surface fail-closed. The dotted-key forms
that would create or shadow the license surface outside the canonical table
(a dotted key in `[package]` whose first segment names a license-relevant
key, and any root-level declaration of the package name) are refused for
the same reason.

## 5. The Maven alignment contract

The Maven row targets `pom.xml`. The project-level `licenses` element is
the license surface; every `license` element carries the `name` of the
license and the `url` of its text (plus the optional `distribution` and
`comments` elements; an SPDX identifier as the `name` is the recommended
form; child poms inherit the declared licenses). The bound organization
projection: the `name` carries a declared `SPDX_LICENSE_IDENTIFIER` itself,
and the file-based custom family carries the LICENSE_ID form; the `url`
carries the tenant `CANONICAL_SOURCE_URL`. A manifest that declares
multiple `license` elements is refused fail-closed — the projection carries
exactly one license family, and the resolution is an explicit tenant
decision.

The adapter's scanner reads the pom skeleton through the standard XML
decoder: the decoder validates the well-formedness fail-closed, and its
input-offset contract tiles every byte to a token, so every alignment span
is exact. A malformed manifest refuses the whole surface fail-closed; a
duplicate `licenses`, `name`, or `url` element is refused for the same
reason. The alignment inserts the missing `licenses` block after the
project start tag, inserts the missing `license` or name/url children
deterministically, and replaces only the char-data spans of diverging
values.

## 6. The .NET alignment contract

The .NET row targets the NuGet packaging manifests: the nuspec `<license>`
element (type `expression`, an SPDX expression per the official ABNF, or
type `file`, the path to the license text inside the package) and the MSBuild
pack properties (`PackageLicenseExpression` / `PackageLicenseFile`). The
manifest names are glob-form: the adapter discovers `*.nuspec` and
`*.csproj` manifests in the render target directory and aligns every
discovered manifest. The seam language token `nuget` selects the row.

The bound organization projection: a declared `SPDX_LICENSE_IDENTIFIER`
projects the expression form (`<license type="expression">` /
`<PackageLicenseExpression>`); the file-based custom family projects the
license-file form pointing at the committed `LICENSE` text at the repository
root (`<license type="file">` / `<PackageLicenseFile>`). The two forms are
exclusive: a form switch replaces the declared surface, and a manifest that
declares the license form more than once is refused fail-closed — the
resolution is an explicit tenant decision.

The deprecated declaration forms — the nuspec `licenseUrl` element and the
MSBuild `PackageLicenseUrl` property — are the license declaration in its
deprecated form: the verify lane reports them as a finding whose remediation
is an explicit tenant decision, and the alignment never rewrites them.

The adapter's scanners read the manifest skeletons through the standard XML
decoder: the decoder validates the well-formedness fail-closed, and its
input-offset contract tiles every byte to a token, so every alignment span
is exact. A malformed manifest refuses the whole surface fail-closed; a
nuspec without a metadata element, a root element that is not package (or
Project), a duplicate metadata or license-url element, and a license element
whose type attribute is neither the expression nor the file form are refused
for the same reason. The alignment inserts the missing license surface
deterministically and replaces only the char-data spans of diverging values.

## 7. The Composer alignment contract

The Composer row targets `composer.json`. The `license` field carries an
SPDX license expression as a string or an array of SPDX license expressions;
the `proprietary` form is the documented non-SPDX value, and the compound
expression forms (`A or B`, `A and B`) are legal string expressions. The
seam language token `composer` selects the row.

The bound organization projection: a declared `SPDX_LICENSE_IDENTIFIER`
projects itself; the file-based custom family projects the SPDX sideload
expression `LicenseRef-<LICENSE_ID>` (the tenant `LICENSE_ID` value) — the
composer field carries SPDX expressions, and the sideload form keeps the
license family identity. The projected form is the string expression: a
diverging string value or a one-element array is aligned to the projected
string form, and the one-element array carrying the projected value is the
value-equal alternative spelling of the same declaration. A license array
that declares more than one expression is refused fail-closed — the
projection carries exactly one license family, and the resolution is an
explicit tenant decision.

The adapter reuses the shared JSON scanner of the package: a manifest the
scanner cannot interpret is refused fail-closed, a license member declared
more than once is refused for the same reason, and the array scan refuses
every non-string entry, the empty array, and every malformed delimiter form
fail-closed.

## 8. The Ruby alignment contract

The Ruby row targets `*.gemspec` files. The license surface is the
assignment pair `spec.license = "MIT"` (the singular string form) and
`spec.licenses = ["MIT"]` (the plural array form) — two value-equal
spellings of the same declaration — and every entry is a single SPDX
identifier: compound expressions (for example `MIT OR Apache-2.0`) and
entries beyond the 64-character limit are not supported per entry, and
`LicenseRef-<idstring>` is the documented non-SPDX entry form. The manifest
names are glob-form: the adapter discovers `*.gemspec` files in the render
target directory and aligns every discovered file. The seam language token
`ruby` selects the row.

The bound organization projection: a declared `SPDX_LICENSE_IDENTIFIER`
projects itself; the file-based custom family projects the SPDX sideload
entry `LicenseRef-<LICENSE_ID>` (the tenant `LICENSE_ID` value). The
projected form is the singular string assignment: a diverging string value
is value-replaced, a diverging one-element array is normalized to the
singular assignment with the observed receiver preserved, and the
one-element array carrying the projected value is the value-equal
alternative spelling. An array that declares more than one entry is refused
fail-closed — the projection carries exactly one license family, and the
resolution is an explicit tenant decision — and so is a file that declares
the assignment more than once, in any mix of the two forms.

The adapter's scanner lexes the file structurally and reads the license
assignment sites strictly: line comments, string literals, percent
literals, heredocs, and `=begin`/`=end` block comments are context, so a
license assignment inside them is never observed, and a structural break
anywhere — an unterminated string, percent literal, heredoc, or block
comment — refuses the whole surface fail-closed. A license value must be an
escape-free, interpolation-free string literal, a bracket array of such
literals, or a `%w`/`%W` word array of such words; every other value form, a
dynamic suffix other than `.freeze`, an empty array, and an assignment
without a value are refused fail-closed. The insertion of a missing
assignment is anchored at the first `Gem::Specification.new do |<receiver>|`
block, uses the block parameter as the receiver and the block body's
indentation, and is proven by a re-scan: an anchor that does not produce a
provable assignment is refused fail-closed.

## 9. The Elixir alignment contract

The Elixir row targets `mix.exs`. The Hex package configuration is the
`package` keyword entry of the project keyword list, and its `licenses`
entry is the license surface: a bracket list of SPDX license identifiers,
where `LicenseRef-<idstring>` is the documented entry form for custom
licenses included in the package (the configuration `mix hex.build`
consumes). The seam language token `elixir` selects the row.

The bound organization projection: a declared `SPDX_LICENSE_IDENTIFIER`
projects itself; the file-based custom family projects the SPDX sideload
entry `LicenseRef-<LICENSE_ID>` (the tenant `LICENSE_ID` value). The
projected form is the single-entry list: a diverging single entry has its
string-literal span replaced with the list formatting preserved, and a
missing entry is inserted as the first member of the package keyword list.
A list that declares more than one entry is refused fail-closed — the
projection carries exactly one license family, and the resolution is an
explicit tenant decision — and so is a mix.exs that declares the package
configuration or the licenses entry more than once.

The adapter's scanner lexes the file structurally and reads the keyword
skeleton strictly: line comments, string literals, charlists, heredoc
strings, and sigils are context, so a license entry inside them is never
observed, and a structural break anywhere — an unterminated string,
charlist, heredoc, sigil, interpolation, or bracket scope, plus a nesting
depth beyond the scanner bound — refuses the whole surface fail-closed.
The `licenses` entry is proven only as a direct member of the `package`
keyword list: a licenses entry outside it (including the `def package`
function form) and one nested deeper inside it are refused for the same
reason, and a package configuration that is not a keyword list is refused
as well. A licenses value must be a bracket list of escape-free,
interpolation-free double-quoted string literals; the empty list, a
non-string entry, and a malformed delimiter are refused fail-closed. The
insertion is anchored at the package bracket the scanner proved and is
   deterministic.

## 10. The Haskell alignment contract

The Haskell row targets `*.cabal` package descriptions. The license surface
is the pair of top-level fields: `license` carries an SPDX license
expression — case-sensitive since `cabal-version: 2.2` — where
`LicenseRef-<idstring>` is the documented expression form for custom
licenses included in the package, and `license-file` carries the single
filename of the license text; the `license-files` list form is the
documented alternative, and a list carrying exactly the projected file is
the value-equal alternative spelling of the same declaration. The manifest
names are glob-form: the adapter discovers `*.cabal` files in the render
target directory and aligns every discovered file. The seam language token
`haskell` selects the row.

The bound organization projection: a declared `SPDX_LICENSE_IDENTIFIER`
projects itself; the file-based custom family projects the SPDX sideload
expression `LicenseRef-<LICENSE_ID>` (the tenant `LICENSE_ID` value). Both
projections pair with the license-file field carrying `LICENSE` — the
canonical license text at the repository root. The projected file form is
the single `license-file` field: a diverging `license-file` value is
value-replaced, a diverging one-entry `license-files` list is normalized to
the projected `license-file` form, and a `license-files` list that declares
more than one entry is refused fail-closed — the projection carries exactly
one license text, and the resolution is an explicit tenant decision. A
package description that declares the license field or the file surface
more than once is refused for the same reason, and so is a license surface
that lives inside a component section instead of the top-level package
description.

The adapter's scanner reads the package description line by line: line
comments and blank lines are transparent, column-0 property lines are
top-level fields, every other column-0 line opens a component section, and
indented lines attach as continuations of the open top-level field. Field
names are matched case-insensitively, quoted values must be escape-free,
and a value that starts with the comment marker, an unterminated quoted
value, an unterminated brace list, or indented content before any
top-level property refuses the whole surface fail-closed. The insertion of
a missing field anchors at the first top-level property block of the
package description and is proven by a re-scan.

## 11. Render alignment discipline

The render is the sanctioned writer of tenant license surfaces, and the same
governed act aligns the ecosystem license fields. Four disciplines bind
every alignment:

- **Declaration-anchored:** the target value is derived from the lock and
  the matrix row, never from a scan of the manifest.
- **Surgical:** only the license projection surface of the manifest is
  written; every other byte — including `name`, `author`, ordering,
  formatting, and unknown fields — is preserved exactly.
- **Idempotent:** a re-run over an aligned manifest is a no-op and reports
  zero deltas; a missing member is inserted deterministically; drift reports
  name exactly the diverging surfaces.
- **Single-owner:** the alignment lives exclusively in the license hub CLI
  (`cmd/license`, commands `render` and `verify`); no provisioning tool, CI
  step, or script may reimplement it.

The render skips the alignment with an explicit report when the declared
manifest is absent, and the dry-run plan previews the alignment without
writing. A non-string license member or an unscannable manifest is refused
fail-closed instead of being rewritten.

## 12. Verification finding classes

Verification is fail-closed in both directions and yields exactly three
finding classes:

1. **Declared, manifest missing or diverging** — the declared ecosystem's
   manifest lacks the license field or carries a diverging value: the
   finding names the observed value, the projected target from the lock, and
   the remediation (run the render).
2. **Manifest without declaration (shadow surface)** — a manifest license
   surface exists where no covering declaration exists: the finding names
   the declaration state; the shadow surface is either removed or the
   declaration is completed, by explicit decision.
3. **Verify proof failure** — the verify lane cannot prove the surface
   (unscannable manifest, unreadable declaration): error; the surface stays
   unproven and merge-blocking controls fail closed.

A diverging field is never adopted from the manifest into the lock; the
correction direction is always lock → manifest through the render.

## 13. Adapter set and growth rule

The npm, Python, Rust, Maven, .NET, Composer, Ruby, Elixir, and Haskell adapters are implemented. Every additional
ecosystem adapter is born content-driven when a consuming project surface
exists, and its matrix row records either the canonical field semantics
(verified against the official ecosystem documentation at specification
time) or the explicit verdict that the ecosystem carries no manifest
field — for those rows the file family is the declared truth and the
adapter aligns nothing.

## 14. Do / Don't

**Do:** derive every expected field value from the lock through the matrix
row; prove declaration, manifest, and field fail-closed in both directions;
align only through the sanctioned render, surgically and idempotently;
report every skip and finding with the observed and target values.

**Don't:** hand-edit a manifest license field outside the render; write the
lock from the manifest or adopt a diverging field value as truth; run a
verify-only process without the sanctioned writer; touch `name`, `author`,
or any non-license field during an alignment; provision license and language
families in one mega-tool; invent field semantics without verification.