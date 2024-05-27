package dsh_utils

import (
	"testing"
)

func TestEvalExprReturnBool(t *testing.T) {
	// return true bool
	program, err := CompileExpr("true")
	if err != nil {
		t.Fatal(err)
	}
	result, err := EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return true bool:", result)

	// return false bool
	program, err = CompileExpr("false")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return false bool:", result)

	// return non-empty string
	program, err = CompileExpr("'aa'")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty string:", result)

	// return empty string
	program, err = CompileExpr("''")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty string:", result)

	// return non-zero int
	program, err = CompileExpr("1 + 2")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-zero int:", result)

	// return zero int
	program, err = CompileExpr("1 - 1")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return zero int:", result)

	// return non-zero float
	program, err = CompileExpr("1.1 + 2.2")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-zero float:", result)

	// return zero float
	program, err = CompileExpr("1.1 - 1.1")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return zero float:", result)

	// return non-empty array
	program, err = CompileExpr("[1, 2, 3]")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty array:", result)

	// return empty array
	program, err = CompileExpr("[]")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty array:", result)

	// return non-empty map
	program, err = CompileExpr("{'a': 1, 'b': 2}")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty map:", result)

	// return empty map
	program, err = CompileExpr("{}")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty map:", result)

	// return exists variables
	program, err = CompileExpr("exists_variable")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, map[string]any{
		"exists_variable": 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return exists variables:", result)

	// return non-exists variables
	program, err = CompileExpr("non_exists_variable")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-exists variables:", result)

	// return nil
	program, err = CompileExpr("nil")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnBool(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return nil:", result)
}

func TestEvalExprReturnString(t *testing.T) {
	// return true bool
	program, err := CompileExpr("true")
	if err != nil {
		t.Fatal(err)
	}
	result, err := EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return true bool:", *result)

	// return false bool
	program, err = CompileExpr("false")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return false bool:", *result)

	// return non-empty string
	program, err = CompileExpr("'aa'")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty string:", *result)

	// return empty string
	program, err = CompileExpr("''")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty string:", *result)

	// return non-zero int
	program, err = CompileExpr("1 + 2")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-zero int:", *result)

	// return zero int
	program, err = CompileExpr("1 - 1")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return zero int:", *result)

	// return non-zero float
	program, err = CompileExpr("1.1 + 2.2")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-zero float:", *result)

	// return zero float
	program, err = CompileExpr("1.1 - 1.1")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return zero float:", *result)

	// return non-empty array
	program, err = CompileExpr("[1, 2, 3]")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty array:", *result)

	// return empty array
	program, err = CompileExpr("[]")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty array:", *result)

	// return non-empty map
	program, err = CompileExpr("{'a': 1, 'b': 2}")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty map:", *result)

	// return empty map
	program, err = CompileExpr("{}")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty map:", *result)

	// return exists variables
	program, err = CompileExpr("exists_variable")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, map[string]any{
		"exists_variable": 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return exists variables:", *result)

	// return non-exists variables
	program, err = CompileExpr("non_exists_variable")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-exists variables:", result)

	// return nil
	program, err = CompileExpr("nil")
	if err != nil {
		t.Fatal(err)
	}
	result, err = EvalExprReturnString(program, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return nil:", result)
}

func TestEvalExprModifyData(t *testing.T) {
	program, err := CompileExpr("setValue(\"a\", 2)")
	if err != nil {
		t.Fatal(err)
	}
	data := map[string]any{
		"a": 1,
	}
	data["setValue"] = func(k string, v any) any {
		data[k] = v
		return v
	}
	result, err := EvalExprReturnString(program, data)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(desc(
		"eval success",
		kv("result", result),
		kv("data", data),
	))
}

func TestEvalData(t *testing.T) {
	data0 := NewEvalData()
	data1 := data0.Data("g1", map[string]any{
		"g1key1": "g1value1",
		"g1key2": "g1value2",
	})
	data2 := data1.Data("g2", map[string]any{
		"g2key1": "g2value1",
		"g2key2": "g2value2",
	})
	data3 := data2.Data("g3", map[string]any{
		"g3key1": "g3value1",
		"g3key2": "g3value2",
	}).Main("g1")
	t.Log(desc("test eval data",
		kv("data0", data0),
		kv("data0-map", data0.Map()),
		kv("data1", data1),
		kv("data1-map", data1.Map()),
		kv("data2", data2),
		kv("data2-map", data2.Map()),
		kv("data3", data3),
		kv("data3-map", data3.Map()),
	))
}
