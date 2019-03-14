package abi

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestEncoding(t *testing.T) {
	cases := []struct {
		Type  string
		Input interface{}
	}{
		{
			"uint40",
			big.NewInt(50),
		},
		{
			"int256",
			big.NewInt(2),
		},
		{
			"int256[]",
			[]*big.Int{big.NewInt(1), big.NewInt(2)},
		},
		{
			"int256",
			big.NewInt(-10),
		},
		{
			"bytes5",
			[5]byte{0x1, 0x2, 0x3, 0x4, 0x5},
		},
		{
			"bytes",
			hexutil.MustDecode("0x12345678911121314151617181920211"),
		},
		{
			"string",
			"foobar",
		},
		{
			"uint8[][2]",
			[2][]uint8{{1}, {1}},
		},
		{
			"address[]",
			[]common.Address{{1}, {2}},
		},
		{
			"bytes10[]",
			[][10]byte{
				[10]byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0x10},
				[10]byte{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0x10},
			},
		},
		{
			"bytes[]",
			[][]byte{
				hexutil.MustDecode("0x11"),
				hexutil.MustDecode("0x22"),
			},
		},
		{
			"uint32[2][3][4]",
			[4][3][2]uint32{{{1, 2}, {3, 4}, {5, 6}}, {{7, 8}, {9, 10}, {11, 12}}, {{13, 14}, {15, 16}, {17, 18}}, {{19, 20}, {21, 22}, {23, 24}}},
		},
		{
			"uint8[]",
			[]uint8{1, 2},
		},
		{
			"string[]",
			[]string{"hello", "foobar"},
		},
		{
			"string[2]",
			[2]string{"hello", "foobar"},
		},
		{
			"bytes32[][]",
			[][][32]uint8{{{1}, {2}}, {{3}, {4}, {5}}},
		},
		{
			"bytes32[][2]",
			[2][][32]uint8{{{1}, {2}}, {{3}, {4}, {5}}},
		},
		{
			"bytes32[3][2]",
			[2][3][32]uint8{{{1}, {2}, {3}}, {{3}, {4}, {5}}},
		},
		{
			"uint16[][2][]",
			[][2][]uint16{
				{{0, 1}, {2, 3}},
				{{4, 5}, {6, 7}},
			},
		},
		{
			"tuple(a bytes[])",
			map[string]interface{}{
				"a": [][]byte{{0xf0, 0xf0, 0xf0}, {0xf0, 0xf0, 0xf0}},
			},
		},
		{
			"tuple(a uint32[2][][])",
			// `[{"type": "uint32[2][][]"}]`,
			map[string]interface{}{
				"a": [][][2]uint32{{{uint32(1), uint32(200)}, {uint32(1), uint32(1000)}}, {{uint32(1), uint32(200)}, {uint32(1), uint32(1000)}}},
			},
		},
		{
			"tuple(a uint64[2])",
			map[string]interface{}{
				"a": [2]uint64{1, 2},
			},
		},
		{
			"tuple(a uint32[2][3][4])",
			map[string]interface{}{
				"a": [4][3][2]uint32{{{1, 2}, {3, 4}, {5, 6}}, {{7, 8}, {9, 10}, {11, 12}}, {{13, 14}, {15, 16}, {17, 18}}, {{19, 20}, {21, 22}, {23, 24}}},
			},
		},
		{
			"tuple(a int32[])",
			map[string]interface{}{
				"a": []int32{1, 2},
			},
		},
		{
			"tuple(a int32, b int32)",
			map[string]interface{}{
				"a": int32(1),
				"b": int32(2),
			},
		},
		{
			"tuple(a string, b int32)",
			map[string]interface{}{
				"a": "Hello Worldxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"b": int32(2),
			},
		},
		{
			"tuple(a int32[2], b int32[])",
			map[string]interface{}{
				"a": [2]int32{1, 2},
				"b": []int32{4, 5, 6},
			},
		},
		{
			// First dynamic second static
			"tuple(a int32[], b int32[2])",
			map[string]interface{}{
				"a": []int32{1, 2, 3},
				"b": [2]int32{4, 5},
			},
		},
		{
			// Both dynamic
			"tuple(a int32[], b int32[])",
			map[string]interface{}{
				"a": []int32{1, 2, 3},
				"b": []int32{4, 5, 6},
			},
		},
		{
			"tuple(a string, b int64)",
			map[string]interface{}{
				"a": "hello World",
				"b": int64(266),
			},
		},
		{
			// tuple array
			"tuple(a int32, b int32)[2]",
			[2]map[string]interface{}{
				map[string]interface{}{
					"a": int32(1),
					"b": int32(2),
				},
				map[string]interface{}{
					"a": int32(3),
					"b": int32(4),
				},
			},
		},

		{
			// tuple array with dynamic content
			"tuple(a int32[])[2]",
			[2]map[string]interface{}{
				map[string]interface{}{
					"a": []int32{1, 2, 3},
				},
				map[string]interface{}{
					"a": []int32{4, 5, 6},
				},
			},
		},
		{
			// tuple slice
			"tuple(a int32, b int32[])[]",
			[]map[string]interface{}{
				map[string]interface{}{
					"a": int32(1),
					"b": []int32{2, 3},
				},
				map[string]interface{}{
					"a": int32(4),
					"b": []int32{5, 6},
				},
			},
		},
		{
			// nested tuple
			"tuple(a tuple(c int32, d int32[]), b int32[])",
			map[string]interface{}{
				"a": map[string]interface{}{
					"c": int32(5),
					"d": []int32{3, 4},
				},
				"b": []int32{1, 2},
			},
		},
		{
			"tuple(a uint8[2], b tuple(e uint8, f uint32)[2], c uint16, d uint64[2][1])",
			map[string]interface{}{
				"a": [2]uint8{uint8(1), uint8(2)},
				"b": [2]map[string]interface{}{
					map[string]interface{}{
						"e": uint8(10),
						"f": uint32(11),
					},
					map[string]interface{}{
						"e": uint8(20),
						"f": uint32(21),
					},
				},
				"c": uint16(3),
				"d": [1][2]uint64{{uint64(4), uint64(5)}},
			},
		},
		{
			"tuple(a uint16, b uint16)[1][]",
			[][1]map[string]interface{}{
				[1]map[string]interface{}{
					map[string]interface{}{
						"a": uint16(1),
						"b": uint16(2),
					},
				},
				[1]map[string]interface{}{
					map[string]interface{}{
						"a": uint16(3),
						"b": uint16(4),
					},
				},
				[1]map[string]interface{}{
					map[string]interface{}{
						"a": uint16(5),
						"b": uint16(6),
					},
				},
				[1]map[string]interface{}{
					map[string]interface{}{
						"a": uint16(7),
						"b": uint16(8),
					},
				},
			},
		},
		{
			"tuple(a uint64[][], b tuple(a uint8, b uint32)[1], c uint64)",
			map[string]interface{}{
				"a": [][]uint64{
					[]uint64{3, 4},
				},
				"b": [1]map[string]interface{}{
					map[string]interface{}{
						"a": uint8(1),
						"b": uint32(2),
					},
				},
				"c": uint64(10),
			},
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			tt, err := NewType(c.Type)
			if err != nil {
				t.Fatal(err)
			}

			if err := testEncodeDecode(tt, c.Input); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func testEncodeDecode(tt *Type, input interface{}) error {
	res1, err := Encode(input, tt)
	if err != nil {
		return err
	}
	res2, err := Decode(tt, res1)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(res2, input) {
		return fmt.Errorf("bad")
	}
	if tt.kind == KindTuple {
		if err := testTypeWithContract(tt); err != nil {
			return err
		}
	}
	return nil
}

func generateRandomArgs(n int) *Type {
	inputs := []*TupleElem{}
	for i := 0; i < randomInt(1, 10); i++ {
		ttt, err := NewType(randomType())
		if err != nil {
			panic(err)
		}
		inputs = append(inputs, &TupleElem{
			Name: fmt.Sprintf("arg%d", i),
			Elem: ttt,
		})
	}
	return &Type{
		kind:  KindTuple,
		tuple: inputs,
	}
}

func TestRandomEncoding(t *testing.T) {
	rand.Seed(time.Now().UTC().UnixNano())

	nStr := os.Getenv("RANDOM_TESTS")
	n, err := strconv.Atoi(nStr)
	if err != nil {
		n = 100
	}

	for i := 0; i < int(n); i++ {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			tt := generateRandomArgs(randomInt(1, 4))
			input := generateRandomType(tt)

			if err := testEncodeDecode(tt, input); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func testTypeWithContract(t *Type) error {
	g := &generateContractImpl{}
	contract := g.run(t)

	client := newClient()
	accounts, err := client.listAccounts()
	if err != nil {
		return nil
	}

	etherbase := accounts[0]

	abi, receipt, err := compileAndDeployContract(contract, etherbase, client)
	if err != nil {
		if strings.Contains(err.Error(), "Stack too deep") {
			return nil
		}
		return err
	}

	method, ok := abi.Methods["set"]
	if !ok {
		return fmt.Errorf("method set not found")
	}

	tt := method.Inputs.Type()
	val := generateRandomType(tt)

	data, err := Encode(val, tt)
	if err != nil {
		return err
	}

	msg := ethereum.CallMsg{
		From: etherbase,
		To:   &receipt.ContractAddress,
		Data: append(method.ID(), data...),
	}

	resp, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		return fmt.Errorf("failed to call contract: %v", err)
	}
	if len(resp) == 0 {
		return fmt.Errorf("empty")
	}

	if !reflect.DeepEqual(resp, data) {
		return fmt.Errorf("bad")
	}
	return nil
}
