package abi

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType(t *testing.T) {
	cases := []struct {
		s   string
		a   *ArgumentStr
		t   *Type
		r   string
		err bool
	}{
		{
			s: "bool",
			a: simpleType("bool"),
			t: &Type{kind: KindBool, t: boolT},
		},
		{
			s: "uint32",
			a: simpleType("uint32"),
			t: &Type{kind: KindUInt, size: 32, t: uint32T},
		},
		{
			s: "int32",
			a: simpleType("int32"),
			t: &Type{kind: KindInt, size: 32, t: int32T},
		},
		{
			s: "int32[]",
			a: simpleType("int32[]"),
			t: &Type{kind: KindSlice, t: reflect.SliceOf(int32T), elem: &Type{kind: KindInt, size: 32, t: int32T}},
		},
		{
			s: "bytes[2]",
			a: simpleType("bytes[2]"),
			t: &Type{
				kind: KindArray,
				t:    reflect.ArrayOf(2, dynamicBytesT),
				size: 2,
				elem: &Type{
					kind: KindBytes,
					t:    dynamicBytesT,
				},
			},
		},
		{
			s: "address[]",
			a: simpleType("address[]"),
			t: &Type{kind: KindSlice, t: reflect.SliceOf(addressT), elem: &Type{kind: KindAddress, size: 20, t: addressT}},
		},
		{
			s: "string[]",
			a: simpleType("string[]"),
			t: &Type{
				kind: KindSlice,
				t:    reflect.SliceOf(stringT),
				elem: &Type{
					kind: KindString,
					t:    stringT,
				},
			},
		},
		{
			s: "string[2]",
			a: simpleType("string[2]"),
			t: &Type{
				kind: KindArray,
				size: 2,
				t:    reflect.ArrayOf(2, stringT),
				elem: &Type{
					kind: KindString,
					t:    stringT,
				},
			},
		},

		{
			s: "string[2][]",
			a: simpleType("string[2][]"),
			t: &Type{
				kind: KindSlice,
				t:    reflect.SliceOf(reflect.ArrayOf(2, stringT)),
				elem: &Type{
					kind: KindArray,
					size: 2,
					t:    reflect.ArrayOf(2, stringT),
					elem: &Type{
						kind: KindString,
						t:    stringT,
					},
				},
			},
		},
		{
			s: "tuple(int64 indexed arg0)",
			a: &ArgumentStr{
				Type: "tuple",
				Components: []*ArgumentStr{
					{
						Name:    "arg0",
						Type:    "int64",
						Indexed: true,
					},
				},
			},
			t: &Type{
				kind: KindTuple,
				t:    tupleT,
				tuple: []*TupleElem{
					{
						Name: "arg0",
						Elem: &Type{
							kind: KindInt,
							size: 64,
							t:    int64T,
						},
						Indexed: true,
					},
				},
			},
		},
		{
			s: "tuple(int64 arg_0)[2]",
			a: &ArgumentStr{
				Type: "tuple[2]",
				Components: []*ArgumentStr{
					{
						Name: "arg_0",
						Type: "int64",
					},
				},
			},
			t: &Type{
				kind: KindArray,
				size: 2,
				t:    reflect.ArrayOf(2, tupleT),
				elem: &Type{
					kind: KindTuple,
					t:    tupleT,
					tuple: []*TupleElem{
						{
							Name: "arg_0",
							Elem: &Type{
								kind: KindInt,
								size: 64,
								t:    int64T,
							},
						},
					},
				},
			},
		},
		{
			s: "tuple(int64 a)[]",
			a: &ArgumentStr{
				Type: "tuple[]",
				Components: []*ArgumentStr{
					{
						Name: "a",
						Type: "int64",
					},
				},
			},
			t: &Type{
				kind: KindSlice,
				t:    reflect.SliceOf(tupleT),
				elem: &Type{
					kind: KindTuple,
					t:    tupleT,
					tuple: []*TupleElem{
						{
							Name: "a",
							Elem: &Type{
								kind: KindInt,
								size: 64,
								t:    int64T,
							},
						},
					},
				},
			},
		},
		{
			s: "tuple(int32 indexed arg0,tuple(int32 c) b_2)",
			a: &ArgumentStr{
				Type: "tuple",
				Components: []*ArgumentStr{
					{
						Name:    "arg0",
						Type:    "int32",
						Indexed: true,
					},
					{
						Name: "b_2",
						Type: "tuple",
						Components: []*ArgumentStr{
							{
								Name: "c",
								Type: "int32",
							},
						},
					},
				},
			},
			t: &Type{
				kind: KindTuple,
				t:    tupleT,
				tuple: []*TupleElem{
					{
						Name: "arg0",
						Elem: &Type{
							kind: KindInt,
							size: 32,
							t:    int32T,
						},
						Indexed: true,
					},
					{
						Name: "b_2",
						Elem: &Type{
							kind: KindTuple,
							t:    tupleT,
							tuple: []*TupleElem{
								{
									Name: "c",
									Elem: &Type{
										kind: KindInt,
										size: 32,
										t:    int32T,
									},
								},
							},
						},
					},
				},
			},
		},
		{
			s: "tuple()",
			a: &ArgumentStr{
				Type:       "tuple",
				Components: []*ArgumentStr{},
			},
			t: &Type{
				kind:  KindTuple,
				t:     tupleT,
				tuple: []*TupleElem{},
			},
		},
		{
			s:   "int[[",
			err: true,
		},
		{
			s:   "int",
			err: true,
		},
		{
			s:   "tuple[](a int32)",
			err: true,
		},
		{
			s:   "int32[a]",
			err: true,
		},
		{
			s:   "tuple(a int32",
			err: true,
		},
		{
			s:   "tuple(a int32,",
			err: true,
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			e0, err := NewType(c.s)
			if err != nil && !c.err {
				t.Fatal(err)
			}
			if err == nil && c.err {
				t.Fatal("it should have failed")
			}

			if !c.err {
				// compare the string
				expected := c.s
				if c.r != "" {
					expected = c.r
				}
				assert.Equal(t, expected, e0.Format(true))

				e1, err := NewTypeFromArgument(c.a)
				if err != nil {
					t.Fatal(err)
				}

				if !reflect.DeepEqual(c.t, e0) {

					// fmt.Println(c.t.t)
					// fmt.Println(e0.t)

					t.Fatal("bad new type")
				}
				if !reflect.DeepEqual(c.t, e1) {
					t.Fatal("bad")
				}
			}
		})
	}
}

func TestSize(t *testing.T) {
	cases := []struct {
		Input string
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
			"tuple(uint8 a, uint32 b)[1]",
			64,
		},
	}

	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			tt, err := NewType(c.Input)
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

func simpleType(s string) *ArgumentStr {
	return &ArgumentStr{
		Type: s,
	}
}
