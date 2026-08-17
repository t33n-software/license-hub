package render

import "testing"

func TestExecuteSubstitutesAllOccurrences(t *testing.T) {
	template := "{{PROJECT_NAME}} is owned by {{COPYRIGHT_HOLDER}}. {{PROJECT_NAME}} rules."
	values := map[string]string{
		"PROJECT_NAME":     "license-hub",
		"COPYRIGHT_HOLDER": "CyberT33N",
	}
	want := "license-hub is owned by CyberT33N. license-hub rules."
	if got := Execute(template, values); got != want {
		t.Fatalf("Execute() = %q, want %q", got, want)
	}
}

func TestExecuteLeavesUnknownPlaceholdersUntouched(t *testing.T) {
	template := "{{PROJECT_NAME}} {{UNKNOWN_KEY}}"
	values := map[string]string{"PROJECT_NAME": "license-hub"}
	want := "license-hub {{UNKNOWN_KEY}}"
	if got := Execute(template, values); got != want {
		t.Fatalf("Execute() = %q, want %q", got, want)
	}
}

func TestExecuteWithEmptyValuesReturnsTemplate(t *testing.T) {
	template := "{{PROJECT_NAME}}"
	if got := Execute(template, map[string]string{}); got != template {
		t.Fatalf("Execute() = %q, want %q", got, template)
	}
}
