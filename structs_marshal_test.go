package web3

import (
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	addr0 = "0x0000000000000000000000000000000000000000"
)

func cleanStr(str string) string {
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\n", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	return str
}

func TestMarshal(t *testing.T) {
	cases := []struct {
		Input  json.Marshaler
		Result string
	}{
		{
			Input: &Transaction{},
			Result: `{
				"from": "` + addr0 + `",
				"gasPrice": "0x0",
				"gas": "0x0"
			}`,
		},
		{
			Input: &Transaction{
				GasPrice: 100,
				Gas:      50,
				Value:    big.NewInt(100),
			},
			Result: `{
				"from": "` + addr0 + `",
				"gasPrice": "0x64",
				"gas": "0x32",
				"value": "0x64"
			}`,
		},
	}

	for _, c := range cases {
		raw, err := c.Input.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, string(raw), cleanStr(c.Result))
	}
}
