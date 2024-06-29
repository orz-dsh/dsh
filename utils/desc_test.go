package utils

import (
	"fmt"
	"testing"
	"unsafe"
)

type StructWithDescKeyValues struct {
	intValue                int
	strValue                string
	nestedStruct            *StructWithDescKeyValues
	mapValue                map[string]any
	emptyDescKeyValues      bool
	emptyDescExtraKeyValues bool
}

func (s *StructWithDescKeyValues) DescKeyValues() KVS {
	if s.emptyDescKeyValues {
		return nil
	}
	return KVS{
		KV("customDescIntValue", s.intValue),
		KV("customDescStrValue", s.strValue),
		KV("customDescNestedStruct", s.nestedStruct),
	}
}

func (s *StructWithDescKeyValues) DescExtraKeyValues() KVS {
	if s.emptyDescExtraKeyValues {
		return nil
	}
	return KVS{
		KV("customDescExtraIntValue", s.intValue),
		KV("customDescExtraStrValue", s.strValue),
		KV("customDescExtraNestedStruct", s.nestedStruct),
	}
}

type StructSlice []*StructWithDescKeyValues

type StructMap map[string]*StructWithDescKeyValues

type StructSliceWithDescKeyValues []*StructWithDescKeyValues

func (s StructSliceWithDescKeyValues) DescKeyValues() KVS {
	kvs := make(KVS, len(s))
	for i, v := range s {
		kvs[i] = KV(fmt.Sprintf("item-%d", i), v)
	}
	return kvs
}

func (s StructSliceWithDescKeyValues) DescExtraKeyValues() KVS {
	return KVS{
		KV("len", len(s)),
	}
}

type StructMapWithDescKeyValues map[string]*StructWithDescKeyValues

func (m StructMapWithDescKeyValues) DescKeyValues() KVS {
	var kvs KVS
	for k, v := range m {
		kvs = append(kvs, KV(fmt.Sprintf("item-%s", k), v))
	}
	return kvs
}

func (m StructMapWithDescKeyValues) DescExtraKeyValues() KVS {
	return KVS{
		KV("len", len(m)),
	}
}

func TestDescBase(t *testing.T) {
	var nilValue any = nil
	str := DescN("test desc base",
		KV("int", 0),
		KV("float", 3.4),
		KV("bool", true),
		KV("string", "abc"),
		KV("complex", complex(1, 2)),
		KV("nil", nilValue),
		KV("array", [3]int{1, 2, 3}),
		KV("slice", []int{1, 2, 3}),
		KV("map", map[string]any{"k1": "a", "k2": "b"}),
	)
	t.Log(str)
}

