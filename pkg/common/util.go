package common

import (
	"errors"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/constant"
	"os"
	"strings"
)

var (
	ErrUnauthorizedEmptyAuthHeader         = errors.New("auth header is empty")
	ErrUnauthorizedIllegalAuthHeaderFormat = errors.New("auth header format is illegal: missing Bearer prefix")
)

// GetCurrentEnvironment returns the current environment.
func GetCurrentEnvironment() string {
	return strings.TrimSpace(os.Getenv(constant.EnvironmentVariableEnv))
}
