package server

import (
	"bytes"
	"encoding/json"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"io"
	"net/http"
)

func LogRequestResponse(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Log request details
		request := c.Request()
		var requestBody []byte

		if request.Body != nil {
			// Read request body
			bodyBytes, err := io.ReadAll(request.Body)
			if err != nil {
				return err
			}
			requestBody = bodyBytes

			// Reset the request body for further use
			request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		logger := GetLoggerFromEchoContext(c)
		var formattedBody string
		if json.Valid(requestBody) {
			var prettyJSON bytes.Buffer
			if err := json.Indent(&prettyJSON, requestBody, "", "  "); err != nil {
				// STEP 7.1: Use zap logger with context for error logging
				if logger != nil {
					logger.Error(c.Request().Context(), "Error formatting request body as JSON",
						zap.Error(err),
					)
				}
			} else {
				formattedBody = prettyJSON.String()
			}
		} else {
			formattedBody = string(requestBody) // Fallback to raw string if not valid JSON
		}

		// STEP 7.2: Log request with trace IDs automatically included
		// The logger.Info() method extracts trace_id, span_id, request.id from context
		if logger != nil {
			logger.Info(c.Request().Context(), "Request received",
				zap.String("method", request.Method),
				zap.String("url", request.URL.String()),
				zap.Any("headers", request.Header),
				zap.String("body", formattedBody),
			)
		}

		// Capture the response body
		response := c.Response()
		responseWriter := &responseLogger{
			ResponseWriter: response.Writer,
			body:           bytes.NewBuffer([]byte{}),
		}
		response.Writer = responseWriter

		// Call the next middleware/handler
		err := next(c)

		// STEP 7.3: Log response with trace IDs automatically included
		// The trace IDs are extracted from context automatically by logger.Info()
		if logger != nil {
			logger.Info(c.Request().Context(), "Response sent",
				zap.Int("status", response.Status),
				zap.Any("headers", response.Header()),
				zap.String("body", responseWriter.body.String()),
			)
		}

		return err
	}
}

// responseLogger is a wrapper to capture response body for logging
type responseLogger struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (rl *responseLogger) Write(b []byte) (int, error) {
	rl.body.Write(b) // Capture response body
	return rl.ResponseWriter.Write(b)
}
