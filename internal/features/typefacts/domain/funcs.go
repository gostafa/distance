// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"fmt"
)

// ProjectFactsString returns a compact debug representation of project facts.
func ProjectFactsString(facts *ProjectFacts) string {
	return fmt.Sprintf(
		"module %q: %d packages, %d types",
		facts.ModulePath,
		len(facts.Packages),
		len(facts.Types),
	)
}

// FormatTypeFacts returns a compact debug representation of type facts.
func FormatTypeFacts(facts *TypeFacts) string {
	return fmt.Sprintf(
		"type %d %q (package %d, kind %d)",
		facts.ID,
		facts.Name,
		facts.PackageID,
		facts.Kind,
	)
}

// DumpFacts returns a compact debug representation of facts.
func DumpFacts(facts *ProjectFacts) string {
	return ProjectFactsString(facts)
}