func TestDescCollection(t *testing.T) {
	mapValue := map[string]any{
		"a": 1,
		"b": 2,
		"c": map[string]any{
			"d": 3,
			"e": nil,
			"f": []int{4, 5, 6},
			"g": StructWithDescKeyValues{
				intValue: 1,
				mapValue: map[string]any{
					"a": 1,
					"b": 2,
				},
			},
			"h": &StructWithDescKeyValues{
				intValue: 1,
				mapValue: map[string]any{
					"a": 1,
					"b": 2,
				},
			},
		},
		"l1": map[string]any{
			"l1-1": "l1-1",
			"l2": map[string]any{
				"l2-1": "l2-1",
				"l3": map[string]any{
					"l3-1": "l3-1",
					"l4": map[string]any{
						"l4-1": "l4-1",
						"l5": map[string]any{
							"l5-1": "l5-1",
							"l6": map[string]any{
								"l6-1": "l6-1",
								"l7": map[string]any{
									"l7-1": "l7-1",
									"l8": map[string]any{
										"l8-1": "l8-1",
										"l9": map[string]any{
											"l9-1": "l9-1",
											"l10": map[string]any{
												"l10-1": "l10-1",
												"l11":   map[string]any{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	sliceValue := []any{
		"l1",
		[]any{
			"l2",
			[]any{
				"l3",
				[]any{
					"l4",
					[]any{
						"l5",
						[]any{
							"l6",
							[]any{
								"l7",
								[]any{
									"l8",
									[]any{
										"l9",
										[]any{
											"l10",
											[]any{
												"l11",
												[]any{},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	structSlice1 := StructSlice{
		{
			intValue: 1,
			strValue: "struct1",
		},
		{
			intValue: 2,
			strValue: "struct2",
		},
		{
			intValue: 3,
			strValue: "struct3",
		},
	}

	structMap1 := StructMap{
		"k1": structSlice1[0],
		"k2": structSlice1[1],
		"k3": structSlice1[2],
	}

	structSlice2 := StructSliceWithDescKeyValues{
		{
			intValue: 1,
			strValue: "struct1",
		},
		{
			intValue: 2,
			strValue: "struct2",
		},
		{
			intValue: 3,
			strValue: "struct3",
		},
	}

	structMap2 := StructMapWithDescKeyValues{
		"k1": structSlice2[0],
		"k2": structSlice2[1],
		"k3": structSlice2[2],
	}

	str := DescN("test desc collection",
		KV("mapValue", mapValue),
		KV("sliceValue", sliceValue),
		KV("structSlice1", structSlice1),
		KV("structMap1", structMap1),
		KV("structSlice2", structSlice2),
		KV("structMap2", structMap2),
		KV("collection", []any{
			mapValue,
			sliceValue,
			structSlice1,
			structMap1,
			structSlice2,
			structMap2,
		}),
	)
	t.Log(str)
}

func TestDescKeyValuesFunc(t *testing.T) {
	struct1 := &StructWithDescKeyValues{
		intValue: 1,
		strValue: "struct1",
	}
	struct2 := &StructWithDescKeyValues{
		intValue: 2,
		strValue: "struct2",
	}
	struct1.nestedStruct = struct2
	struct2.nestedStruct = struct1

	struct3 := &StructWithDescKeyValues{
		intValue:                3,
		strValue:                "struct3",
		emptyDescKeyValues:      true,
		emptyDescExtraKeyValues: true,
	}
	struct4 := &StructWithDescKeyValues{
		intValue:                4,
		strValue:                "struct4",
		emptyDescKeyValues:      false,
		emptyDescExtraKeyValues: true,
	}
	struct5 := &StructWithDescKeyValues{
		intValue:                5,
		strValue:                "struct5",
		emptyDescKeyValues:      true,
		emptyDescExtraKeyValues: false,
	}

	str := DescN("test desc key values func",
		KV("struct1", struct1),
		KV("struct2", struct2),
		KV("struct3", struct3),
		KV("struct4", struct4),
		KV("struct5", struct5),
	)
	t.Log(str)
}

func TestDescNestedDesc(t *testing.T) {
	str := DescN("test desc nested desc",
		KV("desc", DescN("nested desc title",
			KV("int", 0),
		)),
		KV("kvs", KVS{
			KV("int", 0),
			KV("str", "abc"),
			KV("struct", &StructWithDescKeyValues{
				intValue: 1,
				strValue: "struct",
			}),
		}),
		KV("kv", KV("int", 0)),
	)
	t.Log(str)
}

func TestDescStruct(t *testing.T) {
	type Struct1 struct {
		IntValue int `desc:"-"`
		StrValue string
		booValue bool
	}

	type Struct2 struct {
		IntValue int    `desc:"-"`
		StrValue string `desc:"-"`
	}

	struct1 := Struct1{
		IntValue: 1,
		StrValue: "struct1",
		booValue: true,
	}
	struct2 := Struct2{
		IntValue: 2,
		StrValue: "struct2",
	}

	str := DescN("test desc struct",
		KV("struct1", struct1),
		KV("struct2", struct2),
	)
	t.Log(str)
}

func TestDescPointer(t *testing.T) {
	intValue := 1
	intValueP := &intValue
	intValuePP := &intValueP
	intValuePPP := &intValuePP

	strValue := "str"
	strValueP := &strValue
	strValuePP := &strValueP
	strValuePPP := &strValuePP

	arrValue := [3]int{1, 2, 3}
	arrValueP := &arrValue
	arrValuePP := &arrValueP
	arrValuePPP := &arrValuePP

	sliceValue := []int{1, 2, 3}
	sliceValueP := &sliceValue
	sliceValuePP := &sliceValueP
	sliceValuePPP := &sliceValuePP

	mapValue := map[string]any{
		"intValue":      intValue,
		"strValue":      strValue,
		"intValuePPP":   intValuePPP,
		"strValuePPP":   strValuePPP,
		"sliceValue":    sliceValue,
		"sliceValueP":   sliceValueP,
		"sliceValuePP":  sliceValuePP,
		"sliceValuePPP": sliceValuePPP,
	}
	mapValueP := &mapValue
	mapValuePP := &mapValueP
	mapValuePPP := &mapValuePP
	mapValue["mapValue"] = mapValue
	mapValue["mapValueP"] = mapValueP
	mapValue["mapValuePP"] = mapValuePP
	mapValue["mapValuePPP"] = mapValuePPP

	type Struct1 struct {
		IntValuePPP ***int
		StrValuePPP ***string
	}

	type Struct2 struct {
		Struct1ValuePPP ***Struct1
	}

	struct1Value := Struct1{
		IntValuePPP: intValuePPP,
		StrValuePPP: strValuePPP,
	}
	struct1ValueP := &struct1Value
	struct1ValuePP := &struct1ValueP
	struct1ValuePPP := &struct1ValuePP

	struct2Value := Struct2{
		Struct1ValuePPP: struct1ValuePPP,
	}
	struct2ValueP := &struct2Value
	struct2ValuePP := &struct2ValueP
	struct2ValuePPP := &struct2ValuePP

	str := DescN("test desc pointer",
		KV("intValue", intValue),
		KV("intValueP", intValueP),
		KV("intValuePP", intValuePP),
		KV("intValuePPP", intValuePPP),
		KV("strValue", strValue),
		KV("strValueP", strValueP),
		KV("strValuePP", strValuePP),
		KV("strValuePPP", strValuePPP),
		KV("arrValue", arrValue),
		KV("arrValueP", arrValueP),
		KV("arrValuePP", arrValuePP),
		KV("arrValuePPP", arrValuePPP),
		KV("sliceValue", sliceValue),
		KV("sliceValueP", sliceValueP),
		KV("sliceValuePP", sliceValuePP),
		KV("sliceValuePPP", sliceValuePPP),
		KV("mapValue", mapValue),
		KV("mapValueP", mapValueP),
		KV("mapValuePP", mapValuePP),
		KV("mapValuePPP", mapValuePPP),
		KV("struct1Value", struct1Value),
		KV("struct1ValueP", struct1ValueP),
		KV("struct1ValuePP", struct1ValuePP),
		KV("struct1ValuePPP", struct1ValuePPP),
		KV("struct2Value", struct2Value),
		KV("struct2ValueP", struct2ValueP),
		KV("struct2ValuePP", struct2ValuePP),
		KV("struct2ValuePPP", struct2ValuePPP),
	)
	t.Log(str)
}

func TestDescFunc(t *testing.T) {
	func1 := func() {}
	func2 := func(a, b, c int) (int, error) {
		return a + b + c, nil
	}
	func3 := func(a *int, b *string, c any) (*testing.T, error) {
		return nil, nil
	}
	func4 := func(a ...int) ([]int, error) {
		return a, nil
	}
	func5 := (&StructWithDescKeyValues{}).DescExtraKeyValues

	str := DescN("test desc func",
		KV("func1", func1),
		KV("func2", func2),
		KV("func3", func3),
		KV("func4", func4),
		KV("func5", func5),
	)
	t.Log(str)
}

func TestDescChan(t *testing.T) {
	var chan1 chan int
	chan2 := make(chan string)
	chan3 := make(chan<- any)
	chan4 := make(<-chan any)

	str := DescN("test desc chan",
		KV("chan1", chan1),
		KV("chan2", chan2),
		KV("chan3", chan3),
		KV("chan4", chan4),
	)
	t.Log(str)
}

func TestDescUnsafePointer(t *testing.T) {
	var unsafePointer1 unsafe.Pointer

	intValue := 1
	unsafePointer2 := unsafe.Pointer(&intValue)

	str := DescN("test desc unsafe pointer",
		KV("unsafePointer1", unsafePointer1),
		KV("unsafePointer2", unsafePointer2),
	)
	t.Log(str)
}

func TestDescUintptr(t *testing.T) {
	var uintptr1 uintptr
	uintptr2 := uintptr(0)
	uintptr3 := uintptr(unsafe.Pointer(&uintptr2))

	str := DescN("test desc uintptr",
		KV("uintptr1", uintptr1),
		KV("uintptr2", uintptr2),
		KV("uintptr3", uintptr3),
	)
	t.Log(str)
}
