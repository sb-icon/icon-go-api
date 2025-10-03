package api

import (
	"github.com/stretchr/testify/assert"
	"github.com/sb-icon/icon-go-api/config"
	"net/http/httptest"
	"testing"
)

func TestRestApi(t *testing.T) {
	config.ReadEnvironment()

	tests := []struct {
		description  string
		route        string
		expectedCode int
		header       string
	}{
		{
			description:  "Docs",
			route:        "/api/v1/docs/index.html",
			expectedCode: 200,
			header:       "",
		},
		{
			description:  "Version",
			route:        "/version",
			expectedCode: 200,
			header:       "",
		},
		{
			description:  "get HTTP status 404, when route is not exists",
			route:        "/not-found",
			expectedCode: 404,
			header:       "",
		},
		{
			description:  "addresses",
			route:        "/api/v1/addresses",
			expectedCode: 200,
			header:       "",
		},
		{
			description:  "addresses",
			route:        "/api/v1/addresses",
			expectedCode: 200,
			header:       "text/csv",
		},
		{
			description:  "addresses contracts",
			route:        "/api/v1/addresses/contracts",
			expectedCode: 200,
			header:       "",
		},
		{
			description:  "addresses contracts",
			route:        "/api/v1/addresses/contracts",
			expectedCode: 200,
			header:       "text/csv",
		},
		{
			description:  "txs",
			route:        "/api/v1/transactions",
			expectedCode: 200,
			header:       "",
		},
		{
			description:  "txs",
			route:        "/api/v1/transactions",
			expectedCode: 200,
			header:       "text/csv",
		},
		{
			description:  "token-transfers",
			route:        "/api/v1/transactions/token-transfers",
			expectedCode: 200,
			header:       "",
		},
		{
			description:  "token-transfers",
			route:        "/api/v1/transactions/token-transfers",
			expectedCode: 200,
			header:       "text/csv",
		},
		{
			description:  "logs",
			route:        "/api/v1/logs",
			expectedCode: 200,
			header:       "",
		},
	}

	app := Start()

	for _, test := range tests {
		req := httptest.NewRequest("GET", test.route, nil)
		req.Header.Add("Accept", test.header) // Add the header
		resp, err := app.Test(req, 10000)
		assert.Equalf(t, nil, err, "app.Test(req)")
		assert.Equalf(t, test.expectedCode, resp.StatusCode, test.description)
	}
}
