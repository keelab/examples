// Command worker-job demonstrates a broker-neutral Job, explicit ACK
// disposition and application-owned drain ordering.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/keelab/keelith"
	"github.com/keelab/keelith/health"
	"github.com/keelab/keelith/metadata"
	"github.com/keelab/keelith/operation"
	"github.com/keelab/keelith/worker"
)

type manualScheduler struct {
	mu        sync.Mutex
	handler   worker.JobHandler
	accepting bool
	closed    bool
	done      chan struct{}
}

func newManualScheduler() *manualScheduler {
	return &manualScheduler{done: make(chan struct{})}
}

func (scheduler *manualScheduler) Schedule(_ context.Context, handler worker.JobHandler) error {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	if scheduler.handler != nil {
		return fmt.Errorf("manual scheduler: already scheduled")
	}
	scheduler.handler = handler
	scheduler.accepting = true
	return nil
}

func (scheduler *manualScheduler) StopPulling(context.Context) error {
	scheduler.mu.Lock()
	scheduler.accepting = false
	scheduler.mu.Unlock()
	return nil
}

func (scheduler *manualScheduler) Drain(context.Context) error { return nil }

func (scheduler *manualScheduler) Close(context.Context) error {
	scheduler.mu.Lock()
	if !scheduler.closed {
		scheduler.closed = true
		close(scheduler.done)
	}
	scheduler.mu.Unlock()
	return nil
}

func (scheduler *manualScheduler) Wait() error {
	<-scheduler.done
	return nil
}

func (scheduler *manualScheduler) Trigger(ctx context.Context, payload []byte) (worker.Result, error) {
	scheduler.mu.Lock()
	handler := scheduler.handler
	accepting := scheduler.accepting && !scheduler.closed
	scheduler.mu.Unlock()
	if !accepting || handler == nil {
		return worker.Result{}, fmt.Errorf("manual scheduler: not accepting executions")
	}
	return handler(ctx, worker.NewExecution("manual-1", time.Now(), payload, metadata.Metadata{})), nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	readiness := health.NewRegistry()
	scheduler := newManualScheduler()
	target, err := operation.New("job", "examples.v1.ReportJob", "Run", operation.KindJob)
	if err != nil {
		panic(fmt.Errorf("build job operation: %w", err))
	}
	job, err := worker.NewJob(worker.JobConfig{
		Name:      "example-job",
		Operation: target,
		Scheduler: scheduler,
		Handler: func(_ context.Context, execution worker.Execution) worker.Result {
			log.Printf("job execution %s: %s", execution.ID(), string(execution.Payload()))
			return worker.Ack()
		},
		Health: readiness,
	})
	if err != nil {
		panic(fmt.Errorf("build job: %w", err))
	}
	app, err := keelith.New(
		keelith.WithName("worker-job"),
		keelith.WithHTTP(":8090"),
		keelith.WithHealth(readiness),
		keelith.WithServer(job),
		keelith.WithRoute(http.MethodPost, "/run", func(ctx context.Context, _ *http.Request) (any, error) {
			result, triggerErr := scheduler.Trigger(ctx, []byte("triggered by HTTP"))
			if triggerErr != nil {
				return nil, triggerErr
			}
			return map[string]string{"action": string(result.Action())}, nil
		}),
	)
	if err != nil {
		panic(fmt.Errorf("build application: %w", err))
	}
	if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(fmt.Errorf("run application: %w", err))
	}
}
