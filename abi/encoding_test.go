package abi

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/umbracle/ethgo"
	"github.com/umbracle/ethgo/compiler"
	"github.com/umbracle/ethgo/testutil"
)

func mustDecodeHex(str string) []byte {
	buf, err := decodeHex(str)
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
			mustDecodeHex("0x12345678911121314151617181920211"),
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
			[]ethgo.Address{{1}, {2}},
		},
		{
			"bytes10[]",
			[][10]byte{
				{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0x10},
				{0x1, 0x2, 0x3, 0x4, 0x5, 0x6, 0x7, 0x8, 0x9, 0x10},
			},
		},
		{
			"bytes[]",
			[][]byte{
				mustDecodeHex("0x11"),
				mustDecodeHex("0x22"),
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
			// tuple with array slice
			"tuple(address[] a)",
			map[string]interface{}{
				"a": []ethgo.Address{
					{0x1},
				},
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
				{
					"a": int32(1),
					"b": int32(2),
				},
				{
					"a": int32(3),
					"b": int32(4),
				},
			},
		},

		{
			// tuple array with dynamic content
			"tuple(int32[] a)[2]",
			[2]map[string]interface{}{
				{
					"a": []int32{1, 2, 3},
				},
				{
					"a": []int32{4, 5, 6},
				},
			},
		},
		{
			// tuple slice
			"tuple(int32 a, int32[] b)[]",
			[]map[string]interface{}{
				{
					"a": int32(1),
					"b": []int32{2, 3},
				},
				{
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
					{
						"e": uint8(10),
						"f": uint32(11),
					},
					{
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
				{
					{
						"a": uint16(1),
						"b": uint16(2),
					},
				},
				{
					{
						"a": uint16(3),
						"b": uint16(4),
					},
				},
				{
					{
						"a": uint16(5),
						"b": uint16(6),
					},
				},
				{
					{
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
					{3, 4},
				},
				"b": [1]map[string]interface{}{
					{
						"a": uint8(1),
						"b": uint32(2),
					},
				},
				"c": uint64(10),
			},
		},
	}

	server := testutil.NewTestServer(t)

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			t.Parallel()

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

func TestEncodingBestEffort(t *testing.T) {
	strAddress := "0xdbb881a51CD4023E4400CEF3ef73046743f08da3"
	ethAddress := ethgo.HexToAddress(strAddress)
	overflowBigInt, _ := new(big.Int).SetString("50000000000000000000000000000000000000", 10)

	cases := []struct {
		Type     string
		Input    interface{}
		Expected interface{}
	}{
		{
			"uint40",
			float64(50),
			big.NewInt(50),
		},
		{
			"uint40",
			"50",
			big.NewInt(50),
		},
		{
			"uint40",
			"0x32",
			big.NewInt(50),
		},
		{
			"int256",
			float64(2),
			big.NewInt(2),
		},
		{
			"int256",
			"50000000000000000000000000000000000000",
			overflowBigInt,
		},
		{
			"int256",
			"0x259DA6542D43623D04C5112000000000",
			overflowBigInt,
		},
		{
			"int256[]",
			[]interface{}{float64(1), float64(2)},
			[]*big.Int{big.NewInt(1), big.NewInt(2)},
		},
		{
			"int256[]",
			[]interface{}{"1", "2"},
			[]*big.Int{big.NewInt(1), big.NewInt(2)},
		},
		{
			"int256",
			float64(-10),
			big.NewInt(-10),
		},
		{
			"int256",
			"-10",
			big.NewInt(-10),
		},
		{
			"address[]",
			[]interface{}{strAddress, strAddress},
			[]ethgo.Address{ethAddress, ethAddress},
		},
		{
			"uint8[]",
			[]interface{}{float64(1), float64(2)},
			[]uint8{1, 2},
		},
		{
			"uint8[]",
			[]interface{}{"1", "2"},
			[]uint8{1, 2},
		},
		{
			"bytes",
			"0x11",
			[]uint8{17},
		},
		{
			"bytes32",
			"0x11",
			[32]uint8{17},
		},
		{
			"tuple(address a)",
			map[string]interface{}{
				"a": strAddress,
			},
			map[string]interface{}{
				"a": ethAddress,
			},
		},
		{
			"tuple(address[] a)",
			map[string]interface{}{
				"a": []interface{}{strAddress, strAddress},
			},
			map[string]interface{}{
				"a": []ethgo.Address{ethAddress, ethAddress},
			},
		},
		{
			"tuple(address a, int64 b)",
			map[string]interface{}{
				"a": strAddress,
				"b": float64(266),
			},
			map[string]interface{}{
				"a": ethAddress,
				"b": int64(266),
			},
		},
		{
			"tuple(address a, int256 b)",
			map[string]interface{}{
				"a": strAddress,
				"b": "50000000000000000000000000000000000000",
			},
			map[string]interface{}{
				"a": ethAddress,
				"b": overflowBigInt,
			},
		},
		{
			"tuple(address a, int256 b)",
			map[string]interface{}{
				"a": strAddress,
				"b": "0x259DA6542D43623D04C5112000000000",
			},
			map[string]interface{}{
				"a": ethAddress,
				"b": overflowBigInt,
			},
		},
	}

	server := testutil.NewTestServer(t)

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			tt, err := NewType(c.Type)
			if err != nil {
				t.Fatal(err)
			}

			res1, err := Encode(c.Input, tt)
			if err != nil {
				t.Fatal(err)
			}
			res2, err := Decode(tt, res1)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(res2, c.Expected) {
				t.Fatal("bad")
			}
			if tt.kind == KindTuple {
				if err := testTypeWithContract(t, server, tt); err != nil {
					t.Fatal(err)
				}
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
					{
						Name: "",
						Type: "int32",
					},
					{
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
					{
						Name: "a",
						Type: "int32",
					},
					{
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

	server := testutil.NewTestServer(t)

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

	server := testutil.NewTestServer(t)

	for i := 0; i < int(n); i++ {
		t.Run("", func(t *testing.T) {
			t.Parallel()

			tt := generateRandomArgs(randomInt(1, 4))
			input := generateRandomType(tt)

			if err := testEncodeDecode(t, server, tt, input); err != nil {
				t.Fatal(err)
			}

			if err := testDecodePanic(tt, input); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func testDecodePanic(tt *Type, input interface{}) error {
	// test that the encoded input and random permutattions of the response do not cause
	// panics on Decode function
	res1, err := Encode(input, tt)
	if err != nil {
		return err
	}

	buf := make([]byte, len(res1))

	// change each bit of the input with 1
	for i := 0; i < len(res1); i++ {
		copy(buf, res1)
		buf[i] = 0xff

		Decode(tt, buf)
	}

	return nil
}

func testTypeWithContract(t *testing.T, server *testutil.TestServer, typ *Type) error {
	g := &generateContractImpl{}
	source := g.run(typ)

	output, err := compiler.NewSolidityCompiler("solc").CompileCode(source)
	if err != nil {
		return err
	}
	solcContract, ok := output.Contracts["<stdin>:Sample"]
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
	txn := &ethgo.Transaction{
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

	data, err := method.Encode(val)
	if err != nil {
		return err
	}

	res, err := server.Call(&ethgo.CallMsg{
		To:   &receipt.ContractAddress,
		Data: data,
	})
	if err != nil {
		return err
	}
	if res != encodeHex(data[4:]) { // remove funct signature in data
		return fmt.Errorf("bad")
	}
	return nil
}

func TestEncodingStruct(t *testing.T) {
	typ := MustNewType("tuple(address aa, uint256 b)")

	type Obj struct {
		A ethgo.Address `abi:"aa"`
		B *big.Int
	}
	obj := Obj{
		A: ethgo.Address{0x1},
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

func TestEncodingStruct_camcelCase(t *testing.T) {
	typ := MustNewType("tuple(address aA, uint256 b)")

	type Obj struct {
		A ethgo.Address `abi:"aA"`
		B *big.Int
	}
	obj := Obj{
		A: ethgo.Address{0x1},
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

func TestEncodingStructDynamic(t *testing.T) {
	typ := MustNewType("tuple(string A, address B, uint256 C, bytes[] D, bytes[] E, address[] F, int256 G)")

	type Abcdefg struct {
		A string
		B ethgo.Address
		C *big.Int
		D [][]byte
		E [][]byte
		F []ethgo.Address
		G *big.Int
	}

	a := Abcdefg{
		A: "submitKeygen(bytes)",
		B: ethgo.HexToAddress("0xa16E02E87b7454126E5E10d957A927A7F5B5d2be"),
		C: big.NewInt(4),
		D: [][]byte{mustDecodeHex("0x04792730167230add71afb0459dd093980a5dbef6b8cfd2c9eef5f403d8b87a7a03da89bde572e8f564a39ad05452f854fe45328fa8ee7148fb8ee3131b78e6226"),
			mustDecodeHex("0x043770e37d91bbbb001e8c60de87d4fafd44626c8b85e08fbadf8f45778841a0462b0b88cea6cbb10ca931b0cb70d9d2aca23635100e0365bf1e6b07f929b45b32"),
			mustDecodeHex("0x04e397c219c024160ce8c5e35a23dd51ab6b9296cad9f3d6c03f7dbe6b294c4d61c529fd79bd30d1f2dda9a9f70d6f316de01ed9d100e0496cc30a4454215cb726"),
			mustDecodeHex("0x0463b437e92335bf367ab5b3b5bda4ff218cf5e2ac6555b47c187e20ac274476fcf30d1b56ce6fc861c23b8ab147f00df140c53291257ecb58e89e4815803f0f47")},
		E: [][]byte{},
		F: []ethgo.Address{},
		G: big.NewInt(0),
	}

	encoded, err := typ.Encode(&a)
	if err != nil {
		t.Fatal(err)
	}

	hexencoded := hex.EncodeToString(encoded)
	data := "000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000e0000000000000000000000000a16e02e87b7454126e5e10d957a927a7f5b5d2be0000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000012000000000000000000000000000000000000000000000000000000000000003c000000000000000000000000000000000000000000000000000000000000003e0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000137375626d69744b657967656e286279746573290000000000000000000000000000000000000000000000000000000000000000000000000000000000000000040000000000000000000000000000000000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001800000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000004104792730167230add71afb0459dd093980a5dbef6b8cfd2c9eef5f403d8b87a7a03da89bde572e8f564a39ad05452f854fe45328fa8ee7148fb8ee3131b78e6226000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000041043770e37d91bbbb001e8c60de87d4fafd44626c8b85e08fbadf8f45778841a0462b0b88cea6cbb10ca931b0cb70d9d2aca23635100e0365bf1e6b07f929b45b3200000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000004104e397c219c024160ce8c5e35a23dd51ab6b9296cad9f3d6c03f7dbe6b294c4d61c529fd79bd30d1f2dda9a9f70d6f316de01ed9d100e0496cc30a4454215cb7260000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000410463b437e92335bf367ab5b3b5bda4ff218cf5e2ac6555b47c187e20ac274476fcf30d1b56ce6fc861c23b8ab147f00df140c53291257ecb58e89e4815803f0f470000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	if hexencoded != data {
		t.Fatal("encoding failed for Abcdefg")
	}

	var b Abcdefg
	if err := typ.DecodeStruct(encoded, &b); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatal("bad encode-decode cycle in Abcdefg")
	}

	d, err := hex.DecodeString(data)
	if err != nil {
		t.Fatal(err)
	}
	res, err := typ.Decode(d)
	if err != nil {
		t.Fatal(err)
	}

	if (res.(*Abcdefg)).G != big.NewInt(1) {
		t.Fatal("decoding not as expected for Abcdefg")
	}
}
