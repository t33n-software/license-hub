# Tenant verify convention

This document is the binding architectural convention for license
verification in tenant projects. It applies to every repository that carries
a rendered license instance from this hub.

## 1. Rule

Every tenant project MUST execute the license verify proof in its CI/CD build
process as a merge-blocking required check carried by its canonical
conformance check. A tenant release MUST NOT be produced from a state in
which the committed license instance fails verification.

## 2. What verification proves

`license verify` re-renders the canonical template with the tenant values and
the organization defaults and proves three properties fail-closed:

1. **Pin integrity** — the template bytes still match the digest pinned in
   `license.lock.json`.
2. **Drift freedom** — the committed `LICENSE` and `LICENSES/` instance equal
   the canonical render byte for byte; hand edits are detected.
3. **Completeness** — no unresolved `{{PLACEHOLDER}}` anchor remains.

## 3. Integration contract

- The verify proof runs inside the tenant's existing canonical conformance
  check — the merge-blocking `Canonical conformance / Canonical bindings
  verification` check context — not as a separate lane or a dedicated check
  context. When the tenant's binding manifest binds the license-hub family,
  the conformance verifier orchestrates the tenant-pinned `license` CLI and
  proves the three properties of Section 2 fail-closed on every pull request.
- Tenant repositories carry the canonical conformance check as a required
  check through the shared-line rulesets of their quality-gates class, so no
  pull request into a shared line can merge while the instance is unverified.
- The check consumes the `license` CLI through the fleet's tool channel — the
  `tool` directive in the tenant's tooling module at an immutable
  pseudo-version or release-version pin, admitted by the canonical tool
  catalog — and therefore requires no hub release. Source-based execution in
  tenant CI — a hub checkout or `go run` — is not permitted. The verification
  semantics live exactly once in the hub's CLI; the verifier orchestrates the
  pinned tool and never re-implements the proof.

## 4. Failure semantics

Verification failure is fail-closed: the check is red and the merge is
blocked. The remediation is either a re-render from the pinned template or a
governed `license.lock.json` update through a tenant pull request. Editing a
rendered instance by hand to silence the guard is a policy violation.

## 5. Acceptance metrics

| Metric | Target |
|--------|--------|
| Tenants with the license proof as a required merge check | 100 % |
| Tenant CI runs consuming the CLI through the pinned tooling module | 100 % |
| Tenant CI runs fetching the tool from source | 0 % |
