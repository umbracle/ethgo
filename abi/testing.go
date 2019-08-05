package abi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/umbracle/go-web3/compiler"
	"github.com/umbracle/minimal/types"
)

const (
	defaultGasPrice = "0x70000000"
	defaultGasLimit = "0x500000"
)

type jsonRPCRequest struct {
	ID     int             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type jsonRPCResponse struct {
	ID     int                 `json:"id"`
	Result json.RawMessage     `json:"result"`
	Error  *jsonRPCErrorObject `json:"error,omitempty"`
}

type jsonRPCErrorObject struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ethClient struct {
	url string
}

var errNotFound = fmt.Errorf("not found")

func (e *ethClient) call(method string, out interface{}, params ...interface{}) error {
	if e.url == "" {
		e.url = "http://127.0.0.1:8545"
	}

	var err error
	jsonReq := &jsonRPCRequest{
		Method: method,
	}
	if len(params) > 0 {
		jsonReq.Params, err = json.Marshal(params)
		if err != nil {
			return err
		}
	}
	raw, err := json.Marshal(jsonReq)
	if err != nil {
		return err
	}

	resp, err := http.Post(e.url, "application/json", bytes.NewBuffer(raw))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var jsonResp jsonRPCResponse
	d := json.NewDecoder(resp.Body)
	if err := d.Decode(&jsonResp); err != nil {
		return err
	}

	if jsonResp.Error != nil {
		return fmt.Errorf(jsonResp.Error.Message)
	}
	if bytes.Equal(jsonResp.Result, []byte("null")) {
		return errNotFound
	}
	if err := json.Unmarshal(jsonResp.Result, out); err != nil {
		return err
	}
	return nil
}

func (e *ethClient) accounts() ([]string, error) {
	var resp []string
	err := e.call("eth_accounts", &resp)
	return resp, err
}

func compileAndDeployContract(source string, deployer string, client *ethClient) (*ABI, map[string]interface{}, error) {
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

	txn := map[string]string{
		"from":     deployer,
		"data":     "0x" + string(contract.Bin),
		"gasPrice": defaultGasPrice,
		"gas":      defaultGasLimit,
	}

	var txnHash types.Hash
	if err := client.call("eth_sendTransaction", &txnHash, txn); err != nil {
		return nil, nil, err
	}

	c := 0
	for {
		var receipt interface{}
		err := client.call("eth_getTransactionReceipt", &receipt, txnHash.String())
		if err != nil {
			if err != errNotFound {
				return nil, nil, err
			}
		}
		if receipt != nil {
			return abi, receipt.(map[string]interface{}), nil
		}

		if c > 5 {
			return nil, nil, fmt.Errorf("timeout to get the receipt")
		}
		time.Sleep(500 * time.Millisecond)
		c++
	}
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

func randomType() string {
	return pickRandomType(1)
}

func pickRandomType(d int) string {
PICK:
	t := randomTypes[rand.Intn(len(randomTypes))]

	basicTypes := "bool,address,string,bytes,function"
	if strings.Contains(basicTypes, t) {
		return t
	}

	switch t {
	case "int":
		return fmt.Sprintf("int%d", randomNumberBits())

	case "uint":
		return fmt.Sprintf("uint%d", randomNumberBits())

	case "fixedBytes":
		return fmt.Sprintf("bytes%d", randomInt(1, 32))
	}

	if d > 3 {
		// Allow only for 3 levels of depth
		goto PICK
	}

	r := pickRandomType(d + 1)
	switch t {
	case "slice":
		return fmt.Sprintf("%s[]", r)

	case "array":
		s := randomInt(1, 3)
		return fmt.Sprintf("%s[%d]", r, s)

	case "tuple":
		size := randomInt(1, 5)
		elems := []string{}
		for i := 0; i < size; i++ {
			elem := pickRandomType(d + 1)
			elems = append(elems, fmt.Sprintf("arg%d %s", i, elem))
		}
		return fmt.Sprintf("tuple(%s)", strings.Join(elems, ","))

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
		return types.StringToAddress(randString(randomInt(1, 32), hexLetters))

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
