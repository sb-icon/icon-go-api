package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/sudoblockio/icon-go-api/models"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/sudoblockio/icon-go-api/config"
	"github.com/sudoblockio/icon-go-api/crud"
	"github.com/sudoblockio/icon-go-api/redis"
)

type TransactionsQuery struct {
	Limit                int    `query:"limit"`
	Skip                 int    `query:"skip"`
	From                 string `query:"from"`
	To                   string `query:"to"`
	Type                 string `query:"type"`
	Address              string `query:"address"`
	BlockNumber          int    `query:"block_number"`
	StartBlockNumber     int    `query:"start_block_number"`
	EndBlockNumber       int    `query:"end_block_number"`
	Method               string `query:"method"`
	TransactionHash      string `query:"transaction_hash"`
	Sort                 string `query:"sort"`
	TokenContractAddress string `query:"token_contract_address"`
}

func TransactionsAddHandlers(app *fiber.App) {

	prefix := config.Config.RestPrefix + "/transactions"

	app.Get(prefix+"/", handlerGetTransactions)
	app.Get(prefix+"/details/:hash", handlerGetTransaction)
	app.Get(prefix+"/icx/:address", handlerGetIcxTransactionsAddress)
	app.Get(prefix+"/block-number/:block_number", handlerGetTransactionBlockNumber)
	app.Get(prefix+"/address/:address", handlerGetTransactionAddress)
	app.Get(prefix+"/internal/:hash", handlerGetInternalTransactionsByHash)
	app.Get(prefix+"/internal/address/:address", handlerGetInternalTransactionsAddress)
	app.Get(prefix+"/internal/block-number/:block_number", handlerGetInternalTransactionsBlockNumber)
	app.Get(prefix+"/token-transfers", handlerGetTokenTransfers)
	app.Get(prefix+"/token-transfers/address/:address", handlerGetTokenTransfersAddress)
	app.Get(prefix+"/token-transfers/token-contract/:token_contract_address", handlerGetTokenTransfersTokenContract)
	app.Get(prefix+"/token-holders/token-contract/:token_contract_address", handlerGetTokenAddressesTokenContract)
}

type TransactionResult struct {
	Val *[]models.TransactionList
	Err error
}

type TransactionCountResult struct {
	Val int64
	Err error
}

// Transactions
// @Summary Get Transactions
// @Description get historical transactions
// @Tags Transactions
// @BasePath /api/v1
// @Accept application/json,text/csv
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param from query string false "find by from address"
// @Param to query string false "find by to address"
// @Param type query string false "find by type"
// @Param block_number query int false "find by block number"
// @Param start_block_number query int false "find by block number range"
// @Param end_block_number query int false "find by block number range"
// @Param method query string false "find by method"
// @Param sort query string false "desc or asc"
// @Router /api/v1/transactions [get]
// @Success 200 {object} []models.TransactionList
// @Success 200 {string} string "CSV Response"
// @Failure 422 {object} map[string]interface{}
func handlerGetTransactions(c *fiber.Ctx) error {
	params := new(TransactionsQuery)
	err := c.QueryParser(params)
	if err != nil {
		zap.S().Warnf("Transactions Get Handler ERROR: %s", err.Error())

		c.Status(422)
		return c.SendString(`{"error": "could not parse query parameters"}`)
	}

	// Default Params
	if params.Limit <= 0 {
		params.Limit = 25
	}
	if params.Sort == "" {
		params.Sort = "desc"
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
	if params.Sort != "desc" && params.Sort != "asc" {
		params.Sort = "desc"
	}

	// NOTE: casting string types for type field
	if params.Type == "regular" || params.Type == "" {
		params.Type = "transaction"
	} else if params.Type == "internal" {
		params.Type = "log"
	}

	var wg sync.WaitGroup
	transactionResultChan := make(chan TransactionResult)
	transactionCountResultChan := make(chan TransactionCountResult)

	var count int64
	if params.From == "" && params.To == "" && params.BlockNumber == 0 && params.StartBlockNumber == 0 && params.Method == "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count, err = GetRedisCount("transaction_regular_count")
			if err != nil {
				zap.S().Warn("Could not retrieve transaction count: ", err.Error())
			}
			transactionCountResultChan <- TransactionCountResult{Val: count, Err: err}
		}()
	} else {
		wg.Add(1)
		go func() {
			defer wg.Done()
			transactionCount, err := crud.GetTransactionCrud().CountMany(
				params.From,
				params.To,
				params.Type,
				params.BlockNumber,
				params.StartBlockNumber,
				params.EndBlockNumber,
				params.Method,
			)
			if err != nil {
				count, err = GetRedisCount("transaction_regular_count")
				if err != nil {
					zap.S().Warn("Could not retrieve transaction count: ", err.Error())
				}
				transactionCountResultChan <- TransactionCountResult{Val: count, Err: err}
				return
			}
			transactionCountResultChan <- TransactionCountResult{Val: *transactionCount, Err: err}
		}()
	}

	// Get Transactions
	wg.Add(1) // We are not going to do a parallel query
	go func() {
		defer wg.Done()
		transactions, err := crud.GetTransactionCrud().SelectMany(
			params.Limit,
			params.Skip,
			params.From,
			params.To,
			params.Type,
			params.BlockNumber,
			params.StartBlockNumber,
			params.EndBlockNumber,
			params.Method,
			params.Sort,
		)
		transactionResultChan <- TransactionResult{Val: transactions, Err: err}
	}()

	transactionCountResult := <-transactionCountResultChan
	if transactionCountResult.Err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetTransactions",
			" Error=Could not retrieve transactions: ", transactionCountResult.Err,
		)
		return c.SendString(`{"error": "could not retrieve transactions"}`)
	}
	count = transactionCountResult.Val

	transactionResult := <-transactionResultChan
	transactions := transactionResult.Val

	if len(*transactions) == 0 {
		// No Content
		c.Status(204)
	}

	c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))

	if c.Get("Accept") == "text/csv" {
		return respondWithCSV(c, *transactions)
	}

	// Continue with JSON response if not CSV
	body, err := json.Marshal(&transactions)
	if err != nil {
		return c.SendString(`{"error": "parsing error"}`)
	}

	return c.SendString(string(body))
}

