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
		kv("customDescIntValue", s.intValue),
		kv("customDescStrValue", s.strValue),
		kv("customDescNestedStruct", s.nestedStruct),
	}
}

func (s *StructWithDescKeyValues) DescExtraKeyValues() KVS {
	if s.emptyDescExtraKeyValues {
		return nil
	}
	return KVS{
		kv("customDescExtraIntValue", s.intValue),
		kv("customDescExtraStrValue", s.strValue),
		kv("customDescExtraNestedStruct", s.nestedStruct),
	}
}

type StructSlice []*StructWithDescKeyValues

type StructMap map[string]*StructWithDescKeyValues

type StructSliceWithDescKeyValues []*StructWithDescKeyValues

func (s StructSliceWithDescKeyValues) DescKeyValues() KVS {
	kvs := make(KVS, len(s))
	for i, v := range s {
		kvs[i] = kv(fmt.Sprintf("item-%d", i), v)
	}
	return kvs
}

func (s StructSliceWithDescKeyValues) DescExtraKeyValues() KVS {
	return KVS{
		kv("len", len(s)),
	}
}

type StructMapWithDescKeyValues map[string]*StructWithDescKeyValues

func (m StructMapWithDescKeyValues) DescKeyValues() KVS {
	var kvs KVS
	for k, v := range m {
		kvs = append(kvs, kv(fmt.Sprintf("item-%s", k), v))
	}
	return kvs
}

func (m StructMapWithDescKeyValues) DescExtraKeyValues() KVS {
	return KVS{
		kv("len", len(m)),
	}
}

func TestDescBase(t *testing.T) {
	var nilValue any = nil
	str := desc("test desc base",
		kv("int", 0),
		kv("float", 3.4),
		kv("bool", true),
		kv("string", "abc"),
		kv("complex", complex(1, 2)),
		kv("nil", nilValue),
		kv("array", [3]int{1, 2, 3}),
		kv("slice", []int{1, 2, 3}),
		kv("map", map[string]any{"k1": "a", "k2": "b"}),
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

	str := desc("test desc collection",
		kv("mapValue", mapValue),
		kv("sliceValue", sliceValue),
		kv("structSlice1", structSlice1),
		kv("structMap1", structMap1),
		kv("structSlice2", structSlice2),
		kv("structMap2", structMap2),
		kv("collection", []any{
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

	str := desc("test desc key values func",
		kv("struct1", struct1),
		kv("struct2", struct2),
		kv("struct3", struct3),
		kv("struct4", struct4),
		kv("struct5", struct5),
	)
	t.Log(str)
}

func TestDescNestedDesc(t *testing.T) {
	str := desc("test desc nested desc",
		kv("desc", desc("nested desc title",
			kv("int", 0),
		)),
		kv("kvs", KVS{
			kv("int", 0),
			kv("str", "abc"),
			kv("struct", &StructWithDescKeyValues{
				intValue: 1,
				strValue: "struct",
			}),
		}),
		kv("kv", kv("int", 0)),
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

	str := desc("test desc struct",
		kv("struct1", struct1),
		kv("struct2", struct2),
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

	str := desc("test desc pointer",
		kv("intValue", intValue),
		kv("intValueP", intValueP),
		kv("intValuePP", intValuePP),
		kv("intValuePPP", intValuePPP),
		kv("strValue", strValue),
		kv("strValueP", strValueP),
		kv("strValuePP", strValuePP),
		kv("strValuePPP", strValuePPP),
		kv("arrValue", arrValue),
		kv("arrValueP", arrValueP),
		kv("arrValuePP", arrValuePP),
		kv("arrValuePPP", arrValuePPP),
		kv("sliceValue", sliceValue),
		kv("sliceValueP", sliceValueP),
		kv("sliceValuePP", sliceValuePP),
		kv("sliceValuePPP", sliceValuePPP),
		kv("mapValue", mapValue),
		kv("mapValueP", mapValueP),
		kv("mapValuePP", mapValuePP),
		kv("mapValuePPP", mapValuePPP),
		kv("struct1Value", struct1Value),
		kv("struct1ValueP", struct1ValueP),
		kv("struct1ValuePP", struct1ValuePP),
		kv("struct1ValuePPP", struct1ValuePPP),
		kv("struct2Value", struct2Value),
		kv("struct2ValueP", struct2ValueP),
		kv("struct2ValuePP", struct2ValuePP),
		kv("struct2ValuePPP", struct2ValuePPP),
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

	str := desc("test desc func",
		kv("func1", func1),
		kv("func2", func2),
		kv("func3", func3),
		kv("func4", func4),
		kv("func5", func5),
	)
	t.Log(str)
}

func TestDescChan(t *testing.T) {
	var chan1 chan int
	chan2 := make(chan string)
	chan3 := make(chan<- any)
	chan4 := make(<-chan any)

	str := desc("test desc chan",
		kv("chan1", chan1),
		kv("chan2", chan2),
		kv("chan3", chan3),
		kv("chan4", chan4),
	)
	t.Log(str)
}

func TestDescUnsafePointer(t *testing.T) {
	var unsafePointer1 unsafe.Pointer

	intValue := 1
	unsafePointer2 := unsafe.Pointer(&intValue)

	str := desc("test desc unsafe pointer",
		kv("unsafePointer1", unsafePointer1),
		kv("unsafePointer2", unsafePointer2),
	)
	t.Log(str)
}

func TestDescUintptr(t *testing.T) {
	var uintptr1 uintptr
	uintptr2 := uintptr(0)
	uintptr3 := uintptr(unsafe.Pointer(&uintptr2))

	str := desc("test desc uintptr",
		kv("uintptr1", uintptr1),
		kv("uintptr2", uintptr2),
		kv("uintptr3", uintptr3),
	)
	t.Log(str)
}
