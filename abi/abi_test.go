package abi

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbi(t *testing.T) {
	cases := []struct {
		Input  string
		Output *ABI
	}{
		{
			Input: `[
				{
					"name": "abc",
					"type": "function"
				}
			]`,
			Output: &ABI{
				Methods: map[string]*Method{
					"abc": &Method{
						Name:    "abc",
						Inputs:  &Type{kind: KindTuple, raw: "tuple", tuple: []*TupleElem{}},
						Outputs: &Type{kind: KindTuple, raw: "tuple", tuple: []*TupleElem{}},
					},
				},
				Events: map[string]*Event{},
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
		"event Transfer(address from, address to, uint256 amount)",
		"function symbol() returns (string)",
	}
	vv, err := NewABIFromList(cases)
	assert.NoError(t, err)

	fmt.Println(vv.Methods["symbol"].Inputs.String())
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
			input:     "(address)",
			output:    "(address)",
		},
		{
			// no input
			signature: "function approve() returns (address)",
			name:      "approve",
			input:     "()",
			output:    "(address)",
		},
		{
			// no output
			signature: "function approve(address)",
			name:      "approve",
			input:     "(address)",
			output:    "()",
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
