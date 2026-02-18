package server

import (
	"context"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
)

// TraceIDMiddleware extracts and propagates trace IDs from parent requests
// This middleware MUST be added early (before other middleware) to extract trace IDs
func TraceIDMiddleware() echo.MiddlewareFunc {
	// STEP 4.1: Use OpenTelemetry's Echo instrumentation
	// This middleware does several things:
	// 1. Reads HTTP headers (traceparent, X-Amzn-Trace-Id, etc.)
	// 2. Extracts trace ID and span ID from headers using the TextMapPropagator we configured
	// 3. Creates or continues a span
	// 4. Injects the span context into request.Context()
	// 5. This span context is what our logger's WithContext() function reads
	return otelecho.Middleware("doordarshan-kendra",
		otelecho.WithTracerProvider(otel.GetTracerProvider()),
	)
}

// WithRequestIDMiddleware ensures request ID is available in context for logging
// Echo's RequestID middleware sets it in response headers, but we need it in request context
func WithRequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// STEP 4.2: Get request ID from various sources
			// Priority: Response header > Request header > Generate new
			reqID := c.Response().Header().Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = c.Request().Header.Get("X-Request-ID")
			}
			if reqID == "" {
				reqID = c.Request().Header.Get(echo.HeaderXRequestID)
			}

			// STEP 4.3: Store request ID in request context
			// Our logger's WithContext() reads from context.Value(), not Echo's context
			// So we need to explicitly store it in the http.Request's context
			if reqID != "" {
				ctx := c.Request().Context()
				ctx = context.WithValue(ctx, "requestID", reqID)
				ctx = context.WithValue(ctx, "X-Request-ID", reqID)
				c.SetRequest(c.Request().WithContext(ctx))
				c.Set("requestID", reqID) // Also store in Echo context for consistency
			}

			return next(c)
		}
	}
}
