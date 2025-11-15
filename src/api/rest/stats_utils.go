package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sb-icon/icon-go-api/config"
	"github.com/sb-icon/icon-go-api/service"
	"go.uber.org/zap"
)

var CirculatingSupply float64
var TotalSupply float64

func GetCirculatingSupply() (float64, error) {
	totalSupply, err := service.IconNodeServiceGetTotalSupply()
	TotalSupply = totalSupply
	if err != nil {
		return 0, err
	}

	burnBalance, err := service.IconNodeServiceGetBalance("hx1000000000000000000000000000000000000000")
	if err != nil {
		return 0, err
	}
	circulatingSupply := totalSupply - burnBalance
	return circulatingSupply, err
}

var LastUpdatedTimeCirculatingSupply time.Time

func UpdateCirculatingSupply() {
	timeDiff := time.Now().Sub(LastUpdatedTimeCirculatingSupply)
	if timeDiff > config.Config.StatsCirculatingSupplyUpdateTime {
		circulatingSupply, err := GetCirculatingSupply()
		if err != nil {
			zap.S().Info("Error getting circulating-supply: ", err)
		}
		CirculatingSupply = circulatingSupply
		LastUpdatedTimeCirculatingSupply = time.Now()
	}
}

var MarketCap float64

func GetMarketCap() (float64, error) {
	req, err := http.NewRequest("GET", "https://api.coingecko.com/api/v3/coins/icon", nil)
	if err != nil {
		return 0.0, err
	}
	// coingecko is blocking requests without a user agent so spoofing here
	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3",
	)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0.0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0.0, err
	}

	response := make(map[string]interface{})
	if err := json.Unmarshal(body, &response); err != nil {
		// Just return cached value if there is an error
		return MarketCap, err
	}

	// Defensive navigation through response["market_data"]["current_price"]["usd"]
	rawMarketData, ok := response["market_data"]
	if !ok || rawMarketData == nil {
		return MarketCap, errors.New(
			fmt.Sprintf("Error parsing coingecko response: market_data missing or nil (type=%T)", rawMarketData),
		)
	}

	marketData, ok := rawMarketData.(map[string]interface{})
	if !ok {
		return MarketCap, errors.New(
			fmt.Sprintf("Error parsing coingecko response: market_data not an object (type=%T)", rawMarketData),
		)
	}

	rawCurrentPrice, ok := marketData["current_price"]
	if !ok || rawCurrentPrice == nil {
		return MarketCap, errors.New(
			fmt.Sprintf("Error parsing coingecko response: current_price missing or nil (type=%T)", rawCurrentPrice),
		)
	}

	currentPrice, ok := rawCurrentPrice.(map[string]interface{})
	if !ok {
		return MarketCap, errors.New(
			fmt.Sprintf("Error parsing coingecko response: current_price not an object (type=%T)", rawCurrentPrice),
		)
	}

	rawUSD, ok := currentPrice["usd"]
	if !ok || rawUSD == nil {
		return MarketCap, errors.New("Error parsing coingecko response: usd price missing")
	}

	usdPrice, ok := rawUSD.(float64)
	if !ok {
		return MarketCap, errors.New(
			fmt.Sprintf("Error parsing coingecko response: usd is not float64 (type=%T)", rawUSD),
		)
	}

	// Now we have a valid usd price, update circulating supply and market cap
	UpdateCirculatingSupply()
	MarketCap = CirculatingSupply * usdPrice
	return MarketCap, nil
}

var LastUpdatedTimeMarketCap time.Time

func UpdateMarketCap() {
	timeDiff := time.Now().Sub(LastUpdatedTimeMarketCap)
	if timeDiff > config.Config.StatsMarketCapUpdateTime {
		marketCap, err := GetMarketCap()
		if err != nil {
			zap.S().Info("Error getting market-cap: ", err)
		}
		MarketCap = marketCap
		LastUpdatedTimeMarketCap = time.Now()
	}
}
