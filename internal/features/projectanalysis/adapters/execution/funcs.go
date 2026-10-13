// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package execution

import (
	"context"
	"fmt"

	executionport "github.com/gostafa/distance/internal/features/projectanalysis/ports/execution"
	"github.com/gostafa/distance/internal/features/projectanalysis/ports/outbound"
	provider "github.com/gostafa/distance/internal/features/typefacts/public"
	loader "github.com/gostafa/distance/internal/infrastructure/workerpool"
)

// New connects source to the consumer's fact collection port.
func New(source provider.Collector) *Collector { return &Collector{collect: source.Collect} }

// Collect translates provider facts into the consumer contract.
func (c *Collector) Collect(
	ctx context.Context,
	opts *outbound.FactOptions,
) (outbound.ProjectFacts, error) {
	facts, err := c.collect(ctx, opts)
	if err != nil {
		return outbound.ProjectFacts{
			ModulePath: "",
			Packages:   nil,
			Types:      nil,
		}, fmt.Errorf(
			"collect analysis facts: %w",
			err,
		)
	}

	return facts, nil
}

// Run executes indexed tasks and honors cancellation.
func (Executor) Run(ctx context.Context, cfg executionport.WorkerRun, task func(int) error) error {
	err := loader.RunWorkers(ctx, cfg, task)
	if err != nil {
		return fmt.Errorf("analysis workers: %w", err)
	}

	return nil
}
