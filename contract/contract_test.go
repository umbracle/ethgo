package contract

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/abi"
	"github.com/umbracle/ethgo/jsonrpc"
	"github.com/umbracle/ethgo/testutil"
	"github.com/umbracle/ethgo/wallet"
)

var (
	addr0  = "0x0000000000000000000000000000000000000000"
	addr0B = ethgo.HexToAddress(addr0)
)

func TestContract_NoInput(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	cc := &testutil.Contract{}
	cc.AddOutputCaller("set")

	contract, addr := s.DeployContract(cc)

	abi0, err := abi.NewABI(contract.Abi)
	assert.NoError(t, err)

	p, _ := jsonrpc.NewClient(s.HTTPAddr())
	c := NewContract(addr, abi0, WithJsonRPC(p.Eth()))

	vals, err := c.Call("set", ethgo.Latest)
	assert.NoError(t, err)
	assert.Equal(t, vals["0"], big.NewInt(1))

	abi1, err := abi.NewABIFromList([]string{
		"function set() view returns (uint256)",
	})
	assert.NoError(t, err)

	c1 := NewContract(addr, abi1, WithJsonRPC(p.Eth()))
	vals, err = c1.Call("set", ethgo.Latest)
	assert.NoError(t, err)
	assert.Equal(t, vals["0"], big.NewInt(1))
}

func TestContract_IO(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	cc := &testutil.Contract{}
	cc.AddDualCaller("setA", "address", "uint256")

	contract, addr := s.DeployContract(cc)

	abi, err := abi.NewABI(contract.Abi)
	assert.NoError(t, err)

	c := NewContract(addr, abi, WithJsonRPCEndpoint(s.HTTPAddr()))

	resp, err := c.Call("setA", ethgo.Latest, addr0B, 1000)
	assert.NoError(t, err)

	assert.Equal(t, resp["0"], addr0B)
	assert.Equal(t, resp["1"], big.NewInt(1000))
}

func TestContract_From(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	cc := &testutil.Contract{}
	cc.AddCallback(func() string {
		return `function example() public view returns (address) {
			return msg.sender;	
		}`
	})

	contract, addr := s.DeployContract(cc)

	abi, err := abi.NewABI(contract.Abi)
	assert.NoError(t, err)

	from := ethgo.Address{0x1}
	c := NewContract(addr, abi, WithSender(from), WithJsonRPCEndpoint(s.HTTPAddr()))

	resp, err := c.Call("example", ethgo.Latest)
	assert.NoError(t, err)
	assert.Equal(t, resp["0"], from)
}

func TestContract_Deploy(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	// create an address and fund it
	key, _ := wallet.GenerateKey()
	s.Fund(key.Address())

	p, _ := jsonrpc.NewClient(s.HTTPAddr())

	cc := &testutil.Contract{}
	cc.AddConstructor("address", "uint256")

	artifact, err := cc.Compile()
	assert.NoError(t, err)

	abi, err := abi.NewABI(artifact.Abi)
	assert.NoError(t, err)

	bin, err := hex.DecodeString(artifact.Bin)
	assert.NoError(t, err)

	txn, err := DeployContract(abi, bin, []interface{}{ethgo.Address{0x1}, 1000}, WithJsonRPC(p.Eth()), WithSender(key))
	assert.NoError(t, err)

	assert.NoError(t, txn.Do())
	receipt, err := txn.Wait()
	assert.NoError(t, err)

	i := NewContract(receipt.ContractAddress, abi, WithJsonRPC(p.Eth()))
	resp, err := i.Call("val_0", ethgo.Latest)
	assert.NoError(t, err)
	assert.Equal(t, resp["0"], ethgo.Address{0x1})

	resp, err = i.Call("val_1", ethgo.Latest)
	assert.NoError(t, err)
	assert.Equal(t, resp["0"], big.NewInt(1000))
}

func TestContract_Transaction(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	// create an address and fund it
	key, _ := wallet.GenerateKey()
	s.Fund(key.Address())

	cc := &testutil.Contract{}
	cc.AddEvent(testutil.NewEvent("A").Add("uint256", true))
	cc.EmitEvent("setA", "A", "1")

	artifact, addr := s.DeployContract(cc)

	abi, err := abi.NewABI(artifact.Abi)
	assert.NoError(t, err)

	// send multiple transactions
	contract := NewContract(addr, abi, WithJsonRPCEndpoint(s.HTTPAddr()), WithSender(key))

	for i := 0; i < 10; i++ {
		txn, err := contract.Txn("setA")
		assert.NoError(t, err)

		err = txn.Do()
		assert.NoError(t, err)

		receipt, err := txn.Wait()
		assert.NoError(t, err)
		assert.Len(t, receipt.Logs, 1)
	}
}

func TestContract_CallAtBlock(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	// create an address and fund it
	key, _ := wallet.GenerateKey()
	s.Fund(key.Address())

	cc := &testutil.Contract{}
	cc.AddCallback(func() string {
		return `
		uint256 val = 1;
		function getVal() public view returns (uint256) {
			return val;
		}
		function change() public payable {
			val = 2;
		}`
	})

	artifact, addr := s.DeployContract(cc)

	abi, err := abi.NewABI(artifact.Abi)
	assert.NoError(t, err)

	contract := NewContract(addr, abi, WithJsonRPCEndpoint(s.HTTPAddr()), WithSender(key))

	checkVal := func(block ethgo.BlockNumber, expected *big.Int) {
		resp, err := contract.Call("getVal", block)
		assert.NoError(t, err)
		assert.Equal(t, resp["0"], expected)
	}

	// initial value is 1
	checkVal(ethgo.Latest, big.NewInt(1))

	// send a transaction to update the state
	var receipt *ethgo.Receipt
	{
		txn, err := contract.Txn("change")
		assert.NoError(t, err)

		err = txn.Do()
		assert.NoError(t, err)

		receipt, err = txn.Wait()
		assert.NoError(t, err)
	}

	// validate the state at different blocks
	{
		// value at receipt block is 2
		checkVal(ethgo.BlockNumber(receipt.BlockNumber), big.NewInt(2))

		// value at previous block is 1
		checkVal(ethgo.BlockNumber(receipt.BlockNumber-1), big.NewInt(1))
	}
}

func TestContract_SendValueContractCall(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	key, _ := wallet.GenerateKey()
	s.Fund(key.Address())

	cc := &testutil.Contract{}
	cc.AddCallback(func() string {
		return `
		function deposit() public payable {
		}`
	})

	artifact, addr := s.DeployContract(cc)

	abi, err := abi.NewABI(artifact.Abi)
	assert.NoError(t, err)

	contract := NewContract(addr, abi, WithJsonRPCEndpoint(s.HTTPAddr()), WithSender(key))

	balance := big.NewInt(1)

	txn, err := contract.Txn("deposit")
	txn.WithOpts(&TxnOpts{Value: balance})
	assert.NoError(t, err)

	err = txn.Do()
	assert.NoError(t, err)

	_, err = txn.Wait()
	assert.NoError(t, err)

	client, _ := jsonrpc.NewClient(s.HTTPAddr())
	found, err := client.Eth().GetBalance(addr, ethgo.Latest)
	assert.NoError(t, err)
	assert.Equal(t, found, balance)
}
