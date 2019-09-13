package testutil

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/compiler"
	"golang.org/x/crypto/sha3"
)

const (
	DefaultGasPrice = 1879048192 // 0x70000000
	DefaultGasLimit = 5242880    // 0x500000
)

const (
	dummyAddr = "0x015f68893a39b3ba0681584387670ff8b00f4db2"
)

func getOpenPort() string {
	rand.Seed(time.Now().UnixNano())
	min, max := 12000, 15000
	for {
		port := strconv.Itoa(rand.Intn(max-min) + min)
		server, err := net.Listen("tcp", ":"+port)
		if err == nil {
			server.Close()
			return port
		}
	}
}

// TestServerConfig is the configuration of the server
type TestServerConfig struct {
	DataDir  string
	HTTPPort string
	WSPort   string
}

// ServerConfigCallback is the callback to modify the config
type ServerConfigCallback func(c *TestServerConfig)

// TestServer is a Geth test server
type TestServer struct {
	cmd      *exec.Cmd
	config   *TestServerConfig
	accounts []string
	client   *ethClient
	t        *testing.T
}

// NewTestServer creates a new Geth test server
func NewTestServer(t *testing.T, cb ServerConfigCallback) *TestServer {
	path := "geth"

	vcmd := exec.Command(path, "version")
	vcmd.Stdout = nil
	vcmd.Stderr = nil
	if err := vcmd.Run(); err != nil {
		t.Skipf("geth version failed: %v", err)
	}

	dir, err := ioutil.TempDir("/tmp", "geth-")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	config := &TestServerConfig{
		DataDir:  dir,
		HTTPPort: getOpenPort(),
		WSPort:   getOpenPort(),
	}
	if cb != nil {
		cb(config)
	}

	// Build arguments
	args := []string{"--dev"}

	// add data dir
	args = append(args, "--datadir", filepath.Join(dir, "data"))

	// add ipcpath
	args = append(args, "--ipcpath", filepath.Join(dir, "geth.ipc"))

	// enable rpc
	args = append(args, "--rpc", "--rpcport", config.HTTPPort)

	// enable ws
	args = append(args, "--ws", "--wsport", config.WSPort)

	// Start the server
	cmd := exec.Command(path, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		t.Fatalf("err: %s", err)
	}

	server := &TestServer{
		t:      t,
		cmd:    cmd,
		config: config,
	}

	// wait till the jsonrpc endpoint is running
	for {
		if server.testHTTPEndpoint() {
			break
		}
	}

	server.client = &ethClient{server.HTTPAddr()}
	if err := server.client.call("eth_accounts", &server.accounts); err != nil {
		t.Fatal(err)
	}

	return server
}

// Account returns a specific account
func (t *TestServer) Account(i int) string {
	return t.accounts[i]
}

// IPCPath returns the ipc endpoint
func (t *TestServer) IPCPath() string {
	return filepath.Join(filepath.Join(t.config.DataDir, "geth.ipc"))
}

// WSAddr returns the websocket endpoint
func (t *TestServer) WSAddr() string {
	return "ws://localhost:" + t.config.WSPort
}

// HTTPAddr returns the http endpoint
func (t *TestServer) HTTPAddr() string {
	return "http://localhost:" + t.config.HTTPPort
}

// ProcessBlock processes a new block
func (t *TestServer) ProcessBlock() error {
	_, err := t.SendTxn(&web3.Transaction{
		From:  t.accounts[0],
		To:    dummyAddr,
		Value: big.NewInt(10),
	})
	return err
}

// Call sends a contract call
func (t *TestServer) Call(msg *web3.CallMsg) (string, error) {
	if msg.From == "" {
		msg.From = t.Account(0)
	}
	var resp string
	if err := t.client.call("eth_call", &resp, msg, "latest"); err != nil {
		return "", err
	}
	return resp, nil
}

// TxnTo sends a transaction to a given method without any arguments
func (t *TestServer) TxnTo(address string, method string) *web3.Receipt {
	sig := MethodSig(method)
	receipt, err := t.SendTxn(&web3.Transaction{
		To:    address,
		Input: sig,
	})
	if err != nil {
		t.t.Fatal(err)
	}
	return receipt
}

// SendTxn sends a transaction
func (t *TestServer) SendTxn(txn *web3.Transaction) (*web3.Receipt, error) {
	if txn.From == "" {
		txn.From = t.Account(0)
	}
	if txn.GasPrice == 0 {
		txn.GasPrice = DefaultGasPrice
	}
	if txn.Gas == 0 {
		txn.Gas = DefaultGasLimit
	}

	var hash string
	if err := t.client.call("eth_sendTransaction", &hash, txn); err != nil {
		return nil, err
	}

	var receipt *web3.Receipt
	var count uint64
	for {
		err := t.client.call("eth_getTransactionReceipt", &receipt, hash)
		if err != nil {
			if err.Error() != "not found" {
				return nil, err
			}
		}
		if receipt != nil {
			break
		}
		if count > 100 {
			return nil, fmt.Errorf("timeout")
		}
		time.Sleep(50 * time.Millisecond)
		count++
	}
	return receipt, nil
}

// DeployContract deploys a contract with account 0 and returns the address
func (t *TestServer) DeployContract(c *Contract) (*compiler.SolcContract, string) {
	solcContract := compile(c.Print())

	receipt, err := t.SendTxn(&web3.Transaction{
		Data: "0x" + solcContract.Bin,
	})
	if err != nil {
		panic(err)
	}
	return solcContract, receipt.ContractAddress
}

func (t *TestServer) testHTTPEndpoint() bool {
	resp, err := http.Post(t.HTTPAddr(), "application/json", nil)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return true
}

func (t *TestServer) exit(err error) {
	t.Close()
	t.t.Fatal(err)
}

// Close closes the server
func (t *TestServer) Close() {
	defer os.RemoveAll(t.config.DataDir)

	if err := t.cmd.Process.Kill(); err != nil {
		t.t.Errorf("err: %s", err)
	}
	t.cmd.Wait()
}

// Simple jsonrpc client to avoid cycle dependencies

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

func compile(source string) *compiler.SolcContract {
	rawData, err := compiler.NewSolidityCompiler("solc").Compile(source)
	if err != nil {
		panic(err)
	}

	output := rawData.(*compiler.SolcOutput)
	solcContract, ok := output.Contracts["<stdin>:Sample"]
	if !ok {
		panic(fmt.Errorf("Expected the contract to be called Sample"))
	}
	return solcContract
}

// MethodSig returns the signature of a non-parametrized function
func MethodSig(name string) string {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(name + "()"))
	b := h.Sum(nil)
	return "0x" + hex.EncodeToString(b[:4])
}
