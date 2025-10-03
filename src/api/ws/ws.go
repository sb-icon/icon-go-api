package ws

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"

	"github.com/sb-icon/icon-go-api/config"
	"github.com/sb-icon/icon-go-api/redis"
)

// WebsocketsAddHandlers - add fiber endpoint handlers for websocket connections
func WebsocketsAddHandlers(app *fiber.App) {

	prefix := config.Config.WebsocketPrefix

	app.Use(prefix, func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get(prefix+"/blocks", websocket.New(handlerWebsocket(config.Config.RedisBlocksChannel)))
	app.Get(prefix+"/transactions", websocket.New(handlerWebsocket(config.Config.RedisTransactionsChannel)))
	app.Get(prefix+"/logs", websocket.New(handlerWebsocket(config.Config.RedisLogsChannel)))
	app.Get(prefix+"/token-transfers", websocket.New(handlerWebsocket(config.Config.RedisTokenTransfersChannel)))
}

func handlerWebsocket(channelName string) func(*websocket.Conn) {

	return func(c *websocket.Conn) {
		// Add broadcaster
		msgChan := make(chan []byte)

		// If a msg comes into channel name (new subscriber), then send a message to the message channel
		broadcasterID := redis.GetBroadcaster(channelName).AddBroadcastChannel(msgChan)
		defer func() {
			// Remove broadcaster
			redis.GetBroadcaster(channelName).RemoveBroadcastChannel(broadcasterID)
		}()

		// Read for close
		clientCloseSig := make(chan bool)
		go func() {
			for {
				// TODO: If applying filters, this is where you would take the msg, read the filter, and store it outside of
				//  the goroutine.
				_, _, err := c.ReadMessage()
				if err != nil {
					clientCloseSig <- true
					break
				}
			}
		}()

		for {
			// Read
			msg := <-msgChan

			// TODO: If building in filters, this is where you would apply the filter.
			// TODO: If filter is nil, don't broadcast

			// Broadcast
			err := c.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				break
			}

			// check for client close
			select {
			case _ = <-clientCloseSig:
				break
			default:
				continue
			}
		}
	}
}
