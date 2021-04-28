package abi

import (
	"reflect"
	"testing"
)

func TestType(t *testing.T) {
	cases := []struct {
		s   string
		a   *ArgumentStr
		t   *Type
		err bool
	}{
		{
			s: "bool",
			a: simpleType("bool"),
			t: &Type{kind: KindBool, t: boolT, raw: "bool"},
		},
		{
			s: "uint32",
			a: simpleType("uint32"),
			t: &Type{kind: KindUInt, size: 32, t: uint32T, raw: "uint32"},
		},
		{
			s: "int32",
			a: simpleType("int32"),
			t: &Type{kind: KindInt, size: 32, t: int32T, raw: "int32"},
		},
		{
			s: "int32[]",
			a: simpleType("int32[]"),
			t: &Type{kind: KindSlice, t: reflect.SliceOf(int32T), raw: "int32[]", elem: &Type{kind: KindInt, size: 32, t: int32T, raw: "int32"}},
		},
		{
			s: "bytes[2]",
			a: simpleType("bytes[2]"),
			t: &Type{
				kind: KindArray,
				t:    reflect.ArrayOf(2, dynamicBytesT),
				raw:  "bytes[2]",
				size: 2,
				elem: &Type{
					kind: KindBytes,
					t:    dynamicBytesT,
					raw:  "bytes",
				},
			},
		},
		{
			s: "string[]",
			a: simpleType("string[]"),
			t: &Type{
				kind: KindSlice,
				t:    reflect.SliceOf(stringT),
				raw:  "string[]",
				elem: &Type{
					kind: KindString,
					t:    stringT,
					raw:  "string",
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
				raw:  "string[2]",
				elem: &Type{
					kind: KindString,
					t:    stringT,
					raw:  "string",
				},
			},
		},

		{
			s: "string[2][]",
			a: simpleType("string[2][]"),
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
				raw:  "(int64)",
				t:    tupleT,
				tuple: []*TupleElem{
					{
						Name: "arg0",
						Elem: &Type{
							kind: KindInt,
							size: 64,
							t:    int64T,
							raw:  "int64",
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
				raw:  "(int64)[2]",
				t:    reflect.ArrayOf(2, tupleT),
				elem: &Type{
					kind: KindTuple,
					raw:  "(int64)",
					t:    tupleT,
					tuple: []*TupleElem{
						{
							Name: "arg_0",
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
		{
			s: "tuple(int32 indexed arg0, tuple(int32 c) b_2)",
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
				raw:  "(int32,(int32))",
				tuple: []*TupleElem{
					{
						Name: "arg0",
						Elem: &Type{
							kind: KindInt,
							size: 32,
							t:    int32T,
							raw:  "int32",
						},
						Indexed: true,
					},
					{
						Name: "b_2",
						Elem: &Type{
							kind: KindTuple,
							t:    tupleT,
							raw:  "(int32)",
							tuple: []*TupleElem{
								{
									Name: "c",
									Elem: &Type{
										kind: KindInt,
										size: 32,
										t:    int32T,
										raw:  "int32",
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
				raw:   "()",
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
