package api

import (
	"encoding/json"
	"strings"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
	"go.uber.org/zap"

	_ "github.com/sb-icon/icon-go-api/api/docs" // import for swagger docs
	"github.com/sb-icon/icon-go-api/api/rest"
	"github.com/sb-icon/icon-go-api/api/ws"
	"github.com/sb-icon/icon-go-api/config"
	"github.com/sb-icon/icon-go-api/global"
)

// @title Icon Go API
// @version 2.0
// @description The icon tracker API
func Start() *fiber.App {

	app := fiber.New()

	// Logging middleware
	app.Use(func(c *fiber.Ctx) error {
		zap.S().Info(c.Method(), " ", c.Path())

		// Go to next middleware:
		return c.Next()
	})

	// CORS Middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:  config.Config.CORSAllowOrigins,
		AllowHeaders:  config.Config.CORSAllowHeaders,
		AllowMethods:  config.Config.CORSAllowMethods,
		ExposeHeaders: config.Config.CORSExposeHeaders,
	}))

	// Compression Middleware
	app.Use(compress.New(compress.Config{
		// refer to gofiber/fiber/blob/v1.14.6/middleware/compress.go#L17
		Level: compress.Level(config.Config.RestCompressLevel),
		Next: func(c *fiber.Ctx) bool {
			return strings.Contains(c.Path(), "/docs/")
		},
	}))

	// Enforces strict typing for query parameters
	fiber.SetParserDecoder(fiber.ParserConfig{
		SetAliasTag:       "query",
		IgnoreUnknownKeys: false,
	})

	// Swagger docs
	app.Get(config.Config.RestPrefix+"/docs/*", swagger.HandlerDefault)

	// Add version handlers
	app.Get("/version", handlerVersion)
	app.Get("/metadata", handlerMetadata)

	// Add handlers
	rest.BlocksAddHandlers(app)
	rest.TransactionsAddHandlers(app)
	rest.LogsAddHandlers(app)
	rest.AddressesAddHandlers(app)
	rest.StatsAddHandlers(app)
	rest.SuppliesAddHandlers(app)
	ws.WebsocketsAddHandlers(app)

	go app.Listen(":" + config.Config.APIPort)

	return app
}

// Version
// @Summary Show the version of server.
// @Description get the version of server.
// @Tags Version
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /version [get]
func handlerVersion(c *fiber.Ctx) error {
	message := map[string]string{
		"version": global.Version,
	}

	jsonMessage, _ := json.Marshal(message)

	return c.SendString(string(jsonMessage))
}

// Metadata
// @Summary Show the metadata of server.
// @Description get the metadata of server.
// @Tags Version
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /metadata [get]
func handlerMetadata(c *fiber.Ctx) error {
	message := map[string]string{
		"version":     global.Version,
		"name":        config.Config.Name,
		"description": "a go api template",
	}

	jsonMessage, _ := json.Marshal(message)

	return c.SendString(string(jsonMessage))
}
