package server

import (
	"github.com/labstack/echo/v4"
	"time"
)

// MetricsMiddleware Request Response metrics middleware
func (am *AppMiddleware) MetricsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		startTime := time.Now()
		ctx := c.Request().Context()
		// Get logger instance once for use throughout the function
		logger := GetLoggerFromEchoContext(c)
		err := next(c)

		// Create a SynchronousCounter metric to measure request counts
		reqCount, meterErr := am.meter.Count("request_count")
		if meterErr != nil {
			if logger != nil {
				logger.Errorf(ctx, "error creating request count metric: %v", meterErr)
			}
			return meterErr
		}
		reqCount.Add(ctx, 1, am.meter.RequestResponseAddMetricAttributes(c)...)

		// Create a SynchronousHistogram metric to measure request latencies
		reqDuration, meterErr := am.meter.Latency("request_latency")
		if meterErr != nil {
			if logger != nil {
				logger.Errorf(ctx, "error creating request latency metric: %v", meterErr)
			}
			return meterErr
		}
		reqDuration.Record(
			ctx,
			time.Since(startTime).Milliseconds(),
			am.meter.RequestResponseRecordMetricAttributes(c)...,
		)
		return err
	}
}
