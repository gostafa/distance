// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package workerpool

import (
	"context"
	"fmt"
	"runtime"
	"sync"
)

// RunWorkers executes task(i) for each i in [0, cfg.Tasks) using cfg.Workers goroutines.
func RunWorkers(ctx context.Context, cfg WorkerRun, task func(int) error) error {
	runErr := runWorkersEmpty(ctx)

	if cfg.Tasks != emptyLen {
		runErr = runWorkersIndexed(ctx, &cfg, task)
	}

	if runErr != nil {
		return fmt.Errorf(errWrapRunWorkers, runErr)
	}

	return nil
}

// Workers returns how many goroutines to use for taskCount tasks.
func Workers(configured, taskCount int) int {
	workers := min(runtime.GOMAXPROCS(emptyLen), taskCount)

	if configured > emptyLen {
		workers = min(configured, taskCount)
	}

	return max(workers, minWorkers)
}

func drainWorkerTasks(tasks <-chan int, errs []error, task func(int) error) {
	for index := range tasks {
		errs[index] = task(index)
	}
}

func emptyWorkerErrors(count int) []error {
	errs := make([]error, emptyLen, count)

	for range count {
		errs = append(errs, nil)
	}

	return errs
}

func firstWorkerError(errs []error) error {
	for i := range errs {
		if errs[i] != nil {
			return fmt.Errorf("goloader workers firstError: %w", errs[i])
		}
	}

	return nil
}

func foldWorkerErrors(ctx context.Context, errs []error) error {
	err := workerContextError(ctx, "goloader workers foldErrors")
	if err != nil {
		return fmt.Errorf(errWrapFoldWorkerErrors, err)
	}

	first := firstWorkerError(errs)
	if first != nil {
		return fmt.Errorf(errWrapFoldWorkerErrors, first)
	}

	return nil
}

func launchWorkerGoroutines(waitGroup *sync.WaitGroup, count int, start func()) {
	for range count {
		go func() {
			defer waitGroup.Done()

			start()
		}()
	}
}

func runWorkersEmpty(ctx context.Context) error {
	emptyErr := workerContextError(ctx, "goloader workers")
	if emptyErr != nil {
		return fmt.Errorf(errWrapRunWorkers, emptyErr)
	}

	return nil
}

func runWorkersIndexed(ctx context.Context, cfg *WorkerRun, task func(int) error) error {
	errs := emptyWorkerErrors(cfg.Tasks)
	tasks := make(chan int)

	waitGroup := startWorkerDraining(cfg, func() {
		drainWorkerTasks(tasks, errs, task)
	})
	sendWorkerTasks(ctx, tasks, cfg.Tasks)
	close(tasks)
	waitGroup.Wait()

	foldErr := foldWorkerErrors(ctx, errs)
	if foldErr != nil {
		return fmt.Errorf("goloader workers runIndexed: %w", foldErr)
	}

	return nil
}

func sendOneWorkerTask(ctx context.Context, tasks chan<- int, index int) bool {
	select {
	case tasks <- index:
		return true
	case <-ctx.Done():
		return false
	}
}

func sendWorkerTasks(ctx context.Context, tasks chan<- int, count int) {
	for index := range count {
		if !sendOneWorkerTask(ctx, tasks, index) {
			return
		}
	}
}

func startWorkerDraining(cfg *WorkerRun, start func()) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)

	startWorkerPool(waitGroup, Workers(cfg.Workers, cfg.Tasks), start)

	return waitGroup
}

func startWorkerPool(waitGroup *sync.WaitGroup, count int, start func()) {
	waitGroup.Add(count)
	launchWorkerGoroutines(waitGroup, count, start)
}

func workerContextError(ctx context.Context, prefix string) error {
	err := ctx.Err()
	if err != nil {
		return fmt.Errorf("%s: %w", prefix, err)
	}

	return nil
}