// Transaction
// @Summary Get Transaction
// @Description get details of a transaction
// @Tags Transactions
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param hash path string true "transaction hash"
// @Router /api/v1/transactions/details/{hash} [get]
// @Success 200 {object} models.Transaction
// @Failure 422 {object} map[string]interface{}
func handlerGetTransaction(c *fiber.Ctx) error {
	hash := c.Params("hash")

	if hash == "" {
		c.Status(422)
		return c.SendString(`{"error": "hash required"}`)
	}

	transaction, err := crud.GetTransactionCrud().SelectOne(hash, -1)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Status(404)
			return c.SendString(`{"error": "transaction not found"}`)
		}
		c.Status(500)
		zap.S().Warn(err.Error())
		return c.SendString(`{"error": "could not retrieve transaction"}`)
	}

	body, _ := json.Marshal(&transaction)
	return c.SendString(string(body))
}

// Transaction ICX by Address
// @Summary Get ICX Transactions by Address
// @Description get ICX transactions to or from an address
// @Tags Transactions
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param address path string true "address"
// @Router /api/v1/transactions/icx/{address} [get]
// @Success 200 {object} []models.Transaction
// @Failure 422 {object} map[string]interface{}
func handlerGetIcxTransactionsAddress(c *fiber.Ctx) error {
	address := c.Params("address")
	if address == "" {
		c.Status(422)
		return c.SendString(`{"error": "address required"}`)
	}

	params := new(TransactionsQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Transactions Get Handler ERROR: %s", err.Error())

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

	transactions, err := crud.GetTransactionCrud().SelectManyIcxByAddress(
		params.Limit,
		params.Skip,
		address,
	)
	if err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetTransactionAddress",
			" Error=Could not retrieve transactions: ", err.Error(),
		)
		return c.SendString(`{"error": "no transactions found"}`)
	}

	count, err := crud.GetTransactionCrud().CountManyIcxByAddress(address)
	if err != nil {
		c.Status(500)
		return c.SendString(`{"error": "count server error"}`)
	}
	c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))

	if c.Get("Accept") == "text/csv" {
		return respondWithCSV(c, *transactions)
	}

	// Continue with JSON response if not CSV
	body, err := json.Marshal(transactions)
	if err != nil {
		return c.SendString(`{"error": "parsing error"}`)
	}

	return c.SendString(string(body))
}

