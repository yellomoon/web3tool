package ens

import (
	"github.com/yellomoon/web3tool"
	"github.com/yellomoon/web3tool/contract"
	"github.com/yellomoon/web3tool/jsonrpc"
	"strings"
)

type ENSResolver struct {
	e        *ENS
	provider *jsonrpc.Eth
}

func NewENSResolver(addr web3tool.Address, provider *jsonrpc.Client) *ENSResolver {
	return &ENSResolver{NewENS(addr, contract.WithJsonRPC(provider.Eth())), provider.Eth()}
}

func (e *ENSResolver) Resolve(addr string, block ...web3tool.BlockNumber) (res web3tool.Address, err error) {
	addrHash := NameHash(addr)
	resolver, err := e.prepareResolver(addrHash, block...)
	if err != nil {
		return
	}
	res, err = resolver.Addr(addrHash, block...)
	return
}

func addressToReverseDomain(addr web3tool.Address) string {
	return strings.ToLower(strings.TrimPrefix(addr.String(), "0x")) + ".addr.reverse"
}

func (e *ENSResolver) ReverseResolve(addr web3tool.Address, block ...web3tool.BlockNumber) (res string, err error) {
	addrHash := NameHash(addressToReverseDomain(addr))

	resolver, err := e.prepareResolver(addrHash, block...)
	if err != nil {
		return
	}
	res, err = resolver.Name(addrHash, block...)
	return
}

func (e *ENSResolver) prepareResolver(inputHash web3tool.Hash, block ...web3tool.BlockNumber) (*Resolver, error) {
	resolverAddr, err := e.e.Resolver(inputHash, block...)
	if err != nil {
		return nil, err
	}

	resolver := NewResolver(resolverAddr, contract.WithJsonRPC(e.provider))
	return resolver, nil
}
