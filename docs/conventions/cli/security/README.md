# CLI Convention: Security and Governance
[INTENT: SPECIFICATION]

This document is the binding convention for the security and governance of
every CLI tool of the organization.

## 1. Read-only as the default

Reading commands are the normal state; every mutation is explicit (its own
switch, a confirmation, a dry-run path — see
[../interaction/README.md](../interaction/README.md)).

## 2. Documented idempotency

For every mutating command it is documented whether a repetition is safe and
how the repetition case is handled.

## 3. Audit record of governed mutations

Governed mutations produce an auditable record (actor, action, timestamp,
inputs, result) — without secrets. A tool whose mutations are purely local
authoring acts, with the governed mutations living in separate automation
lanes, documents this delimitation explicitly instead of implementing a
local audit trail without governance relevance.

## 4. Least privilege and authentication

The tool demands only the permissions the called operation needs.
Authentication happens through explicit login flows, never through token
arguments (the secret law, see [../README.md](../README.md)).

## 5. Supply-chain-fortress distribution

The artifact is signed and distributed with checksums and attestations;
dependencies are pinned; a bill of materials (SBOM) is attached. Governance
is enforced by the tool — through validation, contract tests, and evidenced
distribution — not merely documented as convention (see
[../testing/README.md](../testing/README.md) and
[../distribution/README.md](../distribution/README.md)).

Positive example: a mutating command without a confirmation flag asks
interactively, shows the plan beforehand (dry-run equivalent), and writes an
audit record without confidential values; the release artifact carries a
signature, a checksum, and an SBOM.

Negative example: a command mutates silently within the same call that also
reads; an access token is accepted as a positional argument; the artifact is
distributed without checksum and signature.
