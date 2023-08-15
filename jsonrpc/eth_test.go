package jsonrpc

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yellomoon/web3tool"
	"github.com/yellomoon/web3tool/abi"
	"github.com/yellomoon/web3tool/testutil"
)

var (
	addr0 = web3tool.Address{0x1}
	addr1 = web3tool.Address{0x2}
)

func TestEthAccounts(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		_, err := c.Eth().Accounts()
		assert.NoError(t, err)
	})
}

func TestEthBlockNumber(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		num, err := c.Eth().BlockNumber()
		require.NoError(t, err)

		for i := 0; i < 10; i++ {
			require.NoError(t, s.ProcessBlock())

			// since it is concurrent, we cannot ensure sequential numbers
			newNum, err := c.Eth().BlockNumber()
			require.NoError(t, err)
			require.Greater(t, newNum, num)

			num = newNum
		}
	})
}

func TestEthGetCode(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}
	cc.AddEvent(testutil.NewEvent("A").
		Add("address", true).
		Add("address", true))

	cc.EmitEvent("setA1", "A", addr0.String(), addr1.String())
	cc.EmitEvent("setA2", "A", addr1.String(), addr0.String())

	_, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	code, err := c.Eth().GetCode(addr, web3tool.Latest)
	assert.NoError(t, err)
	assert.NotEqual(t, code, "0x")

	code2, err := c.Eth().GetCode(addr, web3tool.BlockNumber(0))
	assert.NoError(t, err)
	assert.Equal(t, code2, "0x")
}

func TestEthGetBalance(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	balance, err := c.Eth().GetBalance(s.Account(0), web3tool.Latest)
	assert.NoError(t, err)
	assert.NotEqual(t, balance, big.NewInt(0))

	balance, err = c.Eth().GetBalance(web3tool.Address{}, web3tool.Latest)
	assert.NoError(t, err)
	assert.Equal(t, balance, big.NewInt(0))
}

func TestEthGetBlockByNumber(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	block, err := c.Eth().GetBlockByNumber(0, true)
	assert.NoError(t, err)
	assert.Equal(t, block.Number, uint64(0))

	// query a non-sealed block block 1 has not been processed yet
	// it does not fail but returns nil
	latest, err := c.Eth().BlockNumber()
	require.NoError(t, err)

	block, err = c.Eth().GetBlockByNumber(web3tool.BlockNumber(latest+10000), true)
	assert.NoError(t, err)
	assert.Nil(t, block)
}

func TestEthGetBlockByHash(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		// get block 0 first by number
		block, err := c.Eth().GetBlockByNumber(0, true)
		assert.NoError(t, err)
		assert.Equal(t, block.Number, uint64(0))

		// get block 0 by hash
		block2, err := c.Eth().GetBlockByHash(block.Hash, true)
		assert.NoError(t, err)
		assert.Equal(t, block, block2)
	})
}

func TestEthGasPrice(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		_, err := c.Eth().GasPrice()
		assert.NoError(t, err)
	})
}

func TestEthSendTransaction(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	txn := &web3tool.Transaction{
		From:     s.Account(0),
		GasPrice: testutil.DefaultGasPrice,
		Gas:      testutil.DefaultGasLimit,
		To:       &testutil.DummyAddr,
		Value:    big.NewInt(10),
	}
	hash, err := c.Eth().SendTransaction(txn)
	assert.NoError(t, err)

	var receipt *web3tool.Receipt
	for {
		receipt, err = c.Eth().GetTransactionReceipt(hash)
		if err != nil {
			t.Fatal(err)
		}
		if receipt != nil {
			break
		}
	}
}

