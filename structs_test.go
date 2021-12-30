package web3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddress_Checksum(t *testing.T) {
	cases := []struct {
		src, dst string
	}{
		{
			"0x5aaeb6053f3e94c9b9a09f33669435e7ef1beaed",
			"0x5aAeb6053F3E94C9b9A09f33669435E7Ef1BeAed",
		},
		{
			"0xfb6916095ca1df60bb79ce92ce3ea74c37c5d359",
			"0xfB6916095ca1df60bB79Ce92cE3Ea74c37c5d359",
		},
		{
			"0xdbf03b407c01e7cd3cbea99509d93f8dddc8c6fb",
			"0xdbF03B407c01E7cD3CBea99509d93f8DDDC8C6FB",
		},
		{
			"0xd1220a0cf47c7b9be7a2e6ba89f429762e7b9adb",
			"0xD1220A0cf47c7B9Be7A2E6BA89F429762e7b9aDb",
		},
	}
	for _, c := range cases {
		addr := HexToAddress(c.src)
		assert.Equal(t, addr.String(), c.dst)
	}
}

func TestAddress_HexToString(t *testing.T) {
	assert.Equal(t, HexToAddress("0x1").String(), "0x0000000000000000000000000000000000000001")
	assert.Equal(t, HexToAddress("00000000000000000000000000000000000000001").String(), "0x0000000000000000000000000000000000000001")
	assert.Equal(t, HexToAddress("0000000000000000000000000000000000000001").String(), "0x0000000000000000000000000000000000000001")
}

func TestHash_HexToString(t *testing.T) {
	assert.Equal(t, HexToHash("1").String(), "0x0000000000000000000000000000000000000000000000000000000000000001")
}
