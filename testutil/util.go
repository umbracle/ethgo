package testutil

import (
	"math/big"
	"reflect"

	"github.com/umbracle/go-web3"
)

func CompareLogs(one, two []*web3.Log) bool {
	if len(one) != len(two) {
		return false
	}
	if len(one) == 0 {
		return true
	}
	return reflect.DeepEqual(one, two)
}

func CompareBlocks(one, two []*web3.Block) bool {
	if len(one) != len(two) {
		return false
	}
	if len(one) == 0 {
		return true
	}
	// difficulty is hard to check, set the values to zero
	for _, i := range one {
		i.Difficulty = big.NewInt(0)
	}
	for _, i := range two {
		i.Difficulty = big.NewInt(0)
	}
	return reflect.DeepEqual(one, two)
}
