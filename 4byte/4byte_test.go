package fourbyte

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, i.out, found)
	}

}
