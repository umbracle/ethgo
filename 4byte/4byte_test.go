package fourbyte

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test4Byte(t *testing.T) {
	var cases = []struct {
		in, out string
	}{
		{
			"0xddf252ad",
			"Transfer(address,address,uint256)",
		},
		{
			"0x42842e0e",
			"safeTransferFrom(address,address,uint256)",
		},
	}
	for _, i := range cases {
		found, err := Resolve(i.in)
		assert.NoError(t, err)
		assert.Equal(t, contains(found, i.out), true)
	}

}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
