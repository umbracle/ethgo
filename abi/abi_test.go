package abi

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbiXX(t *testing.T) {
	cases := []struct {
		Input  string
		Output *ABI
	}{
		{
			Input: `[
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
					"balanceOf": {
						Name:    "balanceOf",
						Const:   true,
						Inputs:  MustNewType("tuple(address owner)"),
						Outputs: MustNewType("tuple(uint256 balance)"),
					},
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

func TestAbi_HumanReadable(t *testing.T) {
	cases := []string{
		"event Transfer(address from, address to, uint256 amount)",
		"function symbol() returns (string)",
		"error A(int256 a)",
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

func TestAbi_ParseEventErrorSignature(t *testing.T) {
	cases := []struct {
		signature string
		name      string
		typ       string
	}{
		{
			signature: "event A(int256 a, int256 b)",
			name:      "A",
			typ:       "(int256,int256)",
		},
		{
			signature: "error A(int256 a, int256 b)",
			name:      "A",
			typ:       "(int256,int256)",
		},
	}

	for _, c := range cases {
		name, typ, err := parseEventOrErrorSignature(c.signature)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, name, c.name)
		assert.Equal(t, c.typ, typ.String())
	}
}
