package dsh_utils

import (
	"encoding/json"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"reflect"
	"strconv"
)

func CompileExpr(content string) (*vm.Program, error) {
	program, err := expr.Compile(content)
	if err != nil {
		return nil, errW(err, "compile expr error", kv("content", content))
	}
	return program, nil
}

func EvalExprReturnBool(program *vm.Program, env map[string]any) (bool, error) {
	result, err := expr.Run(program, env)
	if err != nil {
		return false, errW(err, "eval expr return bool error",
			reason("eval expr error"),
			kv("program", program.Source().Content()),
			kv("env", env),
		)
	}
	if result != nil {
		switch result.(type) {
		case bool:
			return result.(bool), nil
		case string:
			return result.(string) != "", nil
		case int:
			return result.(int) != 0, nil
		case uint:
			return result.(uint) != 0, nil
		case float64:
			return result.(float64) != 0, nil
		case []any:
			return len(result.([]any)) > 0, nil
		case map[string]any:
			return len(result.(map[string]any)) > 0, nil
		default:
			return false, errN("eval expr return bool error",
				reason("unsupported result type"),
				kv("result", result),
				kv("resultType", reflect.TypeOf(result)),
			)
		}
	}
	return false, nil
}

func EvalExprReturnString(program *vm.Program, env map[string]any) (*string, error) {
	result, err := expr.Run(program, env)
	if err != nil {
		return nil, errW(err, "eval expr return string error",
			reason("eval expr error"),
			kv("program", program.Source().Content()),
			kv("env", env),
		)
	}
	if result != nil {
		var str string
		switch result.(type) {
		case bool:
			str = strconv.FormatBool(result.(bool))
		case string:
			str = result.(string)
		case int:
			str = strconv.Itoa(result.(int))
		case uint:
			str = strconv.FormatUint(uint64(result.(uint)), 10)
		case float64:
			str = strconv.FormatFloat(result.(float64), 'f', -1, 64)
		case []any:
			if bytes, err := json.Marshal(result.([]any)); err != nil {
				return nil, errW(err, "eval expr return string error",
					reason("array result marshal json error"),
					kv("result", result),
				)
			} else {
				str = string(bytes)
			}
		case map[string]any:
			if bytes, err := json.Marshal(result.(map[string]any)); err != nil {
				return nil, errW(err, "eval expr return string error",
					reason("map result marshal json error"),
					kv("result", result),
				)
			} else {
				str = string(bytes)
			}
		default:
			return nil, errN("eval expr return string error",
				reason("unsupported result type"),
				kv("result", result),
				kv("resultType", reflect.TypeOf(result)),
			)
		}
		return &str, nil
	}
	return nil, nil
}
