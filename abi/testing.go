package abi

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"reflect"
	"strings"
	"time"

	"github.com/umbracle/go-web3/compiler"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
)

const (
	defaultGasPrice = "0x70000000"
	defaultGasLimit = "0x500000"
)

type ethClient struct {
	*ethclient.Client
	rpc *ethrpc.Client

	gasPrice string
	gasLimit string
}

func newClient() *ethClient {
	return newClientWithEndpoint("http://localhost:8545")
}

func newClientWithEndpoint(endpoint string) *ethClient {
	rpc, err := ethrpc.Dial(endpoint)
	if err != nil {
		panic(err)
	}

	return &ethClient{ethclient.NewClient(rpc), rpc, defaultGasPrice, defaultGasLimit}
}

func (c *ethClient) listAccounts() ([]common.Address, error) {
	var accounts []common.Address
	err := c.rpc.Call(&accounts, "eth_accounts")
	return accounts, err
}

func (c *ethClient) sendTx(tx *transaction) (*transactionResult, error) {
	tx.GasPrice = defaultGasPrice
	tx.Gas = defaultGasLimit

	var hash common.Hash
	if err := c.rpc.Call(&hash, "eth_sendTransaction", tx); err != nil {
		return nil, err
	}

	result := &transactionResult{
		Hash:   hash,
		client: c,
	}

	return result, nil
}

func (c *ethClient) SendTxAndWait(tx *transaction) (*types.Receipt, error) {
	result, err := c.sendTx(tx)
	if err != nil {
		return nil, err
	}
	return result.Wait()
}

type transaction struct {
	From     common.Address  `json:"from"`
	To       *common.Address `json:"to"`
	Data     string          `json:"data"`
	Gas      string          `json:"gas"`
	GasPrice string          `json:"gasPrice"`
}

type transactionResult struct {
	Hash   common.Hash
	client *ethClient
}

