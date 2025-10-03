package redis

import (
	"context"
	"time"

	"github.com/sb-icon/icon-go-api/config"
)

func (c *Client) StartSubscribers() {

	go c.startSubscriber(config.Config.RedisBlocksChannel)
	go c.startSubscriber(config.Config.RedisTransactionsChannel)
	go c.startSubscriber(config.Config.RedisLogsChannel)
	go c.startSubscriber(config.Config.RedisTokenTransfersChannel)

}

func (c *Client) startSubscriber(channelName string) {

	// Init pubsub
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	pubsub := c.client.Subscribe(
		ctx,
		channelName,
	)

	subscriberChannel := pubsub.Channel()
	inputChannel := GetBroadcaster(channelName).InputChannel

	for {
		redisMsg := <-subscriberChannel

		inputChannel <- []byte(redisMsg.Payload)
	}
}
