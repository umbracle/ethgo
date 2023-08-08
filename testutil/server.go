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
	"strconv"
	"testing"
	"time"

	"github.com/ory/dockertest"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/compiler"
	"golang.org/x/crypto/sha3"
)

var (
	DefaultGasPrice = uint64(1879048192) // 0x70000000
	DefaultGasLimit = uint64(5242880)    // 0x500000
)

var (
	DummyAddr = ethgo.HexToAddress("0x015f68893a39b3ba0681584387670ff8b00f4db2")
)

func getOpenPort() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	min, max := 12000, 15000
	for {
		port := strconv.Itoa(r.Intn(max-min) + min)
		server, err := net.Listen("tcp", ":"+port)
		if err == nil {
			server.Close()
			return port
		}
	}
}

// MultiAddr creates new servers to test different addresses
func MultiAddr(t *testing.T, c func(s *TestServer, addr string)) {
	s := NewTestServer(t)

	// http addr
	c(s, s.HTTPAddr())

	// ws addr
	// c(s, s.WSAddr())

	// ip addr
	// c(s, s.IPCPath())

	// s.Close()
}

// TestServerConfig is the configuration of the server
type TestServerConfig struct {
	Period int
}

// ServerConfigCallback is the callback to modify the config
type ServerConfigCallback func(c *TestServerConfig)

// TestServer is a Geth test server
type TestServer struct {
	addr     string
	accounts []ethgo.Address
	client   *ethClient
}

// DeployTestServer creates a new Geth test server
func DeployTestServer(t *testing.T, cb ServerConfigCallback) *TestServer {
	tmpDir, err := ioutil.TempDir("/tmp", "geth-")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	config := &TestServerConfig{}
	if cb != nil {
		cb(config)
	}

	args := []string{"--dev"}

	// periodic mining
	if config.Period != 0 {
		args = append(args, "--dev.period", strconv.Itoa(config.Period))
	}

	// add data dir
	args = append(args, "--datadir", "/eth1data")

	// add ipcpath
	args = append(args, "--ipcpath", "/eth1data/geth.ipc")

	// enable rpc
	args = append(args, "--http", "--http.addr", "0.0.0.0", "--http.api", "eth,net,web3,debug")

	// enable ws
	args = append(args, "--ws", "--ws.addr", "0.0.0.0")

	// enable debug verbosity
	args = append(args, "--verbosity", "4")

	opts := &dockertest.RunOptions{
		Repository: "ethereum/client-go",
		Tag:        "v1.10.15",
		Cmd:        args,
		Mounts: []string{
			tmpDir + ":/eth1data",
		},
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("Could not connect to docker: %s", err)
	}
	resource, err := pool.RunWithOptions(opts)
	if err != nil {
		t.Fatalf("Could not start go-ethereum: %s", err)
	}

	closeFn := func() {
		if err := pool.Purge(resource); err != nil {
			t.Fatalf("Could not purge geth: %s", err)
		}
	}

	ipAddr := resource.Container.NetworkSettings.IPAddress
	addr := fmt.Sprintf("http://%s:8545", ipAddr)

	if err := pool.Retry(func() error {
		return testHTTPEndpoint(addr)
	}); err != nil {
		closeFn()
	}

	t.Cleanup(func() {
		closeFn()
	})

	return NewTestServer(t, addr)
}

func NewTestServer(t *testing.T, addrs ...string) *TestServer {
	var addr string
	if len(addrs) != 0 {
		addr = addrs[0]
	} else {
		// default address
		addr = "http://127.0.0.1:8545"
	}

	server := &TestServer{
		addr: addr,
	}

	server.client = &ethClient{addr}
	if err := server.client.call("eth_accounts", &server.accounts); err != nil {
		t.Fatal(err)
	}
	return server
}

// Account returns a specific account
func (t *TestServer) Account(i int) ethgo.Address {
	return t.accounts[i]
}

// IPCPath returns the ipc endpoint
func (t *TestServer) IPCPath() string {
	return ""
	// return t.tmpDir + "/geth.ipc"
}

// WSAddr returns the websocket endpoint
func (t *TestServer) WSAddr() string {
	return "ws://localhost:8546"
}

