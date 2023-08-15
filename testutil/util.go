package testutil

import (
	"math/big"
	"reflect"

	"github.com/yellomoon/web3tool"
)

func CompareLogs(one, two []*web3tool.Log) bool {
	if len(one) != len(two) {
		return false
	}
	if len(one) == 0 {
		return true
	}
	return reflect.DeepEqual(one, two)
}

func CompareBlocks(one, two []*web3tool.Block) bool {
	if len(one) != len(two) {
		return false
	}
	if len(one) == 0 {
		return true
	}
	// difficulty is hard to check, set the values to zero
	for _, i := range one {
		if i.Transactions == nil {
			i.Transactions = []*web3tool.Transaction{}
		}
		i.Difficulty = big.NewInt(0)
	}
	for _, i := range two {
		if i.Transactions == nil {
			i.Transactions = []*web3tool.Transaction{}
		}
		i.Difficulty = big.NewInt(0)
	}
	return reflect.DeepEqual(one, two)
}
