package model

import (
	"strings"
	"testing"
)

func TestMethodFactsString(t *testing.T) {
	facts := &MethodFacts{
		Name:     "Save",
		Exported: true,
		Pos:      Position{File: "store.go", Line: 12, Column: 3},
	}

	got := facts.String()
	for _, want := range []string{
		`method "Save"`,
		"exported true",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("String() = %q, want it to contain %q", got, want)
		}
	}
}
