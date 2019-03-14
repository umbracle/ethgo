package abi

import (
	"fmt"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// batch of predefined reflect types
var (
	boolT         = reflect.TypeOf(bool(false))
	uint8T        = reflect.TypeOf(uint8(0))
	uint16T       = reflect.TypeOf(uint16(0))
	uint32T       = reflect.TypeOf(uint32(0))
	uint64T       = reflect.TypeOf(uint64(0))
	int8T         = reflect.TypeOf(int8(0))
	int16T        = reflect.TypeOf(int16(0))
	int32T        = reflect.TypeOf(int32(0))
	int64T        = reflect.TypeOf(int64(0))
	addressT      = reflect.TypeOf(common.Address{})
	stringT       = reflect.TypeOf("")
	dynamicBytesT = reflect.SliceOf(reflect.TypeOf(byte(0)))
	functionT     = reflect.ArrayOf(24, reflect.TypeOf(byte(0)))
	tupleT        = reflect.TypeOf(map[string]interface{}{})
	bigInt        = reflect.TypeOf(new(big.Int))
)

// Kind represents the kind of abi type
type Kind int

const (
	// KindBool is a boolean
	KindBool Kind = iota

	// KindUInt is an uint
	KindUInt

	// KindInt is an int
	KindInt

	// KindString is a string
	KindString

	// KindArray is an array
	KindArray

	// KindSlice is a slice
	KindSlice

	// KindAddress is an address
	KindAddress

	// KindBytes is a bytes array
	KindBytes

	// KindFixedBytes is a fixed bytes
	KindFixedBytes

	// KindFixedPoint is a fixed point
	KindFixedPoint

	// KindTuple is a tuple
	KindTuple

	// KindFunction is a function
	KindFunction
)

func (k Kind) String() string {
	names := [...]string{
		"Bool",
		"Uint",
		"Int",
		"String",
		"Array",
		"Slice",
		"Address",
		"Bytes",
		"FixedBytes",
		"FixedPoint",
		"Tuple",
		"Function",
	}

	return names[k]
}

// TupleElem is an element of a tuple
type TupleElem struct {
	Name string
	Elem *Type
}

// Type is an ABI type
type Type struct {
	kind  Kind
	size  int
	elem  *Type
	raw   string
	tuple []*TupleElem
	t     reflect.Type
}

func (t *Type) isVariableInput() bool {
	return t.kind == KindSlice || t.kind == KindBytes || t.kind == KindString
}

func (t *Type) isDynamicType() bool {
	if t.kind == KindTuple {
		for _, elem := range t.tuple {
			if elem.Elem.isDynamicType() {
				return true
			}
		}
		return false
	}
	return t.kind == KindString || t.kind == KindBytes || t.kind == KindSlice || (t.kind == KindArray && t.elem.isDynamicType())
}

// NewType parses an abi type string
func NewType(arg *Argument) (*Type, error) {
	tStr := arg.Type
	indx := strings.Index(arg.Type, "[")
	if indx != -1 {
		tStr = arg.Type[:indx]
	}

	t, err := parseType(tStr, arg)
	if err != nil {
		return nil, err
	}
	if t.kind != KindTuple {
		t.raw = tStr
	}

	if indx != -1 {
		return parseList(t, arg.Type[indx:], arg)
	}
	return t, nil
}

func parseList(t *Type, s string, arg *Argument) (*Type, error) {
	l, err := parseBrackets(s)
	if err != nil {
		return nil, err
	}

	for _, size := range l {
		var tAux *Type
		if size == -1 {
			// array
			tAux = &Type{kind: KindSlice, elem: t, raw: fmt.Sprintf("%s[]", t.raw), t: reflect.SliceOf(t.t)}
		} else {
			// slice
			tAux = &Type{kind: KindArray, elem: t, raw: fmt.Sprintf("%s[%d]", t.raw, size), size: size, t: reflect.ArrayOf(size, t.t)}
		}
		t = tAux
	}
	return t, nil
}

var typeRegexp = regexp.MustCompile("^([[:alpha:]]+)([[:digit:]]*)$")

func parseType(s string, arg *Argument) (*Type, error) {
	match := typeRegexp.FindStringSubmatch(s)
	if len(match) == 0 {
		return nil, fmt.Errorf("type format is incorrect. Expected 'type''bytes' but found '%s'", s)
	}
	match = match[1:]

	var err error
	t := match[0]

	b := 0
	if bytesStr := match[1]; bytesStr != "" {
		b, err = strconv.Atoi(bytesStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse bytes '%s': %v", bytesStr, err)
		}
	}

	switch t {
	case "bool":
		if b != 0 {
			return nil, fmt.Errorf("bytes to allowed in bool format")
		}
		return &Type{kind: KindBool, t: boolT}, nil

	case "uint":
		if b == 0 {
			return nil, fmt.Errorf("expected bytes in uint")
		}

		var k reflect.Type
		switch b {
		case 8:
			k = uint8T
		case 16:
			k = uint16T
		case 32:
			k = uint32T
		case 64:
			k = uint64T
		default:
			if b%8 != 0 {
				return nil, fmt.Errorf("number of bytes has to be M mod 8")
			}
			k = bigInt
		}

		return &Type{kind: KindUInt, size: b, t: k}, nil

	case "int":
		if b == 0 {
			return nil, fmt.Errorf("expected bytes in int")
		}

		var k reflect.Type
		switch b {
		case 8:
			k = int8T
		case 16:
			k = int16T
		case 32:
			k = int32T
		case 64:
			k = int64T
		default:
			if b%8 != 0 {
				return nil, fmt.Errorf("number of bytes has to be M mod 8")
			}
			k = bigInt
		}
		return &Type{kind: KindInt, size: b, t: k}, nil

	case "address":
		if b != 0 {
			return nil, fmt.Errorf("bytes to allowed in address")
		}
		return &Type{kind: KindAddress, size: 20, t: addressT}, nil

	case "string":
		if b != 0 {
			return nil, fmt.Errorf("bytes to allowed in string")
		}
		return &Type{kind: KindString, t: stringT}, nil

	case "bytes":
		if b == 0 {
			return &Type{kind: KindBytes, t: dynamicBytesT}, nil
		}
		// fixed bytes
		return &Type{kind: KindFixedBytes, size: b, t: reflect.ArrayOf(b, reflect.TypeOf(byte(0)))}, nil

	case "function":
		if b != 0 {
			return nil, fmt.Errorf("bytes to allowed in function")
		}
		return &Type{kind: KindFunction, size: 24, t: functionT}, nil

	case "tuple":
		elems := []*TupleElem{}
		for _, c := range arg.Components {
			elem, err := NewType(c)
			if err != nil {
				return nil, err
			}
			elems = append(elems, &TupleElem{
				Name: c.Name,
				Elem: elem,
			})
		}

		rawAux := []string{}
		for _, i := range elems {
			rawAux = append(rawAux, i.Elem.raw)
		}
		raw := fmt.Sprintf("(%s)", strings.Join(rawAux, ","))

		return &Type{kind: KindTuple, raw: raw, tuple: elems, t: tupleT}, nil
	}

	return nil, fmt.Errorf("invalid type '%s'", t)
}

func parseBrackets(s string) ([]int, error) {
	res := []int{}
	for {
		if s[0] != '[' {
			return nil, fmt.Errorf("bad")
		}

		indx := strings.Index(s, "]")
		if indx == -1 {
			return nil, fmt.Errorf("close bracket not found")
		}

		val := s[1:indx]
		if val == "" {
			res = append(res, -1)
		} else {
			j, err := strconv.Atoi(val)
			if err != nil {
				return nil, err
			}
			res = append(res, j)
		}

		s = s[indx+1:]
		if s == "" {
			break
		}
	}
	return res, nil
}

func getTypeSize(t *Type) int {
	if t.kind == KindArray && !t.elem.isDynamicType() {
		if t.elem.kind == KindArray || t.elem.kind == KindTuple {
			return t.size * getTypeSize(t.elem)
		}
		return t.size * 32
	} else if t.kind == KindTuple && !t.isDynamicType() {
		total := 0
		for _, elem := range t.tuple {
			total += getTypeSize(elem.Elem)
		}
		return total
	}
	return 32
}
