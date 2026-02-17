package data

import (
	"context"
	"github.com/Allen-Digital-OSS/doordarshan-kendra-oss/pkg/common"
	"github.com/redis/go-redis/v9"
)

// redisRepository is the private repository type to store redis client.
type redisRepository struct {
	ruc       redis.UniversalClient
	appConfig *common.AppConfig
}

// RedisRepository is the interface for redis repository.
type RedisRepository interface {
	PushToStream(ctx context.Context, stream string,
		values map[string]interface{}) (string, error)
}

// NewRedisRepository returns a new instance of redisRepository.
func NewRedisRepository(rc *RedisClient, appConfig *common.AppConfig) RedisRepository {
	return &redisRepository{
		appConfig: appConfig,
		ruc:       rc.ruc,
	}
}

// PushToStream pushes the values to the stream.
func (r *redisRepository) PushToStream(ctx context.Context,
	stream string,
	values map[string]interface{}) (string, error) {
	return r.ruc.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Result()
}
