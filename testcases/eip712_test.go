package testcases

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/umbracle/ethgo/signing"
)

type eip712Testcase struct {
	Name   string
	Domain struct {
		Name              *string
		Version           *string
		VerifyingContract *string
		ChainId           *uint64
		Salt              *string
	}
	Type        string
	Seed        string
	PrimaryType string
	Types       map[string][]*signing.EIP712Type
	Data        map[string]interface{}
	Encoded     string
	Digest      string
}

func (e *eip712Testcase) getDomain() *signing.EIP712Domain {
	d := &signing.EIP712Domain{}

	if name := e.Domain.Name; name != nil {
		d.Name = *name
	}
	if version := e.Domain.Version; version != nil {
		d.Version = *version
	}
	if contract := e.Domain.VerifyingContract; contract != nil {
		d.VerifyingContract = *contract
	}
	if chain := e.Domain.ChainId; chain != nil {
		d.ChainId = new(big.Int).SetUint64(*chain)
	}
	if salt := e.Domain.Salt; salt != nil {
		buf, _ := hex.DecodeString((*salt)[2:])
		d.Salt = buf
	}

	return d
}

func TestEIP712(t *testing.T) {
	var cases []eip712Testcase
	ReadTestCase(t, "eip712", &cases)

	for indx, c := range cases {
		typedData := &signing.EIP712TypedData{
			Types:       c.Types,
			PrimaryType: c.PrimaryType,
			Domain:      c.getDomain(),
			Message:     c.Data,
		}

		digest, err := typedData.Hash()
		require.NoError(t, err)

		if c.Digest != "0x"+hex.EncodeToString(digest) {
			t.Fatalf("wrong digest: %d", indx)
		}
	}
}
