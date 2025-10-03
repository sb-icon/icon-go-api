//go:build unit
// +build unit

package metrics

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/sb-icon/icon-go-api/config"

	"github.com/stretchr/testify/assert"
)

func TestMetricsApiStart(t *testing.T) {
	assert := assert.New(t)

	// Set env
	os.Setenv("METRICS_PORT", "8888")
	os.Setenv("METRICS_PREFIX", "/metrics")

	config.ReadEnvironment()

	// Start metrics server
	Start()

	MaxBlockNumberTransactionsRawGauge.Inc()
	MaxBlockNumberBlocksRawGauge.Set(2)
	MaxBlockNumberLogsRawGauge.Set(2)

	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%s%s", config.Config.MetricsPort, config.Config.MetricsPrefix))
	assert.Equal(nil, err)
	assert.Equal(200, resp.StatusCode)
}
