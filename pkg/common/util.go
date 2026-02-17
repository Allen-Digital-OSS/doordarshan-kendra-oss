package common

import (
	"encoding/json"
	"errors"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/constant"
	"github.com/labstack/gommon/log"
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

// ExtractAccessTokenFromAuthorizationHeader extracts the access token from the auth header.
func ExtractAccessTokenFromAuthorizationHeader(authHeader string) (string, error) {
	if authHeader == constant.EmptyStr {
		return constant.EmptyStr, ErrUnauthorizedEmptyAuthHeader
	}

	if !strings.HasPrefix(authHeader, constant.BearerStr) {
		return constant.EmptyStr, ErrUnauthorizedIllegalAuthHeaderFormat
	}

	// Extract the access token from the auth header.
	accessToken := strings.TrimPrefix(authHeader, constant.BearerStr)
	accessToken = strings.TrimSpace(accessToken)
	return accessToken, nil
}

// ToMap converts the given object to map.
func ToMap(v interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(v)
	if err != nil {
		log.Errorf("[ToMap] Error when trying to marshal the object: %v error: %v", v, err)
		return nil, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Errorf("[ToMap] Error when trying to unmarshal the object: %v error: %v", v, err)
		return nil, err
	}
	return result, nil
}
