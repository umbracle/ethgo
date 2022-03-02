package ethgo

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/umbracle/fastrlp"
)

func TestEncodingRLP_Transaction_Fuzz(t *testing.T) {
	testTransaction := func(t *testing.T, typ TransactionType) {
		obj := &Transaction{}
		err := fastrlp.Fuzz(100, obj,
			fastrlp.WithDefaults(func(obj fastrlp.FuzzObject) {
				obj.(*Transaction).Type = typ
			}),
			fastrlp.WithPostHook(func(obj fastrlp.FuzzObject) error {
				// Test that the hash from unmarshal is the same as the one computed
				txn := obj.(*Transaction)
				cHash, err := txn.GetHash()
				if err != nil {
					return err
				}
				if cHash != txn.Hash {
					return fmt.Errorf("hash not equal")
				}
				return nil
			}),
		)
		assert.NoError(t, err)
	}

	t.Run("legacy", func(t *testing.T) {
		testTransaction(t, TransactionLegacy)
	})
	t.Run("accesslist", func(t *testing.T) {
		testTransaction(t, TransactionAccessList)
	})
	t.Run("dynamicfee", func(t *testing.T) {
		testTransaction(t, TransactionDynamicFee)
	})
}

func TestEncodingRLP_AccessList_Fuzz(t *testing.T) {
	obj := &AccessList{}
	if err := fastrlp.Fuzz(100, obj); err != nil {
		t.Fatal(err)
	}
}
