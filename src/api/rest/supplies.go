package rest

import (
	"encoding/json"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sb-icon/icon-go-api/config"
)

func SuppliesAddHandlers(app *fiber.App) {
	prefix := config.Config.RestPrefix + "/supplies"

	app.Get(prefix+"/", handlerGetSupplies)
	app.Get(prefix+"/circulating-supply", handlerGetSuppliesCirculatingSupply)
	app.Get(prefix+"/total-supply", handlerGetSuppliesTotalSupply)
	app.Get(prefix+"/market-cap", handlerGetSuppliesMarketCap)
}

// Supplies
// @Summary Get Supplies
// @Description get json with a summary of stats
// @Tags Supplies
// @BasePath /api/v1
// @Accept */*
// @Router /api/v1/supplies [get]
// @Success 200 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
func handlerGetSupplies(c *fiber.Ctx) error {
	UpdateCirculatingSupply()
	UpdateMarketCap()

	stats := map[string]interface{}{
		"circulating-supply": CirculatingSupply,
		"market-cap":         MarketCap,
	}
	body, _ := json.Marshal(stats)

	return c.SendString(string(body))
}

// Circulating Supply
// @Summary Get Circulating Supply
// @Description get circulating supply (total supply - burn wallet balance)
// @Tags Supplies
// @BasePath /api/v1
// @Accept */*
// @Router /api/v1/supplies/circulating-supply [get]
// @Success 200 {object} float64
// @Failure 422 {object} map[string]interface{}
func handlerGetSuppliesCirculatingSupply(c *fiber.Ctx) error {
	UpdateCirculatingSupply()
	return c.SendString(strconv.FormatFloat(CirculatingSupply, 'f', -1, 64))
}

// Total Supply
// @Summary Get Total Supply
// @Description get total supply
// @Tags Supplies
// @BasePath /api/v1
// @Accept */*
// @Router /api/v1/supplies/total-supply [get]
// @Success 200 {object} float64
// @Failure 422 {object} map[string]interface{}
func handlerGetSuppliesTotalSupply(c *fiber.Ctx) error {
	UpdateCirculatingSupply()
	return c.SendString(strconv.FormatFloat(TotalSupply, 'f', -1, 64))
}

// Market Cap
// @Summary Get Market Cap
// @Description get mkt cap (Coin Gecko Price * circulating supply)
// @Tags Supplies
// @BasePath /api/v1
// @Accept */*
// @Router /api/v1/supplies/market-cap [get]
// @Success 200 {object} float64
// @Failure 422 {object} map[string]interface{}
func handlerGetSuppliesMarketCap(c *fiber.Ctx) error {
	UpdateMarketCap()
	return c.SendString(strconv.FormatFloat(MarketCap, 'f', -1, 64))
}
