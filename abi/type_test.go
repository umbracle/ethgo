package abi

import (
	"fmt"
	"reflect"
	"testing"
)

func TestType(t *testing.T) {
	cases := []struct {
		s *Argument
		t *Type
	}{
		{
			s: simpleType("bool"),
			t: &Type{kind: KindBool, t: boolT, raw: "bool"},
		},
		{
			s: simpleType("uint32"),
			t: &Type{kind: KindUInt, size: 32, t: uint32T, raw: "uint32"},
		},
		{
			s: simpleType("int32"),
			t: &Type{kind: KindInt, size: 32, t: int32T, raw: "int32"},
		},
		{
			s: simpleType("int32[]"),
			t: &Type{kind: KindSlice, t: reflect.SliceOf(int32T), raw: "int32[]", elem: &Type{kind: KindInt, size: 32, t: int32T, raw: "int32"}},
		},
		{
			s: simpleType("string[2]"),
			t: &Type{
				kind: KindArray,
				size: 2,
				t:    reflect.ArrayOf(2, stringT),
				raw:  "string[2]",
				elem: &Type{
					kind: KindString,
					t:    stringT,
					raw:  "string",
				},
			},
		},
		{
			s: simpleType("string[2][]"),
			t: &Type{
				kind: KindSlice,
				t:    reflect.SliceOf(reflect.ArrayOf(2, stringT)),
				raw:  "string[2][]",
				elem: &Type{
					kind: KindArray,
					size: 2,
					t:    reflect.ArrayOf(2, stringT),
					raw:  "string[2]",
					elem: &Type{
						kind: KindString,
						t:    stringT,
						raw:  "string",
					},
				},
			},
		},
		{
			s: &Argument{
				Type: "tuple",
				Components: []*Argument{
					{
						Name: "a",
						Type: "int64",
					},
				},
			},
			t: &Type{
				kind: KindTuple,
				raw:  "(int64)",
				t:    tupleT,
				tuple: []*TupleElem{
					{
						Name: "a",
						Elem: &Type{
							kind: KindInt,
							size: 64,
							t:    int64T,
							raw:  "int64",
						},
					},
				},
			},
		},
		{
			s: &Argument{
				Type: "tuple[2]",
				Components: []*Argument{
					{
						Name: "a",
						Type: "int64",
					},
				},
			},
			t: &Type{
				kind: KindArray,
				size: 2,
				raw:  "(int64)[2]",
				t:    reflect.ArrayOf(2, tupleT),
				elem: &Type{
					kind: KindTuple,
					raw:  "(int64)",
					t:    tupleT,
					tuple: []*TupleElem{
						{
							Name: "a",
							Elem: &Type{
								kind: KindInt,
								size: 64,
								t:    int64T,
								raw:  "int64",
							},
						},
					},
				},
			},
		},
		{
			s: &Argument{
				Type: "tuple[]",
				Components: []*Argument{
					{
						Name: "a",
						Type: "int64",
					},
				},
			},
			t: &Type{
				kind: KindSlice,
				raw:  "(int64)[]",
				t:    reflect.SliceOf(tupleT),
				elem: &Type{
					kind: KindTuple,
					raw:  "(int64)",
					t:    tupleT,
					tuple: []*TupleElem{
						{
							Name: "a",
							Elem: &Type{
								kind: KindInt,
								size: 64,
								t:    int64T,
								raw:  "int64",
							},
						},
					},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			e, err := NewType(c.s)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(c.t, e) {

				fmt.Println(c.t)
				fmt.Println(e)

				t.Fatal("bad new type")
			}
		})
	}
}

func TestSize(t *testing.T) {
	cases := []struct {
		Input interface{}
		Size  int
	}{
		{
			"int32", 32,
		},
		{
			"int32[]", 32,
		},
		{
			"int32[2]", 32 * 2,
		},
		{
			"int32[2][2]", 32 * 2 * 2,
		},
		{
			"string", 32,
		},
		{
			"string[]", 32,
		},
		{
			&Argument{
				Type: "tuple[1]",
				Components: []*Argument{
					{Name: "a", Type: "uint8"},
					{Name: "b", Type: "uint32"},
				},
			},
			64,
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			var arg *Argument
			if typeStr, ok := c.Input.(string); ok {
				arg = simpleType(typeStr)
			} else if aux, ok := c.Input.(*Argument); ok {
				arg = aux
			} else {
				t.Fatal("unknown input")
			}

			tt, err := NewType(arg)
			if err != nil {
				t.Fatal(err)
			}

			size := getTypeSize(tt)
			if size != c.Size {
				t.Fatalf("expected size %d but found %d", c.Size, size)
			}
		})
	}
}

func TestBrackets(t *testing.T) {
	cases := []struct {
		Input    string
		Expected []int
		Error    bool
	}{
		{
			Input:    "[][][2]",
			Expected: []int{-1, -1, 2},
		},
		{
			Input: "[[[",
			Error: true,
		},
		{
			Input: "]]]",
			Error: true,
		},
		{
			Input: "[[[]]]",
			Error: true,
		},
		{
			Input:    "[][]",
			Expected: []int{-1, -1},
		},
		{
			Input:    "[1]",
			Expected: []int{1},
		},
		{
			Input:    "[]",
			Expected: []int{-1},
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			r, err := parseBrackets(c.Input)
			if err != nil && !c.Error {
				t.Fatalf("Failed not expected: %v", err)
			}
			if err == nil && c.Error {
				t.Fatal("Expected to failed")
			}

			if !c.Error {
				if !reflect.DeepEqual(r, c.Expected) {
					t.Fatal("bad")
				}
			}
		})
	}
}

func simpleType(s string) *Argument {
	return &Argument{
		Type: s,
	}
}
