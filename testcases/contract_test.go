package testcases

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yellomoon/web3tool"
	"github.com/yellomoon/web3tool/abi"
	"github.com/yellomoon/web3tool/testutil"
)

func TestContract_Signatures(t *testing.T) {
	var signatures []struct {
		Name      string `json:"name"`
		Signature string `json:"signature"`
		SigHash   string `json:"sigHash"`
		Abi       string `json:"abi"`
	}
	ReadTestCase(t, "contract-signatures", &signatures)

	for _, c := range signatures {
		m, err := abi.NewMethod(c.Signature)
		assert.NoError(t, err)

		sigHash := "0x" + hex.EncodeToString(m.ID())
		assert.Equal(t, sigHash, c.SigHash)
	}
}

func TestContract_Interface(t *testing.T) {
	t.Skip()

	server := testutil.NewTestServer(t)

	var calls []struct {
		Name      string         `json:"name"`
		Interface string         `json:"interface"`
		Bytecode  web3tool.ArgBytes `json:"bytecode"`
		Result    web3tool.ArgBytes `json:"result"`
		Values    string         `json:"values"`
	}
	ReadTestCase(t, "contract-interface", &calls)

	for _, c := range calls {
		a, err := abi.NewABI(c.Interface)
		assert.NoError(t, err)

		method := a.GetMethod("test")

		receipt, err := server.SendTxn(&web3tool.Transaction{
			Input: c.Bytecode.Bytes(),
		})
		assert.NoError(t, err)

		outputRaw, err := server.Call(&web3tool.CallMsg{
			To:   &receipt.ContractAddress,
			Data: method.ID(),
		})
		assert.NoError(t, err)

		output, err := hex.DecodeString(outputRaw[2:])
		assert.NoError(t, err)

		_, err = method.Decode(output)
		assert.NoError(t, err)
	}

}
