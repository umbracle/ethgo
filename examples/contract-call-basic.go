package examples

import (
	"fmt"
	"math/big"

	"github.com/cloudwalk/ethgo"
	"github.com/cloudwalk/ethgo/abi"
	"github.com/cloudwalk/ethgo/contract"
	"github.com/cloudwalk/ethgo/jsonrpc"
)

func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}

// call a contract
func contractCall() {
	var functions = []string{
		"function totalSupply() view returns (uint256)",
	}

	abiContract, err := abi.NewABIFromList(functions)
	handleErr(err)

	// Matic token
	addr := ethgo.HexToAddress("0x7d1afa7b718fb893db30a3abc0cfc608aacfebb0")

	client, err := jsonrpc.NewClient("https://mainnet.infura.io")
	handleErr(err)

	c := contract.NewContract(addr, abiContract, contract.WithJsonRPC(client.Eth()))
	res, err := c.Call("totalSupply", ethgo.Latest)
	handleErr(err)

	fmt.Printf("TotalSupply: %s", res["totalSupply"].(*big.Int))
}
