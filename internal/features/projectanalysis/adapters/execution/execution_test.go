// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package execution

import (
	"context"
	"errors"
	"testing"

	executionport "github.com/gostafa/distance/internal/features/projectanalysis/ports/execution"
)

func TestExecutorCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	called := false
	err := (Executor{}).Run(
		ctx,
		executionport.WorkerRun{Tasks: 1, Workers: 1},
		func(int) error { called = true; return nil },
	)
	if !errors.Is(err, context.Canceled) || called {
		t.Fatalf("canceled execution: err=%v called=%v", err, called)
	}
}
