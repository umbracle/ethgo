package testutil

import (
	"testing"

	"github.com/cloudwalk/ethgo"
	"github.com/stretchr/testify/require"
)

func TestDeployServer(t *testing.T) {
	srv := DeployTestServer(t, nil)
	require.NotEmpty(t, srv.accounts)

	clt := &ethClient{srv.HTTPAddr()}
	account := []ethgo.Address{}

	err := clt.call("eth_accounts", &account)
	require.NoError(t, err)
}
