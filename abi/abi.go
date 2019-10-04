package abi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"strings"
	"sync"

	"github.com/umbracle/go-web3"
	"golang.org/x/crypto/sha3"
)

// ABI represents the ethereum abi format
type ABI struct {
	Constructor *Method
	Methods     map[string]*Method
	Events      map[string]*Event
}

// NewABI returns a parsed ABI struct
func NewABI(s string) (*ABI, error) {
	return NewABIFromReader(bytes.NewReader([]byte(s)))
}

// NewABIFromReader returns an ABI object from a reader
func NewABIFromReader(r io.Reader) (*ABI, error) {
	var abi *ABI
	dec := json.NewDecoder(r)
	if err := dec.Decode(&abi); err != nil {
		return nil, err
	}
	return abi, nil
}

// UnmarshalJSON implements json.Unmarshaler interface
func (a *ABI) UnmarshalJSON(data []byte) error {
	var fields []struct {
		Type      string
		Name      string
		Constant  bool
		Anonymous bool
		Inputs    Arguments
		Outputs   Arguments
	}

	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	a.Methods = make(map[string]*Method)
	a.Events = make(map[string]*Event)

	for _, field := range fields {
		switch field.Type {
		case "constructor":
			if a.Constructor != nil {
				return fmt.Errorf("multiple constructor declaration")
			}
			a.Constructor = &Method{
				Inputs: field.Inputs,
			}

		case "function", "":
			a.Methods[field.Name] = &Method{
				Name:    field.Name,
				Const:   field.Constant,
				Inputs:  field.Inputs,
				Outputs: field.Outputs,
			}

		case "event":
			a.Events[field.Name] = &Event{
				Name:      field.Name,
				Anonymous: field.Anonymous,
				Inputs:    field.Inputs,
			}

		default:
			return fmt.Errorf("unknown field type '%s'", field.Type)
		}
	}
	return nil
}

// Method is a callable function in the contract
type Method struct {
	Name    string
	Const   bool
	Inputs  Arguments
	Outputs Arguments
}

// Sig returns the signature of the method
func (m *Method) Sig() string {
	return buildSignature(m.Name, m.Inputs)
}

// ID returns the id of the method
func (m *Method) ID() []byte {
	k := acquireKeccak()
	k.Write([]byte(m.Sig()))
	dst := k.Sum(nil)[:4]
	releaseKeccak(k)
	return dst
}

// Event is a triggered log mechanism
type Event struct {
	Name      string
	Anonymous bool
	Inputs    Arguments
}

// Sig returns the signature of the event
func (e *Event) Sig() string {
	return buildSignature(e.Name, e.Inputs)
}

// ID returns the id of the event used during logs
func (e *Event) ID() (res web3.Hash) {
	k := acquireKeccak()
	k.Write([]byte(e.Sig()))
	dst := k.Sum(nil)
	releaseKeccak(k)
	copy(res[:], dst)
	return
}

func buildSignature(name string, args Arguments) string {
	types := make([]string, len(args))
	for i, input := range args {
		types[i] = input.Type.raw
	}
	return fmt.Sprintf("%v(%v)", name, strings.Join(types, ","))
}

// Argument is a solidity argument for functions and events
type Argument struct {
	Name    string
	Type    *Type
	Indexed bool
}

// Arguments is a list of arguments
type Arguments []*Argument

// Type returns the type of the argument in tuple form
func (a *Arguments) Type() *Type {
	inputs := []*TupleElem{}
	for _, i := range *a {
		inputs = append(inputs, &TupleElem{
			Name: i.Name,
			Elem: i.Type,
		})
	}

	tt := &Type{
		kind:  KindTuple,
		raw:   "tuple",
		tuple: inputs,
	}
	return tt
}

// UnmarshalJSON implements the unmarshal interface
func (a *Argument) UnmarshalJSON(data []byte) error {
	var arg *ArgumentStr
	if err := json.Unmarshal(data, &arg); err != nil {
		return fmt.Errorf("argument json err: %v", err)
	}

	t, err := NewTypeFromArgument(arg)
	if err != nil {
		return err
	}

	a.Type = t
	a.Name = arg.Name
	a.Indexed = arg.Indexed
	return nil
}

// ArgumentStr encodes a type object
type ArgumentStr struct {
	Name       string
	Type       string
	Indexed    bool
	Components []*ArgumentStr
}

var keccakPool = sync.Pool{
	New: func() interface{} {
		return sha3.NewLegacyKeccak256()
	},
}

func acquireKeccak() hash.Hash {
	return keccakPool.Get().(hash.Hash)
}

func releaseKeccak(k hash.Hash) {
	k.Reset()
	keccakPool.Put(k)
}
