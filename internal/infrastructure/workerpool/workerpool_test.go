// Gostafa 2026.
// SPDX-License-Identifier: Apache-2.0.

package workerpool

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
)

func TestWorkerPool(t *testing.T) {
	var calls atomic.Int32
	for _, workers := range []int{0, 1, 8} {
		rows := make([]int, 7)
		err := RunWorkers(
			t.Context(),
			WorkerRun{Workers: workers, Tasks: len(rows)},
			func(index int) error { rows[index] = index + 1; calls.Add(1); return nil },
		)
		if err != nil {
			t.Fatal(err)
		}
		for index, value := range rows {
			if value != index+1 {
				t.Fatalf("task %d ran incorrectly", index)
			}
		}
	}
	if calls.Load() != 21 {
		t.Fatalf("duplicate or missing work: %d", calls.Load())
	}
	failure := errors.New("task failed")
	if err := RunWorkers(
		t.Context(),
		WorkerRun{Tasks: 1},
		func(int) error { return failure },
	); !errors.Is(
		err,
		failure,
	) {
		t.Fatalf("failure lost: %v", err)
	}
	if err := RunWorkers(t.Context(), WorkerRun{}, func(int) error { return nil }); err != nil {
		t.Fatal(err)
	}
	for _, tasks := range []int{0, 3} {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		if err := RunWorkers(
			ctx,
			WorkerRun{Tasks: tasks},
			func(int) error { return nil },
		); !errors.Is(
			err,
			context.Canceled,
		) {
			t.Fatalf("cancellation lost: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if sendOneWorkerTask(ctx, make(chan int), 0) {
		t.Fatal("sent a canceled task")
	}
	sendWorkerTasks(ctx, make(chan int), 2)
	if Workers(0, 0) != 1 || Workers(3, 7) != 3 {
		t.Fatal("invalid worker bound")
	}
}
