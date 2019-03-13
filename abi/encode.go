package abi

import (
	"fmt"
	"math/big"
	"reflect"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
)

// Encode encodes a value
func Encode(v interface{}, t *Type) ([]byte, error) {
	return encode(reflect.ValueOf(v), t)
}

func encode(v reflect.Value, t *Type) ([]byte, error) {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	switch t.kind {
	case KindSlice, KindArray:
		return encodeSliceAndArray(v, t)

	case KindTuple:
		return encodeTuple(v, t)

	case KindString:
		return encodeString(v)

	case KindBool:
		return encodeBool(v)

	case KindAddress:
		return encodeAddress(v)

	case KindInt, KindUInt:
		return encodeNum(v)

	case KindBytes:
		return encodeBytes(v)

	case KindFixedBytes, KindFunction:
		return encodeFixedBytes(v)

	default:
		return nil, fmt.Errorf("encoding not available for type '%s'", t.kind)
	}
}

func encodeSliceAndArray(v reflect.Value, t *Type) ([]byte, error) {
	var ret, tail []byte

	if t.isVariableInput() {
		ret = append(ret, packNum(v.Len())...)
	}

	offset := 0
	isDynamic := t.elem.isDynamicType()
	if isDynamic {
		offset = getTypeSize(t.elem) * v.Len()
	}

	for i := 0; i < v.Len(); i++ {
		val, err := encode(v.Index(i), t.elem)
		if err != nil {
			return nil, err
		}
		if !isDynamic {
			ret = append(ret, val...)
		} else {
			ret = append(ret, packNum(offset)...)
			offset += len(val)
			tail = append(tail, val...)
		}
	}
	return append(ret, tail...), nil
}

func encodeTuple(v reflect.Value, t *Type) ([]byte, error) {
	offset := 0
	for _, elem := range t.tuple {
		offset += getTypeSize(elem.Elem)
	}

	var ret, tail []byte
	for i, elem := range t.tuple {
		val, err := encode(v.Index(i), elem.Elem)
		if err != nil {
			return nil, err
		}
		if elem.Elem.isDynamicType() {
			ret = append(ret, packNum(offset)...)
			tail = append(tail, val...)
			offset += len(val)
		} else {
			ret = append(ret, val...)
		}
	}

	return append(ret, tail...), nil
}

func convertArrayToBytes(value reflect.Value) reflect.Value {
	slice := reflect.MakeSlice(reflect.TypeOf([]byte{}), value.Len(), value.Len())
	reflect.Copy(slice, value)
	return slice
}

func encodeFixedBytes(v reflect.Value) ([]byte, error) {
	if v.Kind() == reflect.Array {
		v = convertArrayToBytes(v)
	}
	res := common.RightPadBytes(v.Bytes(), 32)
	return res, nil
}

func encodeAddress(v reflect.Value) ([]byte, error) {
	if v.Kind() == reflect.Array {
		v = convertArrayToBytes(v)
	}
	return common.LeftPadBytes(v.Bytes(), 32), nil
}

func encodeBytes(v reflect.Value) ([]byte, error) {
	if v.Kind() == reflect.Array {
		v = convertArrayToBytes(v)
	}
	return packBytesSlice(v.Bytes(), v.Len())
}

func encodeString(v reflect.Value) ([]byte, error) {
	if v.Kind() != reflect.String {
		return nil, encodeErr(v, "string")
	}
	return packBytesSlice([]byte(v.String()), v.Len())
}

func packBytesSlice(bytes []byte, l int) ([]byte, error) {
	len, err := encodeNum(reflect.ValueOf(l))
	if err != nil {
		return nil, err
	}
	return append(len, common.RightPadBytes(bytes, (l+31)/32*32)...), nil
}

func packNum(offset int) []byte {
	n, _ := encodeNum(reflect.ValueOf(offset))
	return n
}

func encodeNum(v reflect.Value) ([]byte, error) {
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return U256(new(big.Int).SetUint64(v.Uint())), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return U256(big.NewInt(v.Int())), nil

	case reflect.Ptr:
		if v.Type() != bigInt {
			return nil, encodeErr(v.Elem(), "number")
		}
		return U256(v.Interface().(*big.Int)), nil

	default:
		return nil, encodeErr(v, "number")
	}
}

func encodeBool(v reflect.Value) ([]byte, error) {
	if v.Kind() != reflect.Bool {
		return nil, encodeErr(v, "bool")
	}
	if v.Bool() {
		return math.PaddedBigBytes(common.Big1, 32), nil
	}
	return math.PaddedBigBytes(common.Big0, 32), nil
}

func encodeErr(v reflect.Value, t string) error {
	return fmt.Errorf("failed to encode %s as %s", v.Kind().String(), t)
}
