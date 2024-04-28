package dsh_utils

import (
	"encoding/json"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"reflect"
	"strconv"
)

func CompileExpr(code string) (*vm.Program, error) {
	program, err := expr.Compile(code)
	if err != nil {
		return nil, WrapError(err, "expr compile failed", map[string]any{
			"code": code,
		})
	}
	return program, nil
}

func EvalExprReturnBool(program *vm.Program, env map[string]any) (bool, error) {
	result, err := expr.Run(program, env)
	if err != nil {
		return false, WrapError(err, "expr eval failed", map[string]any{
			"program": program.Source().Content(),
			"env":     env,
		})
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
			return false, NewError("unsupported type", map[string]any{
				"result":     result,
				"resultType": reflect.TypeOf(result),
			})
		}
	}
	return false, nil
}

func EvalExprReturnString(program *vm.Program, env map[string]any) (*string, error) {
	result, err := expr.Run(program, env)
	if err != nil {
		return nil, WrapError(err, "expr eval failed", map[string]any{
			"program": program.Source().Content(),
			"env":     env,
		})
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
				return nil, WrapError(err, "expr eval array result marshal json failed", map[string]any{
					"result": result,
				})
			} else {
				str = string(bytes)
			}
		case map[string]any:
			if bytes, err := json.Marshal(result.(map[string]any)); err != nil {
				return nil, WrapError(err, "expr eval map result marshal json failed", map[string]any{
					"result": result,
				})
			} else {
				str = string(bytes)
			}
		default:
			return nil, NewError("unsupported type", map[string]any{
				"result":     result,
				"resultType": reflect.TypeOf(result),
			})
		}
		return &str, nil
	}
	return nil, nil
}
