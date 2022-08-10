package abi

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbi(t *testing.T) {
	methodOutput := &Method{
		Name:    "abc",
		Inputs:  MustNewType("tuple()"),
		Outputs: MustNewType("tuple()"),
	}
	balanceFunc := &Method{
		Name:    "balanceOf",
		Const:   true,
		Inputs:  MustNewType("tuple(address owner)"),
		Outputs: MustNewType("tuple(uint256 balance)"),
	}

	cases := []struct {
		Input  string
		Output *ABI
	}{
		{
			Input: `[
				{
					"name": "abc",
					"type": "function"
				},
				{
					"name": "cde",
					"type": "event",
					"inputs": [
						{
							"indexed": true,
							"name": "a",
							"type": "address"
						}
					]
				},
				{
					"name": "def",
					"type": "error",
					"inputs": [
						{
							"indexed": true,
							"name": "a",
							"type": "address"
						}
					]
				},
				{
					"type": "function",
					"name": "balanceOf",
					"constant": true,
					"stateMutability": "view",
				 	"payable": false,
					"inputs": [
						{
					    	"type": "address",
					    	"name": "owner"
					   	}
					],
					"outputs": [
						{
					    	"type": "uint256",
					    	"name": "balance"
					   	}
					]
				}
			]`,
			Output: &ABI{
				Events: map[string]*Event{
					"cde": {
						Name:   "cde",
						Inputs: MustNewType("tuple(address indexed a)"),
					},
				},
				Methods: map[string]*Method{
					"abc":       methodOutput,
					"balanceOf": balanceFunc,
				},
				MethodsBySignature: map[string]*Method{
					"abc()":              methodOutput,
					"balanceOf(address)": balanceFunc,
				},
				Errors: map[string]*Error{
					"def": {
						Name:   "def",
						Inputs: MustNewType("tuple(address indexed a)"),
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			abi, err := NewABI(c.Input)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(abi, c.Output) {
				fmt.Println(reflect.DeepEqual(abi.Methods["balanceOf"].Outputs, c.Output.Methods["balanceOf"].Outputs))
				t.Fatal("bad")
			}

		})
	}
}

func TestAbi_Polymorphism(t *testing.T) {
	// This ABI contains 2 "transfer" functions (polymorphism)
	const polymorphicABI = `[
        {
            "inputs": [
                {
                    "internalType": "address",
                    "name": "_to",
                    "type": "address"
                },
                {
                    "internalType": "address",
                    "name": "_token",
                    "type": "address"
                },
                {
                    "internalType": "uint256",
                    "name": "_amount",
                    "type": "uint256"
                }
            ],
            "name": "transfer",
            "outputs": [
                {
                    "internalType": "bool",
                    "name": "",
                    "type": "bool"
                }
            ],
            "stateMutability": "nonpayable",
            "type": "function"
        },
		{
            "inputs": [
                {
                    "internalType": "address",
                    "name": "_to",
                    "type": "address"
                },
                {
                    "internalType": "uint256",
                    "name": "_amount",
                    "type": "uint256"
                }
            ],
            "name": "transfer",
            "outputs": [
                {
                    "internalType": "bool",
                    "name": "",
                    "type": "bool"
                }
            ],
            "stateMutability": "nonpayable",
            "type": "function"
        }
    ]`

	abi, err := NewABI(polymorphicABI)
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, abi.Methods, 2)
	assert.Equal(t, abi.GetMethod("transfer").Sig(), "transfer(address,address,uint256)")
	assert.Equal(t, abi.GetMethod("transfer0").Sig(), "transfer(address,uint256)")
	assert.NotEmpty(t, abi.GetMethodBySignature("transfer(address,address,uint256)"))
	assert.NotEmpty(t, abi.GetMethodBySignature("transfer(address,uint256)"))
}

func TestAbi_HumanReadable(t *testing.T) {
	cases := []string{
		"constructor(string symbol, string name)",
		"function transferFrom(address from, address to, uint256 value)",
		"function balanceOf(address owner) view returns (uint256 balance)",
		"function balanceOf() view returns ()",
		"event Transfer(address indexed from, address indexed to, address value)",
		"error InsufficientBalance(address owner, uint256 balance)",
		"function addPerson(tuple(string name, uint16 age) person)",
		"function addPeople(tuple(string name, uint16 age)[] person)",
		"function getPerson(uint256 id) view returns (tuple(string name, uint16 age))",
		"event PersonAdded(uint256 indexed id, tuple(string name, uint16 age) person)",
	}
	vv, err := NewABIFromList(cases)
	assert.NoError(t, err)

	// make it nil to not compare it and avoid writing each method twice for the test
	vv.MethodsBySignature = nil

	expect := &ABI{
		Constructor: &Method{
			Inputs: MustNewType("tuple(string symbol, string name)"),
		},
		Methods: map[string]*Method{
			"transferFrom": &Method{
				Name:    "transferFrom",
				Inputs:  MustNewType("tuple(address from, address to, uint256 value)"),
				Outputs: MustNewType("tuple()"),
			},
			"balanceOf": &Method{
				Name:    "balanceOf",
				Inputs:  MustNewType("tuple(address owner)"),
				Outputs: MustNewType("tuple(uint256 balance)"),
			},
			"balanceOf0": &Method{
				Name:    "balanceOf",
				Inputs:  MustNewType("tuple()"),
				Outputs: MustNewType("tuple()"),
			},
			"addPerson": &Method{
				Name:    "addPerson",
				Inputs:  MustNewType("tuple(tuple(string name, uint16 age) person)"),
				Outputs: MustNewType("tuple()"),
			},
			"addPeople": &Method{
				Name:    "addPeople",
				Inputs:  MustNewType("tuple(tuple(string name, uint16 age)[] person)"),
				Outputs: MustNewType("tuple()"),
			},
			"getPerson": &Method{
				Name:    "getPerson",
				Inputs:  MustNewType("tuple(uint256 id)"),
				Outputs: MustNewType("tuple(tuple(string name, uint16 age))"),
			},
		},
		Events: map[string]*Event{
			"Transfer": &Event{
				Name:   "Transfer",
				Inputs: MustNewType("tuple(address indexed from, address indexed to, address value)"),
			},
			"PersonAdded": &Event{
				Name:   "PersonAdded",
				Inputs: MustNewType("tuple(uint256 indexed id, tuple(string name, uint16 age) person)"),
			},
		},
		Errors: map[string]*Error{
			"InsufficientBalance": &Error{
				Name:   "InsufficientBalance",
				Inputs: MustNewType("tuple(address owner, uint256 balance)"),
			},
		},
	}
	assert.Equal(t, expect, vv)
}

func TestAbi_ParseMethodSignature(t *testing.T) {
	cases := []struct {
		signature string
		name      string
		input     string
		output    string
	}{
		{
			// both input and output
			signature: "function approve(address to) returns (address)",
			name:      "approve",
			input:     "tuple(address)",
			output:    "tuple(address)",
		},
		{
			// no input
			signature: "function approve() returns (address)",
			name:      "approve",
			input:     "tuple()",
			output:    "tuple(address)",
		},
		{
			// no output
			signature: "function approve(address)",
			name:      "approve",
			input:     "tuple(address)",
			output:    "tuple()",
		},
		{
			// multiline
			signature: `function a(
				uint256 b,
				address[] c
			)
				returns
				(
				uint256[] d
			)`,
			name:   "a",
			input:  "tuple(uint256,address[])",
			output: "tuple(uint256[])",
		},
	}

	for _, c := range cases {
		name, input, output, err := parseMethodSignature(c.signature)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, name, c.name)

		if input != nil {
			assert.Equal(t, c.input, input.String())
		} else {
			assert.Equal(t, c.input, "")
		}

		if input != nil {
			assert.Equal(t, c.output, output.String())
		} else {
			assert.Equal(t, c.output, "")
		}
	}
}
