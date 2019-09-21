package web3

import (
	"fmt"
	"math/big"
)

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
	BlockHash       string `json:"blockHash"`
}

type CallMsg struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Data     string `json:"data"`
	GasPrice string `json:"gasPrice"`
}

type LogFilter struct {
	Address   []string `json:"address"`
	Topics    []string `json:"topics"`
	BlockHash string   `json:"blockhash"`
}

type Log struct {
	Address string   `json:"address"`
	Topics  []string `json:"topics"`
	Data    string   `json:"data"`
}

type BlockNumber int

const (
	Latest   BlockNumber = -1
	Earliest             = -2
	Pending              = -3
)

func (b BlockNumber) String() string {
	switch b {
	case Latest:
		return "latest"
	case Earliest:
		return "earliest"
	case Pending:
		return "pending"
	}
	if b < 0 {
		panic("internal. blocknumber is negative")
	}
	return fmt.Sprintf("0x%x", uint64(b))
}

func EncodeBlock(block ...BlockNumber) BlockNumber {
	if len(block) != 1 {
		return Latest
	}
	return block[0]
}
