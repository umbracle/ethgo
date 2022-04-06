package etherscan

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/ethgo"
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

func TestContract(t *testing.T) {
	e := testEtherscanMainnet(t)

	// uniswap v2. router
	code, err := e.GetContractCode(ethgo.HexToAddress("0x7a250d5630b4cf539739df2c5dacb4c659f2488d"))
	assert.NoError(t, err)
	assert.Equal(t, code.Runs, "999999")
}

func TestGetLogs(t *testing.T) {
	e := testEtherscanMainnet(t)

	from := ethgo.BlockNumber(379224)
	to := ethgo.Latest

	filter := &ethgo.LogFilter{
		From: &from,
		To:   &to,
		Address: []ethgo.Address{
			ethgo.HexToAddress("0x33990122638b9132ca29c723bdf037f1a891a70c"),
		},
	}
	logs, err := e.GetLogs(filter)
	assert.NoError(t, err)
	assert.NotEmpty(t, logs)
}

func TestGasPrice(t *testing.T) {
	e := testEtherscanMainnet(t)

	gas, err := e.GasPrice()
	assert.NoError(t, err)
	assert.NotZero(t, gas)
}
