// Package logger provides unified logging components.
package logger

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"go.opentelemetry.io/contrib/bridges/otelslog"

	slogmulti "github.com/samber/slog-multi"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Global logger
var (
	logger *slog.Logger
	once   sync.Once
)

// createDirIfNotExist checks and creates the directory.
func createDirIfNotExist(filePath string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0o755) // 创建目录
	}
	return nil
}

// Init initializes the global logger.
func Init() {
	fmt.Printf("\033[1;30;42m[info]\033[0m init log %s\n", viper.GetString("log.path"))
	once.Do(func() {
		// 检查并创建目录
		if err := createDirIfNotExist(viper.GetString("log.path")); err != nil {
			log.Fatalf("failed to create log directory: %v", err)
		}

		// Configure Lumberjack for log rotation
		fileWriter := &lumberjack.Logger{
			Filename:   viper.GetString("log.path"),
			MaxSize:    viper.GetInt("log.max_size"), // megabytes
			MaxBackups: viper.GetInt("log.max_backups"),
			MaxAge:     viper.GetInt("log.max_age"), // days
			Compress:   true,
		}

		// JSON handler for logging to file
		fileHandler := slog.NewJSONHandler(fileWriter, nil)

		// Text handler for logging to console
		consoleHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

		otelLogHandler := otelslog.NewHandler("otlp")

		// Combine handlers to log to both console and file
		logger = slog.New(slogmulti.Fanout(fileHandler, consoleHandler, otelLogHandler))
	})

	fmt.Printf("\033[1;30;42m[info]\033[0m init log done \n")
}

// Logger returns the global logger.
func Logger() *slog.Logger {
	if logger == nil {
		log.Fatal("Logger is not initialized. Call InitLogger() first.")
	}
	return logger
}

// WithNamespace returns a namespaced child logger.
func WithNamespace(namespace string) *slog.Logger {
	return Logger().With("namespace", namespace)
}

// https://segmentfault.com/a/1190000044992581
