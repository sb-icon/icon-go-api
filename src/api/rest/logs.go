package rest

import (
	"encoding/json"
	"errors"
	"gorm.io/gorm"
	"strconv"

	fiber "github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/sb-icon/icon-go-api/config"
	"github.com/sb-icon/icon-go-api/crud"
	"github.com/sb-icon/icon-go-api/redis"
)

type LogsQuery struct {
	Limit int `query:"limit"`
	Skip  int `query:"skip"`

	BlockNumber     uint32 `query:"block_number"`
	BlockStart      uint32 `query:"block_start"`
	BlockEnd        uint32 `query:"block_end"`
	TransactionHash string `query:"transaction_hash"`
	Address         string `query:"address"`
	Method          string `query:"method"`
}

func LogsAddHandlers(app *fiber.App) {

	prefix := config.Config.RestPrefix + "/logs"

	app.Get(prefix+"/", handlerGetLogs)
}

// Logs
// @Summary Get Logs
// @Description get historical logs
// @Tags Logs
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param block_number query int false "skip to a block"
// @Param block_start query int false "For block range queries, a start block. Invalid with block_number"
// @Param block_end query int false "For block range queries, an end block. Invalid with block_number"
// @Param transaction_hash query string false "find by transaction hash"
// @Param address query string false "find by address"
// @Param method query string false "find by method"
// @Router /api/v1/logs [get]
// @Success 200 {object} []models.Log
// @Failure 422 {object} map[string]interface{}
func handlerGetLogs(c *fiber.Ctx) error {
	params := new(LogsQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Logs Get Handler ERROR: %s", err.Error())

		c.Status(422)
		return c.SendString(`{"error": "could not parse query parameters"}`)
	}

	// Default Params
	if params.Limit <= 0 {
		params.Limit = 25
	}

	// Check Params
	if params.Limit < 1 || params.Limit > config.Config.MaxPageSize {
		c.Status(422)
		return c.SendString(`{"error": "limit must be greater than 0 and less than 101"}`)
	}
	if params.Skip < 0 || params.Skip > config.Config.MaxPageSkip {
		c.Status(422)
		return c.SendString(`{"error": "invalid skip"}`)
	}

	if params.BlockNumber != 0 && (params.BlockEnd != 0 || params.BlockStart != 0) {
		c.Status(422)
		return c.SendString(`{"error": "can't supply both block_number and block_start or block_end'"}`)
	}

	// Get Logs
	logs, err := crud.GetLogCrud().SelectMany(
		params.Limit,
		params.Skip,
		params.BlockNumber,
		params.BlockStart,
		params.BlockEnd,
		params.TransactionHash,
		params.Address,
		params.Method,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Status(404)
			return c.SendString(`{"error": "logs not found"}`)
		}
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetLogs", " Error=Could not retrieve logs: ", err.Error(),
		)
		return c.SendString(`{"error": "could not retrieve logs"}`)
	}

	if len(*logs) == 0 {
		// No Content
		c.Status(204)
	}

	// Set X-TOTAL-COUNT
	if params.TransactionHash != "" {
		// By Transaction
		transaction, err := crud.GetTransactionCrud().SelectOne(params.TransactionHash, -1)
		count := int64(0)
		if err != nil {
			zap.S().Warn("Logs CRUD ERROR: ", err.Error())
		} else {
			count = transaction.LogCount
		}

		c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))
	} else if params.Address != "" {
		// By Address
		count, err := redis.GetRedisClient().GetCount(config.Config.RedisKeyPrefix + "log_count_by_address_" + params.Address)
		if err != nil {
			count = 0
			zap.S().Warn("Could not retrieve log count by address: ", params.Address, " Error: ", err.Error())
		}

		c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))
	} else {
		// All
		count, err := redis.GetRedisClient().GetCount(config.Config.RedisKeyPrefix + "log_count")
		if err != nil {
			count = 0
			zap.S().Warn("Could not retrieve log count: ", err.Error())
		}
		c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))
	}
	if c.Get("Accept") == "text/csv" {
		return respondWithCSV(c, *logs)
	}

	// Continue with JSON response if not CSV
	body, err := json.Marshal(&logs)
	if err != nil {
		return c.SendString(`{"error": "parsing error"}`)
	}

	return c.SendString(string(body))
}
