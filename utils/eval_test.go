package utils

import (
	"testing"
)

func TestEvalBoolExpr(t *testing.T) {
	// return true bool
	result, err := EvalExpr("true", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return true bool:", result)

	// return false bool
	result, err = EvalExpr("false", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return false bool:", result)

	// return non-empty string
	result, err = EvalExpr("'aa'", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty string:", result)

	// return empty string
	result, err = EvalExpr("''", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty string:", result)

	// return non-zero int
	result, err = EvalExpr("1 + 2", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-zero int:", result)

	// return zero int
	result, err = EvalExpr("1 - 1", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return zero int:", result)

	// return non-zero float
	result, err = EvalExpr("1.1 + 2.2", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-zero float:", result)

	// return zero float
	result, err = EvalExpr("1.1 - 1.1", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return zero float:", result)

	// return non-empty array
	result, err = EvalExpr("[1, 2, 3]", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty array:", result)

	// return empty array
	result, err = EvalExpr("[]", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty array:", result)

	// return non-empty map
	result, err = EvalExpr("{'a': 1, 'b': 2}", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty map:", result)

	// return empty map
	result, err = EvalExpr("{}", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty map:", result)

	// return exists variables
	result, err = EvalExpr("exists_variable", map[string]any{
		"exists_variable": 1,
	}, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return exists variables:", result)

	// return non-exists variables
	result, err = EvalExpr("'non_exists_variable' in $env", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-exists variables:", result)

	// return nil
	result, err = EvalExpr("nil", nil, CastTypeBool)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return nil:", result)
}

func TestEvalStringExpr(t *testing.T) {
	// return true bool
	result, err := EvalExpr("true", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return true bool:", result)

	// return false bool
	result, err = EvalExpr("false", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return false bool:", result)

	// return non-empty string
	result, err = EvalExpr("'aa'", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty string:", result)

	// return empty string
	result, err = EvalExpr("''", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty string:", result)

	// return non-zero int
	result, err = EvalExpr("1 + 2", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-zero int:", result)

	// return zero int
	result, err = EvalExpr("1 - 1", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return zero int:", result)

	// return non-zero float
	result, err = EvalExpr("1.1 + 2.2", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-zero float:", result)

	// return zero float
	result, err = EvalExpr("1.1 - 1.1", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return zero float:", result)

	// return non-empty array
	result, err = EvalExpr("[1, 2, 3]", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty array:", result)

	// return empty array
	result, err = EvalExpr("[]", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty array:", result)

	// return non-empty map
	result, err = EvalExpr("{'a': 1, 'b': 2}", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-empty map:", result)

	// return empty map
	result, err = EvalExpr("{}", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return empty map:", result)

	// return exists variables
	result, err = EvalExpr("exists_variable", map[string]any{
		"exists_variable": 1,
	}, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return exists variables:", result)

	// return non-exists variables
	result, err = EvalExpr("'non_exists_variable' in $env", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return non-exists variables:", result)

	// return nil
	result, err = EvalExpr("nil", nil, CastTypeString)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("return nil:", result)
}

func TestEvalExprModifyData(t *testing.T) {
	data := map[string]any{
		"a": 1,
	}
	data["setValue"] = func(k string, v any) any {
		data[k] = v
		return v
	}
	result, err := EvalExpr("setValue(\"a\", 2)", data, CastTypeInteger)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(DescN(
		"eval success",
		KV("result", result),
		KV("data", data),
	))
}

func TestEvaluator(t *testing.T) {
	evaluator0 := NewEvaluator()
	evaluator1 := evaluator0.SetData("g1", map[string]any{
		"g1key1": "g1value1",
		"g1key2": "g1value2",
	})
	evaluator2 := evaluator1.SetData("g2", map[string]any{
		"g2key1": "g2value1",
		"g2key2": "g2value2",
	})
	evaluator3 := evaluator2.SetData("g3", map[string]any{
		"g3key1": "g3value1",
		"g3key2": "g3value2",
		"g3":     "conflict key",
	})
	evaluator4 := evaluator3.SetRoot("g3")
	evaluator5 := evaluator4.SetFunc("test", func(input string) bool {
		return input == "g3value1"
	})

	template1 := "{{test .g3key1}} / {{.g3key1}} / {{.g3key2}} / {{.g3.g3}}"
	template2 := "{{test .g3key2}} / {{.g3key1}} / {{.g3key2}} / {{.g3.g3}}"
	expr1Result, err := evaluator5.EvalBoolExpr("funcs.test(g3key1)")
	if err != nil {
		t.Fatal(err)
	}
	expr2Result, err := evaluator5.EvalBoolExpr("funcs.test(g3key2)")
	if err != nil {
		t.Fatal(err)
	}
	template1Result, err := evaluator5.EvalStringTemplate(template1)
	if err != nil {
		t.Fatal(err)
	}
	template2Result, err := evaluator5.EvalStringTemplate(template2)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(DescN("test evaluator",
		KV("evaluator0", evaluator0),
		KV("evaluator0-map", evaluator0.GetMap(true)),
		KV("evaluator1", evaluator1),
		KV("evaluator1-map", evaluator1.GetMap(true)),
		KV("evaluator2", evaluator2),
		KV("evaluator2-map", evaluator2.GetMap(true)),
		KV("evaluator3", evaluator3),
		KV("evaluator3-map", evaluator3.GetMap(true)),
		KV("evaluator4", evaluator4),
		KV("evaluator4-map", evaluator4.GetMap(true)),
		KV("evaluator5", evaluator5),
		KV("evaluator5-map", evaluator5.GetMap(true)),
		KV("expr1Result", expr1Result),
		KV("expr2Result", expr2Result),
		KV("template1Result", template1Result),
		KV("template2Result", template2Result),
	))
}
