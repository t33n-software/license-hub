# CLI Convention: Error Philosophy
[INTENT: SPECIFICATION]

This document is the binding convention for the error philosophy of every CLI
tool of the organization. The structure of the error records themselves is
owned by [../output/README.md](../output/README.md) and is not restated here.

## 1. Fail-closed at the earliest boundary

Invalid input is rejected hard at the earliest possible point — with a named
remediation. A precondition that fails only at a distant boundary is a gate
at the wrong place.

## 2. Discoverable-closed

Fail-closed alone is not sufficient; all validation knowledge is additionally
discoverable before the call through the help and the discovery endpoint (the
core principle, see [../README.md](../README.md)). A tool that reveals its
value domains only in error messages forces every consumer into a costly
guess-and-error cycle.

## 3. Truthful remediation

A remediation reference may point only to surfaces that actually carry the
answer. "See `--help`" is forbidden when the help does not contain the
referenced information.

## 4. No partial success as success

A partially executed operation is never reported as an overall success; the
reported state names the exact stand and the next step.

## 5. Missing evidence is never a pass

A check that was not executed, or whose proof is missing, never counts as
passed.

Positive example: the input violation is rejected before any mutation; the
error record names field, actual, expected, example, and a remediation whose
reference target carries the answer.

Negative example: the invalid input is noticed only after half of the
execution; the error references a help that does not list the value set; the
final report says "done" although one sub-step failed. All three behaviors
violate the error philosophy independently of one another.
