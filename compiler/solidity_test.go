package compiler

import (
	"reflect"
	"strings"
	"testing"
)

func TestSolidityCompiler(t *testing.T) {
	solc := NewSolidityCompiler("solc")

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
			data, err := solc.Compile(c.code)
			if err != nil {
				t.Fatal(err)
			}

			output := data.(*SolcOutput)
			result := map[string]struct{}{}
			for i := range output.Contracts {
				result[strings.TrimPrefix(i, "<stdin>:")] = struct{}{}
			}

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
