package compiler

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	solcDir  = "/tmp/ethgo-solc"
	solcPath = solcDir + "/solidity"
)

func init() {
	_, err := os.Stat(solcDir)
	if err == nil {
		// already exists
		return
	}
	if !os.IsNotExist(err) {
		panic(err)
	}
	// solc folder does not exists
	if err := DownloadSolidity("0.5.5", solcDir, false); err != nil {
		panic(err)
	}
}

func TestSolidityInline(t *testing.T) {
	solc := NewSolidityCompiler(solcPath)

	cases := []struct {
		code      string
		contracts []string
	}{
		{
			`
		pragma solidity >0.0.0;
		contract foo{}
			`,
			[]string{
				"foo",
			},
		},
		{
			`
		pragma solidity >0.0.0;
		contract foo{}
		contract bar{}
			`,
			[]string{
				"bar",
				"foo",
			},
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			output, err := solc.CompileCode(c.code)
			if err != nil {
				t.Fatal(err)
			}

			result := map[string]struct{}{}
			for i := range output.Contracts {
				result[strings.TrimPrefix(i, "<stdin>:")] = struct{}{}
			}

			// only one source file
			assert.Len(t, output.Sources, 1)

			expected := map[string]struct{}{}
			for _, i := range c.contracts {
				expected[i] = struct{}{}
			}

			if !reflect.DeepEqual(result, expected) {
				t.Fatal("bad")
			}
		})
	}
}

func TestSolidity(t *testing.T) {
	solc := NewSolidityCompiler(solcPath)

	files := []string{
		"./fixtures/ballot.sol",
		"./fixtures/simple_auction.sol",
	}
	output, err := solc.Compile(files...)
	if err != nil {
		t.Fatal(err)
	}
	if len(output.Contracts) != 2 {
		t.Fatal("two expected")
	}
}

func existsSolidity(t *testing.T, path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		t.Fatal(err)
	}

	cmd := exec.Command(path, "--version")
	var stderr, stdout bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("solidity version failed: %s", stderr.String())
	}
	if len(stdout.Bytes()) == 0 {
		t.Fatal("empty output")
	}
	return true
}

func TestDownloadSolidityCompiler(t *testing.T) {
	dst1, err := ioutil.TempDir("/tmp", "ethgo-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dst1)

	if err := DownloadSolidity("0.5.5", dst1, true); err != nil {
		t.Fatal(err)
	}
	if existsSolidity(t, filepath.Join(dst1, "solidity")) {
		t.Fatal("it should not exist")
	}
	if !existsSolidity(t, filepath.Join(dst1, "solidity-0.5.5")) {
		t.Fatal("it should exist")
	}

	dst2, err := ioutil.TempDir("/tmp", "ethgo-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dst2)

	if err := DownloadSolidity("0.5.5", dst2, false); err != nil {
		t.Fatal(err)
	}
	if !existsSolidity(t, filepath.Join(dst2, "solidity")) {
		t.Fatal("it should exist")
	}
	if existsSolidity(t, filepath.Join(dst2, "solidity-0.5.5")) {
		t.Fatal("it should not exist")
	}
}
