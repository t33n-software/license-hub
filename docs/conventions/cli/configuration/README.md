# CLI Convention: Configuration and Twelve-Factor Mapping
[INTENT: SPECIFICATION]

This document is the binding convention for the configuration of every CLI
tool of the organization, mapped to the twelve-factor principles.

## 1. One precedence order, organization-wide

`Flag > Environment variable > Configuration file > Default` — documented
once, identical for every CLI tool of the organization. Every flag owns an
environment mapping by naming convention (for example `--output` →
`<TOOL>_OUTPUT`) or a documented reason why not. A tool without a
configuration file documents this explicitly as a reasoned reduction of the
order to `Flag > Environment variable > Default`.

## 2. Configuration outside the code

Deployment-specific configuration lives outside the artifact and is
type-validated at startup. No environment-specific value is compiled into the
artifact.

## 3. Secrets never as CLI arguments

Secrets are passed exclusively as references (environment variable, file,
broker) — arguments land in process lists and shell histories (see the
`secret-reference` class in [../values/README.md](../values/README.md)).

## 4. Bounded, configurable timeouts

External processes and network access own bounded, configurable timeouts with
a documented default; no operation is unbounded. A tool that starts no
external processes and performs no network access documents this as not
applicable.

## 5. Declared offline capability

Per command it is declared whether it works offline or requires network. The
declaration is part of the command's help area and the tool's contract
documentation.

Positive example:

```text
TOOL_OUTPUT=json tool validate --context ci
# flag would beat env, env would beat file, file would beat default —
# one order, identical everywhere
```

Negative example: two tools of the organization apply different precedence
orders; a token is accepted as a `--token <value>` argument; a network call
owns no timeout and hangs unboundedly.
