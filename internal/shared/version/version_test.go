// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package version

import (
	"runtime/debug"
	"testing"
)

// White-box: the embedded version string is always populated.
func TestVersionIsSet(t *testing.T) {
	t.Parallel()

	if Version() == "" {
		t.Fatal("Version must not be empty")
	}
}

func TestMissingBuildInfo(t *testing.T) {
	got := versionWith(func() (*debug.BuildInfo, bool) { return nil, false })
	if got != "dev" {
		t.Fatalf("missing build info version = %q", got)
	}
}

func TestBuildInfoVersion(t *testing.T) {
	got := versionWith(
		func() (*debug.BuildInfo, bool) { return &debug.BuildInfo{Main: debug.Module{Version: "v2.3.4"}}, true },
	)
	if got != "v2.3.4" {
		t.Fatalf("version = %q", got)
	}
}
