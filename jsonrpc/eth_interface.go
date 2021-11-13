package jsonrpc

import (
	"math/big"

	"github.com/panyanyany/go-web3"
)

type IEth interface {
	GetCode(addr web3.Address, block web3.BlockNumberOrHash) (string, error)
	Accounts() ([]web3.Address, error)
	GetStorageAt(addr web3.Address, slot web3.Hash, block web3.BlockNumberOrHash) (web3.Hash, error)
	BlockNumber() (uint64, error)
	GetBlockByNumber(i web3.BlockNumber, full bool) (*web3.Block, error)
	GetBlockByHash(hash web3.Hash, full bool) (*web3.Block, error)
	GetFilterChanges(id string) ([]*web3.Log, error)
	GetTransactionByHash(hash web3.Hash) (*web3.Transaction, error)
	GetFilterChangesBlock(id string) ([]web3.Hash, error)
	NewFilter(filter *web3.LogFilter) (string, error)
	NewBlockFilter() (string, error)
	UninstallFilter(id string) (bool, error)
	SendRawTransaction(data []byte) (web3.Hash, error)
	SendTransaction(txn *web3.Transaction) (web3.Hash, error)
	GetTransactionReceipt(hash web3.Hash) (*web3.Receipt, error)
	GetNonce(addr web3.Address, blockNumber web3.BlockNumberOrHash) (uint64, error)
	GetBalance(addr web3.Address, blockNumber web3.BlockNumberOrHash) (*big.Int, error)
	GasPrice() (uint64, error)
	Call(msg *web3.CallMsg, block web3.BlockNumber) (string, error)
	EstimateGasContract(bin []byte) (uint64, error)
	EstimateGas(msg *web3.CallMsg) (uint64, error)
	GetLogs(filter *web3.LogFilter) ([]*web3.Log, error)
	ChainID() (*big.Int, error)
}