// HTTPAddr returns the http endpoint
func (t *TestServer) HTTPAddr() string {
	return fmt.Sprintf(t.addr)
}

// ProcessBlock processes a new block
func (t *TestServer) ProcessBlockWithReceipt() (*ethgo.Receipt, error) {
	receipt, err := t.SendTxn(&ethgo.Transaction{
		From:  t.accounts[0],
		To:    &DummyAddr,
		Value: big.NewInt(10),
	})
	return receipt, err
}

func (t *TestServer) ProcessBlock() error {
	_, err := t.ProcessBlockWithReceipt()
	return err
}

var emptyAddr ethgo.Address

func isEmptyAddr(w ethgo.Address) bool {
	return bytes.Equal(w[:], emptyAddr[:])
}

// Call sends a contract call
func (t *TestServer) Call(msg *ethgo.CallMsg) (string, error) {
	if isEmptyAddr(msg.From) {
		msg.From = t.Account(0)
	}
	var resp string
	if err := t.client.call("eth_call", &resp, msg, "latest"); err != nil {
		return "", err
	}
	return resp, nil
}

func (t *TestServer) Fund(address ethgo.Address) (*ethgo.Receipt, error) {
	return t.Transfer(address, big.NewInt(1000000000000000000))
}

func (t *TestServer) Transfer(address ethgo.Address, value *big.Int) (*ethgo.Receipt, error) {
	return t.SendTxn(&ethgo.Transaction{
		From:  t.accounts[0],
		To:    &address,
		Value: value,
	})
}

// TxnTo sends a transaction to a given method without any arguments
func (t *TestServer) TxnTo(address ethgo.Address, method string) (*ethgo.Receipt, error) {
	return t.SendTxn(&ethgo.Transaction{
		To:    &address,
		Input: MethodSig(method),
	})
}

// SendTxn sends a transaction
func (t *TestServer) SendTxn(txn *ethgo.Transaction) (*ethgo.Receipt, error) {
	if isEmptyAddr(txn.From) {
		txn.From = t.Account(0)
	}
	if txn.GasPrice == 0 {
		txn.GasPrice = DefaultGasPrice
	}
	if txn.Gas == 0 {
		txn.Gas = DefaultGasLimit
	}

	var hash ethgo.Hash
	if err := t.client.call("eth_sendTransaction", &hash, txn); err != nil {
		return nil, err
	}

	return t.WaitForReceipt(hash)
}

// WaitForReceipt waits for the receipt
func (t *TestServer) WaitForReceipt(hash ethgo.Hash) (*ethgo.Receipt, error) {
	var receipt *ethgo.Receipt
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
		if count > 300 {
			return nil, fmt.Errorf("timeout waiting for receipt")
		}
		time.Sleep(500 * time.Millisecond)
		count++
	}
	return receipt, nil
}

// DeployContract deploys a contract with account 0 and returns the address
func (t *TestServer) DeployContract(c *Contract) (*compiler.Artifact, ethgo.Address, error) {
	// solcContract := compile(c.Print())
	solcContract, err := c.Compile()
	if err != nil {
		return nil, ethgo.Address{}, err
	}
	buf, err := hex.DecodeString(solcContract.Bin)
	if err != nil {
		return nil, ethgo.Address{}, err
	}

	receipt, err := t.SendTxn(&ethgo.Transaction{
		Input: buf,
	})
	if err != nil {
		return nil, ethgo.Address{}, err
	}
	return solcContract, receipt.ContractAddress, nil
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

// MethodSig returns the signature of a non-parametrized function
func MethodSig(name string) []byte {
	h := sha3.NewLegacyKeccak256()
	h.Write([]byte(name + "()"))
	b := h.Sum(nil)
	return b[:4]
}

// TestInfuraEndpoint returns the testing infura endpoint to make testing requests
func TestInfuraEndpoint(t *testing.T) string {
	url := os.Getenv("INFURA_URL")
	if url == "" {
		t.Skip("Infura url not set")
	}
	return url
}

func testHTTPEndpoint(endpoint string) error {
	resp, err := http.Post(endpoint, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
