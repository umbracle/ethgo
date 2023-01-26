package abi

import "testing"

func TestDecode_BytesBound(t *testing.T) {
	typ := MustNewType("tuple(string)")
	decodeTuple(typ, nil) // it should not panic
}

func TestDecode_Unbound(t *testing.T) {
	input := []byte("00000000000000000000000000000000\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00 \x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x0000000000000000000000000000")
	inputDataABIType := MustNewType("tuple(bytes32, bytes, bytes)")
	_, _ = Decode(inputDataABIType, input)
}
