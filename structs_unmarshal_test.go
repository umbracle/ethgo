package web3

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	hash1 = Hash{0x1}
	hash2 = Hash{0x2}
	hash3 = Hash{0x3}

	addr1 = Address{0x1}
)

func TestUnmarshalBlock(t *testing.T) {
	cases := []struct {
		Input  string
		Result interface{}
	}{
		{
			Input: `{
				"hash": "` + hash1.String() + `",
				"parentHash": "` + hash2.String() + `",
				"sha3Uncles": "` + hash3.String() + `",
				"transactionsRoot": "` + hash1.String() + `",
				"receiptsRoot": "` + hash2.String() + `",
				"stateRoot": "` + hash3.String() + `",
				"miner": "` + addr1.String() + `",
				"number": "0x1",
				"gasLimit": "0x2",
				"gasUsed": "0x3",
				"timestamp": "0x4",
				"difficulty": "0x5",
				"extraData": "0x01",
				"uncles": [
					"` + hash1.String() + `",
					"` + hash2.String() + `"
				]
			}`,
			Result: &Block{
				Hash:             hash1,
				ParentHash:       hash2,
				Sha3Uncles:       hash3,
				TransactionsRoot: hash1,
				ReceiptsRoot:     hash2,
				StateRoot:        hash3,
				Miner:            addr1,
				Number:           1,
				GasLimit:         2,
				GasUsed:          3,
				Timestamp:        4,
				Difficulty:       big.NewInt(5),
				ExtraData:        []byte{0x1},
				Uncles: []Hash{
					hash1,
					hash2,
				},
			},
		},
	}

	for _, c := range cases {
		var b *Block
		assert.NoError(t, json.Unmarshal([]byte(c.Input), &b))
		assert.Equal(t, b, c.Result)
	}
}
