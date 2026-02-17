package common

import (
	"fmt"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/constant"
	"github.com/spf13/viper"
)

// GetUsernameAndPasswordFromCredFile reads the username and password from the given file.
func GetUsernameAndPasswordFromCredFile(filename string) (username, password string, err error) {
	viper.SetConfigFile(filename)
	viper.SetConfigType("json")
	err = viper.ReadInConfig()

	if err != nil {
		return constant.EmptyStr, constant.EmptyStr, fmt.Errorf("error reading secret file, %w", err)
	}

	username = viper.GetString("username")
	password = viper.GetString("password")
	return username, password, nil
}

// GetJWTSecretFromCredFile reads the jwt secret from the given file.
func GetJWTSecretFromCredFile(filename string) (jwtSecret string, err error) {
	viper.SetConfigFile(filename)
	viper.SetConfigType("json")
	err = viper.ReadInConfig()

	if err != nil {
		return constant.EmptyStr, fmt.Errorf("error reading secret file, %w", err)
	}

	jwtSecret = viper.GetString("jwt_secret")
	return jwtSecret, nil
}
