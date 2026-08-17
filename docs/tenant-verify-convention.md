# Tenant verify convention

This document is the binding architectural convention for license
verification in tenant projects. It applies to every repository that carries
a rendered license instance from this hub.

## 1. Rule

Every tenant project MUST execute the license verify step in its CI/CD build
process as a merge-blocking required check. A tenant release MUST NOT be
produced from a state in which the committed license instance fails
verification.

## 2. What verification proves

`license verify` re-renders the canonical template with the tenant values and
the organization defaults and proves three properties fail-closed:

1. **Pin integrity** — the template bytes still match the digest pinned in
   `license.lock.json`.
2. **Drift freedom** — the committed `LICENSE` and `LICENSES/` instance equal
   the canonical render byte for byte; hand edits are detected.
3. **Completeness** — no unresolved `{{PLACEHOLDER}}` anchor remains.

## 3. Integration contract

- The verify step runs as a required status check with the stable check
  context **`License instance verification`**.
- Tenant repositories list this context in their `02-develop` and `03-main`
  rulesets, so no pull request into a shared line can merge while the
  instance is unverified.
- The step consumes the `license` CLI exclusively as a digest-pinned,
  signature- and attestation-verified release binary of this hub.
  Source-based execution in tenant CI — a hub checkout or `go run` — is not
  permitted.

## 4. Failure semantics

Verification failure is fail-closed: the check is red and the merge is
blocked. The remediation is either a re-render from the pinned template or a
governed `license.lock.json` update through a tenant pull request. Editing a
rendered instance by hand to silence the guard is a policy violation.

## 5. Acceptance metrics

| Metric | Target |
|--------|--------|
| Tenants with verify as a required merge check | 100 % |
| Tenant CI runs consuming the digest-pinned release binary | 100 % |
| Tenant CI runs fetching the tool from source | 0 % |
