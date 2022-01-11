package abi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hash"
	"io"
	"regexp"
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
	Errors      map[string]*Error
}

func (a *ABI) GetMethod(name string) *Method {
	m := a.Methods[name]
	return m
}

func (a *ABI) addError(e *Error) {
	if len(a.Errors) == 0 {
		a.Errors = map[string]*Error{}
	}
	a.Errors[e.Name] = e
}

func (a *ABI) addEvent(e *Event) {
	if len(a.Events) == 0 {
		a.Events = map[string]*Event{}
	}
	a.Events[e.Name] = e
}

func (a *ABI) addMethod(m *Method) {
	if len(a.Methods) == 0 {
		a.Methods = map[string]*Method{}
	}
	a.Methods[m.Name] = m
}

// NewABI returns a parsed ABI struct
func NewABI(s string) (*ABI, error) {
	return NewABIFromReader(bytes.NewReader([]byte(s)))
}

// MustNewABI returns a parsed ABI contract or panics if fails
func MustNewABI(s string) *ABI {
	a, err := NewABI(s)
	if err != nil {
		panic(err)
	}
	return a
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
		Type            string
		Name            string
		Constant        bool
		Anonymous       bool
		StateMutability string
		Inputs          []*ArgumentStr
		Outputs         []*ArgumentStr
	}

	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}

	a.Methods = make(map[string]*Method)
	a.Events = make(map[string]*Event)
	a.Errors = make(map[string]*Error)

	for _, field := range fields {
		switch field.Type {
		case "constructor":
			if a.Constructor != nil {
				return fmt.Errorf("multiple constructor declaration")
			}
			/*
				a.Constructor = &Method{
					Inputs: field.Inputs.Type(),
				}
			*/

		case "function", "":
			c := field.Constant
			if field.StateMutability == "view" || field.StateMutability == "pure" {
				c = true
			}

			inputs, err := NewTupleTypeFromArgs(field.Inputs)
			if err != nil {
				panic(err)
			}
			outputs, err := NewTupleTypeFromArgs(field.Outputs)
			if err != nil {
				panic(err)
			}
			a.Methods[field.Name] = &Method{
				Name:    field.Name,
				Const:   c,
				Inputs:  inputs,
				Outputs: outputs,
			}

		case "event":
			input, err := NewTupleTypeFromArgs(field.Inputs)
			if err != nil {
				panic(err)
			}
			a.Events[field.Name] = &Event{
				Name:      field.Name,
				Anonymous: field.Anonymous,
				Inputs:    input,
			}

		case "error":
			input, err := NewTupleTypeFromArgs(field.Inputs)
			if err != nil {
				panic(err)
			}
			a.Errors[field.Name] = &Error{
				Name:   field.Name,
				Inputs: input,
			}

		case "fallback":
		case "receive":
			// do nothing

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
	Inputs  *Type
	Outputs *Type
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

// Encode encodes the inputs with this function
func (m *Method) Encode(args interface{}) ([]byte, error) {
	data, err := Encode(args, m.Inputs)
	if err != nil {
		return nil, err
	}
	data = append(m.ID(), data...)
	return data, nil
}

// Decode decodes the output with this function
func (m *Method) Decode(data []byte) (map[string]interface{}, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty response")
	}
	respInterface, err := Decode(m.Outputs, data)
	if err != nil {
		return nil, err
	}
	resp := respInterface.(map[string]interface{})
	return resp, nil
}

func NewMethod(name string) (*Method, error) {
	name, inputs, outputs, err := parseMethodSignature(name)
	if err != nil {
		return nil, err
	}
	m := &Method{Name: name, Inputs: inputs, Outputs: outputs}
	return m, nil
}

var (
	funcRegexpWithReturn    = regexp.MustCompile(`(.*)\((.*)\)(.*) returns \((.*)\)`)
	funcRegexpWithoutReturn = regexp.MustCompile(`(.*)\((.*)\)(.*)`)
)

func parseMethodSignature(name string) (string, *Type, *Type, error) {
	name = strings.TrimPrefix(name, "function ")
	name = strings.TrimSpace(name)

	var funcName, inputArgs, outputArgs string

	if strings.Contains(name, "returns") {
		matches := funcRegexpWithReturn.FindAllStringSubmatch(name, -1)
		if len(matches) == 0 {
			return "", nil, nil, fmt.Errorf("no matches found")
		}
		funcName = strings.TrimSpace(matches[0][1])
		inputArgs = strings.TrimSpace(matches[0][2])
		outputArgs = strings.TrimSpace(matches[0][4])
	} else {
		matches := funcRegexpWithoutReturn.FindAllStringSubmatch(name, -1)
		if len(matches) == 0 {
			return "", nil, nil, fmt.Errorf("no matches found")
		}
		funcName = strings.TrimSpace(matches[0][1])
		inputArgs = strings.TrimSpace(matches[0][2])
	}

	input, err := NewType("tuple(" + inputArgs + ")")
	if err != nil {
		return "", nil, nil, err
	}
	output, err := NewType("tuple(" + outputArgs + ")")
	if err != nil {
		return "", nil, nil, err
	}
	return funcName, input, output, nil
}

// Event is a triggered log mechanism
type Event struct {
	Name      string
	Anonymous bool
	Inputs    *Type
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

// MustNewEvent creates a new solidity event object or fails
func MustNewEvent(name string) *Event {
	evnt, err := NewEvent(name)
	if err != nil {
		panic(err)
	}
	return evnt
}

// NewEvent creates a new solidity event object using the signature
func NewEvent(name string) (*Event, error) {
	name, typ, err := parseEventOrErrorSignature(name)
	if err != nil {
		return nil, err
	}
	return NewEventFromType(name, typ), nil
}

// Error is a solidity error
type Error struct {
	Name   string
	Inputs *Type
}

// NewError creates a new solidity error object using the signature
func NewError(name string) (*Error, error) {
	name, typ, err := parseEventOrErrorSignature(name)
	if err != nil {
		return nil, err
	}
	return &Error{Name: name, Inputs: typ}, nil
}

func parseEventOrErrorSignature(name string) (string, *Type, error) {
	// the prefix can be either for an 'event' or an 'error'
	prefix := []string{
		"event ",
		"error ",
	}
	found := false
	for _, p := range prefix {
		if strings.HasPrefix(name, p) {
			name = strings.TrimPrefix(name, p)
			found = true
		}
	}
	if !found {
		return "", nil, fmt.Errorf("signature is neither for 'error' nor 'event'")
	}

	if !strings.HasSuffix(name, ")") {
		return "", nil, fmt.Errorf("failed to parse input, expected 'name(types)'")
	}
	indx := strings.Index(name, "(")
	if indx == -1 {
		return "", nil, fmt.Errorf("failed to parse input, expected 'name(types)'")
	}

	funcName, signature := name[:indx], name[indx:]
	signature = "tuple" + signature

	typ, err := NewType(signature)
	if err != nil {
		return "", nil, err
	}
	return funcName, typ, nil
}

// NewEventFromType creates a new solidity event object using the name and type
func NewEventFromType(name string, typ *Type) *Event {
	return &Event{Name: name, Inputs: typ}
}

// Match checks wheter the log is from this event
func (e *Event) Match(log *web3.Log) bool {
	if len(log.Topics) == 0 {
		return false
	}
	if log.Topics[0] != e.ID() {
		return false
	}
	return true
}

// ParseLog parses a log with this event
func (e *Event) ParseLog(log *web3.Log) (map[string]interface{}, error) {
	if !e.Match(log) {
		return nil, fmt.Errorf("log does not match this event")
	}
	return e.Inputs.ParseLog(log)
}

func buildSignature(name string, typ *Type) string {
	types := make([]string, len(typ.tuple))
	for i, input := range typ.tuple {
		types[i] = input.Elem.String()
	}
	return fmt.Sprintf("%v(%v)", name, strings.Join(types, ","))
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

func NewABIFromList(humanReadableAbi []string) (*ABI, error) {
	res := &ABI{}
	for _, c := range humanReadableAbi {
		if strings.HasPrefix(c, "function ") {
			method, err := NewMethod(c)
			if err != nil {
				return nil, err
			}
			res.addMethod(method)

		} else if strings.HasPrefix(c, "event ") {
			evnt, err := NewEvent(c)
			if err != nil {
				return nil, err
			}
			res.addEvent(evnt)

		} else if strings.HasPrefix(c, "error ") {
			errTyp, err := NewError(c)
			if err != nil {
				return nil, err
			}
			res.addError(errTyp)

		} else {
			return nil, fmt.Errorf("either event or function expected")
		}
	}
	return res, nil
}
