# Contributing to license-hub

## Development prerequisites

- Go 1.26.6; `go.mod` pins `toolchain go1.26.6`
- Git 2.53 or newer recommended

```powershell
$env:GOTOOLCHAIN = "local"
$env:GOFLAGS = "-mod=readonly"
$env:GOVCS = "*:off"
go version
git --version
go mod download
go -C tools mod download
```

Do not add a runtime requirement for end users. Go belongs to development and
CI; released artifacts are native binaries.

These variables affect only the current shell. Do not use `go env -w` for
repository policy because it mutates a developer's global Go configuration.

## Local development loop

```powershell
go test -mod=readonly ./...
go tool -modfile tools/go.mod check-coverage
go vet -mod=readonly ./...
go run -mod=readonly .\cmd\license version
```

Format changed Go files:

```powershell
gofmt -w (Get-ChildItem -Recurse -Filter *.go | ForEach-Object FullName)
```

Use `gofmt` before every proposed commit.

## Architecture rules

- Domain packages (`internal/domain/...`) must not import filesystem adapters,
  process execution, environment variables, or hosting-provider APIs.
- Application packages own the ports consumed by the use cases.
- Adapters implement those ports and never contain a second template,
  placeholder, or digest grammar.
- The CLI must stay free of external module dependencies; the standard
  library meets the render and verify contract.

## Testing requirements

Every behavior change requires tests at the lowest meaningful boundary:

| Change | Required evidence |
|---|---|
| Template, placeholder, or digest grammar | same-package whitebox table tests and fuzz seed |
| Values or lockfile decoding | decoder whitebox test and fuzz seed |
| CLI flags or command surface | command contract test |

Run the full local gate:

```powershell
go tool -modfile tools/go.mod quality-gate
```

The pinned quality-gate orchestrator owns the complete ordered quality
sequence and resolves every development tool from `tools/go.mod`; do not
require globally installed linters, vulnerability scanners, or Lefthook
binaries.

The pinned `check-coverage` tool runs its coverage tests uncached. It rejects
every Go package reported without a `_test.go` file and every package with
executable statements below `100.0%` coverage.

Dependency updates belong in a separately reviewed update lane. That lane is
the only place allowed to run `go get` or a mutating `go mod tidy`; normal
development, CI, and release lanes must keep `go.mod` and `go.sum` read-only.

Run bounded fuzzing before changes to parser or decoder code:

```powershell
go test ./internal/domain/placeholder -run=^$ -fuzz=FuzzUnresolved -fuzztime=2s -parallel=1
go test ./internal/domain/values -run=^$ -fuzz=FuzzParse -fuzztime=2s -parallel=1
go test ./internal/domain/lockfile -run=^$ -fuzz=FuzzParse -fuzztime=2s -parallel=1
```

## Repository-local quality gates

`git-governance` is project- and language-agnostic. This repository opts in
through `git-governance.quality.json`, using one executable and an argument
array per gate. No file means `qualityStatus=unconfigured`, not a successful
project-quality result.

## Commit and branch conventions

Use the contracts enforced by the CLI:

```text
feature/ABC-123-add-export-button
feat(ABC-123): add export button
```

Official published ticket branches are append-only: do not amend, do not
force push, and merge an updated target base when synchronization is
necessary. For normal ticket work, create exactly one official regular branch
per ticket.

## Lefthook

Install the approved Lefthook binary, then:

```powershell
lefthook install
lefthook validate
```

The hook configuration is deliberately thin. Do not add regexes, live network
calls, direct `git pull`, rebase, merge, or business logic to `lefthook.yml`.

## Documentation

When a product contract changes, update the relevant local document:

- `README.md` for user-facing behavior;
- `POLICY.md` for family, versioning, and placeholder governance;
- `docs/adoption-guide.md` for the tenant consumption contract;
- `templates/README.md` for the family taxonomy.

The repository must remain self-contained. Do not add dependencies on
external rule files or unpublished documentation.

## Release handoff

The local CLI prepares release promotion only. After a protected
`release/<semver> -> main` merge, the `Tag Promoted Release` workflow creates
the annotated immutable `v<semver>` tag at the merge commit and dispatches the
artifact workflow (`release.yml` for CLI releases, `release-template.yml` for
template family releases). Do not create release tags from a developer
workstation.

The governed control lanes (`Governed Release Control`, `Execute
Protected-Line Request`, `Recover Protected-Line Request`,
`Tag Promoted Release`) consume the pinned, checksum- and
signature-verified `git-governance` release binary; they never build the tool
from source at control time. The protected environments `release-request`,
`release-execution`, and `release-delivery` gate them.

Broker-backed lanes from the reference project (credential-broker smoke and
server-side reconciliation publishing) are deliberately excluded until a
credential broker exists for this organization; reconciliation candidates are
prepared locally through the CLI with the governed pull-request session.
