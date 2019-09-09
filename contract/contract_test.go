package contract

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/abi"
	"github.com/umbracle/go-web3/jsonrpc"
	"github.com/umbracle/go-web3/testutil"
)

func mustDecode(data string) []byte {
	raw, err := hex.DecodeString(data[2:])
	if err != nil {
		panic(err)
	}
	return raw
}

func mustDecodeAddr(data string) [20]byte {
	raw := mustDecode(data)
	buf := [20]byte{}
	copy(buf[:], raw)
	return buf
}

var (
	addr0  = "0x0000000000000000000000000000000000000000"
	addr0B = mustDecodeAddr(addr0)
)

func TestContract(t *testing.T) {
	s := testutil.NewTestServer(t, nil)
	defer s.Close()

	cc := &testutil.Contract{}
	cc.AddDualCaller("setA", "address", "uint256")

	contract, addr := s.DeployContract(cc)

	abi, err := abi.NewABI(contract.Abi)
	assert.NoError(t, err)

	p := jsonrpc.NewClient(s.HTTPAddr())
	c := NewContract(addr, abi, p)
	c.SetFrom(s.Account(0))

	resp, err := c.Call("setA", web3.Latest, addr0B, 1000)
	assert.NoError(t, err)

	assert.Equal(t, resp["0"], addr0B)
	assert.Equal(t, resp["1"], big.NewInt(1000))
}
