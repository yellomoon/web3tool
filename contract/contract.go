package contract

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/yellomoon/web3tool"
	"github.com/yellomoon/web3tool/abi"
	"github.com/yellomoon/web3tool/jsonrpc"
	"github.com/yellomoon/web3tool/wallet"
)

// Provider handles the interactions with the Ethereum 1x node
type Provider interface {
	Call(web3tool.Address, []byte, *CallOpts) ([]byte, error)
	Txn(web3tool.Address, web3tool.Key, []byte) (Txn, error)
}

type jsonRPCNodeProvider struct {
	client  *jsonrpc.Eth
	eip1559 bool
}

func (j *jsonRPCNodeProvider) Call(addr web3tool.Address, input []byte, opts *CallOpts) ([]byte, error) {
	msg := &web3tool.CallMsg{
		To:   &addr,
		Data: input,
	}
	if opts.From != web3tool.ZeroAddress {
		msg.From = opts.From
	}
	rawStr, err := j.client.Call(msg, opts.Block)
	if err != nil {
		return nil, err
	}
	raw, err := hex.DecodeString(rawStr[2:])
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func (j *jsonRPCNodeProvider) Txn(addr web3tool.Address, key web3tool.Key, input []byte) (Txn, error) {
	txn := &jsonrpcTransaction{
		opts:    &TxnOpts{},
		input:   input,
		client:  j.client,
		key:     key,
		to:      addr,
		eip1559: j.eip1559,
	}
	return txn, nil
}

type jsonrpcTransaction struct {
	to      web3tool.Address
	input   []byte
	hash    web3tool.Hash
	opts    *TxnOpts
	key     web3tool.Key
	client  *jsonrpc.Eth
	txn     *web3tool.Transaction
	txnRaw  []byte
	eip1559 bool
}

func (j *jsonrpcTransaction) Hash() web3tool.Hash {
	return j.hash
}

func (j *jsonrpcTransaction) WithOpts(opts *TxnOpts) {
	j.opts = opts
}

func (j *jsonrpcTransaction) Build() error {
	var err error
	from := j.key.Address()

	// estimate gas price
	if j.opts.GasPrice == 0 && !j.eip1559 {
		j.opts.GasPrice, err = j.client.GasPrice()
		if err != nil {
			return err
		}
	}
	// estimate gas limit
	if j.opts.GasLimit == 0 {
		msg := &web3tool.CallMsg{
			From:     from,
			To:       nil,
			Data:     j.input,
			Value:    j.opts.Value,
			GasPrice: j.opts.GasPrice,
		}
		if j.to != web3tool.ZeroAddress {
			msg.To = &j.to
		}
		j.opts.GasLimit, err = j.client.EstimateGas(msg)
		if err != nil {
			return err
		}
	}
	// calculate the nonce
	if j.opts.Nonce == 0 {
		j.opts.Nonce, err = j.client.GetNonce(from, web3tool.Latest)
		if err != nil {
			return fmt.Errorf("failed to calculate nonce: %v", err)
		}
	}

	chainID, err := j.client.ChainID()
	if err != nil {
		return err
	}

	// send transaction
	rawTxn := &web3tool.Transaction{
		From:     from,
		Input:    j.input,
		GasPrice: j.opts.GasPrice,
		Gas:      j.opts.GasLimit,
		Value:    j.opts.Value,
		Nonce:    j.opts.Nonce,
		ChainID:  chainID,
	}
	if j.to != web3tool.ZeroAddress {
		rawTxn.To = &j.to
	}

	if j.eip1559 {
		rawTxn.Type = web3tool.TransactionDynamicFee

		// use gas price as fee data
		gasPrice, err := j.client.GasPrice()
		if err != nil {
			return err
		}
		rawTxn.MaxFeePerGas = new(big.Int).SetUint64(gasPrice)
		rawTxn.MaxPriorityFeePerGas = new(big.Int).SetUint64(gasPrice)
	}

	j.txn = rawTxn
	return nil
}

func (j *jsonrpcTransaction) Do() error {
	if j.txn == nil {
		if err := j.Build(); err != nil {
			return err
		}
	}

	signer := wallet.NewEIP155Signer(j.txn.ChainID.Uint64())
	signedTxn, err := signer.SignTx(j.txn, j.key)
	if err != nil {
		return err
	}
	txnRaw, err := signedTxn.MarshalRLPTo(nil)
	if err != nil {
		return err
	}

	j.txnRaw = txnRaw
	hash, err := j.client.SendRawTransaction(j.txnRaw)
	if err != nil {
		return err
	}
	j.hash = hash
	return nil
}

func (j *jsonrpcTransaction) Wait() (*web3tool.Receipt, error) {
	if (j.hash == web3tool.Hash{}) {
		panic("transaction not executed")
	}

	for {
		receipt, err := j.client.GetTransactionReceipt(j.hash)
		if err != nil {
			if err.Error() != "not found" {
				return nil, err
			}
		}
		if receipt != nil {
			return receipt, nil
		}
	}
}

// Txn is the transaction object returned
type Txn interface {
	Hash() web3tool.Hash
	WithOpts(opts *TxnOpts)
	Do() error
	Wait() (*web3tool.Receipt, error)
}

type Opts struct {
	JsonRPCEndpoint string
	JsonRPCClient   *jsonrpc.Eth
	Provider        Provider
	Sender          web3tool.Key
	EIP1559         bool
}

type ContractOption func(*Opts)

func WithJsonRPCEndpoint(endpoint string) ContractOption {
	return func(o *Opts) {
		o.JsonRPCEndpoint = endpoint
	}
}

func WithJsonRPC(client *jsonrpc.Eth) ContractOption {
	return func(o *Opts) {
		o.JsonRPCClient = client
	}
}

func WithProvider(provider Provider) ContractOption {
	return func(o *Opts) {
		o.Provider = provider
	}
}

func WithSender(sender web3tool.Key) ContractOption {
	return func(o *Opts) {
		o.Sender = sender
	}
}

func WithEIP1559() ContractOption {
	return func(o *Opts) {
		o.EIP1559 = true
	}
}

func DeployContract(abi *abi.ABI, bin []byte, args []interface{}, opts ...ContractOption) (Txn, error) {
	a := NewContract(web3tool.Address{}, abi, opts...)
	a.bin = bin
	return a.Txn("constructor", args...)
}

func NewContract(addr web3tool.Address, abi *abi.ABI, opts ...ContractOption) *Contract {
	opt := &Opts{
		JsonRPCEndpoint: "http://localhost:8545",
	}
	for _, c := range opts {
		c(opt)
	}

	var provider Provider
	if opt.Provider != nil {
		provider = opt.Provider
	} else if opt.JsonRPCClient != nil {
		provider = &jsonRPCNodeProvider{client: opt.JsonRPCClient, eip1559: opt.EIP1559}
	} else {
		client, _ := jsonrpc.NewClient(opt.JsonRPCEndpoint)
		provider = &jsonRPCNodeProvider{client: client.Eth(), eip1559: opt.EIP1559}
	}

	a := &Contract{
		addr:     addr,
		abi:      abi,
		provider: provider,
		key:      opt.Sender,
	}

	return a
}

// Contract is a wrapper to make abi calls to contract with a state provider
type Contract struct {
	addr     web3tool.Address
	abi      *abi.ABI
	bin      []byte
	provider Provider
	key      web3tool.Key
}

func (a *Contract) GetABI() *abi.ABI {
	return a.abi
}

type TxnOpts struct {
	Value    *big.Int
	GasPrice uint64
	GasLimit uint64
	Nonce    uint64
}

func (a *Contract) Txn(method string, args ...interface{}) (Txn, error) {
	if a.key == nil {
		return nil, fmt.Errorf("no key selected")
	}

	isContractDeployment := method == "constructor"

	var input []byte
	if isContractDeployment {
		input = append(input, a.bin...)
	}

	var abiMethod *abi.Method
	if isContractDeployment {
		if a.abi.Constructor != nil {
			abiMethod = a.abi.Constructor
		}
	} else {
		if abiMethod = a.abi.GetMethod(method); abiMethod == nil {
			return nil, fmt.Errorf("method %s not found", method)
		}
	}
	if abiMethod != nil {
		data, err := abi.Encode(args, abiMethod.Inputs)
		if err != nil {
			return nil, fmt.Errorf("failed to encode arguments: %v", err)
		}
		if isContractDeployment {
			input = append(input, data...)
		} else {
			input = append(abiMethod.ID(), data...)
		}
	}

	txn, err := a.provider.Txn(a.addr, a.key, input)
	if err != nil {
		return nil, err
	}
	return txn, nil
}

type CallOpts struct {
	Block web3tool.BlockNumber
	From  web3tool.Address
}

func (a *Contract) Call(method string, block web3tool.BlockNumber, args ...interface{}) (map[string]interface{}, error) {
	m := a.abi.GetMethod(method)
	if m == nil {
		return nil, fmt.Errorf("method %s not found", method)
	}

	data, err := m.Encode(args)
	if err != nil {
		return nil, err
	}

	opts := &CallOpts{
		Block: block,
	}
	if a.key != nil {
		opts.From = a.key.Address()
	}
	rawOutput, err := a.provider.Call(a.addr, data, opts)
	if err != nil {
		return nil, err
	}

	resp, err := m.Decode(rawOutput)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
