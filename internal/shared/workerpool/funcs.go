// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package workerpool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

// Run executes task(i) for each i in [0, cfg.TaskCount()) using cfg.WorkerLimit().
func Run(ctx context.Context, cfg PoolConfig, task func(int) error) error {
	runErr := runEmpty(ctx)

	if cfg.TaskCount() != zero {
		runErr = runIndexed(ctx, cfg, task)
	}

	if runErr != nil {
		return fmt.Errorf(errWrapRun, runErr)
	}

	return nil
}

func runEmpty(ctx context.Context) error {
	emptyErr := contextError(ctx, "workerpool")
	if emptyErr != nil {
		return fmt.Errorf(errWrapRun, emptyErr)
	}

	return nil
}

// TaskCount returns the number of indexed tasks.
func (cfg *Config) TaskCount() int {
	return cfg.Tasks
}

// WorkerLimit returns the configured worker cap.
func (cfg *Config) WorkerLimit() int {
	return cfg.Workers
}

// Workers returns how many goroutines to use for taskCount tasks.
func Workers(configured, taskCount int) int {
	workers := min(runtime.GOMAXPROCS(zero), taskCount)

	if configured > zero {
		workers = min(configured, taskCount)
	}

	return max(workers, minWorkers)
}

func contextError(ctx context.Context, prefix string) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("%s: %w", prefix, err)
	}

	return nil
}

func emptyErrors(count int) []error {
	errs := make([]error, zero, count)

	for range count {
		errs = append(errs, nil)
	}

	return errs
}

func foldErrors(ctx context.Context, errs []error) error {
	err := contextError(ctx, "workerpool foldErrors")
	if err != nil {
		return fmt.Errorf(errWrapFoldErrors, err)
	}

	first := firstError(errs)
	if first != nil {
		return fmt.Errorf(errWrapFoldErrors, first)
	}

	return nil
}

func firstError(errs []error) error {
	for i := range errs {
		if errs[i] != nil {
			return fmt.Errorf("workerpool firstError: %w", errs[i])
		}
	}

	return nil
}

func runIndexed(ctx context.Context, cfg PoolConfig, task func(int) error) error {
	errs := emptyErrors(cfg.TaskCount())
	tasks := make(chan int)

	waitGroup := startDraining(cfg, func() {
		drainTasks(tasks, errs, task)
	})
	sendTasks(ctx, tasks, cfg.TaskCount())
	close(tasks)
	waitGroup.Wait()

	foldErr := foldErrors(ctx, errs)
	if foldErr != nil {
		return fmt.Errorf("workerpool runIndexed: %w", foldErr)
	}

	return nil
}

func startDraining(cfg PoolConfig, start func()) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)

	startPool(waitGroup, Workers(cfg.WorkerLimit(), cfg.TaskCount()), start)

	return waitGroup
}

func startPool(waitGroup *sync.WaitGroup, count int, start func()) {
	waitGroup.Add(count)
	launchWorkers(waitGroup, count, start)
}

func sendOne(ctx context.Context, tasks chan<- int, index int) bool {
	select {
	case tasks <- index:
		return true
	case <-ctx.Done():
		return false
	}
}

func sendTasks(ctx context.Context, tasks chan<- int, count int) {
	for index := range count {
		if !sendOne(ctx, tasks, index) {
			return
		}
	}
}

func launchWorkers(waitGroup *sync.WaitGroup, count int, start func()) {
	for range count {
		go func() {
			defer waitGroup.Done()

			start()
		}()
	}
}

func drainTasks(tasks <-chan int, errs []error, task func(int) error) {
	for index := range tasks {
		errs[index] = task(index)
	}
}
