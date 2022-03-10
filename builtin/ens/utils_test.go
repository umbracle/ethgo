package ens

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNameHash(t *testing.T) {
	cases := []struct {
		Name     string
		Expected string
	}{
		{
			Name:     "eth",
			Expected: "0x93cdeb708b7545dc668eb9280176169d1c33cfd8ed6f04690a0bcc88a93fc4ae",
		},
		{
			Name:     "foo.eth",
			Expected: "0xde9b09fd7c5f901e23a3f19fecc54828e9c848539801e86591bd9801b019f84f",
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			found := NameHash(c.Name)
			assert.Equal(t, c.Expected, found.String())
		})
	}
}
