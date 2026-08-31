package contract

import "testing"

// Convention: docs/conventions/cli/testing/README.md

func TestBinaryNameIsTheStableContractName(t *testing.T) {
	if BinaryName != "license" {
		t.Fatalf("BinaryName = %q, want %q", BinaryName, "license")
	}
}

func TestSchemaVersionIsTheInitialContractVersion(t *testing.T) {
	if SchemaVersion != 1 {
		t.Fatalf("SchemaVersion = %d, want 1", SchemaVersion)
	}
}

func TestOutputModesAreTheClosedEnumValues(t *testing.T) {
	if OutputHuman != "human" || OutputJSON != "json" {
		t.Fatalf("output modes = %q, %q", OutputHuman, OutputJSON)
	}
}

func TestExitCodesAreTheSemanticContract(t *testing.T) {
	codes := map[int]string{
		ExitSuccess:    "success",
		ExitExecution:  "execution failure",
		ExitUsage:      "usage error",
		ExitGovernance: "governance rejection",
	}
	want := map[string]int{
		"success":              0,
		"execution failure":    1,
		"usage error":          2,
		"governance rejection": 3,
	}
	for code, name := range codes {
		if want[name] != code {
			t.Fatalf("exit code %s = %d, want %d", name, code, want[name])
		}
	}
	if len(codes) != len(want) {
		t.Fatalf("exit codes are not distinct: %v", codes)
	}
}

func TestValueClassesAreExactlyTheEightCanonicalClasses(t *testing.T) {
	classes := []ValueClass{
		ClassClosedEnum,
		ClassShaped,
		ClassFreeConstrained,
		ClassStructuralReference,
		ClassScalarBounded,
		ClassBooleanSwitch,
		ClassCompositeToken,
		ClassSecretReference,
	}
	want := []string{
		"closed-enum",
		"shaped",
		"free-constrained",
		"structural-reference",
		"scalar-bounded",
		"boolean-switch",
		"composite-token",
		"secret-reference",
	}
	if len(classes) != 8 {
		t.Fatalf("value class count = %d, want 8", len(classes))
	}
	seen := make(map[ValueClass]bool, len(classes))
	for i, class := range classes {
		if string(class) != want[i] {
			t.Fatalf("value class %d = %q, want %q", i, class, want[i])
		}
		if seen[class] {
			t.Fatalf("value class %q is declared twice", class)
		}
		seen[class] = true
	}
}

func TestStabilityLevelsAreTheCanonicalLevels(t *testing.T) {
	levels := map[Stability]string{
		StabilityStable:       "stable",
		StabilityExperimental: "experimental",
		StabilityInternal:     "internal",
	}
	for level, want := range levels {
		if string(level) != want {
			t.Fatalf("stability level = %q, want %q", level, want)
		}
	}
	if len(levels) != 3 {
		t.Fatalf("stability level count = %d, want 3", len(levels))
	}
}

func TestExitCodeNamesCoverExactlyTheSemanticContract(t *testing.T) {
	want := map[int]string{
		0: "success",
		1: "execution failure",
		2: "usage error",
		3: "governance rejection",
	}
	if len(ExitCodeNames) != len(want) {
		t.Fatalf("ExitCodeNames = %v, want %v", ExitCodeNames, want)
	}
	for code, name := range want {
		if ExitCodeNames[code] != name {
			t.Fatalf("ExitCodeNames[%d] = %q, want %q", code, ExitCodeNames[code], name)
		}
	}
}

func TestErrorCodesAreStableAndDistinct(t *testing.T) {
	codes := map[ErrorCode]string{
		ErrUsage:                "USAGE_ERROR",
		ErrValueMissing:         "VALUE_MISSING",
		ErrValueInvalid:         "VALUE_INVALID",
		ErrConfirmationRequired: "CONFIRMATION_REQUIRED",
		ErrExecution:            "EXECUTION_FAILED",
		ErrGovernanceRejected:   "GOVERNANCE_REJECTED",
	}
	for code, want := range codes {
		if string(code) != want {
			t.Fatalf("error code = %q, want %q", code, want)
		}
	}
	if len(codes) != 6 {
		t.Fatalf("error code count = %d, want 6", len(codes))
	}
}
