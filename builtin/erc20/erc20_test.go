package erc20

import (
	"testing"

	"github.com/Ethernal-Tech/ethgo"
	"github.com/Ethernal-Tech/ethgo/contract"
	"github.com/Ethernal-Tech/ethgo/jsonrpc"
	"github.com/Ethernal-Tech/ethgo/testutil"
	"github.com/stretchr/testify/assert"
)

var (
	zeroX = ethgo.HexToAddress("0xe41d2489571d322189246dafa5ebde1f4699f498")
)

func TestERC20Decimals(t *testing.T) {
	c, _ := jsonrpc.NewClient(testutil.TestInfuraEndpoint(t))
	erc20 := NewERC20(zeroX, contract.WithJsonRPC(c.Eth()))

	decimals, err := erc20.Decimals()
	assert.NoError(t, err)
	if decimals != 18 {
		t.Fatal("bad")
	}
}

func TestERC20Name(t *testing.T) {
	c, _ := jsonrpc.NewClient(testutil.TestInfuraEndpoint(t))
	erc20 := NewERC20(zeroX, contract.WithJsonRPC(c.Eth()))

	name, err := erc20.Name()
	assert.NoError(t, err)
	assert.Equal(t, name, "0x Protocol Token")
}

func TestERC20Symbol(t *testing.T) {
	c, _ := jsonrpc.NewClient(testutil.TestInfuraEndpoint(t))
	erc20 := NewERC20(zeroX, contract.WithJsonRPC(c.Eth()))

	symbol, err := erc20.Symbol()
	assert.NoError(t, err)
	assert.Equal(t, symbol, "ZRX")
}

func TestTotalSupply(t *testing.T) {
	c, _ := jsonrpc.NewClient(testutil.TestInfuraEndpoint(t))
	erc20 := NewERC20(zeroX, contract.WithJsonRPC(c.Eth()))

	supply, err := erc20.TotalSupply()
	assert.NoError(t, err)
	assert.Equal(t, supply.String(), "1000000000000000000000000000")
}
