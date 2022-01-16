package gaspricer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/go-web3/jsonrpc"
	"github.com/umbracle/go-web3/testutil"
)

func TestGasPricer_Network(t *testing.T) {
	srv := testutil.NewTestServer(t, nil)
	defer srv.Close()

	client, err := jsonrpc.NewClient(srv.HTTPAddr())
	assert.NoError(t, err)

	pricer, err := NewNetworkGasPricer(nil, client.Eth())
	assert.NoError(t, err)

	assert.Equal(t, pricer.GasPrice(), uint64(1))
}
