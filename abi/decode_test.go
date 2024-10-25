package abi

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecode_BytesBound(t *testing.T) {
	typ := MustNewType("tuple(string)")
	decodeTuple(typ, nil) // it should not panic
}

func TestDecode_DynamicLengthOutOfBounds(t *testing.T) {
	input := []byte("00000000000000000000000000000000\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00 00000000000000000000000000")
	typ := MustNewType("tuple(bytes32, bytes, bytes)")

	_, err := Decode(typ, input)
	require.Error(t, err)
}