func (t *transactionResult) Wait() (*types.Receipt, error) {
	for {
		receipt, err := t.client.TransactionReceipt(context.Background(), t.Hash)
		if err != nil {
			if err != ethereum.NotFound {
				return nil, err
			}
		}
		if receipt != nil {
			return receipt, nil
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func compileAndDeployContract(source string, deployer common.Address, client *ethClient) (*ABI, *types.Receipt, error) {
	data, err := compiler.NewSolidityCompiler("solc").Compile(source)
	if err != nil {
		return nil, nil, err
	}

	output := data.(*compiler.SolcOutput)
	contract, ok := output.Contracts["<stdin>:Sample"]
	if !ok {
		return nil, nil, fmt.Errorf("Expected the contract to be called Sample")
	}

	abi, err := NewABI(string(contract.Abi))
	if err != nil {
		return nil, nil, err
	}

	tx := &transaction{
		From: deployer,
		Data: "0x" + string(contract.Bin),
	}

	rr, err := client.SendTxAndWait(tx)
	if err != nil {
		return nil, nil, err
	}
	return abi, rr, nil
}

func randomInt(min, max int) int {
	return min + rand.Intn(max-min)
}

var randomTypes = []string{
	"bool",
	"int",
	"uint",
	"array",
	"slice",
	"tuple",
	"address",
	"string",
	"bytes",
	"fixedBytes",
}

func randomNumberBits() int {
	return randomInt(1, 31) * 8
}

func randomType() *Argument {
	return pickRandomType(1)
}

func pickRandomType(d int) *Argument {
PICK:
	t := randomTypes[rand.Intn(len(randomTypes))]

	basicTypes := "bool,address,string,bytes,function"
	if strings.Contains(basicTypes, t) {
		return &Argument{Type: t}
	}

	switch t {
	case "int":
		return &Argument{Type: fmt.Sprintf("int%d", randomNumberBits())}

	case "uint":
		return &Argument{Type: fmt.Sprintf("uint%d", randomNumberBits())}

	case "fixedBytes":
		return &Argument{Type: fmt.Sprintf("bytes%d", randomInt(1, 32))}
	}

	if d > 3 {
		// Allow only for 3 levels of depth
		goto PICK
	}

	r := pickRandomType(d + 1)
	switch t {
	case "slice":
		return &Argument{Type: fmt.Sprintf("%s[]", r.Type), Components: r.Components}

	case "array":
		s := randomInt(1, 3)
		return &Argument{Type: fmt.Sprintf("%s[%d]", r.Type, s), Components: r.Components}

	case "tuple":
		size := randomInt(1, 5)
		elems := make([]*Argument, size)
		for i := 0; i < size; i++ {
			elem := pickRandomType(d + 1)
			elem.Name = fmt.Sprintf("arg%d", i)
			elems[i] = elem
		}
		return &Argument{Type: "tuple", Components: elems}

	default:
		panic(fmt.Errorf("type not implemented: %s", t))
	}
}

func generateNumber(t *Type) interface{} {
	b := make([]byte, t.size/8)
	if t.kind == KindUInt {
		rand.Read(b)
	} else {
		rand.Read(b[1:])
	}

	num := big.NewInt(1).SetBytes(b)
	if t.size == 8 || t.size == 16 || t.size == 32 || t.size == 64 {
		return reflect.ValueOf(num.Int64()).Convert(t.t).Interface()
	}
	return num
}

func generateRandomType(t *Type) interface{} {

	switch t.kind {
	case KindInt:
		fallthrough
	case KindUInt:
		return generateNumber(t)

	case KindBool:
		if randomInt(0, 1) == 1 {
			return true
		}
		return false

	case KindAddress:
		return common.HexToAddress(randString(randomInt(1, 32), hexLetters))

	case KindString:
		return randString(randomInt(1, 100), letters)

	case KindBytes:
		buf := make([]byte, randomInt(1, 100))
		rand.Read(buf)
		return buf

	case KindFixedBytes, KindFunction:
		buf := make([]byte, t.size)
		rand.Read(buf)

		val := reflect.New(t.t).Elem()
		for i := 0; i < len(buf); i++ {
			val.Index(i).Set(reflect.ValueOf(buf[i]))
		}
		return val.Interface()

	case KindSlice:
		size := randomInt(0, 5)
		val := reflect.MakeSlice(t.t, size, size)
		for i := 0; i < size; i++ {
			val.Index(i).Set(reflect.ValueOf(generateRandomType(t.elem)))
		}
		return val.Interface()

	case KindArray:
		val := reflect.New(t.t).Elem()
		for i := 0; i < t.size; i++ {
			val.Index(i).Set(reflect.ValueOf(generateRandomType(t.elem)))
		}
		return val.Interface()

	case KindTuple:
		vals := map[string]interface{}{}
		for _, i := range t.tuple {
			vals[i.Name] = generateRandomType(i.Elem)
		}
		return vals

	default:
		panic(fmt.Errorf("type not implemented: %s", t.kind.String()))
	}
}

const hexLetters = "0123456789abcdef"

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randString(n int, dict string) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = dict[rand.Intn(len(dict))]
	}
	return string(b)
}

type generateContractImpl struct {
	structs []string
}

func (g *generateContractImpl) run(t *Type) string {

	var input, output, body []string
	for indx, i := range t.tuple {
		val := g.getValue(i.Elem)
		memory := ""
		if val == "bytes" || strings.Contains(val, "[") || strings.Contains(val, "struct") || strings.Contains(val, "string") {
			memory = " memory"
		}

		input = append(input, fmt.Sprintf("%s%s arg%d", val, memory, indx))
		output = append(output, fmt.Sprintf("%s%s", val, memory))
		body = append(body, fmt.Sprintf("arg%d", indx))
	}

	contractTemplate := `pragma solidity ^0.5.5;
pragma experimental ABIEncoderV2;

contract Sample {
	// structs
	%s
	function set(%s) public view returns (%s) {
		return (%s);
	}
}`

	contract := fmt.Sprintf(
		contractTemplate,
		strings.Join(g.structs, "\n"),
		strings.Join(input, ","),
		strings.Join(output, ","),
		strings.Join(body, ","))

	return contract
}

func (g *generateContractImpl) getValue(t *Type) string {
	switch t.kind {
	case KindTuple:
		attrs := []string{}
		for indx, i := range t.tuple {
			attrs = append(attrs, fmt.Sprintf("%s attr%d;", g.getValue(i.Elem), indx))
		}
		id := len(g.structs)
		str := fmt.Sprintf("struct struct%d {\n%s\n}\n", id, strings.Join(attrs, "\n"))
		g.structs = append(g.structs, str)
		return fmt.Sprintf("struct%d", id)

	case KindSlice:
		return fmt.Sprintf("%s[]", g.getValue(t.elem))

	case KindArray:
		return fmt.Sprintf("%s[%d]", g.getValue(t.elem), t.size)

	default:
		return t.raw
	}
}
