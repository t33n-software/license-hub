# CLI Convention: Identity and Discovery
[INTENT: SPECIFICATION]

This document is the binding convention for the identity and the discovery of
every CLI tool of the organization.

## 1. One stable binary name

The tool owns exactly one stable, organization-wide uniform name (lowercase,
kebab-case); repository-specific alias names or special binaries are
forbidden.

## 2. Machine-readable version

The root answers `--version` machine-readably; the version identifies the CLI
contract (see [../lifecycle/README.md](../lifecycle/README.md)). A human
command form of the version output may exist in addition; both forms render
from the same version source.

## 3. Consistent grouping law

The command tree follows one uniform ordering law (consistent noun/verb
grouping across all levels); the same vocabulary applies in every tool of the
organization.

## 4. No hidden consumer commands

Visibility is the default rule; `hidden` is reserved exclusively for
operator- or machine-internal endpoints (see
[../help/README.md](../help/README.md)).

Positive example:

```text
tool --version            # 2.8.0
tool <noun> <verb> ...    # consistent grouping on every level
```

Negative example: the same tool is shipped in two repositories under
different names or with repository-specific special commands; a
consumer-relevant command is `hidden` and appears in no parent navigation.
