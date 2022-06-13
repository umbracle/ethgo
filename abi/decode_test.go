package abi

import "testing"

func TestDecode_BytesBound(t *testing.T) {
	typ := MustNewType("tuple(string)")
	decodeTuple(typ, nil) // it should not panic
}
