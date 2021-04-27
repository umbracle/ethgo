package abi

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/umbracle/go-web3"
	"github.com/umbracle/go-web3/compiler"
	"github.com/umbracle/go-web3/testutil"
)

func encodeHex(b []byte) string {
	return "0x" + hex.EncodeToString(b)
}

func decodeHex(str string) []byte {
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
	}
	buf, err := hex.DecodeString(str)
	if err != nil {
		panic(fmt.Errorf("could not decode hex: %v", err))
	}
	return buf
}

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
			decodeHex("0x12345678911121314151617181920211"),
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
			[]web3.Address{{1}, {2}},
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
				decodeHex("0x11"),
				decodeHex("0x22"),
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
			"tuple(bytes[] a)",
			map[string]interface{}{
				"a": [][]byte{{0xf0, 0xf0, 0xf0}, {0xf0, 0xf0, 0xf0}},
			},
		},
		{
			"tuple(uint32[2][][] a)",
			// `[{"type": "uint32[2][][]"}]`,
			map[string]interface{}{
				"a": [][][2]uint32{{{uint32(1), uint32(200)}, {uint32(1), uint32(1000)}}, {{uint32(1), uint32(200)}, {uint32(1), uint32(1000)}}},
			},
		},
		{
			"tuple(uint64[2] a)",
			map[string]interface{}{
				"a": [2]uint64{1, 2},
			},
		},
		{
			"tuple(uint32[2][3][4] a)",
			map[string]interface{}{
				"a": [4][3][2]uint32{{{1, 2}, {3, 4}, {5, 6}}, {{7, 8}, {9, 10}, {11, 12}}, {{13, 14}, {15, 16}, {17, 18}}, {{19, 20}, {21, 22}, {23, 24}}},
			},
		},
		{
			"tuple(int32[] a)",
			map[string]interface{}{
				"a": []int32{1, 2},
			},
		},
		{
			"tuple(int32 a, int32 b)",
			map[string]interface{}{
				"a": int32(1),
				"b": int32(2),
			},
		},
		{
			"tuple(string a, int32 b)",
			map[string]interface{}{
				"a": "Hello Worldxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
				"b": int32(2),
			},
		},
		{
			"tuple(int32[2] a, int32[] b)",
			map[string]interface{}{
				"a": [2]int32{1, 2},
				"b": []int32{4, 5, 6},
			},
		},
		{
			// First dynamic second static
			"tuple(int32[] a, int32[2] b)",
			map[string]interface{}{
				"a": []int32{1, 2, 3},
				"b": [2]int32{4, 5},
			},
		},
		{
			// Both dynamic
			"tuple(int32[] a, int32[] b)",
			map[string]interface{}{
				"a": []int32{1, 2, 3},
				"b": []int32{4, 5, 6},
			},
		},
		{
			"tuple(string a, int64 b)",
			map[string]interface{}{
				"a": "hello World",
				"b": int64(266),
			},
		},
		{
			// tuple array
			"tuple(int32 a, int32 b)[2]",
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
			"tuple(int32[] a)[2]",
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
			"tuple(int32 a, int32[] b)[]",
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
			"tuple(tuple(int32 c, int32[] d) a, int32[] b)",
			map[string]interface{}{
				"a": map[string]interface{}{
					"c": int32(5),
					"d": []int32{3, 4},
				},
				"b": []int32{1, 2},
			},
		},
		{
			"tuple(uint8[2] a, tuple(uint8 e, uint32 f)[2] b, uint16 c, uint64[2][1] d)",
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
			"tuple(uint16 a, uint16 b)[1][]",
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
			"tuple(uint64[][] a, tuple(uint8 a, uint32 b)[1] b, uint64 c)",
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

	server := testutil.NewTestServer(t, nil)
	defer server.Close()

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			tt, err := NewType(c.Type)
			if err != nil {
				t.Fatal(err)
			}

			if err := testEncodeDecode(t, server, tt, c.Input); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEncodingArguments(t *testing.T) {
	cases := []struct {
		Arg   *ArgumentStr
		Input interface{}
	}{
		{
			&ArgumentStr{
				Type: "tuple",
				Components: []*ArgumentStr{
					&ArgumentStr{
						Name: "",
						Type: "int32",
					},
					&ArgumentStr{
						Name: "",
						Type: "int32",
					},
				},
			},
			map[string]interface{}{
				"0": int32(1),
				"1": int32(2),
			},
		},
		{
			&ArgumentStr{
				Type: "tuple",
				Components: []*ArgumentStr{
					&ArgumentStr{
						Name: "a",
						Type: "int32",
					},
					&ArgumentStr{
						Name: "",
						Type: "int32",
					},
				},
			},
			map[string]interface{}{
				"a": int32(1),
				"1": int32(2),
			},
		},
	}

	server := testutil.NewTestServer(t, nil)
	defer server.Close()

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			tt, err := NewTypeFromArgument(c.Arg)
			if err != nil {
				t.Fatal(err)
			}

			if err := testEncodeDecode(t, server, tt, c.Input); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func testEncodeDecode(t *testing.T, server *testutil.TestServer, tt *Type, input interface{}) error {
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
		if err := testTypeWithContract(t, server, tt); err != nil {
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

	server := testutil.NewTestServer(t, nil)
	defer server.Close()

	for i := 0; i < int(n); i++ {
		t.Run("", func(t *testing.T) {
			tt := generateRandomArgs(randomInt(1, 4))
			input := generateRandomType(tt)

			if err := testEncodeDecode(t, server, tt, input); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func testTypeWithContract(t *testing.T, server *testutil.TestServer, typ *Type) error {
	g := &generateContractImpl{}
	source := g.run(typ)

	output, err := compiler.NewSolidityCompiler("solc").(*compiler.Solidity).CompileCode(source)
	if err != nil {
		return err
	}
	solcContract, ok := output["<stdin>:Sample"]
	if !ok {
		return fmt.Errorf("Expected the contract to be called Sample")
	}

	abi, err := NewABI(string(solcContract.Abi))
	if err != nil {
		return err
	}

	binBuf, err := hex.DecodeString(solcContract.Bin)
	if err != nil {
		return err
	}
	txn := &web3.Transaction{
		Input: binBuf,
	}
	receipt, err := server.SendTxn(txn)
	if err != nil {
		return err
	}

	method, ok := abi.Methods["set"]
	if !ok {
		return fmt.Errorf("method set not found")
	}

	tt := method.Inputs
	val := generateRandomType(tt)

	data, err := Encode(val, tt)
	if err != nil {
		return err
	}

	res, err := server.Call(&web3.CallMsg{
		To:   &receipt.ContractAddress,
		Data: append(method.ID(), data...),
	})
	if err != nil {
		return err
	}
	if res != encodeHex(data) {
		return fmt.Errorf("bad")
	}
	return nil
}

func TestEncodingStruct(t *testing.T) {
	typ := MustNewType("tuple(address a, uint256 b)")

	type Obj struct {
		A web3.Address
		B *big.Int
	}
	obj := Obj{
		A: web3.Address{0x1},
		B: big.NewInt(1),
	}

	encoded, err := typ.Encode(&obj)
	if err != nil {
		t.Fatal(err)
	}

	var obj2 Obj
	if err := typ.DecodeStruct(encoded, &obj2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(obj, obj2) {
		t.Fatal("bad")
	}
}
