package dsh_utils

import (
	"fmt"
	"testing"
)

func TestDesc1(t *testing.T) {
	type testStruct struct {
		a int
		b string
		*testStruct
	}

	fmt.Print(NewDesc("desc title", []DescKeyValue{
		NewDescKeyValue("int", 0),
		NewDescKeyValue("float", 3.4),
		NewDescKeyValue("bool", true),
		NewDescKeyValue("string", "abc"),
		NewDescKeyValue("nil", nil),
		NewDescKeyValue("array", []int{1, 2, 3}),
		NewDescKeyValue("map", map[string]any{
			"a": 1,
			"b": 2,
		}),
		NewDescKeyValue("struct", testStruct{
			a: 1,
			b: "abc",
		}),
		NewDescKeyValue("struct-point", &testStruct{
			a: 1,
			b: "abc",
			testStruct: &testStruct{
				a: 2,
				b: "def",
			},
		}),
		NewDescKeyValue("func", func() {}),
		NewDescKeyValue("desc", NewDesc("nested desc title", []DescKeyValue{
			NewDescKeyValue("int", 0),
		})),
	}))
}
