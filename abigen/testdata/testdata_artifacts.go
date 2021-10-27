package testdata

import (
	"encoding/hex"
	"fmt"

	"github.com/umbracle/go-web3/abi"
)

var abiTestdata *abi.ABI

// TestdataAbi returns the abi of the Testdata contract
func TestdataAbi() *abi.ABI {
	return abiTestdata
}

var binTestdata []byte

func init() {
	var err error
	abiTestdata, err = abi.NewABI(abiTestdataStr)
	if err != nil {
		panic(fmt.Errorf("cannot parse Testdata abi: %v", err))
	}
	if len(binTestdataStr) != 0 {
		binTestdata, err = hex.DecodeString(binTestdataStr[2:])
		if err != nil {
			panic(fmt.Errorf("cannot parse Testdata bin: %v", err))
		}
	}
}

var binTestdataStr = ""

var abiTestdataStr = `[
    {
        "constant": false,
        "inputs": [
            {
                "name": "_val1",
                "type": "address"
            },
            {
                "name": "_val2",
                "type": "uint256"
            }
        ],
        "name": "txnBasicInput",
        "outputs": [
            {
                "name": "",
                "type": "bool"
            }
        ],
        "payable": false,
        "stateMutability": "nonpayable",
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "callBasicInput",
        "outputs": [
            {
                "name": "",
                "type": "uint256"
            },
            {
                "name": "",
                "type": "address"
            }
        ],
        "payable": false,
        "stateMutability": "view",
        "type": "function"
    },
    {
        "anonymous": false,
        "inputs": [
            {
                "indexed": true,
                "name": "owner",
                "type": "address"
            },
            {
                "indexed": true,
                "name": "spender",
                "type": "address"
            },
            {
                "indexed": false,
                "name": "value",
                "type": "uint256"
            }
        ],
        "name": "eventBasic",
        "type": "event"
    }
]`
