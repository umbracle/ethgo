package ens

import (
	"encoding/hex"
	"testing"

	"github.com/Ethernal-Tech/ethgo"
	"github.com/Ethernal-Tech/ethgo/jsonrpc"
	"github.com/Ethernal-Tech/ethgo/testutil"
	"github.com/stretchr/testify/assert"
)

var (
	mainnetAddr = ethgo.HexToAddress("0x314159265dD8dbb310642f98f50C066173C1259b")
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
