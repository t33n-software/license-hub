# CLI Convention: Test and Verification Law
[INTENT: SPECIFICATION]

This document is the binding convention for the test and verification law of
every CLI tool of the organization. It pins the projections defined in
[../values/single-source-of-truth.md](../values/single-source-of-truth.md)
against their canonical source.

## 1. Help contract tests

Every help text is pinned by contract or golden tests against the command
registry and the value-domain source; a deviation between help and source
turns CI red (drift blocks the release).

## 2. Property-based domain tests

For every `closed-enum` flag it is proven that every documented value is
accepted and every undocumented value is rejected; for every
`free-constrained` and `shaped` flag there is at least one pass and one fail
test per documented rule.

## 3. The help-first consumer test

A consumer-simulating test proves that a valid call is derivable from the
help alone — without an error path and without reading the source code. The
test passes only when help values and validation match.

## 4. Drift blocks CI

Every deviation between a projection and the source — help, errors, prompts,
completion, discovery — is a CI-blocking defect, not a warning.

Positive example: a test iterates over the canonical value list of the source
and verifies that every value appears in the help text, is accepted, and that
a control value outside the list is rejected with the correct error code.

Negative example: the help is only manually "read and approved"; a test
claims coverage without comparing the help against the source; a line
coverage percentage is reported as a substitute for the semantic alignment of
help and validation.
