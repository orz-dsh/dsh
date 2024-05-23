package dsh_utils

import (
	"encoding/json"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

func CompileExpr(content string) (*vm.Program, error) {
	program, err := expr.Compile(content)
	if err != nil {
		return nil, errW(err, "compile expr error", kv("content", content))
	}
	return program, nil
}

func EvalExprReturnBool(program *vm.Program, data map[string]any) (bool, error) {
	result, err := expr.Run(program, data)
	if err != nil {
		return false, errW(err, "eval expr return bool error",
			reason("eval expr error"),
			kv("program", program.Source().Content()),
			kv("data", data),
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

func EvalExprReturnString(program *vm.Program, data map[string]any) (*string, error) {
	result, err := expr.Run(program, data)
	if err != nil {
		return nil, errW(err, "eval expr return string error",
			reason("eval expr error"),
			kv("program", program.Source().Content()),
			kv("data", data),
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

func EvalFileTemplate(inputPath string, libraryPaths []string, outputPath string, data map[string]any, funcs template.FuncMap) error {
	tpl := template.New(filepath.Base(inputPath)).Option("missingkey=error")
	if funcs != nil {
		tpl = tpl.Funcs(funcs)
	}
	files := append([]string{inputPath}, libraryPaths...)
	tpl, err := tpl.ParseFiles(files...)
	if err != nil {
		return errW(err, "eval file template error",
			reason("parse template error"),
			kv("inputPath", inputPath),
			kv("libraryPaths", libraryPaths),
		)
	}

	if err = os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return errW(err, "eval file template error",
			reason("make target dir error"),
			kv("outputPath", outputPath),
		)
	}

	targetFile, err := os.Create(outputPath)
	if err != nil {
		return errW(err, "eval file template error",
			reason("create target file error"),
			kv("outputPath", outputPath),
		)
	}
	defer targetFile.Close()

	err = tpl.Execute(targetFile, data)
	if err != nil {
		return errW(err, "eval file template error",
			reason("execute template error"),
			kv("inputPath", inputPath),
			kv("libraryPaths", libraryPaths),
			kv("outputPath", outputPath),
			kv("data", data),
			kv("funcs", funcs),
		)
	}
	return nil
}

func EvalStringTemplate(str string, data map[string]any, funcs template.FuncMap) (string, error) {
	tpl := template.New("StringTemplate").Option("missingkey=error")
	if funcs != nil {
		tpl = tpl.Funcs(funcs)
	}
	tpl, err := tpl.Parse(str)
	if err != nil {
		return "", errW(err, "eval string template error",
			reason("parse template error"),
			kv("str", str),
			kv("data", data),
			kv("funcs", funcs),
		)
	}
	var writer strings.Builder
	err = tpl.Execute(&writer, data)
	if err != nil {
		return "", errW(err, "eval string template error",
			reason("execute template error"),
			kv("str", str),
			kv("data", data),
			kv("funcs", funcs),
		)
	}
	return strings.TrimSpace(writer.String()), nil
}

type EvalMatcher struct {
	data map[string]any
}

func (m *EvalMatcher) Match(expr *vm.Program) (bool, error) {
	if expr == nil {
		return true, nil
	}
	return EvalExprReturnBool(expr, m.data)
}

func NewEvalMatcher(data map[string]any) *EvalMatcher {
	return &EvalMatcher{
		data: data,
	}
}
