# CLI Convention: Output and Error Contract
[INTENT: SPECIFICATION]

This document is the binding convention for the output and error contract of
every CLI tool of the organization: the stable output contract, the
structured error records, the semantic exit codes, the quiet mode, and the
secret-free output rule.

## 1. The stable output contract

The tool owns a documented output-mode switch (for example `--output
human|json`); the machine-readable form carries a versioned schema that
evolves only additively within one schema version (see
[../lifecycle/README.md](../lifecycle/README.md)). The output mode is a
`closed-enum` value domain and follows
[../values/README.md](../values/README.md).

## 2. Structured error records

Every error is a coded record carrying at least:

- the error code,
- the affected field,
- the actual value,
- the expected value,
- the violated rule,
- a valid example,
- the remediation.

The human and the machine form carry **content-wise identical** information.
Error records never carry value domains as handwritten literals; they render
the canonical source (see
[../values/single-source-of-truth.md](../values/single-source-of-truth.md)).

Positive example:

```text
Error [VALUE_INVALID]
Field: strategy
Actual value: rebase-and-merge
Expected: check, auto, rebase, or merge
Valid example: --strategy rebase
How to fix it: select a supported strategy
(exit code: usage class)
```

Negative example:

```text
Error: invalid input, see --help
```

without field, actual value, expected set, and remediation — and doubly
problematic when the referenced help does not actually carry the values
(untrue remediation, see [../errors/README.md](../errors/README.md)). Equally
forbidden: errors as prose only without a stable code, or JSON and human
forms with diverging information content.

## 3. Semantic exit codes

Exit codes are stable, documented, and carry semantic classes, for example:
success, usage error, governance rejection, external error, internal error.
The exact numeric mapping of a tool is part of its versioned CLI contract
(see [../lifecycle/README.md](../lifecycle/README.md)) and is documented in
the root help and the tool's contract documentation.

## 4. Quiet mode

Successful human output is suppressible through `--quiet`; machine-readable
output is never chatty. Errors are never suppressed: the quiet mode governs
success output only.

## 5. No secrets in outputs

Neither errors nor logs nor audit records contain credentials, tokens, keys,
or headers. This rule binds every output surface of the tool without
exception (see [../security/README.md](../security/README.md)).
