// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package browser

import (
	"os/exec"
)

type (
	commandStarter interface {
		start(cmd *exec.Cmd) error
	}

	browserRuntime struct {
		startCommand func(*exec.Cmd) error
	}
)
