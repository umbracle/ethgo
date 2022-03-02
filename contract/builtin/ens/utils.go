package ens

import (
	"strings"

	web3 "github.com/umbracle/ethgo"
	"golang.org/x/crypto/sha3"
)

// NameHash returns the hash of an ENS name
func NameHash(str string) (node web3.Hash) {
	if str == "" {
		return
	}

	aux := make([]byte, 32)
	hash := sha3.NewLegacyKeccak256()

	labels := strings.Split(str, ".")
	for i := len(labels) - 1; i >= 0; i-- {
		label := labels[i]

		hash.Write([]byte(label))
		aux = hash.Sum(aux) // append the hash of the label to node
		hash.Reset()

		hash.Write(aux)
		aux = hash.Sum(aux[:0])
		hash.Reset()
	}

	copy(node[:], aux)
	return
}
