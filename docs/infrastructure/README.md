# Infrastructure Documentation

This directory is the canonical home of the operational architecture of the
license hub: how templates are contracted, which control files exist, how the
CLI renders and verifies instances, and how CI integrates the verification
gate and the release lanes.

The hub follows one rule: **definition once (hub), instantiation everywhere
(tenants).** This project holds the canonical license templates as the single
source of truth; tenant projects produce rendered, committed instances through
the `license` CLI with digest pinning. Only the template is maintained;
instances are re-rendered, never hand-edited.

## Component map

| Component | Role | Canonical document |
|-----------|------|--------------------|
| `templates/` | Exactly one versioned directory per license template, released as immutable digest-pinned template releases | [template-contract.md](template-contract.md) and [../../templates/README.md](../../templates/README.md) |
| `org-defaults.json` | Organization constants injected into every render | [tenant-files.md](tenant-files.md) |
| `license.values.json` (tenant) | Project facts of the adopting repository | [tenant-files.md](tenant-files.md) |
| `license.lock.json` (tenant) | Template pin: path, version, SHA-256 digest | [tenant-files.md](tenant-files.md) |
| `license` CLI | Zero-dependency render, verify, digest, and version tool | [render-and-verify.md](render-and-verify.md) |
| CI verify step | Merge-blocking drift guard on the tenant shared lines | [ci-integration.md](ci-integration.md) and [../tenant-verify-convention.md](../tenant-verify-convention.md) |
| Release lanes | CLI releases and immutable template family releases | [ci-integration.md](ci-integration.md) |
| Tenant adoption | The four-step adoption contract | [../adoption-guide.md](../adoption-guide.md) |
| License knowledge plane | Families, types, conventions, per-template documentation | [../licensing/README.md](../licensing/README.md) |

## Architectural invariants

1. Templates are never adjusted inside the generation process; only input
   values are adjusted. Organization constants live in the hub, project facts
   in the tenant.
2. A license instance is a generated, committed artifact. Generation happens
   at maintenance time, never at tenant build time, because the license text
   must exist in the source tree itself: hosting detection, release archives,
   and clones read the tree, not build outputs.
3. Referencing a license by URL instead of committing the full text is
   forbidden: the notice becomes fragile, the version ambiguous, and scanners
   and platforms see no license.
4. The hub is organization- and tenant-agnostic: it carries no tenant
   inventory and no references to adopting projects; adoption audit views live
   in the consuming organization's instance layer, never in the hub.
5. The pull model needs no cross-repository credentials: every tenant acts
   only on itself, and template updates propagate as tenant-controlled pull
   requests.
6. This architecture requires no cloud resources: immutability comes from
   protected releases, integrity from SHA-256 digests, provenance from
   platform-native attestation, and minimal rights from repository-local
   tokens.
7. The `templates/` tree is legal-governance content: changes require the
   review boundary recorded in `../../.github/CODEOWNERS` and the
   legal-counsel gate declared in [../../POLICY.md](../../POLICY.md).
