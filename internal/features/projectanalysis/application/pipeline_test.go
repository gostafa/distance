// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package application

import (
	"testing"
)

// White-box: the request→fact-options mapping.
func TestCollectOptionsMapping(t *testing.T) {
	t.Parallel()

	fo := collectOptions(&Options{
		Directory: "d", Patterns: []string{"./..."}, IncludeTests: true,
		IncludeGenerated: true, BuildTags: []string{"tag"}, Workers: 3, ContinueOnError: true,
	})

	if fo.Directory != "d" || !fo.IncludeTests || !fo.IncludeGenerated ||
		fo.Workers != 3 || !fo.ContinueOnError || len(fo.BuildTags) != 1 {

		t.Fatalf("collectOptions = %+v", fo)
	}
}
