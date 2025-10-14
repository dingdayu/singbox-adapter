// Package jwt provides JWT issue and parse utilities.
package jwt

import "go.opentelemetry.io/otel"

var tracer = otel.Tracer("github.com/dingdayu/go-project-template/pkg/jwt")