// Transactions by Block Number
// @Summary Get Transactions by block_number
// @Description get transactions by block_number
// @Tags Transactions
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param block_number path string true "block_number"
// @Router /api/v1/transactions/block-number/{block_number} [get]
// @Success 200 {object} models.TransactionList
// @Failure 422 {object} map[string]interface{}
func handlerGetTransactionBlockNumber(c *fiber.Ctx) error {
	blockNumberRaw := c.Params("block_number")
	if blockNumberRaw == "" {
		c.Status(422)
		return c.SendString(`{"error": "block_number required"}`)
	}

	blockNumber, err := strconv.Atoi(blockNumberRaw)
	if err != nil {
		c.Status(422)
		return c.SendString(`{"error": "invalid block_number"}`)
	}

	params := new(TransactionsQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Transactions Get Handler ERROR: %s", err.Error())

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

	// Get Transactions
	transactions, err := crud.GetTransactionCrud().SelectMany(
		params.Limit,
		params.Skip,
		"",
		"",
		"transaction",
		blockNumber,
		0,
		0,
		"",
		"desc",
	)
	if err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetTransactionBlockNumber",
			" Error=Could not retrieve transactions: ", err.Error(),
		)
		return c.SendString(`{"error": "could not retrieve transactions"}`)
	}

	// X-TOTAL-COUNT
	block, err := crud.GetBlockCrud().SelectOne(uint32(blockNumber))
	count := int64(0)
	if err != nil {
		zap.S().Warn("Could not retrieve transaction count: ", err.Error())
	} else {
		count = int64(block.TransactionCount)
	}
	c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))

	body, _ := json.Marshal(&transactions)
	return c.SendString(string(body))
}

// Transactions by Address
// @Summary Get Transactions by address
// @Description get transactions by address
// @Tags Transactions
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param address path string true "address"
// @Router /api/v1/transactions/address/{address} [get]
// @Success 200 {object} models.TransactionList
// @Failure 422 {object} map[string]interface{}
func handlerGetTransactionAddress(c *fiber.Ctx) error {
	address := c.Params("address")
	if address == "" {
		c.Status(422)
		return c.SendString(`{"error": "address required"}`)
	}

	params := new(SkipLimitQuery)
	if err := c.QueryParser(params); err != nil {
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

	transactions, err := crud.GetTransactionCrud().SelectManyByAddress(
		params.Limit, params.Skip, address,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Status(404)
			return c.SendString(`{"error": "no transactions found"}`)
		}
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetTransactionAddress",
			" Error=Could not retrieve transactions: ",
			err.Error(),
		)
		return c.SendString(`{"error": "could not retrieve transactions"}`)
	}

	// X-TOTAL-COUNT
	count, err := redis.GetRedisClient().GetCount(config.Config.RedisKeyPrefix + "transaction_regular_count_by_address_" + address)
	if err != nil {
		count = 0
		zap.S().Warn("Could not retrieve transaction count: ", err.Error())
	}
	c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))

	body, _ := json.Marshal(&transactions)
	return c.SendString(string(body))
}

// Internal transactions by hash
// @Summary Get Internal Transactions By Hash
// @Description Get internal transactions by hash
// @Tags Transactions
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param hash path string true "find by hash"
// @Router /api/v1/transactions/internal/{hash} [get]
// @Success 200 {object} []models.TransactionInternalList
// @Failure 422 {object} map[string]interface{}
func handlerGetInternalTransactionsByHash(c *fiber.Ctx) error {
	hash := c.Params("hash")
	if hash == "" {
		c.Status(422)
		return c.SendString(`{"error": "hash required"}`)
	}

	params := new(TransactionsQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Transactions Get Handler ERROR: %s", err.Error())

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

	if hash == "" {
		c.Status(422)
		return c.SendString(`{"error": "hash required"}`)
	}

	internalTransactions, err := crud.GetTransactionCrud().SelectManyInternal(
		params.Limit,
		params.Skip,
		hash,
		0,
	)
	if err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetInternalTransactionsByHash",
			" Error=Could not retrieve transactions: ", err.Error(),
		)
		return c.SendString(`{"error": "no internal transaction found"}`)
	}

	if len(*internalTransactions) == 0 {
		// No Content
		c.Status(204)
	}

	body, _ := json.Marshal(&internalTransactions)
	return c.SendString(string(body))
}

// Internal transactions by address
// @Summary Get Internal Transactions By Address
// @Description Get internal transactions by address
// @Tags Transactions
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param address path string true "find by address"
// @Router /api/v1/transactions/internal/address/{address} [get]
// @Success 200 {object} []models.TransactionInternalList
// @Failure 422 {object} map[string]interface{}
func handlerGetInternalTransactionsAddress(c *fiber.Ctx) error {
	address := c.Params("address")
	if address == "" {
		c.Status(422)
		return c.SendString(`{"error": "address required"}`)
	}

	params := new(TransactionsQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Transactions Get Handler ERROR: %s", err.Error())

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
		return c.SendString(fmt.Sprintf(`{"error": "invalid skip, must be less than %d"}`, config.Config.MaxPageSkip))
	}

	internalTransactions, err := crud.GetTransactionCrud().SelectManyInternalByAddress(
		params.Limit,
		params.Skip,
		address,
	)
	if err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetInternalTransactionsAddress",
			" Error=Could not retrieve transactions: ", err.Error(),
		)
		return c.SendString(`{"error": "no internal transaction found"}`)
	}

	if len(*internalTransactions) == 0 {
		// No Content
		c.Status(204)
	}

	// X-TOTAL-COUNT
	count, err := redis.GetRedisClient().GetCount(config.Config.RedisKeyPrefix + "transaction_internal_count_by_address_" + address)
	if err != nil {
		count = 0
		zap.S().Warn("Could not retrieve transaction count: ", err.Error())
	}
	c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))

	body, _ := json.Marshal(&internalTransactions)
	return c.SendString(string(body))
}

