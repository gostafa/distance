// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package domain

import (
	"strings"
	"testing"

	"github.com/gostafa/distance/internal/features/typefacts/domain/model"
)

// White-box: the debug Stringers stay informative and panic-free.
func TestStringers(t *testing.T) {
	t.Parallel()

	pf := &ProjectFacts{
		ModulePath: "m",
		Packages:   make([]PackageFacts, 2),
		Types:      make([]TypeFacts, 3),
	}

	if s := pf.String(); !strings.Contains(s, "2 packages") || !strings.Contains(s, "3 types") {
		t.Errorf("ProjectFacts.String = %q", s)
	}

	tf := &TypeFacts{Name: "W", Kind: KindInterface}

	if s := tf.String(); !strings.Contains(s, `"W"`) {
		t.Errorf("TypeFacts.String = %q", s)
	}

	named := model.FormatNamed("E", model.KindOther)

	if !strings.Contains(named, `"E"`) {
		t.Errorf("FormatNamed = %q", named)
	}
}
