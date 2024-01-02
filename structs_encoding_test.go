package ethgo

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func compactJSON(s string) string {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, []byte(s)); err != nil {
		panic(err)
	}
	return buffer.String()
}

func TestDecodeL2Block(t *testing.T) {
	c := readTestsuite(t, "./testsuite/arbitrum-block-full.json")

	block := new(Block)
	require.NoError(t, block.UnmarshalJSON(c[0].content))

	for _, txn := range block.Transactions {
		require.NotEqual(t, txn.Type, 0)
	}
}

func TestEncodingJSON_Block(t *testing.T) {
	for _, c := range readTestsuite(t, "./testsuite/block-*.json") {
		content := []byte(compactJSON(string(c.content)))
		txn := new(Block)

		// unmarshal
		err := txn.UnmarshalJSON(content)
		assert.NoError(t, err)

		// marshal back
		res2, err := txn.MarshalJSON()
		assert.NoError(t, err)

		assert.Equal(t, content, res2)
	}
}

func TestEncodingJSON_Transaction(t *testing.T) {
	for _, c := range readTestsuite(t, "./testsuite/transaction-*.json") {
		content := []byte(compactJSON(string(c.content)))
		txn := new(Transaction)

		// unmarshal
		err := txn.UnmarshalJSON(content)
		assert.NoError(t, err)

		// marshal back
		res2, err := txn.MarshalJSON()
		assert.NoError(t, err)

		assert.Equal(t, content, res2)
	}
}

type testFile struct {
	name    string
	content []byte
}

func readTestsuite(t *testing.T, pattern string) (res []*testFile) {
	files, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no test files found")
	}
	for _, f := range files {
		data, err := ioutil.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		res = append(res, &testFile{
			name:    f,
			content: data,
		})
	}
	return
}
