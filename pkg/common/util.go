package common

import (
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/constant"
	"os"
	"strings"
)

// GetCurrentEnvironment returns the current environment.
func GetCurrentEnvironment() string {
	return strings.TrimSpace(os.Getenv(constant.EnvironmentVariableEnv))
}
