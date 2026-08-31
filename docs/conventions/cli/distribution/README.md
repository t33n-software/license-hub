# CLI Convention: Distribution and Operations
[INTENT: SPECIFICATION]

This document is the binding convention for the distribution and the
operation of every CLI tool of the organization.

## 1. Dependency-free artifact form

A standalone, runtime-independent artifact form (for example a static binary)
is preferred, so the tool can be operated without ecosystem coupling.

## 2. Cross-platform parity

All supported platforms offer the same commands, flags, and semantics;
unavoidable platform deviations are documented.

## 3. Documented lifecycle

Installation, upgrade, uninstallation, and version pinning are documented;
reproducible builds produce versioned, evidenced artifacts (signed, with
checksums, attestations, and an SBOM — see
[../security/README.md](../security/README.md)).

## 4. Telemetry off or opt-in

Telemetry is deactivated by default or absent; if present, it is documented
and explicitly opt-in.

Positive example: the same command behaves identically on Windows, Linux, and
macOS; the installation guide names pinning to an exact version including
checksum verification.

Negative example: a flag exists on only one operating system without the help
naming this; the tool sends usage data by default; an upgrade path is
undocumented.
