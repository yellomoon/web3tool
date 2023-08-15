// Code generated by web3tool/abigen. DO NOT EDIT.
// Hash: bfee2618a5908e1a24f19dcce873d3b8e797374138dd7604f7b593db3cca5c17
// Version: 0.1.1
package ens

import (
	"fmt"
	"math/big"

	"github.com/yellomoon/web3tool"
	"github.com/yellomoon/web3tool/contract"
	"github.com/yellomoon/web3tool/jsonrpc"
)

var (
	_ = big.NewInt
	_ = jsonrpc.NewClient
)

// ENS is a solidity contract
type ENS struct {
	c *contract.Contract
}

// DeployENS deploys a new ENS contract
func DeployENS(provider *jsonrpc.Client, from web3tool.Address, args []interface{}, opts ...contract.ContractOption) (contract.Txn, error) {
	return contract.DeployContract(abiENS, binENS, args, opts...)
}

// NewENS creates a new instance of the contract at a specific address
func NewENS(addr web3tool.Address, opts ...contract.ContractOption) *ENS {
	return &ENS{c: contract.NewContract(addr, abiENS, opts...)}
}

// calls

// Owner calls the owner method in the solidity contract
func (e *ENS) Owner(node [32]byte, block ...web3tool.BlockNumber) (retval0 web3tool.Address, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = e.c.Call("owner", web3tool.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["0"].(web3tool.Address)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Resolver calls the resolver method in the solidity contract
func (e *ENS) Resolver(node [32]byte, block ...web3tool.BlockNumber) (retval0 web3tool.Address, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = e.c.Call("resolver", web3tool.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["0"].(web3tool.Address)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// Ttl calls the ttl method in the solidity contract
func (e *ENS) Ttl(node [32]byte, block ...web3tool.BlockNumber) (retval0 uint64, err error) {
	var out map[string]interface{}
	var ok bool

	out, err = e.c.Call("ttl", web3tool.EncodeBlock(block...), node)
	if err != nil {
		return
	}

	// decode outputs
	retval0, ok = out["0"].(uint64)
	if !ok {
		err = fmt.Errorf("failed to encode output at index 0")
		return
	}
	
	return
}

// txns

// SetOwner sends a setOwner transaction in the solidity contract
func (e *ENS) SetOwner(node [32]byte, owner web3tool.Address) (contract.Txn, error) {
	return e.c.Txn("setOwner", node, owner)
}

// SetResolver sends a setResolver transaction in the solidity contract
func (e *ENS) SetResolver(node [32]byte, resolver web3tool.Address) (contract.Txn, error) {
	return e.c.Txn("setResolver", node, resolver)
}

// SetSubnodeOwner sends a setSubnodeOwner transaction in the solidity contract
func (e *ENS) SetSubnodeOwner(node [32]byte, label [32]byte, owner web3tool.Address) (contract.Txn, error) {
	return e.c.Txn("setSubnodeOwner", node, label, owner)
}

// SetTTL sends a setTTL transaction in the solidity contract
func (e *ENS) SetTTL(node [32]byte, ttl uint64) (contract.Txn, error) {
	return e.c.Txn("setTTL", node, ttl)
}

// events

func (e *ENS) NewOwnerEventSig() web3tool.Hash {
	return e.c.GetABI().Events["NewOwner"].ID()
}

func (e *ENS) NewResolverEventSig() web3tool.Hash {
	return e.c.GetABI().Events["NewResolver"].ID()
}

func (e *ENS) NewTTLEventSig() web3tool.Hash {
	return e.c.GetABI().Events["NewTTL"].ID()
}

func (e *ENS) TransferEventSig() web3tool.Hash {
	return e.c.GetABI().Events["Transfer"].ID()
}