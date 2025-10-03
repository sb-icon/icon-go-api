package rest

import (
	"github.com/sb-icon/icon-go-api/config"
	"github.com/sb-icon/icon-go-api/redis"
	"go.uber.org/zap"
)

func GetRedisCount(key string) (int64, error) {
	var count int64
	count, err := redis.GetRedisClient().GetCount(config.Config.RedisKeyPrefix + key)
	if err != nil {
		count = 0
		zap.S().Warn("Could not retrieve count: ", err.Error())
	}
	return count, err
}
