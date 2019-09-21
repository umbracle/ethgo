package ens

import (
	"strings"

	"golang.org/x/crypto/sha3"
)

// NameHash returns the hash of an ENS name
func NameHash(str string) []byte {
	node := make([]byte, 32)
	if str == "" {
		return node
	}

	hash := sha3.NewLegacyKeccak256()
	
	labels := strings.Split(str, ".")
	for i := len(labels) - 1; i >= 0; i-- {
		label := labels[i]

		hash.Write([]byte(label))
		node = hash.Sum(node) // append the hash of the label to node
		hash.Reset()

		hash.Write(node)
		node = hash.Sum(node[:0])
		hash.Reset()
	}

	return node
}
