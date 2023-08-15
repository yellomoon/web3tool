package etherscan

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/yellomoon/web3tool"
	"github.com/yellomoon/web3tool/jsonrpc/codec"
	"github.com/valyala/fasthttp"
)

// Etherscan is a provider using the Etherscan api
type Etherscan struct {
	client fasthttp.Client
	url    string
	apiKey string
}

// NewEtherscanFromNetwork creates a new client from the network id
func NewEtherscanFromNetwork(n web3tool.Network, apiKey string) (*Etherscan, error) {
	var url string
	switch n {
	case web3tool.Mainnet:
		url = "https://api.etherscan.io"

	case web3tool.Ropsten:
		url = "https://ropsten.etherscan.io"

	case web3tool.Rinkeby:
		url = "https://rinkeby.etherscan.io"

	case web3tool.Goerli:
		url = "https://goerli.etherscan.io"

	default:
		return nil, fmt.Errorf("unknwon network id %d", n)
	}
	return NewEtherscan(url, apiKey), nil
}

// NewEtherscan creates a new Etherscan service from a url
func NewEtherscan(url, apiKey string) *Etherscan {
	return &Etherscan{url: url}
}

type proxyResponse struct {
	Status  string
	Message string
	Result  json.RawMessage
}

// Query sends a query to Etherscan
func (e *Etherscan) Query(module, action string, out interface{}, params map[string]string) error {
	url := fmt.Sprintf("%s/api?module=%s&action=%s", e.url, module, action)
	if len(params) != 0 {
		res := []string{}
		for k, v := range params {
			res = append(res, k+"="+v)
		}
		url += "&" + strings.Join(res, "&")
	}
	if e.apiKey != "" {
		url = url + "&apikey=" + e.apiKey
	}

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	req.SetRequestURI(url)
	req.Header.SetMethod("GET")

	if err := e.client.Do(req, res); err != nil {
		return err
	}

	// Decode json-rpc response
	var result json.RawMessage

	if module == "proxy" {
		var response codec.Response
		if err := json.Unmarshal(res.Body(), &response); err != nil {
			return err
		}
		if response.Error != nil {
			return response.Error
		}
		result = response.Result
	} else {
		var response proxyResponse
		if err := json.Unmarshal(res.Body(), &response); err != nil {
			return err
		}
		result = response.Result
	}

	if err := json.Unmarshal(result, out); err != nil {
		return err
	}
	return nil
}

// BlockNumber returns the number of most recent block.
func (e *Etherscan) BlockNumber() (uint64, error) {
	var out string
	if err := e.Query("proxy", "eth_blockNumber", &out, nil); err != nil {
		return 0, err
	}
	return parseUint64orHex(out)
}

// GetBlockByNumber returns information about a block by block number.
func (e *Etherscan) GetBlockByNumber(i web3tool.BlockNumber, full bool) (*web3tool.Block, error) {
	var b *web3tool.Block
	params := map[string]string{
		"tag":     i.String(),
		"boolean": strconv.FormatBool(full),
	}
	if err := e.Query("proxy", "eth_getBlockByNumber", &b, params); err != nil {
		return nil, err
	}
	return b, nil
}

type ContractCode struct {
	SourceCode           string
	ContractName         string
	Runs                 string
	CompilerVersion      string
	ConstructorArguments string
}

func (e *Etherscan) GetContractCode(addr web3tool.Address) (*ContractCode, error) {
	var out []*ContractCode
	err := e.Query("contract", "getsourcecode", &out, map[string]string{
		"address": addr.String(),
	})
	if err != nil {
		return nil, err
	}
	if len(out) != 1 {
		return nil, fmt.Errorf("incorrect values")
	}
	return out[0], nil
}

func (e *Etherscan) GasPrice() (uint64, error) {
	var out struct {
		LastBlock string `json:"LastBlock"`
	}
	if err := e.Query("gastracker", "gasoracle", &out, map[string]string{}); err != nil {
		return 0, err
	}
	num, err := strconv.Atoi(out.LastBlock)
	if err != nil {
		return 0, err
	}
	return uint64(num), nil
}

func (e *Etherscan) GetLogs(filter *web3tool.LogFilter) ([]*web3tool.Log, error) {
	if len(filter.Address) == 0 {
		return nil, fmt.Errorf("an address to filter is required")
	}
	strBlockNumber := func(b web3tool.BlockNumber) string {
		switch b {
		case web3tool.Latest:
			return "latest"
		case web3tool.Earliest:
			return "earliest"
		case web3tool.Pending:
			return "pending"
		}
		if b < 0 {
			panic("internal. blocknumber is negative")
		}
		return fmt.Sprintf("%d", uint64(b))
	}

	params := map[string]string{
		"address": filter.Address[0].String(),
	}
	if filter.From != nil {
		params["fromBlock"] = strBlockNumber(*filter.From)
	}
	if filter.To != nil {
		params["toBlock"] = strBlockNumber(*filter.To)
	}
	var out []*web3tool.Log
	if err := e.Query("logs", "getLogs", &out, params); err != nil {
		return nil, err
	}
	return out, nil
}

func parseUint64orHex(str string) (uint64, error) {
	base := 10
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
		base = 16
	}
	return strconv.ParseUint(str, base, 64)
}
