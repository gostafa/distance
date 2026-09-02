// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package browser

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
)

// White-box: Open launches the platform command through the seam with the
// path as the final argument. (Not parallel — it swaps the seam.)
func TestOpenLaunchesCommand(t *testing.T) {
	var got []string

	runtime := defaultBrowserRuntime()

	runtime.startCommand = func(cmd *exec.Cmd) error {
		got = cmd.Args

		return nil
	}

	err := runtime.open("report.html")
	if err != nil {
		t.Fatal(err)
	}

	if len(got) == 0 || got[len(got)-1] != "report.html" {
		t.Errorf("launched args = %v, want the path last", got)
	}
}

// White-box: a launch failure is wrapped and names the path. (Not parallel —
// it swaps the seam.)
func TestOpenWrapsLaunchError(t *testing.T) {
	sentinel := errors.New("boom")
	runtime := defaultBrowserRuntime()

	runtime.startCommand = func(*exec.Cmd) error { return sentinel }

	err := runtime.open("report.html")

	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want the wrapped launch error", err)
	}

	if !strings.Contains(err.Error(), "report.html") {
		t.Errorf("error %q does not name the path", err)
	}
}

// White-box: the default startCommand body calls cmd.Start().
func TestDefaultStartCommand(t *testing.T) {
	err := defaultBrowserRuntime().startCommand(exec.Command("true"))
	if err != nil {
		t.Fatalf("startCommand(true) = %v", err)
	}
}
