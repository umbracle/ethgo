package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeployServer(t *testing.T) {
	srv := DeployTestServer(t, nil)
	require.NotEmpty(t, srv.accounts)
}
