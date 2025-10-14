// Package tracker contains registration and implementations of async tasks.
package tracker

import (
	"context"
	"fmt"
	"time"

	"github.com/dingdayu/async/v4"
)

func init() {
	// Register your async tasks here.
	if err := async.Register(&ExampleTimerAsync{}); err != nil {
		fmt.Printf("\u001B[1;30;41m[error]\u001B[0m register ExampleTimerAsync failed: %v\n", err)
	}
}

// ExampleTimerAsync is a sample scheduled async task.
type ExampleTimerAsync struct {
	ticker *time.Ticker
}

// OnPreRun runs before Handle; a panic here causes registration failure.
func (a *ExampleTimerAsync) OnPreRun() {
	a.ticker = time.NewTicker(1 * time.Minute)
	fmt.Printf("\u001B[1;30;42m[info]\u001B[0m ExampleTimerAsync registered and running!\n")
}

// Name async name
func (a ExampleTimerAsync) Name() string {
	return "ExampleTimerAsync"
}

// Handle async logical
func (a *ExampleTimerAsync) Handle(ctx async.Context) {
	defer ctx.Exit()
	for {
		select {
		case <-ctx.Done():
			fmt.Println("ExampleTimerAsync received done signal, exiting...")
			return
		case t := <-a.ticker.C:
			fmt.Printf("ExampleTimerAsync timer fired at %v\n", t)
			// Reset timer
			a.ticker.Reset(1 * time.Minute)
		}
	}
}

// OnShutdown is called when async is shutting down.
func (a *ExampleTimerAsync) OnShutdown(_ context.Context) {
	a.ticker.Stop()
	fmt.Printf("\u001B[1;30;42m[info]\u001B[0m ExampleTimerAsync preparing to exit!\n")
}
