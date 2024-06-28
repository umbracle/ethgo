package signing

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"

	"github.com/Ethernal-Tech/ethgo"
	"github.com/Ethernal-Tech/ethgo/abi"
)

type EIP712MessageBuilder[T any] struct {
	Types       map[string][]*EIP712Type
	PrimaryType string
	Domain      *EIP712Domain
}

func (e *EIP712MessageBuilder[T]) GetEncodedType() string {
	return encodeType(e.PrimaryType, e.Types)
}

func NewEIP712MessageBuilder[T any](domain *EIP712Domain) *EIP712MessageBuilder[T] {
	var t T
	types := map[string][]*EIP712Type{}
	primaryType := decodeStructType(reflect.ValueOf(t).Type(), &types)

	builder := &EIP712MessageBuilder[T]{
		Types:       types,
		PrimaryType: primaryType,
		Domain:      domain,
	}
	return builder
}

func decodeStructType(typ reflect.Type, result *map[string][]*EIP712Type) string {
	if typ.Kind() != reflect.Struct {
		panic(fmt.Sprintf("struct expected but found %s", typ.Kind()))
	}

	name := typ.Name()

	types := []*EIP712Type{}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldName := field.Name

		// use a tag as a name (if any)
		if tagVal := field.Tag.Get("eip712"); tagVal != "" {
			fieldName = tagVal
		}

		fieldType := decodeTypes(field.Type, result)

		types = append(types, &EIP712Type{
			Type: fieldType,
			Name: fieldName,
		})
	}

	(*result)[name] = types
	return name
}

func isByteSlice(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8
}

func isByteArray(t reflect.Type) bool {
	return t.Kind() == reflect.Array && t.Elem().Kind() == reflect.Uint8
}

var (
	addressT = reflect.TypeOf(ethgo.Address{})
	bigIntT  = reflect.TypeOf(new(big.Int))
)

func decodeTypes(val reflect.Type, result *map[string][]*EIP712Type) string {
	if val == addressT {
		return "address"
	}
	if val == bigIntT {
		return "uint256"
	}

	switch val.Kind() {
	case reflect.Array:
		if val.Elem().Kind() == reflect.Uint8 {
			// [x]byte
			return fmt.Sprintf("[%d]byte", val.Len())
		}
		return fmt.Sprintf("%s[%d]", decodeTypes(val.Elem(), result), val.Len())

	case reflect.Slice:
		if val.Elem().Kind() == reflect.Uint8 {
			// []byte
			return "bytes"
		}
		return decodeTypes(val.Elem(), result) + "[]"

	case reflect.Struct:
		return decodeStructType(val, result)

	case reflect.String:
		return "string"

	case reflect.Uint8:
		return "uint8"

	case reflect.Uint16:
		return "uint16"

	case reflect.Uint32:
		return "uint32"

	case reflect.Uint64:
		return "uint64"

	case reflect.Ptr:
		return decodeTypes(val.Elem(), result)

	default:
	}

	panic(fmt.Sprintf("type %s not found", val.Kind()))
}

func (e *EIP712MessageBuilder[T]) Build(obj *T) *EIP712TypedData {
	message := structToMap(reflect.ValueOf(obj))

	res := &EIP712TypedData{
		Types:       e.Types,
		PrimaryType: e.PrimaryType,
		Message:     message,
		Domain:      e.Domain,
	}
	return res
}

