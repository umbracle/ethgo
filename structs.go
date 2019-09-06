package web3

import "math/big"

type Block struct {
	Number           uint64
	Hash             string
	ParentHash       []byte
	Sha3Uncles       []byte
	TransactionsRoot []byte
	StateRoot        []byte
	ReceiptsRoot     []byte
	Miner            []byte
	Difficulty       *big.Int
	TotalDifficulty  *big.Int
	ExtraData        []byte
	GasLimit         uint64
	GasUsed          uint64
	Timestamp        uint64
	Transactions     []*Transaction
	Uncles           [][]byte
}

type Transaction struct {
	From     string
	To       string
	Data     string
	Input    string
	GasPrice uint64
	Gas      uint64
	Value    *big.Int
}

type Receipt struct {
	TransactionHash string `json:"transactionHash"`
	ContractAddress string `json:"contractAddress"`
}