func TestEthEstimateGas(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}
	cc.AddEvent(testutil.NewEvent("A").Add("address", true))
	cc.EmitEvent("setA", "A", addr0.String())

	// estimate gas to deploy the contract
	solcContract, err := cc.Compile()
	assert.NoError(t, err)

	input, err := hex.DecodeString(solcContract.Bin)
	assert.NoError(t, err)

	gas, err := c.Eth().EstimateGasContract(input)
	assert.NoError(t, err)
	assert.Greater(t, gas, uint64(140000))

	_, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	msg := &web3tool.CallMsg{
		From: s.Account(0),
		To:   &addr,
		Data: testutil.MethodSig("setA"),
	}

	gas, err = c.Eth().EstimateGas(msg)
	assert.NoError(t, err)
	assert.NotEqual(t, gas, 0)
}

func TestEthGetLogs(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}
	cc.AddEvent(testutil.NewEvent("A").
		Add("address", true).
		Add("address", true))

	cc.EmitEvent("setA1", "A", addr0.String(), addr1.String())
	cc.EmitEvent("setA2", "A", addr1.String(), addr0.String())

	_, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	r, err := s.TxnTo(addr, "setA2")
	require.NoError(t, err)

	filter := &web3tool.LogFilter{
		BlockHash: &r.BlockHash,
	}
	logs, err := c.Eth().GetLogs(filter)
	assert.NoError(t, err)
	assert.Len(t, logs, 1)

	log := logs[0]
	assert.Len(t, log.Topics, 3)
	assert.Equal(t, log.Address, addr)

	// first topic is the signature of the event
	assert.Equal(t, log.Topics[0].String(), cc.GetEvent("A").Sig())

	// topics have 32 bytes and the addr are 20 bytes, then, assert.Equal wont work.
	// this is a workaround until we build some helper function to test this better
	assert.True(t, bytes.HasSuffix(log.Topics[1][:], addr1[:]))
	assert.True(t, bytes.HasSuffix(log.Topics[2][:], addr0[:]))
}

func TestEthChainID(t *testing.T) {
	testutil.MultiAddr(t, func(s *testutil.TestServer, addr string) {
		c, _ := NewClient(addr)
		defer c.Close()

		num, err := c.Eth().ChainID()
		assert.NoError(t, err)
		assert.Equal(t, num.Uint64(), uint64(1337)) // chainid of geth-dev
	})
}

func TestEthCall(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())
	cc := &testutil.Contract{}

	// add global variables
	cc.AddCallback(func() string {
		return "uint256 val = 1;"
	})

	// add setter method
	cc.AddCallback(func() string {
		return `function getValue() public returns (uint256) {
			return val;
		}`
	})

	_, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	input := abi.MustNewMethod("function getValue() public returns (uint256)").ID()

	resp, err := c.Eth().Call(&web3tool.CallMsg{To: &addr, Data: input}, web3tool.Latest)
	require.NoError(t, err)

	require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000001", resp)

	nonce := uint64(1)

	// override the state
	override := &web3tool.StateOverride{
		addr: web3tool.OverrideAccount{
			Nonce:   &nonce,
			Balance: big.NewInt(1),
			StateDiff: &map[web3tool.Hash]web3tool.Hash{
				// storage slot 0 stores the 'val' uint256 value
				{0x0}: {0x3},
			},
		},
	}

	resp, err = c.Eth().Call(&web3tool.CallMsg{To: &addr, Data: input}, web3tool.Latest, override)
	require.NoError(t, err)

	require.Equal(t, "0x0300000000000000000000000000000000000000000000000000000000000000", resp)
}

func TestEthGetNonce(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	receipt, err := s.ProcessBlockWithReceipt()
	assert.NoError(t, err)

	// query the balance with different options
	cases := []web3tool.BlockNumberOrHash{
		web3tool.Latest,
		receipt.BlockHash,
		web3tool.BlockNumber(receipt.BlockNumber),
	}
	for _, ca := range cases {
		num, err := c.Eth().GetNonce(s.Account(0), ca)
		assert.NoError(t, err)
		assert.NotEqual(t, num, uint64(0))
	}
}

