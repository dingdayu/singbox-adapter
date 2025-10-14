// Package api provides entry points to start HTTP, Cron, and Async services.
package api

import (
	"context"
	"fmt"

	"github.com/dingdayu/async/v4"

	_ "github.com/dingdayu/go-project-template/api/async" // register async tasks
)

// AsyncRun starts the async scheduler and blocks until it exits.
func AsyncRun(ctx context.Context) {
	if err := async.Run(ctx); err != nil {
		fmt.Printf("\u001B[1;30;41m[error]\u001B[0m async run failed: %v\n", err)
	}
	fmt.Println("\u001B[1;30;42m[info]\u001B[0m Task exited")
}
