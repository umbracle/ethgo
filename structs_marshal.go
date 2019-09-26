package web3

import (
	"encoding/hex"
	"fmt"

	"github.com/valyala/fastjson"
)

var defaultArena fastjson.ArenaPool

// MarshalJSON implements the Marshal interface.
func (t *Transaction) MarshalJSON() ([]byte, error) {
	a := defaultArena.Get()

	o := a.NewObject()
	o.Set("from", a.NewString(t.From.String()))
	if t.To != "" {
		o.Set("to", a.NewString(t.To))
	}
	if len(t.Input) != 0 {
		o.Set("input", a.NewString("0x"+hex.EncodeToString(t.Input)))
	}
	o.Set("gasPrice", a.NewString(fmt.Sprintf("0x%x", t.GasPrice)))
	o.Set("gas", a.NewString(fmt.Sprintf("0x%x", t.Gas)))
	if t.Value != nil {
		o.Set("value", a.NewString(fmt.Sprintf("0x%x", t.Value)))
	}

	res := o.MarshalTo(nil)
	defaultArena.Put(a)
	return res, nil
}

// MarshalJSON implements the Marshal interface.
func (c *CallMsg) MarshalJSON() ([]byte, error) {
	a := defaultArena.Get()
	defer defaultArena.Put(a)

	o := a.NewObject()
	o.Set("from", a.NewString(c.From.String()))
	o.Set("to", a.NewString(c.To.String()))
	if len(c.Data) != 0 {
		o.Set("data", a.NewString("0x"+hex.EncodeToString(c.Data)))
	}
	if c.GasPrice != 0 {
		o.Set("gasPrice", a.NewString(fmt.Sprintf("0x%x", c.GasPrice)))
	}

	res := o.MarshalTo(nil)
	defaultArena.Put(a)
	return res, nil
}
