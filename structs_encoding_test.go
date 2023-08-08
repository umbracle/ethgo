package ethgo

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func compactJSON(s string) string {
	buffer := new(bytes.Buffer)
	if err := json.Compact(buffer, []byte(s)); err != nil {
		panic(err)
	}
	return buffer.String()
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
		data, err := os.ReadFile(f)
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
