package main

import (
	"log"

	"github.com/sb-icon/icon-go-api/api"
	"github.com/sb-icon/icon-go-api/config"
	"github.com/sb-icon/icon-go-api/global"
	"github.com/sb-icon/icon-go-api/healthcheck"
	"github.com/sb-icon/icon-go-api/logging"
	"github.com/sb-icon/icon-go-api/metrics"
	_ "github.com/sb-icon/icon-go-api/models" // for swagger docs
	"github.com/sb-icon/icon-go-api/redis"
)

func main() {
	config.ReadEnvironment()

	logging.Init()
	log.Printf("Main: Starting logging with level %s", config.Config.LogLevel)

	// Start Prometheus client
	metrics.Start()

	// Start Redis Client
	// NOTE: redis is used for websockets
	redis.GetRedisClient().StartSubscribers()

	// Start API server
	api.Start()

	// Start Health server
	healthcheck.Start()

	global.WaitShutdownSig()
}
