package service

import (
	"github.com/stretchr/testify/require"
	"github.com/sb-icon/icon-go-api/config"
	"testing"
)

func TestIconNodeServiceGetTotalSupply(t *testing.T) {
	config.ReadEnvironment()
	body, err := IconNodeServiceGetTotalSupply()
	require.Nil(t, err)
	require.NotEmpty(t, body)
}

func TestIconNodeServiceGetBalance(t *testing.T) {
	config.ReadEnvironment()
	body, err := IconNodeServiceGetBalance("hx1000000000000000000000000000000000000000")

	require.Nil(t, err)
	require.NotEmpty(t, body)
}
