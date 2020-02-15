package etherscan

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testEtherscanMainnet(t *testing.T) *Etherscan {
	apiKey := os.Getenv("ETHERSCAN_APIKEY")
	if apiKey == "" {
		t.Skip("Etherscan APIKey not specified")
	}
	return &Etherscan{url: "https://api.etherscan.io", apiKey: apiKey}
}

func TestBlockByNumber(t *testing.T) {
	e := testEtherscanMainnet(t)
	n, err := e.BlockNumber()
	assert.NoError(t, err)
	assert.NotEqual(t, n, 0)
}

func TestGetBlockByNumber(t *testing.T) {
	e := testEtherscanMainnet(t)
	b, err := e.GetBlockByNumber(1, false)
	assert.NoError(t, err)
	assert.Equal(t, b.Number, uint64(1))
}