func structToMap(v reflect.Value) map[string]interface{} {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	typ := v.Type()
	result := make(map[string]interface{})

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := v.Field(i)

		fieldName := field.Name
		// use a tag as a name (if any)
		if tagVal := field.Tag.Get("eip712"); tagVal != "" {
			fieldName = tagVal
		}

		if field.Type == bigIntT {
			result[fieldName] = fieldValue.Interface()
			continue
		}

		if fieldValue.Kind() == reflect.Ptr {
			fieldValue = fieldValue.Elem()
		}

		if fieldValue.Kind() == reflect.Struct {
			// the field is a struct, handle recursively as another map
			result[fieldName] = structToMap(fieldValue)

		} else if fieldValue.Kind() == reflect.Slice || fieldValue.Kind() == reflect.Array {
			if field.Type.Elem().Kind() == reflect.Uint8 {
				// []byte return as it is
				result[fieldName] = fieldValue.Interface()
			} else {
				// the field is an slice, return a list of interfaces
				var arr []interface{}
				for j := 0; j < fieldValue.Len(); j++ {
					arr = append(arr, structToMap(fieldValue.Index(j)))
				}

				if fieldValue.Kind() == reflect.Array {
					// conver to array
					result[fieldName] = sliceToArray(arr)
				} else {
					result[fieldName] = arr
				}

			}

		} else {
			// primary field, return its interface form
			result[fieldName] = fieldValue.Interface()
		}
	}

	return result
}

func sliceToArray(slice interface{}) interface{} {
	sliceValue := reflect.ValueOf(slice)
	elemType := sliceValue.Type().Elem()
	length := sliceValue.Len()

	// Create a new slice with the same length and capacity as the original slice
	newSlice := reflect.MakeSlice(reflect.SliceOf(elemType), length, length)

	// Copy the elements from the original slice to the new slice
	reflect.Copy(newSlice, sliceValue)

	// Convert the new slice to an array
	arrayType := reflect.ArrayOf(length, elemType)
	arrayValue := reflect.New(arrayType).Elem()
	reflect.Copy(arrayValue, newSlice)

	// Return the array value as an interface{}
	return arrayValue.Interface()
}

type EIP712Type struct {
	Name string
	Type string
}

type EIP712TypedData struct {
	Types       map[string][]*EIP712Type `json:"types"`
	PrimaryType string                   `json:"primaryType"`
	Domain      *EIP712Domain            `json:"domain"`
	Message     map[string]interface{}   `json:"message"`
}

func (t *EIP712TypedData) Hash() ([]byte, error) {
	a, err := t.Domain.hashStruct()
	if err != nil {
		return nil, err
	}

	b, err := hashStruct(t.PrimaryType, t.Types, t.Message)
	if err != nil {
		return nil, err
	}

	res := []byte{}
	res = append(res, 0x19, 0x1)
	res = append(res, a...)
	res = append(res, b...)

	return ethgo.Keccak256(res), nil
}

type EIP712Domain struct {
	Name              string   `json:"name"`
	Version           string   `json:"version"`
	VerifyingContract string   `json:"verifyingContract"`
	ChainId           *big.Int `json:"chainId"`
	Salt              []byte   `json:"salt"`
}

func hashStruct(primary string, types map[string][]*EIP712Type, data map[string]interface{}) ([]byte, error) {
	a1 := encodeType(primary, types)

	a2, err := encodeData(primary, types, data)
	if err != nil {
		return nil, err
	}

	typeHash := ethgo.Keccak256([]byte(a1))

	input := []byte{}
	input = append(input, typeHash...)
	input = append(input, a2...)

	result := ethgo.Keccak256(input)
	return result, nil
}

func (e *EIP712Domain) hashStruct() ([]byte, error) {
	a1, a2 := e.getObjs()

	return hashStruct("EIP712Domain", map[string][]*EIP712Type{"EIP712Domain": a1}, a2)
}

func (e *EIP712Domain) getObjs() ([]*EIP712Type, map[string]interface{}) {
	var types []*EIP712Type
	data := map[string]interface{}{}

	addType := func(name, typ string) {
		types = append(types, &EIP712Type{Name: name, Type: typ})
	}

	if len(e.Name) != 0 {
		addType("name", "string")
		data["name"] = e.Name
	}
	if len(e.Version) != 0 {
		addType("version", "string")
		data["version"] = e.Version
	}
	if e.ChainId != nil {
		addType("chainId", "uint256")
		data["chainId"] = e.ChainId
	}
	if len(e.VerifyingContract) != 0 {
		addType("verifyingContract", "address")
		data["verifyingContract"] = e.VerifyingContract
	}
	if len(e.Salt) != 0 {
		addType("salt", "bytes32")
		data["salt"] = e.Salt
	}

	return types, data
}

