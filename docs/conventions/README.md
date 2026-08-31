# Conventions
[INTENT: REFERENCE]

This directory is the canonical conventions plane of this repository. Every
convention documented here is binding for this project. Each convention domain
owns exactly one document location, and no convention is defined twice:
reference replaces copy across documentation, code, and channels (single
source of truth).

## Domains

| Domain | Scope | Owning documents |
|---|---|---|
| CLI conventions | Language-agnostic and CLI-theme-agnostic infrastructure conventions for every command-line tool: help, value domains, output and error contracts, interaction, error philosophy, configuration, security, lifecycle, testing, distribution, and identity | [cli/README.md](cli/README.md) |
| Hosting platform | Platform-specific operational conventions bound by this repository | [hosting-platforms/github/rule-sets/README.md](hosting-platforms/github/rule-sets/README.md) |

## Naming and language rules

1. Every directory and file name in this tree is lowercase kebab-case;
   Markdown documentation files may carry the name `README.md`.
2. All convention content is written in English.
3. The business code references the owning convention document in a comment
   at every point where a convention is implemented; the code never restates
   convention content.
4. A deviation from a convention in an individual case is an explicitly
   documented, reasoned exception — never a silent one.