// Internal transactions by block number
// @Summary Get Internal Transactions By Block Number
// @Description Get internal transactions by block number
// @Tags Transactions
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param block_number path string true "block_number"
// @Router /api/v1/transactions/internal/block-number/{block_number} [get]
// @Success 200 {object} []models.TransactionInternalList
// @Failure 422 {object} map[string]interface{}
func handlerGetInternalTransactionsBlockNumber(c *fiber.Ctx) error {
	blockNumberRaw := c.Params("block_number")
	if blockNumberRaw == "" {
		c.Status(422)
		return c.SendString(`{"error": "block number required"}`)
	}

	blockNumber, err := strconv.Atoi(blockNumberRaw)

	params := new(TransactionsQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Transactions Get Handler ERROR: %s", err.Error())

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

	internalTransactions, err := crud.GetTransactionCrud().SelectManyInternal(
		params.Limit,
		params.Skip,
		"",
		blockNumber,
	)
	if err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetInternalTransactionsBlock",
			" Error=Could not retrieve transactions: ", err.Error(),
		)
		return c.SendString(`{"error": "no internal transaction found"}`)
	}

	if len(*internalTransactions) == 0 {
		// No Content
		c.Status(204)
	}

	// X-TOTAL-COUNT
	block, err := crud.GetBlockCrud().SelectOne(uint32(blockNumber))
	count := int64(0)
	if err != nil {
		zap.S().Warn("Could not retrieve transaction count: ", err.Error())
	} else {
		count = block.InternalTransactionCount
	}
	c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))

	body, _ := json.Marshal(&internalTransactions)
	return c.SendString(string(body))
}

// TokenTransfers
// @Summary Get Token Transfers
// @Description get historical token transfers
// @Tags Transactions
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param from query string false "find by from address"
// @Param to query string false "find by to address"
// @Param block_number query int false "find by block number"
// @Param start_block_number query int false "find by block number range"
// @Param end_block_number query int false "find by block number range"
// @Param token_contract_address query string false "find by token contract"
// @Param transaction_hash query string false "find by transaction hash"
// @Router /api/v1/transactions/token-transfers [get]
// @Success 200 {object} []models.TokenTransfer
// @Failure 422 {object} map[string]interface{}
func handlerGetTokenTransfers(c *fiber.Ctx) error {
	params := new(TransactionsQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Transactions Get Handler ERROR: %s", err.Error())

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

	// Get Transactions
	tokenTransfers, err := crud.GetTokenTransferCrud().SelectMany(
		params.Limit,
		params.Skip,
		params.From,
		params.To,
		params.BlockNumber,
		params.StartBlockNumber,
		params.EndBlockNumber,
		params.TransactionHash,
		params.TokenContractAddress,
	)
	if err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetTokenTransfers",
			" Error=Could not retrieve token transfers: ", err.Error(),
		)
		return c.SendString(`{"error": "could not retrieve transactions"}`)
	}

	if len(*tokenTransfers) == 0 {
		// No Content
		c.Status(204)
	}

	// X-TOTAL-COUNT
	count, err := redis.GetRedisClient().GetCount(config.Config.RedisKeyPrefix + "token_transfer_count")
	if err != nil {
		count = 0
		zap.S().Warn("Could not retrieve transaction count: ", err.Error())
	}
	c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))

	if c.Get("Accept") == "text/csv" {
		return respondWithCSV(c, *tokenTransfers)
	}

	// Continue with JSON response if not CSV
	body, err := json.Marshal(&tokenTransfers)
	if err != nil {
		return c.SendString(`{"error": "parsing error"}`)
	}

	return c.SendString(string(body))
}

