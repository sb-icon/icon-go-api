package rest

import (
	"encoding/json"
	"errors"
	"strconv"

	"gorm.io/gorm"

	fiber "github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"github.com/sb-icon/icon-go-api/config"
	"github.com/sb-icon/icon-go-api/crud"
	"github.com/sb-icon/icon-go-api/redis"
)

type AddressesQuery struct {
	Limit         int    `query:"limit"`
	Skip          int    `query:"skip"`
	Address       string `query:"address"`
	IsContract    *bool  `query:"is_contract"`
	IsToken       *bool  `query:"is_token"`
	IsNft         *bool  `query:"is_nft"`
	TokenStandard string `query:"token_standard"`
	Sort          string `query:"sort"`
}

func AddressesAddHandlers(app *fiber.App) {

	prefix := config.Config.RestPrefix + "/addresses"

	app.Get(prefix+"/", handlerGetAddresses)
	app.Get(prefix+"/details/:address", handlerGetAddressDetails)
	app.Get(prefix+"/contracts", handlerGetContracts)
	app.Get(prefix+"/token-addresses/:address", handlerGetTokenAddresses)
}

// Addresses
// @Summary Get Addresses
// @Description get list of addresses
// @Tags Addresses
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param limit query int false "Amount of records"
// @Param skip query int false "Skip to a record"
// @Param address query string false "Find by address"
// @Param is_contract query bool false "Contract addresses only"
// @Param is_token query bool false "Token addresses only"
// @Param is_nft query bool false "NFT addresses only"
// @Param token_standard query string false "Token standard, either irc2, irc3, irc31"
// @Param sort query string false "Field to sort by. name, balance, transaction_count, transaction_internal_count, token_transfer_count. Use leading `-` (ie -balance) for sort direction or omit for descending."
// @Router /api/v1/addresses [get]
// @Success 200 {object} []models.AddressList
// @Failure 422 {object} map[string]interface{}
func handlerGetAddresses(c *fiber.Ctx) error {
	params := new(AddressesQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Addresses Get Handler ERROR: %s", err.Error())

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

	if params.Sort != "" {
		// Check if the sort is valid. Needed so that unindexed params are not sorted on.
		var sortParam string
		sortFirstChar := params.Sort[0:1]
		if sortFirstChar == "-" {
			sortParam = params.Sort[1:]
		} else {
			sortParam = params.Sort
		}

		if !stringInSlice(sortParam, addressSortParams) {
			c.Status(422)
			return c.SendString(`{"error": "invalid sort parameter"}`)
		}
	}

	// Get Addresses
	addresses, err := crud.GetAddressCrud().SelectMany(
		params.Limit,
		params.Skip,
		params.Address,
		params.IsContract,
		params.IsToken,
		params.IsNft,
		params.TokenStandard,
		params.Sort,
	)
	if err != nil {
		zap.S().Warnf("Addresses CRUD ERROR: %s", err.Error())
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetAddresses",
			" Error=Could not retrieve addresses: ", err.Error(),
		)
		return c.SendString(`{"error": "could not retrieve addresses"}`)
	}

	if len(*addresses) == 0 {
		// No Content
		c.Status(204)
	}

	// TODO: RM and replace with concurrent count?
	// Set X-TOTAL-COUNT
	count, err := redis.GetRedisClient().GetCount(config.Config.RedisKeyPrefix + "address_count")
	if err != nil {
		count = 0
		zap.S().Warn("Error: ", err.Error())
	}
	c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))

	if c.Get("Accept") == "text/csv" {
		return respondWithCSV(c, *addresses)
	}

	// Continue with JSON response if not CSV
	body, err := json.Marshal(addresses)
	if err != nil {
		return c.SendString(`{"error": "parsing error"}`)
	}

	return c.SendString(string(body))
}

// Address Details
// @Summary Get Address Details
// @Description get details of an address
// @Tags Addresses
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param address path string true "find by address"
// @Router /api/v1/addresses/details/{address} [get]
// @Success 200 {object} models.Address
// @Failure 422 {object} map[string]interface{}
func handlerGetAddressDetails(c *fiber.Ctx) error {
	addressString := c.Params("address")
	if addressString == "" {
		c.Status(422)
		return c.SendString(`{"error": "address required"}`)
	}

	params := new(SkipLimitQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Addresses Get Handler ERROR: %s", err.Error())
		c.Status(422)
		return c.SendString(`{"error": "could not parse query parameters"}`)
	}

	// Get Addresses
	address, err := crud.GetAddressCrud().SelectOne(addressString)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Status(404)
			return c.SendString(`{"error": "address not found"}`)
		}
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetAddressDetails",
			" Error=Could not retrieve addresses: ", err.Error(),
		)
		return c.SendString(`{"error": "could not retrieve addresses"}`)
	}

	// Continue with JSON response if not CSV
	body, err := json.Marshal(address)
	if err != nil {
		return c.SendString(`{"error": "parsing error"}`)
	}
	return c.SendString(string(body))
}

