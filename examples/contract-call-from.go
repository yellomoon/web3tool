package examples

import (
	"fmt"
	"math/big"

	"github.com/yellomoon/web3tool"
	"github.com/yellomoon/web3tool/abi"
	"github.com/yellomoon/web3tool/contract"
	"github.com/yellomoon/web3tool/jsonrpc"
)

// call a contract and use a custom `from` address
func contractCallFrom() {
	var functions = []string{
		"function totalSupply() view returns (uint256)",
	}

	abiContract, err := abi.NewABIFromList(functions)
	handleErr(err)

	// Matic token
	addr := web3tool.HexToAddress("0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0")

	client, err := jsonrpc.NewClient("https://mainnet.infura.io")
	handleErr(err)

	// from address (msg.sender in solidity)
	from := web3tool.Address{0x1}

	c := contract.NewContract(addr, abiContract, contract.WithSender(from), contract.WithJsonRPC(client.Eth()))
	res, err := c.Call("totalSupply", web3tool.Latest)
	handleErr(err)

	fmt.Printf("TotalSupply: %s", res["totalSupply"].(*big.Int))
}
