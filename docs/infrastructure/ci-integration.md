# CI Integration

This document defines how the licensing architecture integrates into
continuous integration and delivery: the tenant verify gate, the release
lanes of this repository, and the pull-propagation model.

## 1. The tenant verify gate

Every tenant project executes the license verify step in its CI/CD build
process as a merge-blocking required check. A tenant release must not be
produced from a state in which the committed license instance fails
verification.

The binding contract for this gate — including the stable check context
`License instance verification`, the shared-line ruleset binding, the
digest-pinned binary consumption rule, and the failure semantics — is owned by
[../tenant-verify-convention.md](../tenant-verify-convention.md) and is not
restated here.

Verification failure is fail-closed: the check is red and the merge is
blocked. The remediation is either a re-render from the pinned template or a
governed `license.lock.json` update through a tenant pull request.

## 2. Quality gates of this repository

This repository runs its own governed quality suite. The suite is configured
in `git-governance.quality.json` and executed through the pinned
quality-gate orchestrator resolved from `tools/go.mod`; it covers the full
test suite, 100.0% statement coverage for every executable package, static
analysis, vulnerability scanning, bounded fuzzing of the parser and decoder
packages, and a binary smoke check. The template contract test renders every
`templates/**/*.hbs` against complete fixture values and proves the
placeholder gate, the changelog presence, the sibling documentation reference,
and the registry completeness for the entire inventory on every change.

## 3. Release lanes

Two release lanes exist, deliberately separated.

**CLI releases.** The `license` binary is released through the governed
release lifecycle of this repository: after a protected `release/<semver>`
promotion to the default branch, the governed tag and artifact lanes create
the annotated immutable `v<semver>` tag and deliver the release artifact sets
(cross-platform binaries for Linux, macOS, and Windows, checksums, SBOMs, and
Sigstore signatures). Release tags are never created from a developer
workstation.

**Template family releases.** Template families stay with the
repository-specific `release-template.yml` lane. A template release is an
immutable `<template>/v<semver>` tag whose lane validates the tag shape,
packages the family templates with a `SHA256SUMS` checksum file and the family
changelog, creates a draft release, verifies the artifact set, attaches a
provenance attestation, and publishes exactly once. Every template release
requires legal-counsel review before publication; the lane runs in the
protected `release-delivery` environment.

## 4. Pull propagation

Template updates propagate as tenant-controlled pull requests; no tenant
updates automatically. A tenant propagation lane checks hub releases, verifies
digest and attestation, renders, and opens a pull request in its own
repository with its own token. The pull approach needs no cross-repository
credentials: every tenant acts only on itself.

Rollback is trivial: pin the previous version and digest in
`license.lock.json` and re-render.
