package testcases

import (
	"math/big"
	"testing"

	"github.com/Ethernal-Tech/ethgo"
	"github.com/Ethernal-Tech/ethgo/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getUint64FromBigInt(b *ethgo.ArgBig) (uint64, bool) {
	g := (*big.Int)(b)
	if !g.IsUint64() {
		return 0, false
	}
	return g.Uint64(), true
}

func TestTransactions(t *testing.T) {
	var transactions []struct {
		Name              string         `json:"name"`
		AccountAddress    ethgo.Address  `json:"accountAddress"`
		PrivateKey        ethgo.ArgBytes `json:"privateKey"`
		SignedTransaction ethgo.ArgBytes `json:"signedTransactionChainId5"`

		Data     *ethgo.ArgBytes  `json:"data,omitempty"`
		Value    *ethgo.ArgBig    `json:"value,omitempty"`
		To       *ethgo.Address   `json:"to,omitempty"`
		GasLimit *ethgo.ArgBig    `json:"gasLimit,omitempty"`
		Nonce    *ethgo.ArgUint64 `json:"nonce,omitempty"`
		GasPrice *ethgo.ArgBig    `json:"gasPrice,omitempty"`
	}
	ReadTestCase(t, "transactions", &transactions)

	for _, c := range transactions {
		key, err := wallet.NewWalletFromPrivKey(c.PrivateKey)
		assert.NoError(t, err)
		assert.Equal(t, key.Address(), c.AccountAddress)

		txn := &ethgo.Transaction{
			ChainID: big.NewInt(5),
		}
		if c.Data != nil {
			txn.Input = *c.Data
		}
		if c.Value != nil {
			txn.Value = (*big.Int)(c.Value)
		}
		if c.To != nil {
			txn.To = c.To
		}
		if c.GasLimit != nil {
			gasLimit, isUint64 := getUint64FromBigInt(c.GasLimit)
			if !isUint64 {
				return
			}
			txn.Gas = gasLimit
		}
		if c.Nonce != nil {
			txn.Nonce = c.Nonce.Uint64()
		}
		if c.GasPrice != nil {
			gasPrice, isUint64 := getUint64FromBigInt(c.GasPrice)
			if !isUint64 {
				return
			}
			txn.GasPrice = gasPrice
		}

		signer := wallet.NewEIP155Signer(5)
		signedTxn, err := signer.SignTx(txn, key)
		assert.NoError(t, err)

		txnRaw, err := signedTxn.MarshalRLPTo(nil)
		assert.NoError(t, err)
		assert.Equal(t, txnRaw, c.SignedTransaction.Bytes())

		sender, err := signer.RecoverSender(signedTxn)
		require.NoError(t, err)
		require.Equal(t, sender, key.Address())
	}
}

func TestTypedTransactions(t *testing.T) {
	var transactions []struct {
		Name           string         `json:"name"`
		AccountAddress ethgo.Address  `json:"address"`
		Key            ethgo.ArgBytes `json:"key"`
		Signed         ethgo.ArgBytes `json:"signed"`

		Tx struct {
			Type                 ethgo.TransactionType
			Data                 *ethgo.ArgBytes  `json:"data,omitempty"`
			GasLimit             *ethgo.ArgBig    `json:"gasLimit,omitempty"`
			MaxPriorityFeePerGas *ethgo.ArgBig    `json:"maxPriorityFeePerGas,omitempty"`
			MaxFeePerGas         *ethgo.ArgBig    `json:"maxFeePerGas,omitempty"`
			Nonce                uint64           `json:"nonce,omitempty"`
			To                   *ethgo.Address   `json:"to,omitempty"`
			Value                *ethgo.ArgBig    `json:"value,omitempty"`
			GasPrice             *ethgo.ArgBig    `json:"gasPrice,omitempty"`
			ChainID              uint64           `json:"chainId,omitempty"`
			AccessList           ethgo.AccessList `json:"accessList,omitempty"`
		}
	}
	ReadTestCase(t, "typed-transactions", &transactions)

	for _, c := range transactions {
		key, err := wallet.NewWalletFromPrivKey(c.Key)
		assert.NoError(t, err)
		assert.Equal(t, key.Address(), c.AccountAddress)

		chainID := big.NewInt(int64(c.Tx.ChainID))

		txn := &ethgo.Transaction{
			ChainID:              chainID,
			Type:                 c.Tx.Type,
			MaxPriorityFeePerGas: (*big.Int)(c.Tx.MaxPriorityFeePerGas),
			MaxFeePerGas:         (*big.Int)(c.Tx.MaxFeePerGas),
			AccessList:           c.Tx.AccessList,
		}
		if c.Tx.Data != nil {
			txn.Input = *c.Tx.Data
		}
		if c.Tx.Value != nil {
			txn.Value = (*big.Int)(c.Tx.Value)
		}
		if c.Tx.To != nil {
			txn.To = c.Tx.To
		}
		if c.Tx.GasLimit != nil {
			gasLimit, isUint64 := getUint64FromBigInt(c.Tx.GasLimit)
			if !isUint64 {
				return
			}
			txn.Gas = gasLimit
		}
		txn.Nonce = c.Tx.Nonce
		if c.Tx.GasPrice != nil {
			gasPrice, isUint64 := getUint64FromBigInt(c.Tx.GasPrice)
			if !isUint64 {
				return
			}
			txn.GasPrice = gasPrice
		}

		signer := wallet.NewEIP155Signer(chainID.Uint64())
		signedTxn, err := signer.SignTx(txn, key)
		assert.NoError(t, err)

		txnRaw, err := signedTxn.MarshalRLPTo(nil)
		assert.NoError(t, err)

		assert.Equal(t, txnRaw, c.Signed.Bytes())

		sender, err := signer.RecoverSender(signedTxn)
		require.NoError(t, err)
		require.Equal(t, sender, key.Address())
	}
}
