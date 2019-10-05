package etherscan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testEtherscanMainnet() *Etherscan {
	return &Etherscan{url: "https://api.etherscan.io"}
}

func TestBlockByNumber(t *testing.T) {
	e := testEtherscanMainnet()
	n, err := e.BlockNumber()
	assert.NoError(t, err)
	assert.NotEqual(t, n, 0)
}

func TestGetBlockByNumber(t *testing.T) {
	e := testEtherscanMainnet()
	b, err := e.GetBlockByNumber(1, false)
	assert.NoError(t, err)
	assert.Equal(t, b.Number, uint64(1))
}
