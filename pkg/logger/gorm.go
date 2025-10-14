// Package logger provides unified logging components.
package logger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// GormLogger is a slog-backed GORM logger implementation.
type GormLogger struct {
	log   *slog.Logger
	level gormLogger.LogLevel
}

// NewGormLogger creates a new GORM logger adapter.
func NewGormLogger(log *slog.Logger) gormLogger.Interface {
	if log == nil {
		log = logger
	}
	return &GormLogger{
		log:   log,
		level: gormLogger.Warn, // 默认级别
	}
}

// LogMode implements gorm.Logger to set log level.
func (l *GormLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	newLogger := *l
	newLogger.level = level
	return &newLogger
}

// Info implements gorm logger interface.
func (l *GormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormLogger.Info {
		l.log.InfoContext(ctx, fmt.Sprintf(msg, args...))
	}
}

// Warn implements gorm logger interface.
func (l *GormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormLogger.Warn {
		l.log.WarnContext(ctx, fmt.Sprintf(msg, args...))
	}
}

// Error implements gorm logger interface.
func (l *GormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.level >= gormLogger.Error {
		l.log.ErrorContext(ctx, fmt.Sprintf(msg, args...))
	}
}

// Trace implements gorm logger interface: record SQL details.
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && l.level >= gormLogger.Error:
		l.log.ErrorContext(ctx, sql,
			slog.String("source", utils.FileWithLineNum()),
			slog.Duration("du", elapsed),
			slog.Int64("rows", rows),
			slog.Any("error", err),
		)
	case l.level >= gormLogger.Info:
		l.log.InfoContext(ctx, sql,
			slog.String("source", utils.FileWithLineNum()),
			slog.Duration("du", elapsed),
			slog.Int64("rows", rows),
		)
	}
}
