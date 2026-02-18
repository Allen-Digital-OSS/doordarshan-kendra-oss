package common

import (
	"github.com/labstack/gommon/log"
	"strings"
)

// GetLogLevelFromString returns the log level from the given string.
func GetLogLevelFromString(level string) log.Lvl {
	level = strings.TrimSpace(level)
	switch strings.ToUpper(level) {
	case "DEBUG":
		return log.DEBUG
	case "INFO":
		return log.INFO
	case "WARN":
		return log.WARN
	case "ERROR":
		return log.ERROR
	case "OFF":
		return log.OFF
	default:
		return log.INFO
	}
}
