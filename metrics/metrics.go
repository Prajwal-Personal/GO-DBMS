package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/unidb/unidb-go/api"
	"github.com/unidb/unidb-go/internal"
)

func recordMetrics(query string, duration time.Duration, err error) {
	// mock metrics collection, e.g., to prometheus
	status := "success"
	if err != nil {
		status = "error"
	}
	_ = fmt.Sprintf("Query: %v, Duration: %v, Status: %s", len(query), duration, status)
	
	if duration > 500*time.Millisecond {
		// Log slow query
		fmt.Printf("SLOW QUERY: %v, Duration: %v\n", query, duration)
	}
}

// MetricsMiddleware creates a middleware that records query metrics
func MetricsMiddleware() api.Middleware {
	return func(next api.Handler) api.Handler {
		return func(ctx context.Context, query string, args ...any) (internal.Result, error) {
			start := time.Now()
			res, err := next(ctx, query, args...)
			duration := time.Since(start)

			recordMetrics(query, duration, err)

			return res, err
		}
	}
}
