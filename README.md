# license-hub

Canonical home of the organization's license template families and the
zero-dependency `license` CLI that renders and verifies per-project license
instances.

**License of this repository:** Source-Available
`LicenseRef-license-hub-NoRepublish-1.0` — free to use, clone, build, and
modify for your own use, commercially and non-commercially; republishing this
project or substantially similar forks is prohibited. See `LICENSE`.

## Layout

| Path | Content |
|------|---------|
| `templates/` | Canonical license template families (see `templates/README.md`) |
| `org-defaults.json` | Organization constants injected into every render |
| `registry.json` | Audit index of known tenant instances |
| `POLICY.md` | Family assignment, legal SemVer, adoption, placeholder contract |
| `cmd/license` | `render`, `verify`, `digest`, `version` CLI |
| `cmd/build` | Ordered local quality gate runner |
| `cmd/check-coverage` | 100.0% statement-coverage gate |
| `docs/hosting-platforms/github/rulesets/` | Versioned GitHub ruleset definitions and activation state |

## Tenant adoption in four steps

1. Add `license.values.json` and `license.lock.json` to the tenant repository.
2. Render the instance:

   ```bash
   license render \
     --template templates/custom/norepublish/NoRepublish-1.0.0.hbs \
     --org-defaults org-defaults.json \
     --values license.values.json \
     --out .
   ```

3. Commit the rendered `LICENSE` and `LICENSES/LicenseRef-<ID>.txt`.
4. Let CI run `license verify` as the drift guard.

The full contract is documented in `docs/adoption-guide.md` and `POLICY.md`.

## Development

See `CONTRIBUTING.md`. The full local gate is:

```bash
go run -mod=readonly ./cmd/build
```
