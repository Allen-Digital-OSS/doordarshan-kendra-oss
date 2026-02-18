package common

import (
	"fmt"
	"github.com/labstack/gommon/log"
	"github.com/spf13/viper"
	"strings"
)

type ServerConfig struct {
	Name                  string   `mapstructure:"SERVER_NAME"`
	Port                  int      `mapstructure:"SERVER_PORT"`
	LogLevel              string   `mapstructure:"SERVER_LOG_LEVEL"`
	ReadTimeoutInSeconds  int      `mapstructure:"SERVER_READ_TIMEOUT_IN_SECONDS"`
	WriteTimeoutInSeconds int      `mapstructure:"SERVER_WRITE_TIMEOUT_IN_SECONDS"`
	MaxHeaderBytes        int      `mapstructure:"SERVER_MAX_HEADER_BYTES"`
	CORSOrigins           []string `mapstructure:"SERVER_CORS_ALLOWED_ORIGINS"`
}

type OpenTelemetryConfig struct {
	ServiceName               string `mapstructure:"OPEN_TELEMETRY_SERVICE_NAME"`
	ServiceVersion            string `mapstructure:"OPEN_TELEMETRY_SERVICE_VERSION"`
	ExporterOLTPEndpoint      string `mapstructure:"OPEN_TELEMETRY_EXPORTER_OTLP_ENDPOINT"`
	ExporterIntervalInSeconds int    `mapstructure:"OPEN_TELEMETRY_EXPORTER_INTERVAL_IN_SECONDS"`
}

type RateLimitConfig struct {
	DefaultApiRateLimitQps int `mapstructure:"RATE_LIMIT_DEFAULT_API_RATE_LIMIT_QPS"`
}

type MySQLConfig struct {
	ConnectionString string `mapstructure:"MYSQL_CONNECTION_STRING"`
	Endpoint         string `mapstructure:"MYSQL_ENDPOINT"`
	Port             string `mapstructure:"MYSQL_PORT"`
	Username         string `mapstructure:"MYSQL_USERNAME"`
	Password         string `mapstructure:"MYSQL_PASSWORD"`
	DBName           string `mapstructure:"MYSQL_DB_NAME"`
	SecretFile       string `mapstructure:"MYSQL_CREDENTIALS_SECRET_FILE_PATH"`
	OpenConnections  int    `mapstructure:"MYSQL_MAX_CONNECTIONS"`
}
type RedisClusterConfig struct {
	ClusterModeOn               bool   `mapstructure:"REDIS_CLUSTER_MODE_ON"`
	Addresses                   string `mapstructure:"REDIS_CLUSTER_ADDRESSES"`
	Name                        string `mapstructure:"REDIS_CLUSTER_NAME"`
	SecretFile                  string `mapstructure:"REDIS_CLUSTER_CREDENTIALS_SECRET_FILE_PATH"`
	Username                    string `mapstructure:"REDIS_CLUSTER_USERNAME"`
	Password                    string `mapstructure:"REDIS_CLUSTER_PASSWORD"`
	ReadTimeoutInSecs           int    `mapstructure:"REDIS_CLUSTER_READ_TIMEOUT_IN_SECS"`
	WriteTimeoutInSecs          int    `mapstructure:"REDIS_CLUSTER_WRITE_TIMEOUT_IN_SECS"`
	DialTimeoutInSecs           int    `mapstructure:"REDIS_CLUSTER_DIAL_TIMEOUT_IN_SECS"`
	PoolTimeoutInSecs           int    `mapstructure:"REDIS_CLUSTER_POOL_TIMEOUT_IN_SECS"`
	PoolSize                    int    `mapstructure:"REDIS_CLUSTER_POOL_SIZE"`
	PoolFIFO                    bool   `mapstructure:"REDIS_CLUSTER_POOL_FIFO"`
	MinIdleConnections          int    `mapstructure:"REDIS_CLUSTER_MIN_IDLE_CONNECTIONS"`
	MaxIdleConnections          int    `mapstructure:"REDIS_CLUSTER_MAX_IDLE_CONNECTIONS"`
	MaxActiveConnections        int    `mapstructure:"REDIS_CLUSTER_MAX_ACTIVE_CONNECTIONS"`
	ConnectionMaxIdleTimeInSecs int    `mapstructure:"REDIS_CLUSTER_CONNECTION_MAX_IDLE_TIME_IN_SECS"`
	ConnectionMaxLifetimeInSecs int    `mapstructure:"REDIS_CLUSTER_CONNECTION_MAX_LIFETIME_IN_SECS"`
	MaxRetries                  int    `mapstructure:"REDIS_CLUSTER_MAX_RETRIES"`
	PodRoleArn                  string `mapstructure:"REDIS_CLUSTER_POD_ROLE_ARN"`
	AWSRegion                   string `mapstructure:"REDIS_CLUSTER_AWS_REGION"`
	TLS                         bool   `mapstructure:"REDIS_CLUSTER_TLS"`
	SkipVerify                  bool   `mapstructure:"REDIS_CLUSTER_SKIP_VERIFY"`
	TTLInHours                  int    `mapstructure:"REDIS_CLUSTER_TTL_IN_HOURS"`
}

