package cmd

import (
	"flag"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/constant"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/data"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/sfu"
	"github.com/labstack/gommon/log"
)

func Run(envToConfigPathMap map[string]string, tenantClusterHandlerMap map[string]sfu.IClusterHandler) {
	flag.Parse()

	// Try to get the environment from the environment variable.
	environment := common.GetCurrentEnvironment()

	// Check that the config for the past environment is present.
	_, ok := envToConfigPathMap[environment]

	if environment == constant.EmptyStr || !ok {
		// If the environment is not set, then use the mode flag.
		environment = "local"
	}

	log.Infof("Starting the server in %s mode", environment)

	configPath, configName, err := common.SplitConfigPathAndNameAndRemoveExtension(envToConfigPathMap[environment])
	if err != nil {
		log.Errorf("panic: %v", err)
		panic(err)
	}

	// Load the application config from the environment.
	appConfig, err := common.LoadAppConfigFromEnv(configPath, configName)
	if err != nil {
		log.Errorf("panic: %v", err)
		panic(err)
	}

	// Read JWT secret from environment if not set.
	if appConfig.Server.JWTSecret == constant.EmptyStr {
		log.Infof("JWT secret is not set in the config. Trying to read from the file: %s", appConfig.Server.JWTCredentialsSecretFile)
		var jwtSecret string
		jwtSecret, err = common.GetJWTSecretFromCredFile(appConfig.Server.JWTCredentialsSecretFile)
		if err != nil {
			log.Errorf("panic: %v", err)
			panic("JWT secret is not set in config or via file. Please set it in the config or in the file.")
		}
		appConfig.Server.JWTSecret = jwtSecret
	}
	mySQLClient := data.NewMySQLClient(appConfig)
	for _, v := range tenantClusterHandlerMap {
		v.SetDb(mySQLClient.MysqlDb)
	}

	// Initialize the OpenTelemetry Meter.
	meter := common.InitializeOpenTelemetry(appConfig)

	// Initialize the application.
	application, err := InitializeApp(appConfig, meter, tenantClusterHandlerMap)
	if err != nil {
		panic(err)
	}

	// Start the application.
	application.Start()
}
