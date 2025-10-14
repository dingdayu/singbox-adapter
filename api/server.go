// Package api provides entry points to start HTTP, Cron, and Async services.
package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dingdayu/go-project-template/api/router"
	"github.com/dingdayu/go-project-template/pkg/otel"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// Run starts the HTTP server and gracefully shuts down.
func Run(ctx context.Context) {
	// ✅ 一行搞定信号监听 + context 取消
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ---------- OpenTelemetry ----------
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		otelShutdown, err := otel.Setup(ctx, otel.Options{
			Environment:  gin.Mode(),
			Insecure:     true,
			MetricPeriod: 10 * time.Second,
		})
		if err != nil {
			log.Fatalf("\033[1;30;41m[error]\033[0m failed to setup otel: %v", err)
			return
		}
		// Handle shutdown properly so nothing leaks.
		defer func() {
			err = errors.Join(err, otelShutdown(context.Background()))
		}()
	}

	addr := net.JoinHostPort(viper.GetString("app.host"), viper.GetString("app.port"))

	srv := &http.Server{
		Addr:           addr,
		Handler:        router.Handler(),
		WriteTimeout:   6 * time.Minute,
		ReadTimeout:    15 * time.Second,
		IdleTimeout:    20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	fmt.Printf("\033[1;30;42m[info]\033[0m start http server listening %s\n", addr)

	// start HTTP server (non-blocking)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("\033[1;30;41m[error]\033[0m HTTP Server failed: %v", err)
			os.Exit(1)
		}
	}()

	// ✅ wait for signal (ctx canceled)
	<-ctx.Done()
	fmt.Println("\n\033[1;30;42m[info]\033[0m Shutdown Server...")

	// ✅ create a timeout context for graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // ensure cancel is called

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("\033[1;30;41m[error]\033[0m Server forced to shutdown: %v\n", err)
		os.Exit(1) // optional: force exit
	}
	fmt.Println("\033[1;30;42m[info]\033[0m HTTP Server exited gracefully")
}
