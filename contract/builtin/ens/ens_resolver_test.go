package ens

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	web3 "github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/testutil"
)

var (
	url         = "https://mainnet.infura.io"
	mainnetAddr = web3.HexToAddress("0x314159265dD8dbb310642f98f50C066173C1259b")
)

func TestResolveAddr(t *testing.T) {
	c, _ := jsonrpc.NewClient(testutil.TestInfuraEndpoint(t))
	r := NewENSResolver(mainnetAddr, c)

	cases := []struct {
		Addr     string
		Expected string
	}{
		{
			Addr:     "arachnid.eth",
			Expected: "0xfdb33f8ac7ce72d7d4795dd8610e323b4c122fbb",
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			found, err := r.Resolve(c.Addr)
			assert.NoError(t, err)
			assert.Equal(t, "0x"+hex.EncodeToString(found[:]), c.Expected)
		})
	}
}
