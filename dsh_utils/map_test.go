package dsh_utils

import "testing"

func TestMergeMap1(t *testing.T) {
	map1 := map[string]any{
		"key1": "map1 value1",
		"key2": "map1 value2",
		"key3": "map1 value3",
		"subMap1": map[string]any{
			"subKey1": "subMap1 value1",
			"subKey2": "subMap1 value2",
		},
		"subList1": []any{
			"map1 subList1 value1",
			"map1 subList1 value2",
		},
		"subList2": []any{
			"map1 subList2 value1",
			"map1 subList2 value2",
		},
	}
	map2 := map[string]any{
		"key2": "map2 value2",
		"key3": "map2 value3",
		"key4": "map2 value4",
		"subMap1": map[string]any{
			"subKey2": "subMap2 value2",
			"subKey3": "subMap2 value3",
		},
		"subList1": []any{
			"map2 subList1 value1",
			"map2 subList1 value2",
		},
		"subList2": []any{
			"map2 subList2 value1",
			"map2 subList2 value2",
		},
	}
	map3 := map[string]any{
		"subMap1": []any{},
	}

	result1, trace1, err := MapMerge(nil, map1, nil, "map1", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = MapMerge(result1, map2, nil, "map2", trace1)
	if err != nil {
		t.Fatal(err)
	}

	result2, trace2, err := MapMerge(nil, map1, nil, "map1", nil)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = MapMerge(result2, map2, map[string]MapMergeMode{
		"subMap1":  MapMergeModeReplace,
		"subList1": MapMergeModeReplace,
		"subList2": MapMergeModeInsert,
	}, "map2", trace2)

	t.Log(desc("test merge map 1",
		kv("result1", result1),
		kv("trace1", trace1),
		kv("result2", result2),
		kv("trace2", trace2),
	))

	_, _, err = MapMerge(result2, map3, nil, "map3", nil)
	if err != nil {
		t.Log(err)
	}

	_, _, err = MapMerge(result2, map2, map[string]MapMergeMode{
		"subMap1": MapMergeModeInsert,
	}, "map2", trace2)
	if err != nil {
		t.Log(err)
	}
}
