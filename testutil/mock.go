package testutil

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"sync"

	"github.com/umbracle/ethgo"
)

type mockCall int

const (
	blockByNumberCall mockCall = iota
	blockByHashCall
	blockNumberCall
	getLogsCall
)

type MockClient struct {
	lock     sync.Mutex
	num      uint64
	blockNum map[uint64]ethgo.Hash
	blocks   map[ethgo.Hash]*ethgo.Block
	logs     map[ethgo.Hash][]*ethgo.Log
	chainID  *big.Int
}

func (m *MockClient) SetChainID(id *big.Int) {
	m.chainID = id
}

func (d *MockClient) ChainID() (*big.Int, error) {
	if d.chainID == nil {
		d.chainID = big.NewInt(1337)
	}
	return d.chainID, nil
}

func (d *MockClient) GetLastBlocks(n uint64) (res []*ethgo.Block) {
	if d.num == 0 {
		return
	}
	num := n
	if d.num < num {
		num = d.num + 1
	}

	for i := int(num - 1); i >= 0; i-- {
		res = append(res, d.blocks[d.blockNum[d.num-uint64(i)]])
	}
	return
}

func (d *MockClient) GetAllLogs() (res []*ethgo.Log) {
	if d.num == 0 {
		return
	}
	for i := uint64(0); i <= d.num; i++ {
		res = append(res, d.logs[d.blocks[d.blockNum[i]].Hash]...)
	}
	return
}

func (d *MockClient) AddScenario(m MockList) {
	d.lock.Lock()
	defer d.lock.Unlock()

	// add the logs
	for _, b := range m {
		block := &ethgo.Block{
			Hash:   b.Hash(),
			Number: uint64(b.num),
		}

		if b.num != 0 {
			bb, err := d.blockByNumberLock(uint64(b.num) - 1)
			if err != nil {
				// This happens during reconcile tests because we include only partial blocks
				block.ParentHash = encodeHash(strconv.Itoa(b.num - 1))
			} else {
				block.ParentHash = bb.Hash
			}
		}

		// add history block
		d.addBlocks(block)

		// add logs
		// remove any other logs for this block in case there are any
		delete(d.logs, block.Hash)

		d.AddLogs(b.GetLogs())
	}
}

func (d *MockClient) AddLogs(logs []*ethgo.Log) {
	if d.logs == nil {
		d.logs = map[ethgo.Hash][]*ethgo.Log{}
	}
	for _, log := range logs {
		entry, ok := d.logs[log.BlockHash]
		if ok {
			entry = append(entry, log)
		} else {
			entry = []*ethgo.Log{log}
		}
		d.logs[log.BlockHash] = entry
	}
}

func (d *MockClient) addBlocks(bb ...*ethgo.Block) {
	if d.blocks == nil {
		d.blocks = map[ethgo.Hash]*ethgo.Block{}
	}
	if d.blockNum == nil {
		d.blockNum = map[uint64]ethgo.Hash{}
	}
	for _, b := range bb {
		if b.Number > d.num {
			d.num = b.Number
		}
		d.blocks[b.Hash] = b
		d.blockNum[b.Number] = b.Hash
	}
}

func (d *MockClient) BlockNumber() (uint64, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	return d.num, nil
}

func (d *MockClient) GetBlockByHash(hash ethgo.Hash, full bool) (*ethgo.Block, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	b := d.blocks[hash]
	if b == nil {
		return nil, fmt.Errorf("hash %s not found", hash)
	}
	return b, nil
}

func (d *MockClient) blockByNumberLock(i uint64) (*ethgo.Block, error) {
	hash, ok := d.blockNum[i]
	if !ok {
		return nil, fmt.Errorf("number %d not found", i)
	}
	return d.blocks[hash], nil
}

func (d *MockClient) GetBlockByNumber(i ethgo.BlockNumber, full bool) (*ethgo.Block, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if i < 0 {
		switch i {
		case ethgo.Latest:
			if d.num == 0 {
				return &ethgo.Block{Number: 0}, nil
			}
			return d.blockByNumberLock(d.num)
		default:
			return nil, fmt.Errorf("getBlockByNumber query not supported")
		}
	}
	return d.blockByNumberLock(uint64(i))
}

func (d *MockClient) GetLogs(filter *ethgo.LogFilter) ([]*ethgo.Log, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if filter.BlockHash != nil {
		return d.logs[*filter.BlockHash], nil
	}

	from, to := uint64(*filter.From), uint64(*filter.To)
	if from > to {
		return nil, fmt.Errorf("from higher than to")
	}
	if int(to) > len(d.blocks) {
		return nil, fmt.Errorf("out of bounds")
	}

	logs := []*ethgo.Log{}
	for i := from; i <= to; i++ {
		b, err := d.blockByNumberLock(i)
		if err != nil {
			return nil, err
		}
		elems, ok := d.logs[b.Hash]
		if ok {
			logs = append(logs, elems...)
		}
	}
	return logs, nil
}

type MockLog struct {
	data string
}

type MockBlock struct {
	hash   string
	extra  string
	parent string
	num    int
	logs   []*MockLog
}

func mustDecodeHash(str string) []byte {
	str = strings.TrimPrefix(str, "0x")
	if len(str)%2 == 1 {
		str = str + "0"
	}
	buf, err := hex.DecodeString(str)
	if err != nil {
		panic(err)
	}
	return buf
}

func (m *MockBlock) Extra(data string) *MockBlock {
	m.extra = data
	return m
}

func (m *MockBlock) GetLogs() (logs []*ethgo.Log) {
	for _, log := range m.logs {
		logs = append(logs, &ethgo.Log{Data: mustDecodeHash(log.data), BlockNumber: uint64(m.num), BlockHash: m.Hash()})
	}
	return
}

func (m *MockBlock) Log(data string) *MockBlock {
	m.logs = append(m.logs, &MockLog{data})
	return m
}

func (m *MockBlock) GetNum() int {
	return m.num
}

func (m *MockBlock) Num(i int) *MockBlock {
	m.num = i
	return m
}

func (m *MockBlock) Parent(i int) *MockBlock {
	m.parent = strconv.Itoa(i)
	m.num = i + 1
	return m
}

func encodeHash(str string) (h ethgo.Hash) {
	tmp := ""
	for i := 0; i < 64-len(str); i++ {
		tmp += "0"
	}
	str = "0x" + tmp + str
	if err := h.UnmarshalText([]byte(str)); err != nil {
		panic(err)
	}
	return
}

func (m *MockBlock) Hash() ethgo.Hash {
	return encodeHash(m.extra + m.hash)
}

func (m *MockBlock) Block() *ethgo.Block {
	b := &ethgo.Block{
		Hash:   m.Hash(),
		Number: uint64(m.num),
	}
	if m.num != 0 {
		b.ParentHash = encodeHash(m.parent)
	}
	return b
}

func Mock(number int) *MockBlock {
	return &MockBlock{hash: strconv.Itoa(number), num: number, parent: strconv.Itoa(number - 1)}
}

type MockList []*MockBlock

func (m *MockList) Create(from, to int, callback func(b *MockBlock)) {
	for i := from; i < to; i++ {
		b := Mock(i)
		callback(b)
		*m = append(*m, b)
	}
}

func (m *MockList) GetLogs() (res []*ethgo.Log) {
	for _, log := range *m {
		res = append(res, log.GetLogs()...)
	}
	return
}

func (m *MockList) ToBlocks() []*ethgo.Block {
	e := []*ethgo.Block{}
	for _, i := range *m {
		e = append(e, i.Block())
	}
	return e
}