func encodeData(primary string, types map[string][]*EIP712Type, data map[string]interface{}) ([]byte, error) {
	fields := types[primary]

	result := []byte{}
	for _, field := range fields {
		val, ok := data[field.Name]
		if !ok {
			return nil, fmt.Errorf("field '%s' not found", field.Name)
		}

		res, err := encodeItem(field.Type, types, val)
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	}

	return result, nil
}

func encodeItem(typ string, types map[string][]*EIP712Type, val interface{}) ([]byte, error) {
	var res []byte

	// handle array
	if typ[len(typ)-1:] == "]" {
		subType := typ[:strings.LastIndex(typ, "[")]

		var subElem []byte
		v := reflect.ValueOf(val)

		for i := 0; i < v.Len(); i++ {
			elemRes, err := encodeItem(subType, types, v.Index(i).Interface())
			if err != nil {
				return nil, err
			}
			subElem = append(subElem, elemRes...)
		}
		res = ethgo.Keccak256(subElem)

	} else if _, ok := types[typ]; ok {
		// if the item is a struct, handle it
		var err error
		if res, err = hashStruct(typ, types, val.(map[string]interface{})); err != nil {
			return nil, err
		}
	} else if typ == "string" {
		// dynamic string type
		valStr, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("string type not found")
		}
		res = ethgo.Keccak256([]byte(valStr))
	} else if typ == "bytes" {
		// dynamic length bytes, it can be either a string or []byte
		if valStr, ok := val.(string); ok {
			// the string must start with 0x
			valBytes, err := decodeHexString(valStr)
			if err != nil {
				return nil, err
			}
			res = ethgo.Keccak256(valBytes)
		} else if valBytes, ok := val.([]byte); ok {
			res = ethgo.Keccak256(valBytes)
		}
	} else {
		// encode basic item
		typ, err := abi.NewType(typ)
		if err != nil {
			return nil, err
		}

		res, err = abi.Encode(val, typ)
		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func getDependencies(primary string, types map[string][]*EIP712Type) []string {
	// two types cannot be encoded twice
	visited := map[string]struct{}{}

	// a queue of the types to encode
	parseTypes := []string{
		primary,
	}

	deps := sort.StringSlice{}

	for len(parseTypes) != 0 {
		var typ string
		typ, parseTypes = parseTypes[0], parseTypes[1:]

		for _, field := range types[typ] {
			typ := field.Type

			// remove any array items from the name (i.e. Struct[][3])
			if indx := strings.Index(typ, "["); indx != -1 {
				typ = typ[:indx]
			}

			if _, ok := types[typ]; ok {
				if _, ok := visited[typ]; !ok {
					// its a type and not visited yet
					deps = append(deps, typ)
					parseTypes = append(parseTypes, typ)
					visited[typ] = struct{}{}
				}
			}
		}
	}

	deps.Sort()

	// the primary is always the first field
	res := []string{primary}
	res = append(res, deps...)

	return res
}

func encodeType(primary string, types map[string][]*EIP712Type) string {
	deps := getDependencies(primary, types)

	encodedType := ""
	for _, dep := range deps {
		strFields := []string{}
		for _, field := range types[dep] {
			strFields = append(strFields, field.Type+" "+field.Name)
		}
		encodedType += fmt.Sprintf("%s(%s)", dep, strings.Join(strFields, ","))
	}

	return encodedType
}

func decodeHexString(str string) ([]byte, error) {
	if !strings.HasPrefix(str, "0x") {
		return nil, fmt.Errorf("0x prefix not found")
	}
	buf, err := hex.DecodeString(str[2:])
	if err != nil {
		return nil, err
	}
	return buf, nil
}