type AppConfig struct {
	Server          ServerConfig
	OpenTelemetry   OpenTelemetryConfig
	MySQL           MySQLConfig
	RateLimitConfig RateLimitConfig
	RedisCluster    RedisClusterConfig
}

// LoadAppConfigFromEnv loads the app config from the given path and name.
func LoadAppConfigFromEnv(path string, name string) (appConfig *AppConfig, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType("env")

	err = viper.ReadInConfig()
	if err != nil {
		log.Errorf("Unable to read and load config: %v", err)
		return &AppConfig{}, err
	}

	var serverConfig ServerConfig
	err = viper.Unmarshal(&serverConfig)
	if err != nil {
		log.Errorf("Unable to unmarshal server config: %v", err)
		return &AppConfig{}, err
	}

	var openTelemetryConfig OpenTelemetryConfig
	err = viper.Unmarshal(&openTelemetryConfig)
	if err != nil {
		log.Errorf("Unable to unmarshal open telemetry config: %v", err)
		return &AppConfig{}, err
	}

	var rateLimitConfig RateLimitConfig
	err = viper.Unmarshal(&rateLimitConfig)
	if err != nil {
		log.Errorf("Unable to unmarshal rate limit config: %v", err)
		return &AppConfig{}, err
	}

	var mysqlConfig MySQLConfig
	err = viper.Unmarshal(&mysqlConfig)
	if err != nil {
		log.Errorf("Unable to unmarshal mysql config: %v", err)
		return &AppConfig{}, err
	}

	var redisClusterConfig RedisClusterConfig
	err = viper.Unmarshal(&redisClusterConfig)
	if err != nil {
		log.Errorf("Unable to unmarshal redis cluster config: %v", err)
		return &AppConfig{}, err
	}

	return &AppConfig{
		serverConfig,
		openTelemetryConfig,
		mysqlConfig,
		rateLimitConfig,
		redisClusterConfig,
	}, nil
}

// SplitConfigPathAndNameAndRemoveExtension splits the config path and name from the given string and removes the extension.
func SplitConfigPathAndNameAndRemoveExtension(s string) (configPath, configName string, err error) {
	if len(s) < len(".env") {
		log.Errorf("invalid config path: %s", s)
		return "", "", fmt.Errorf("invalid config path: %s", s)
	}

	if s[len(s)-len(".env"):] != ".env" {
		log.Errorf("unsupported config extension: %s", s)
		return "", "", fmt.Errorf("unsupported config extension: %s", s)
	}

	configPath = s[:strings.LastIndex(s, "/")]
	configName = s[strings.LastIndex(s, "/")+1:]
	configName = configName[:len(configName)-len(".env")]
	return configPath, configName, nil
}