func TestEthTransactionsInBlock(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	// block 0 does not have transactions
	_, err := c.Eth().GetBlockByNumber(0, false)
	assert.NoError(t, err)

	// Process a block with a transaction
	assert.NoError(t, s.ProcessBlock())

	latest, err := c.Eth().BlockNumber()
	require.NoError(t, err)

	num := web3tool.BlockNumber(latest)

	// get a non-full block
	block0, err := c.Eth().GetBlockByNumber(num, false)
	assert.NoError(t, err)

	assert.NotEmpty(t, block0.TransactionsHashes, 1)
	assert.Empty(t, block0.Transactions, 0)

	// get a full block
	block1, err := c.Eth().GetBlockByNumber(num, true)
	assert.NoError(t, err)

	assert.Empty(t, block1.TransactionsHashes, 0)
	assert.NotEmpty(t, block1.Transactions, 1)

	for indx := range block0.TransactionsHashes {
		assert.Equal(t, block0.TransactionsHashes[indx], block1.Transactions[indx].Hash)
	}
}

func TestEthGetStorageAt(t *testing.T) {
	s := testutil.NewTestServer(t)

	c, _ := NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}

	// add global variables
	cc.AddCallback(func() string {
		return "uint256 val;"
	})

	// add setter method
	cc.AddCallback(func() string {
		return `function setValue() public payable {
			val = 10;
		}`
	})

	_, addr, err := s.DeployContract(cc)
	require.NoError(t, err)

	receipt, err := s.TxnTo(addr, "setValue")
	require.NoError(t, err)

	cases := []web3tool.BlockNumberOrHash{
		web3tool.Latest,
		receipt.BlockHash,
		web3tool.BlockNumber(receipt.BlockNumber),
	}
	for _, ca := range cases {
		res, err := c.Eth().GetStorageAt(addr, web3tool.Hash{}, ca)
		assert.NoError(t, err)
		assert.True(t, strings.HasSuffix(res.String(), "a"))
	}
}

func TestEthFeeHistory(t *testing.T) {
	c, _ := NewClient(testutil.TestInfuraEndpoint(t))

	lastBlock, err := c.Eth().BlockNumber()
	assert.NoError(t, err)

	fee, err := c.Eth().FeeHistory(1, web3tool.BlockNumber(lastBlock), []float64{25, 75})
	assert.NoError(t, err)
	assert.NotNil(t, fee)
}

func TestEthMaxPriorityFeePerGas(t *testing.T) {
	s := testutil.NewTestServer(t)
	c, err := NewClient(s.HTTPAddr())
	require.NoError(t, err)

	initialMaxPriorityFee, err := c.Eth().MaxPriorityFeePerGas()
	require.NoError(t, err)

	// wait for 2 blocks
	require.NoError(t, s.ProcessBlock())
	require.NoError(t, s.ProcessBlock())

	txn := &web3tool.Transaction{
		To:                   &testutil.DummyAddr,
		Value:                web3tool.Gwei(1),
		Type:                 web3tool.TransactionDynamicFee,
		MaxPriorityFeePerGas: web3tool.Gwei(1),
	}

	latestBlock, err := c.Eth().BlockNumber()
	require.NoError(t, err)

	feeHistory, err := c.Eth().FeeHistory(1, web3tool.BlockNumber(latestBlock), nil)
	require.NoError(t, err)

	latestBaseFee := feeHistory.BaseFee[len(feeHistory.BaseFee)-1]
	txn.MaxFeePerGas = new(big.Int).Add(latestBaseFee, txn.MaxPriorityFeePerGas)

	receipt, err := s.SendTxn(txn)
	require.NoError(t, err)
	require.Equal(t, uint64(1), receipt.Status)

	newMaxPriorityFee, err := c.Eth().MaxPriorityFeePerGas()
	t.Log(initialMaxPriorityFee)
	t.Log(newMaxPriorityFee)
	require.NoError(t, err)
	require.True(t, initialMaxPriorityFee.Cmp(newMaxPriorityFee) <= 0)
}
