package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yellomoon/web3tool"
)

func TestDeployServer(t *testing.T) {
	srv := DeployTestServer(t, nil)
	require.NotEmpty(t, srv.accounts)

	clt := &ethClient{srv.HTTPAddr()}
	account := []web3tool.Address{}

	err := clt.call("eth_accounts", &account)
	require.NoError(t, err)
}
