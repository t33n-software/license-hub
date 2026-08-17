# GitHub Rulesets

Versioned ruleset definitions for this repository, modeled on the canonical
GitHub rulesets blueprint. Import order and activation contract:

1. `01-ticket-working-branches.json`
2. `02-develop.json`
3. `03-main.json`
4. `04-release.json`
5. `05-support.json`

Import via **Settings → Rules → Rulesets → New ruleset → Import a ruleset**,
after the CI workflows have reported their check contexts at least once and
`.github/CODEOWNERS` is merged.

## Activation state

| File | State | Reason |
|------|-------|--------|
| `00-push-protections.json` | **Not importable** | Push rulesets exist only for private/internal repositories on the Team plan. This public repository keeps the file as the versioned boundary definition; its secret-material boundary is secret scanning with push protection plus the local quality gates. |
| `01`–`05` | Pending import | Imported by the repository administrator after the first successful CI run. |

Required status-check contexts referenced by the shared-line rulesets:

- `Quality gates (linux-amd64)` (`.github/workflows/ci.yml`)
- `Dependency admission review` (`.github/workflows/dependency-review.yml`)

The `code_scanning` rule binds the CodeQL tool (`.github/workflows/codeql.yml`).

## `do_not_enforce_on_create` contract

The flag is assigned per branch class and is never left implicit:

| Ruleset | Branch class | `do_not_enforce_on_create` |
|---------|--------------|----------------------------|
| `02-develop.json` | `develop` | `false` |
| `03-main.json` | `main` | `false` |
| `04-release.json` | `release/*` | `true` |
| `05-support.json` | `support/*` | `true` |

Only release and support lines carry `true`, so a frozen line can be created
before any check has reported on it; every pull request into it is then gated
normally. `01-ticket-working-branches.json` carries no `required_status_checks`
rule, so the flag is not applicable there.
