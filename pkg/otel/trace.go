package otel

import (
	"context"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var (
	neverTP     = sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample()))
	neverTracer = neverTP.Tracer("trace-off") // 专门产出“不采样”的父 span
)

// NeverSampleContext 返回一个永远不采样的 Context 和结束函数
func NeverSampleContext() (context.Context, func()) {
	ctx, span := neverTracer.Start(context.Background(), "skip", trace.WithNewRoot())
	return ctx, func() { span.End() }
}

// NeverLogSessionWithGORM 返回一个永远不打印 SQL 的 GORM Session
func NeverLogSessionWithGORM(tx *gorm.DB) *gorm.DB {
	return tx.Session(&gorm.Session{
		Logger: tx.Logger.LogMode(gormLogger.Silent),
	})
}
