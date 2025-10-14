// Package api provides entry points to start HTTP, Cron, and Async services.
package api

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/dingdayu/go-project-template/api/cron"
	"github.com/dingdayu/go-project-template/pkg/logger"
	pkgCron "github.com/robfig/cron/v3"
)

var c *pkgCron.Cron

// CronRun starts the Cron scheduler and gracefully stops on shutdown signals.
func CronRun(ctx context.Context) {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	c = pkgCron.New(pkgCron.WithLogger(logger.NewCronLogger(logger.Logger())))
	// add job to scheduler
	if _, err := c.AddFunc("@every 5m", cron.Timer); err != nil {
		fmt.Printf("\u001B[1;30;41m[error]\u001B[0m cron add func failed: %v\n", err)
	}

	// start Cron scheduler
	c.Start()
	fmt.Println("\u001B[1;30;42m[info]\u001B[0m Cron started.")

	// wait for shutdown signal
	<-ctx.Done()

	// stop Cron scheduler
	cronCtx := c.Stop()
	// wait for jobs to finish (or set a timeout if needed)
	<-cronCtx.Done()
	fmt.Println("\u001B[1;30;42m[info]\u001B[0m CRON exited")
}
