// Package otel provides OpenTelemetry middlewares and metrics.
package otel

import (
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// GinMiddleware returns gin's otel middleware.
func GinMiddleware() gin.HandlerFunc {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	return otelgin.Middleware(serviceName)
}

// Prometheus exposes the Prometheus metrics endpoint.
func Prometheus(c *gin.Context) {
	// register promhttp.HandlerOpts DisableCompression
	promhttp.InstrumentMetricHandler(prometheus.DefaultRegisterer, promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{
		DisableCompression: true,
		EnableOpenMetrics:  true,
	})).ServeHTTP(c.Writer, c.Request)
}

var meter metric.Meter

// HTTPRequestMetrics records HTTP request metrics.
func HTTPRequestMetrics() gin.HandlerFunc {
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	meter = otel.Meter(serviceName+"/api", metric.WithInstrumentationVersion("1.0.0"))

	reqHist, _ := meter.Float64Histogram(
		"http.server.duration", // 与 View 匹配；如不使用 View，可改为 "http_request_duration_seconds"
		metric.WithDescription("Histogram of response latency (seconds) of HTTP handlers."),
		metric.WithUnit("s"),
	)
	reqCounter, _ := meter.Int64Counter(
		"http.server.requests",
		metric.WithDescription("Counter of HTTP requests made."),
		metric.WithUnit("1"),
	)

	return func(c *gin.Context) {
		start := time.Now()
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		// Process request first
		c.Next()

		attrs := []attribute.KeyValue{
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", route),
			attribute.String("http.status_code", strconv.Itoa(c.Writer.Status())),
		}

		if reqCounter != nil {
			reqCounter.Add(c.Request.Context(), 1, metric.WithAttributes(attrs...))
		}

		if reqHist != nil {
			reqHist.Record(
				c.Request.Context(),
				time.Since(start).Seconds(),
				metric.WithAttributes(attrs...),
			)
		}
	}
}
