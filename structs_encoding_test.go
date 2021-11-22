package web3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

type jsonEncoder interface {
	json.Marshaler
	json.Unmarshaler
}

func TestJSONEncodingStd(t *testing.T) {
	b := new(Block)
	res, err := json.Marshal(b)
	assert.NoError(t, err)

	b1 := new(Block)
	assert.NoError(t, json.Unmarshal(res, b1))
	assert.Equal(t, b, b1)
}

func TestJSONEncoding(t *testing.T) {
	var (
		block = func() jsonEncoder { return new(Block) }
		txn   = func() jsonEncoder { return new(Transaction) }
	)

	cases := []struct {
		Input string
		build func() jsonEncoder
	}{
		{
			Input: `{
				"number": "0x1",
				"hash": "{{.Hash1}}",
				"parentHash": "{{.Hash2}}",
				"sha3Uncles": "{{.Hash3}}",
				"transactionsRoot": "{{.Hash1}}",
				"stateRoot": "{{.Hash3}}",
				"receiptsRoot": "{{.Hash2}}",
				"miner": "{{.Addr1}}",
				"gasLimit": "0x2",
				"gasUsed": "0x3",
				"timestamp": "0x4",
				"difficulty": "0x5",
				"extraData": "0x01",
				"uncles": [
					"{{.Hash1}}",
					"{{.Hash2}}"
				],
				"transactions": [
					"{{.Hash1}}"
				]
			}`,
			build: block,
		},
		{
			Input: `{
				"hash": "{{.Hash1}}",
				"from": "{{.Addr1}}",
				"input": "0x00",
				"value": "0x0",
				"gasPrice": "0x0",
				"gas": "0x0",
				"nonce": "0x10",
				"to": null,
				"v":"0x25",
				"r":"{{.Hash1}}",
				"s":"{{.Hash1}}",
				"blockHash": "{{.Hash0}}",
				"blockNumber": "0x0",
				"transactionIndex": "0x0"
				}`,
			build: txn,
		},
		{
			Input: `{
				"hash": "{{.Hash1}}",
				"from": "{{.Addr1}}",
				"input": "0x00",
				"value": "0x0",
				"gasPrice": "0x0",
				"gas": "0x0",
				"nonce": "0x10",
				"to": "{{.Addr1}}",
				"v":"0x25",
				"r":"{{.Hash1}}",
				"s":"{{.Hash1}}",
				"blockHash": "{{.Hash0}}",
				"blockNumber": "0x0",
				"transactionIndex": "0x0"
				}`,
			build: txn,
		},
	}

	for _, c := range cases {
		tmpl, err := template.New("test").Parse(c.Input)
		assert.NoError(t, err)

		config := map[string]string{}
		for i := 0; i <= 3; i++ {
			config[fmt.Sprintf("Hash%d", i)] = (Hash{byte(i)}).String()
			config[fmt.Sprintf("Addr%d", i)] = (Address{byte(i)}).String()
		}

		buffer := new(bytes.Buffer)
		assert.NoError(t, tmpl.Execute(buffer, config))

		input := compactJSON(buffer.String())
		obj := c.build()

		// unmarshal
		err = obj.UnmarshalJSON([]byte(input))
		assert.NoError(t, err)

		// now marshal
		res2, err := obj.MarshalJSON()
		assert.NoError(t, err)
		assert.Equal(t, string(res2), input)
	}
}

func compactJSON(s string) string {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, []byte(s)); err != nil {
		panic(err)
	}
	return buffer.String()
}
