package server

import (
	appLog "github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/log"
	"github.com/labstack/echo/v4"
)

// GetLoggerFromEchoContext returns the logger instance
// This is a convenience function that retrieves the global logger
func GetLoggerFromEchoContext(c echo.Context) *appLog.Logger {
	// STEP 6: Get the global logger instance
	// The logger was initialized in server.Start() and stored globally
	logger := appLog.GetGlobalLogger()

	// Fallback: If global logger not initialized (shouldn't happen), create one
	if logger == nil {
		l, err := appLog.NewLogger("doordarshan-kendra", appLog.LevelInfo)
		if err != nil {
			// This really shouldn't happen, but return nil if it does
			return nil
		}
		return l
	}

	return logger
}
