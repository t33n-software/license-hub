# CLI Convention: Compatibility and Lifecycle
[INTENT: SPECIFICATION]

This document is the binding convention for the compatibility and the
lifecycle of every CLI tool of the organization.

## 1. Semantic versioning on the CLI contract

The command, flag, and output surface is a versioned contract; a breaking
change to commands, flags, value domains, or exit codes is a major change.
The version identifies the CLI contract and is answerable machine-readably
(see [../identity/README.md](../identity/README.md)).

## 2. The deprecation pipeline

Deprecation follows the fixed sequence "notice in the help → warning at
runtime → removal", each stage with a date or version horizon; silent removal
is forbidden.

## 3. Stability levels in the help

Commands and flags carry their stability level (for example `stable`,
`experimental`, `internal`) visibly in the help.

## 4. Additive JSON evolution

Machine-readable outputs evolve within one schema version exclusively
additively; removing or renaming changes require a new schema version (see
[../output/README.md](../output/README.md)).

Positive example:

```text
--old-flag string   DEPRECATED since 2.4.0, removal in 3.0.0; use --new-flag
```

Negative example: a flag disappears in a minor release without prior notice;
a JSON field is renamed within the same schema version; a command meant as
`experimental` carries no stability marking at all.
