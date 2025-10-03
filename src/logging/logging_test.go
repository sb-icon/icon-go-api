package logging

import (
	"os"
	"testing"

	"go.uber.org/zap"

	"github.com/sb-icon/icon-go-api/config"
)

func TestInit(t *testing.T) {
	os.Setenv("LOG_LEVEL", "Info")
	config.ReadEnvironment()

	Init()

	zap.S().Info("Test Log")
}
