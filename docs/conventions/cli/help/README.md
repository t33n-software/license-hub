# CLI Convention: Help
[INTENT: SPECIFICATION]

This document is the binding convention for the help system of every CLI
tool of the organization. It fixes the universal help contract on every
command node, the help level model, and the generation mandate. The core
principle discoverable-closed and the consumer equality law are owned by
[../README.md](../README.md) and are not restated here.

## 1. The universal help contract on every node

Every CLI tool owns a help argument, and **every** command and subcommand
node — recursively, independent of nesting depth — owns its own complete help
area. There is no depth or importance discount: every node is its own
responsibility boundary. A leaf command is a use-case boundary; a command
group is a navigation boundary.

Binding rules:

1. `--help` and `-h` are answered on every level; additionally a `help`
   subcommand is recommended as a second access form; the root additionally
   carries a machine-readable `--version` (see
   [../identity/README.md](../identity/README.md)).
2. **Parent/child role split.** The parent node answers "*what exists
   here?*" — it lists its children with a one-line purpose and the reference
   for reaching their help. The child node answers "*how do I call it
   correctly?*" — it carries the full input contract: usage line, flags with
   value domains, canonical examples, and exit behavior. Both directions are
   mandatory; neither replaces the other.
3. **No hidden consumer commands.** Commands intended for end users or agents
   appear in the parent navigation. The `hidden` marker is permitted
   exclusively for operator- or machine-internal endpoints.

Positive example:

```text
$ tool --help
# shows: purpose, command families with one-line purpose, global flags,
# the note "tool <command> --help"

$ tool release --help
# shows: purpose of the group, the list of subcommands (cut, promote, ...)
# with one-line purpose

$ tool release cut --help
# shows: the full input contract of the leaf node including all flag domains
```

Negative example:

```text
$ tool release cut --help
Error: unknown command "cut"
```

A registered subcommand that answers `--help` with an error or merely repeats
the parent text violates the recursion claim: the node owns no help area of
its own and its input contract is not discoverable.

## 2. The help level model

The help system owns four levels:

```text
Level 0  Root help       Purpose of the tool, command families, global flags,
                         discovery law
Level 1  Group help      Navigation: children + one-line purpose (recursive
                         per group level)
Level 2  Leaf help       Full contract: usage, flags with value domains,
                         examples, exit behavior
Level 3  Deep reference  Specification/manpage — generated from the same
                         source, linked by identity
```

A tool with a flat command tree has no group nodes; level 1 is then **not
applicable** and this is documented as such in the tool's own contract
documentation — the levels 0, 2, and 3 remain mandatory.

## 3. The generation mandate

1. **Generation over hand maintenance.** Help texts are generated from the
   command registry; per-command hand-maintained help texts are forbidden
   because they drift away from the registry and the validation.
2. **Examples are valid and canonical.** Every example in the help would pass
   the tool's own validation — consumers copy examples verbatim.
3. **Deprecations in the help.** Deprecated flags and commands carry the
   deprecation notice and the removal horizon directly in the help text (see
   [../lifecycle/README.md](../lifecycle/README.md)).

Positive example: the leaf help of a mutating command shows usage, every flag
with its class information, two canonical call examples (which would
validate), and the reference to the level-3 specification of the underlying
contract.

Negative example: the help of a command is maintained as free prose text and
names a flag the registry no longer owns — or an example the tool's own
validation would reject. Both destroy the trust foundation of the level-2
surface.