type ContractsQuery struct {
	Limit         int    `query:"limit"`
	Skip          int    `query:"skip"`
	Search        string `query:"search"`
	IsToken       *bool  `query:"is_token"`
	IsNft         *bool  `query:"is_nft"`
	TokenStandard string `query:"token_standard"`
	Status        string `query:"status"`
	Sort          string `query:"sort"`
}

// Contract
// @Summary Get Contract Addresses
// @Description get list of contracts
// @Tags Addresses
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param search query string false "contract name search"
// @Param is_token query bool false "tokens only"
// @Param is_nft query bool false "NFTs only"
// @Param token_standard query string false "token standard, one of irc2,irc3,irc31"
// @Param status query string false "contract status, one of active, rejected, or pending"
// @Param limit query int false "amount of records"
// @Param skip query int false "skip to a record"
// @Param sort query string false "Field to sort by. name, balance, transaction_count, transaction_internal_count, token_transfer_count. Use leading `-` (ie -balance) for sort direction or omit for descending."
// @Router /api/v1/addresses/contracts [get]
// @Success 200 {object} []models.ContractList
// @Failure 422 {object} map[string]interface{}
func handlerGetContracts(c *fiber.Ctx) error {
	params := new(ContractsQuery)
	if err := c.QueryParser(params); err != nil {
		zap.S().Warnf("Addresses Get Handler ERROR: %s", err.Error())

		c.Status(422)
		return c.SendString(`{"error": "could not parse query parameters"}`)
	}

	if params.TokenStandard != "" {
		params.IsToken = newTrue()
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

	if params.Sort != "" {
		// Check if the sort is valid. Needed so that unindexed params are not sorted on.
		var sortParam string
		sortFirstChar := params.Sort[0:1]
		if sortFirstChar == "-" {
			sortParam = params.Sort[1:]
		} else {
			sortParam = params.Sort
		}

		if !stringInSlice(sortParam, addressSortParams) {
			c.Status(422)
			return c.SendString(`{"error": "invalid sort parameter"}`)
		}
	}

	// Get contracts
	contracts, err := crud.GetAddressCrud().SelectManyContracts(
		params.Search,
		params.TokenStandard,
		params.IsToken,
		params.IsNft,
		params.Status,
		params.Limit,
		params.Skip,
		params.Sort,
	)
	if err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetContracts",
			" Error=Could not retrieve contracts: ", err.Error(),
		)
		return c.SendString(`{"error": "could not retrieve contracts"}`)
	}

	if len(*contracts) == 0 {
		// No Content
		c.Status(204)
	}

	if params.Search != "" || params.TokenStandard != "" || params.IsToken != nil || params.IsNft != nil {
		count, err := crud.GetAddressCrud().CountWithParamsSearch(
			params.Search,
			params.TokenStandard,
			params.Status,
			params.IsToken,
			params.IsNft,
			newTrue(),
		)
		if err != nil {
			c.Status(500)
			zap.S().Warn(
				"Endpoint=handlerGetContracts",
				" Error=Could not retrieve contracts: ", err.Error(),
			)
			return c.SendString(`{"error": "could not count contracts"}`)
		}
		c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))
	} else {
		// Set X-TOTAL-COUNT
		count, err := redis.GetRedisClient().GetCount(config.Config.RedisKeyPrefix + "address_contract_count")
		if err != nil {
			count = 0
			zap.S().Warn("Error: ", err.Error())
		}
		c.Append("X-TOTAL-COUNT", strconv.FormatInt(count, 10))
	}

	if c.Get("Accept") == "text/csv" {
		return respondWithCSV(c, *contracts)
	}

	// Continue with JSON response if not CSV
	body, err := json.Marshal(contracts)
	if err != nil {
		return c.SendString(`{"error": "parsing error"}`)
	}

	return c.SendString(string(body))
}

// Token Addresses
// @Summary Get Token Addresses
// @Description get list of token contracts by address
// @Tags Addresses
// @BasePath /api/v1
// @Accept */*
// @Produce json
// @Param address path string true "address"
// @Router /api/v1/addresses/token-addresses/{address} [get]
// @Success 200 {object} []string
// @Failure 422 {object} map[string]interface{}
func handlerGetTokenAddresses(c *fiber.Ctx) error {
	addressString := c.Params("address")

	params := new(SkipLimitQuery)
	if err := c.QueryParser(params); err != nil {
		c.Status(422)
		return c.SendString(`{"error": "could not parse query parameters. Only 'skip' and 'limit' accepted."}`)
	}

	// Get TokenAddresses
	tokenAddress, err := crud.GetTokenAddressCrud().SelectManyByAddress(addressString)
	if err != nil {
		c.Status(500)
		zap.S().Warn(
			"Endpoint=handlerGetContracts",
			" Error=Could not retrieve token addresses: ", err.Error(),
		)
		return c.SendString(`{"error": "could not retrieve addresses"}`)
	}

	if len(*tokenAddress) == 0 {
		// No Content
		c.Status(204)
	}

	// []models.TokenAddress -> []string
	var tokenContractAddresses []string
	for _, a := range *tokenAddress {
		tokenContractAddresses = append(tokenContractAddresses, a.TokenContractAddress)
	}

	body, _ := json.Marshal(&tokenContractAddresses)
	return c.SendString(string(body))
}
