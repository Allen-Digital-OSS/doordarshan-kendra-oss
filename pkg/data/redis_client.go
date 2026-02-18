package data

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/redis/go-redis/v9"

	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/constant"
)

const (
	DefaultRedisPoolSize = 10

	ErrConfigNil                       = "appConfig is nil; cannot create redis client"
	ErrClientNil                       = "Error creating redis client"
	ErrPingingRedisCluster             = "Error pinging redis cluster (name: %s): %v"
	ErrGettingUsernamePasswordCredFile = "Error getting username and password from cred file: %v"

	SuccessPingingRedisCluster = "Successfully pinged redis cluster (name: %s)"
)

// RedisClient is the client for accessing redis store.
type RedisClient struct {
	appConfig *common.AppConfig
	ruc       redis.UniversalClient
}

// NewRedisClient is the client for accessing redis store.
func NewRedisClient(appConfig *common.AppConfig) *RedisClient {
	if appConfig == nil {
		log.Error(ErrConfigNil)
		panic(ErrConfigNil)
	}

	once := &sync.Once{}
	rc := &RedisClient{}
	once.Do(func() {
		rc = createRedisClusterClient(appConfig)
	})
	if (rc != nil) && (rc.ruc != nil) {
		return rc
	}

	log.Error(ErrClientNil)
	panic(ErrClientNil)
}

// createRedisClusterClient creates a redis client for a redis cluster.
func createRedisClusterClient(config *common.AppConfig) *RedisClient {
	log.Info(fmt.Sprintf("Connecting redis cluster: %s (%s)", config.RedisCluster.Name, config.RedisCluster.Addresses))
	var username, password string
	var err error
	if config.RedisCluster.SecretFile != constant.EmptyStr {
		username, password, err = common.GetUsernameAndPasswordFromCredFile(config.RedisCluster.SecretFile)
		if err != nil {
			log.Errorf(ErrGettingUsernamePasswordCredFile, err)
			panic(fmt.Sprintf(ErrGettingUsernamePasswordCredFile, err))
		}
	} else {
		username = config.RedisCluster.Username
		password = config.RedisCluster.Password
	}
	poolSize := config.RedisCluster.PoolSize
	if poolSize == 0 {
		poolSize = DefaultRedisPoolSize
	}
	addresses := strings.Split(config.RedisCluster.Addresses, ",")
	opts := &redis.ClusterOptions{
		Addrs:          addresses,
		PoolSize:       poolSize,
		ReadTimeout:    time.Duration(config.RedisCluster.ReadTimeoutInSecs) * time.Second,
		DialTimeout:    time.Duration(config.RedisCluster.DialTimeoutInSecs) * time.Second,
		WriteTimeout:   time.Duration(config.RedisCluster.WriteTimeoutInSecs) * time.Second,
		RouteByLatency: true,
	}
	if username != constant.EmptyStr {
		opts.Username = username
	}
	if password != constant.EmptyStr {
		opts.Password = password
	}
	if config.RedisCluster.TLS {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: config.RedisCluster.SkipVerify,
		}
	}
	ruc := redis.NewClusterClient(opts)
	var ctx = context.Background()
	_, err = ruc.Ping(ctx).Result()
	if err != nil {
		log.Errorf(ErrPingingRedisCluster, config.RedisCluster.Name, err)
		panic(fmt.Sprintf(ErrPingingRedisCluster, config.RedisCluster.Name, err))
	}

	log.Info(fmt.Sprintf(SuccessPingingRedisCluster, config.RedisCluster.Name))
	rc := &RedisClient{
		appConfig: config,
		ruc:       ruc,
	}
	log.Info(fmt.Sprintf("Completed connecting with redis cluster (name: %s) successfully", config.RedisCluster.Name))
	return rc
}
