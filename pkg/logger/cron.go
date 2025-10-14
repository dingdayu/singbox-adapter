// Package logger provides unified logging components.
package logger

import "log/slog"

// CronLogger adapts robfig/cron logging interface.
type CronLogger struct {
	logger *slog.Logger
}

// NewCronLogger creates a CronLogger.
func NewCronLogger(logger *slog.Logger) *CronLogger {
	return &CronLogger{
		logger: logger,
	}
}

// Info logs info level message.
func (l *CronLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

func (l *CronLogger) Error(err error, msg string, args ...any) {
	l.logger.Error(msg, append(args, slog.Any("error", err))...)
}
