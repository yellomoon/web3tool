package ens

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yellomoon/web3tool"
	"github.com/yellomoon/web3tool/testutil"
)

func TestENS_Resolve(t *testing.T) {
	ens, err := NewENS(WithAddress(testutil.TestInfuraEndpoint(t)))
	assert.NoError(t, err)

	addr, err := ens.Resolve("nick.eth")
	assert.NoError(t, err)
	assert.Equal(t, web3tool.HexToAddress("0xb8c2C29ee19D8307cb7255e1Cd9CbDE883A267d5"), addr)

	name, err := ens.ReverseResolve(web3tool.HexToAddress("0xb8c2C29ee19D8307cb7255e1Cd9CbDE883A267d5"))
	assert.NoError(t, err)
	assert.Equal(t, "nick.eth", name)
}
