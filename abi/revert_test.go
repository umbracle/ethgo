package abi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnpackRevertError(t *testing.T) {
	data := "08c379a00000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000000d72657665727420726561736f6e00000000000000000000000000000000000000"

	reason, err := UnpackRevertError(decodeHex(data))
	assert.NoError(t, err)
	assert.Equal(t, "revert reason", reason)
}
