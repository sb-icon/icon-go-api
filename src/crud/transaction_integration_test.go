package crud

import (
	"github.com/stretchr/testify/assert"
	"github.com/sb-icon/icon-go-api/config"
	"github.com/sb-icon/icon-go-api/redis"
	"testing"
)

func TestGetTransactionCrud(t *testing.T) {
	config.ReadEnvironment()

	transactions, err := GetTransactionCrud().SelectManyByAddress(
		100,
		0,
		"hx9f0c84a113881f0617172df6fc61a8278eb540f5",
	)

	assert.Equal(t, nil, err)
	println(transactions)

	count, err := redis.GetRedisClient().GetCount(config.Config.RedisKeyPrefix + "transaction_regular_count_by_address_hx9f0c84a113881f0617172df6fc61a8278eb540f5")
	assert.Equal(t, nil, err)
	println(count)
	//assert.IsType(t, int64(), count)
}
