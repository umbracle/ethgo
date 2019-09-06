package web3

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/valyala/fastjson"
)

var blockPool fastjson.ParserPool

func (b *Block) UnmarshalJSON(buf []byte) error {
	p := blockPool.Get()
	defer blockPool.Put(p)

	v, err := p.Parse(string(buf))
	if err != nil {
		return nil
	}
	o, err := v.Object()
	if err != nil {
		return fmt.Errorf("object not found")
	}

	// hash objects
	b.Hash, err = decodeString(o, "hash")
	if err != nil {
		return err
	}
	b.StateRoot, err = decodeHash(o, "stateRoot")
	if err != nil {
		return err
	}
	b.Sha3Uncles, err = decodeHash(o, "sha3Uncles")
	if err != nil {
		return err
	}
	b.ReceiptsRoot, err = decodeHash(o, "receiptsRoot")
	if err != nil {
		return err
	}
	b.TransactionsRoot, err = decodeHash(o, "transactionsRoot")
	if err != nil {
		return err
	}
	b.ParentHash, err = decodeHash(o, "parentHash")
	if err != nil {
		return err
	}

	// address object
	b.Miner, err = decodeAddr(o, "miner")
	if err != nil {
		return err
	}

	// uint64 objects
	b.Number, err = decodeUint64(o, "number")
	if err != nil {
		return err
	}
	b.GasLimit, err = decodeUint64(o, "gasLimit")
	if err != nil {
		return err
	}
	b.GasUsed, err = decodeUint64(o, "gasUsed")
	if err != nil {
		return err
	}
	b.Timestamp, err = decodeUint64(o, "timestamp")
	if err != nil {
		return err
	}

	return nil
}

var txnArena fastjson.ArenaPool

// MarshalJSON implements the Marshal interface
func (t *Transaction) MarshalJSON() ([]byte, error) {
	a := txnArena.Get()
	defer txnArena.Put(a)

	o := a.NewObject()
	o.Set("from", a.NewString(t.From))
	if t.To != "" {
		o.Set("to", a.NewString(t.To))
	}
	if t.Input != "" {
		o.Set("input", a.NewString(t.Input))
	}
	if t.Data != "" {
		o.Set("data", a.NewString(t.Data))
	}

	o.Set("gasPrice", a.NewString(fmt.Sprintf("0x%x", t.GasPrice)))
	o.Set("gas", a.NewString(fmt.Sprintf("0x%x", t.Gas)))

	if t.Value != nil {
		o.Set("value", a.NewString(fmt.Sprintf("0x%x", t.Value)))
	}
	res := o.MarshalTo(nil)
	return res, nil
}

func decodeAddr(o *fastjson.Object, n string) ([]byte, error) {
	return decodeBytes(o, n, 20)
}

func decodeHash(o *fastjson.Object, n string) ([]byte, error) {
	return decodeBytes(o, n, 32)
}

func decodeString(o *fastjson.Object, n string) (string, error) {
	m := o.Get(n)
	if m == nil {
		return "", fmt.Errorf("field %s does not exists", n)
	}
	str := m.String()
	str = strings.Trim(str, "\"")
	return str, nil
}

func decodeBytes(o *fastjson.Object, n string, bits int) ([]byte, error) {
	m := o.Get(n)
	if m == nil {
		return nil, fmt.Errorf("field %s does not exists", n)
	}

	str := m.String()
	str = strings.Trim(str, "\"")

	if !strings.HasPrefix(str, "0x") {
		return nil, fmt.Errorf("field %s does not have 0x prefix", n)
	}

	buf, err := hex.DecodeString(str[2:])
	if err != nil {
		return nil, err
	}
	if len(buf) != bits {
		return nil, fmt.Errorf("field %s invalid length, expected %d but found %d", n, bits, len(buf))
	}
	return buf, nil
}

func decodeUint64(o *fastjson.Object, n string) (uint64, error) {
	m := o.Get(n)
	if m == nil {
		return 0, fmt.Errorf("field %s does not exists", n)
	}

	str := m.String()
	str = strings.Trim(str, "\"")

	if !strings.HasPrefix(str, "0x") {
		return 0, fmt.Errorf("field %s does not have 0x prefix", n)
	}
	return strconv.ParseUint(str[2:], 16, 64)
}
