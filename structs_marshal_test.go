package ethgo

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateHashPtr(input string) *Hash {
	res := HexToHash(input)
	return &res
}

func TestLogFilter_MarshalJSON(t *testing.T) {
	testTable := []struct {
		name   string
		topics [][]*Hash
	}{
		{
			"match any topic",
			[][]*Hash{
				nil,
				{},
			},
		},
		{
			"match single topic in pos. 1",
			[][]*Hash{
				{
					generateHashPtr("0xa"),
				},
			},
		},
		{
			"match single topic in pos. 2",
			[][]*Hash{
				{},
				{
					generateHashPtr("0xb"),
				},
			},
		},
		{
			"match topic in pos. 1 AND pos. 2",
			[][]*Hash{
				{
					generateHashPtr("0xa"),
				},
				{
					generateHashPtr("0xb"),
				},
			},
		},
		{
			"match topic A or B in pos. 1 AND C or D in pos. 2",
			[][]*Hash{
				{
					generateHashPtr("0xa"),
					generateHashPtr("0xb"),
				},
				{
					generateHashPtr("0xc"),
					generateHashPtr("0xd"),
				},
			},
		},
	}

	defaultLogFilter := &LogFilter{
		Address:   []Address{HexToAddress("0x123")},
		Topics:    nil,
		BlockHash: generateHashPtr("0xabc"),
	}
	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			defaultLogFilter.Topics = testCase.topics

			// Marshal it to JSON
			output, marshalErr := defaultLogFilter.MarshalJSON()
			if marshalErr != nil {
				t.Fatalf("Unable to marshal value, %v", marshalErr)
			}

			// Unmarshal it from JSON
			reverseOutput := &LogFilter{}
			unmarshalErr := json.Unmarshal(output, reverseOutput)
			if unmarshalErr != nil {
				t.Fatalf("Unable to unmarshal value, %v", unmarshalErr)
			}

			// Assert that the original and unmarshalled values match
			assert.Equal(t, defaultLogFilter, reverseOutput)
		})
	}
}

func TestMarshal_StateOverride(t *testing.T) {
	nonce := uint64(1)
	code := []byte{0x1}

	o := StateOverride{
		{0x0}: OverrideAccount{
			Nonce:   &nonce,
			Balance: big.NewInt(1),
			Code:    &code,
			State: &map[Hash]Hash{
				{0x1}: {0x1},
			},
			StateDiff: &map[Hash]Hash{
				{0x1}: {0x1},
			},
		},
	}

	res, err := o.MarshalJSON()
	require.NoError(t, err)

	expected := `{"0x0000000000000000000000000000000000000000":{"nonce":"0x1","balance":"0x1","code":"0x01","state":{"0x0100000000000000000000000000000000000000000000000000000000000000":"0x0100000000000000000000000000000000000000000000000000000000000000"},"stateDiff":{"0x0100000000000000000000000000000000000000000000000000000000000000":"0x0100000000000000000000000000000000000000000000000000000000000000"}}}`
	require.Equal(t, expected, string(res))
}
