// Package middleware defines gin middlewares.
package middleware

import (
	"log/slog"

	libredis "github.com/redis/go-redis/v9"

	limiter "github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	sredis "github.com/ulule/limiter/v3/drivers/store/redis"

	"github.com/gin-gonic/gin"
)

// LimiterLogNamed is the logger namespace for the rate limiter.
const LimiterLogNamed = "limiter"

// Limiter rate limit middleware
// rate examples: "1-M", "10-H", "100-D".
func Limiter(logger *slog.Logger, client *libredis.Client, rate string) gin.HandlerFunc {
	// Define a limit rate to 4 requests per hour.
	rateFormatted, err := limiter.NewRateFromFormatted(rate)
	if err != nil {
		logger.Error("limiter rate error", slog.Any("error", err))
	}

	// Create a store with the redis client.
	store, err := sredis.NewStoreWithOptions(client, limiter.StoreOptions{
		Prefix:   "resource:limiter:",
		MaxRetry: 3,
	})
	if err != nil {
		logger.Error("limiter store redis error", slog.Any("error", err))
	}
	// Create a new middleware with the limiter instance.
	return mgin.NewMiddleware(limiter.New(store, rateFormatted))
}