// TokenTransfersAddress
// @Summary Get Token Transfer By Address
// @Description get historical token transfers by address
// @Tags Transactions
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param start_block_number query int false "find by block number range"
// @Param end_block_number query int false "find by block number range"
// @Param address path string true "find by address"
// @Router /api/v1/transactions/token-transfers/address/{address} [get]
// @Success 200 {object} []models.TokenTransfer
// @Failure 422 {object} map[string]interface{}
func handlerGetTokenTransfersAddress(c *fiber.Ctx) error {
	address := c.Params("address")
	if address == "" {
		c.Status(422)
		return c.SendString(`{"error": "address required"}`)
	}

	params := new(SkipLimitQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Transactions Get Handler ERROR: %s", err.Error())

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

	// Get Transactions
	tokenTransfers, err := crud.GetTokenTransferCrud().SelectManyByAddressBlockRange(
		params.Limit,
		params.Skip,
		address,
		params.StartBlockNumber,
		params.EndBlockNumber,
	)
	if err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetTokenTransfersAddress",
			" Error=Could not retrieve token transfers: ", err.Error(),
		)
		return c.SendString(`{"error": "could not retrieve transactions"}`)
	}

	if len(*tokenTransfers) == 0 {
		// No Content
		c.Status(204)
	}

	// X-TOTAL-COUNT
	count, err := redis.GetRedisClient().GetCount(config.Config.RedisKeyPrefix + "token_transfer_count_by_address_" + address)
	if err != nil {
		count = 0
		zap.S().Warn("Could not retrieve transaction count: ", err.Error())
	}
	c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))

	body, _ := json.Marshal(&tokenTransfers)
	return c.SendString(string(body))
}

// TokenTransfersTokenContract
// @Summary Get Token Transfers By Token Contract
// @Description get historical token transfers by token contract
// @Tags Transactions
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param token_contract_address path string true "find by token contract address"
// @Router /api/v1/transactions/token-transfers/token-contract/{token_contract_address} [get]
// @Success 200 {object} []models.TokenTransfer
// @Failure 422 {object} map[string]interface{}
func handlerGetTokenTransfersTokenContract(c *fiber.Ctx) error {
	tokenContractAddress := c.Params("token_contract_address")
	if tokenContractAddress == "" {
		c.Status(422)
		return c.SendString(`{"error": "token_contract_address required"}`)
	}

	params := new(TransactionsQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Transactions Get Handler ERROR: %s", err.Error())

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

	// Get Transactions
	tokenTransfers, err := crud.GetTokenTransferCrud().SelectManyByTokenContractAddress(
		params.Limit,
		params.Skip,
		tokenContractAddress,
	)
	if err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetTokenTransfersTokenContract",
			" Error=Could not retrieve token transfers: ", err.Error(),
		)
		return c.SendString(`{"error": "could not retrieve transactions"}`)
	}

	if len(*tokenTransfers) == 0 {
		// No Content
		c.Status(204)
	}

	// X-TOTAL-COUNT
	count, err := redis.GetRedisClient().GetCount(config.Config.RedisKeyPrefix + "token_transfer_count_by_token_contract_" + tokenContractAddress)
	if err != nil {
		count = 0
		zap.S().Warn("Could not retrieve transaction count: ", err.Error())
	}
	c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))

	body, _ := json.Marshal(&tokenTransfers)
	return c.SendString(string(body))
}

// TokenAddressesTokenContract
// @Summary Get Token Holders By Token Contract
// @Description get token holders
// @Tags Transactions
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param token_contract_address path string true "find by token contract address"
// @Router /api/v1/transactions/token-holders/token-contract/{token_contract_address} [get]
// @Success 200 {object} []models.TokenAddress
// @Failure 422 {object} map[string]interface{}
func handlerGetTokenAddressesTokenContract(c *fiber.Ctx) error {
	tokenContractAddress := c.Params("token_contract_address")
	if tokenContractAddress == "" {
		c.Status(422)
		return c.SendString(`{"error": "token_contract_address required"}`)
	}

	params := new(TransactionsQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Transactions Get Handler ERROR: %s", err.Error())

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

	// Get Transactions
	tokenAddresses, err := crud.GetTokenAddressCrud().SelectManyByTokenContractAddress(
		params.Limit,
		params.Skip,
		tokenContractAddress,
	)
	if err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetTokenAddressesTokenContract",
			" Error=Could not retrieve token transfers: ", err.Error(),
		)
		return c.SendString(`{"error": "could not retrieve transactions"}`)
	}

	if len(*tokenAddresses) == 0 {
		// No Content
		c.Status(204)
	}

	// Get Transactions
	count, err := crud.GetTokenAddressCrud().CountBy(
		"",
		tokenContractAddress,
	)
	if err != nil {
		count = 0
		zap.S().Warn("Could not retrieve token contract address holders count: ", err.Error())
	}
	c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))

	body, _ := json.Marshal(&tokenAddresses)
	return c.SendString(string(body))
}
