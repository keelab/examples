// Command continuation demonstrates a durable call, a frozen Machine
// registry and one bounded Runtime.RunOnce transition.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/keelab/keelith/programmable/continuation"
	continuationMemory "github.com/keelab/keelith/programmable/continuation/memory"
)

func main() {
	ctx := context.Background()
	callOperation, err := continuation.NewOperation("/examples.v1.Report/Run")
	if err != nil {
		panic(fmt.Errorf("build continuation operation: %w", err))
	}
	callID, err := continuation.NewCallID("report-001")
	if err != nil {
		panic(fmt.Errorf("build continuation call ID: %w", err))
	}
	registry := continuation.NewRegistry()
	if err := registry.Register(callOperation, continuation.MachineFunc(
		func(_ context.Context, snapshot continuation.Snapshot) (continuation.Transition, error) {
			frame, frameErr := continuation.NewFrame(
				continuation.FrameCompleted,
				[]byte(`{"status":"done"}`),
			)
			if frameErr != nil {
				return continuation.Transition{}, frameErr
			}
			return continuation.Move(continuation.StatusCompleted, snapshot.Fence(), frame), nil
		},
	)); err != nil {
		panic(fmt.Errorf("register continuation machine: %w", err))
	}
	if err := registry.Freeze(); err != nil {
		panic(fmt.Errorf("freeze continuation registry: %w", err))
	}
	store := continuationMemory.New()
	runtime, err := continuation.NewRuntime(continuation.RuntimeConfig{
		Store:             store,
		Registry:          registry,
		ExecutorID:        "example-executor",
		LeaseDuration:     3 * time.Second,
		HeartbeatInterval: time.Second,
		PollInterval:      time.Millisecond,
	})
	if err != nil {
		panic(fmt.Errorf("build continuation runtime: %w", err))
	}
	if _, err := runtime.Create(ctx, callID, callOperation); err != nil {
		panic(fmt.Errorf("create durable call: %w", err))
	}
	processed, err := runtime.RunOnce(ctx)
	if err != nil {
		panic(fmt.Errorf("run continuation: %w", err))
	}
	snapshot, err := store.Load(ctx, callID)
	if err != nil {
		panic(fmt.Errorf("load continuation snapshot: %w", err))
	}
	fmt.Printf("processed=%d status=%s revision=%d frames=%d\n",
		processed,
		snapshot.Status(),
		snapshot.Revision(),
		len(snapshot.Frames()),
	)
}
