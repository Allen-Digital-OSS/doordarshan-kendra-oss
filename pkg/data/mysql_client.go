package data

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/constant"
	appLog "github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewMySQLClient, NewRedisClient, NewRedisRepository)

type MySQLClient struct {
	appConfig *common.AppConfig
	MysqlDb   *sql.DB
}

func NewMySQLClient(config *common.AppConfig) *MySQLClient {
	// Get logger instance (this is initialization, so no request context available)
	// Use background context since these logs occur during startup, not during a request
	logger := appLog.GetGlobalLogger()
	ctx := context.Background()

	var dsn, username, password string
	var err error
	if config.MySQL.ConnectionString == constant.EmptyStr {
		// Generating connection string from the individual config lines.
		if config.MySQL.SecretFile != constant.EmptyStr {
			username, password, err = common.GetUsernameAndPasswordFromCredFile(config.MySQL.SecretFile)
			if err != nil {
				panic(fmt.Sprintf("Error getting username and password from cred file: %v", err))
			}
		} else {
			if logger != nil {
				logger.Infof(ctx, "MySQL credentials secret file not provided : %s", config.MySQL.SecretFile)
				logger.Infof(ctx, "Using MySQL credentials from config file")
			}
			username = config.MySQL.Username
			password = config.MySQL.Password
		}

		if username == constant.EmptyStr || password == constant.EmptyStr {
			panic("MySQL username or password is empty")
		}

		dsn = username + ":" + password + "@tcp(" + config.MySQL.Endpoint + ":" + config.MySQL.Port + ")/" +
			config.MySQL.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
		maskedDsn := "MASKED:MASKED" + "@tcp(" + config.MySQL.Endpoint + ":" + config.MySQL.Port + ")/" +
			config.MySQL.DBName + "?charset=utf8mb4&parseTime=True&loc=Local"
		if logger != nil {
			logger.Infof(ctx, "Using MySQL connection string: %s", maskedDsn)
		}
	} else {
		dsn = config.MySQL.ConnectionString
	}

	mysql_db, err := sql.Open("mysql", dsn)
	if err != nil {
		if logger != nil {
			logger.Errorf(ctx, "panic: Unable to open db connection: %v", err)
		}
		panic(err)
	}
	mysql_db.SetMaxOpenConns(config.MySQL.OpenConnections)
	return &MySQLClient{appConfig: config,
		MysqlDb: mysql_db}
}
