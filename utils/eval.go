package utils

import (
	"encoding/json"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
)

type EvalExpr = vm.Program

func CompileExpr(content string) (*EvalExpr, error) {
	program, err := expr.Compile(content)
	if err != nil {
		return nil, ErrW(err, "compile expr error", KV("content", content))
	}
	return program, nil
}

func EvalBoolExpr(program *EvalExpr, data map[string]any) (bool, error) {
	result, err := expr.Run(program, data)
	if err != nil {
		return false, ErrW(err, "eval bool expr error",
			Reason("eval expr error"),
			KV("program", program.Source().Content()),
			KV("data", data),
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
			return false, ErrN("eval bool expr error",
				Reason("unsupported result type"),
				KV("result", result),
				KV("resultType", reflect.TypeOf(result)),
			)
		}
	}
	return false, nil
}

func EvalStringExpr(program *EvalExpr, data map[string]any) (*string, error) {
	result, err := expr.Run(program, data)
	if err != nil {
		return nil, ErrW(err, "eval string expr error",
			Reason("eval expr error"),
			KV("program", program.Source().Content()),
			KV("data", data),
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
				return nil, ErrW(err, "eval string expr error",
					Reason("array result marshal json error"),
					KV("result", result),
				)
			} else {
				str = string(bytes)
			}
		case map[string]any:
			if bytes, err := json.Marshal(result.(map[string]any)); err != nil {
				return nil, ErrW(err, "eval string expr error",
					Reason("map result marshal json error"),
					KV("result", result),
				)
			} else {
				str = string(bytes)
			}
		default:
			return nil, ErrN("eval string expr error",
				Reason("unsupported result type"),
				KV("result", result),
				KV("resultType", reflect.TypeOf(result)),
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
		return ErrW(err, "eval file template error",
			Reason("parse template error"),
			KV("inputPath", inputPath),
			KV("libraryPaths", libraryPaths),
		)
	}

	if err = os.MkdirAll(filepath.Dir(outputPath), os.ModePerm); err != nil {
		return ErrW(err, "eval file template error",
			Reason("make target dir error"),
			KV("outputPath", outputPath),
		)
	}

	targetFile, err := os.Create(outputPath)
	if err != nil {
		return ErrW(err, "eval file template error",
			Reason("create target file error"),
			KV("outputPath", outputPath),
		)
	}
	defer targetFile.Close()

	err = tpl.Execute(targetFile, data)
	if err != nil {
		return ErrW(err, "eval file template error",
			Reason("execute template error"),
			KV("inputPath", inputPath),
			KV("libraryPaths", libraryPaths),
			KV("outputPath", outputPath),
			KV("data", data),
			KV("funcs", funcs),
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
		return "", ErrW(err, "eval string template error",
			Reason("parse template error"),
			KV("str", str),
			KV("data", data),
			KV("funcs", funcs),
		)
	}
	var writer strings.Builder
	err = tpl.Execute(&writer, data)
	if err != nil {
		return "", ErrW(err, "eval string template error",
			Reason("execute template error"),
			KV("str", str),
			KV("data", data),
			KV("funcs", funcs),
		)
	}
	return strings.TrimSpace(writer.String()), nil
}

type EvalData map[string]any

type EvalDataset map[string]EvalData

func (ds EvalDataset) SetData(name string, data map[string]any) EvalDataset {
	dataset := EvalDataset{}
	maps.Copy(dataset, ds)
	dataset[name] = data
	return dataset
}

func (ds EvalDataset) MergeDataset(dataset EvalDataset) EvalDataset {
	result := EvalDataset{}
	maps.Copy(result, ds)
	maps.Copy(result, dataset)
	return result
}

func (ds EvalDataset) ToMap(root string, funcs EvalFuncs) map[string]any {
	result := map[string]any{}
	if funcs != nil {
		result["funcs"] = funcs
	}
	for k, v := range ds {
		result[k] = v
	}
	if root != "" {
		for k, v := range ds[root] {
			if _, exist := result[k]; !exist {
				result[k] = v
			}
		}
	}
	return result
}

type EvalFuncs map[string]any

func (fs EvalFuncs) SetFunc(name string, fn any) EvalFuncs {
	result := EvalFuncs{}
	maps.Copy(result, fs)
	result[name] = fn
	return result
}

func (fs EvalFuncs) MergeFuncs(funcs EvalFuncs) EvalFuncs {
	result := EvalFuncs{}
	maps.Copy(result, fs)
	maps.Copy(result, funcs)
	return result
}

func (fs EvalFuncs) ToTemplateFuncMap() template.FuncMap {
	result := template.FuncMap{}
	maps.Copy(result, fs)
	return result
}

type Evaluator struct {
	root    string
	dataset EvalDataset
	funcs   EvalFuncs
}

func NewEvaluator() *Evaluator {
	return &Evaluator{
		root:    "",
		dataset: EvalDataset{},
		funcs:   EvalFuncs{},
	}
}

func (e *Evaluator) SetRoot(name string) *Evaluator {
	return &Evaluator{
		root:    name,
		dataset: e.dataset,
		funcs:   e.funcs,
	}
}

func (e *Evaluator) ClearRoot() *Evaluator {
	return e.SetRoot("")
}

func (e *Evaluator) SetData(name string, data map[string]any) *Evaluator {
	return &Evaluator{
		root:    e.root,
		dataset: e.dataset.SetData(name, data),
		funcs:   e.funcs,
	}
}

func (e *Evaluator) SetRootData(name string, data map[string]any) *Evaluator {
	return &Evaluator{
		root:    name,
		dataset: e.dataset.SetData(name, data),
		funcs:   e.funcs,
	}
}

func (e *Evaluator) SetDataset(dataset EvalDataset) *Evaluator {
	return &Evaluator{
		root:    e.root,
		dataset: dataset,
		funcs:   e.funcs,
	}
}

func (e *Evaluator) MergeDataset(dataset EvalDataset) *Evaluator {
	return &Evaluator{
		root:    e.root,
		dataset: e.dataset.MergeDataset(dataset),
		funcs:   e.funcs,
	}
}

func (e *Evaluator) SetFunc(name string, fn any) *Evaluator {
	return &Evaluator{
		root:    e.root,
		dataset: e.dataset,
		funcs:   e.funcs.SetFunc(name, fn),
	}
}

func (e *Evaluator) SetFuncs(funcs EvalFuncs) *Evaluator {
	return &Evaluator{
		root:    e.root,
		dataset: e.dataset,
		funcs:   funcs,
	}
}

func (e *Evaluator) MergeFuncs(funcs EvalFuncs) *Evaluator {
	return &Evaluator{
		root:    e.root,
		dataset: e.dataset,
		funcs:   e.funcs.MergeFuncs(funcs),
	}
}

func (e *Evaluator) GetMap(includeFuncs bool) map[string]any {
	return e.dataset.ToMap(e.root, ValT(includeFuncs, e.funcs, nil))
}

func (e *Evaluator) GetFieldValue(name, field string) any {
	if data, exist := e.dataset[name]; exist {
		if value, exist := data[field]; exist {
			return value
		}
	}
	return nil
}

func (e *Evaluator) GetFieldString(name, field string) string {
	value := e.GetFieldValue(name, field)
	if value != nil {
		return value.(string)
	}
	return ""
}

func (e *Evaluator) DescExtraKeyValues() KVS {
	return KVS{
		KV("root", e.root),
		KV("dataset", e.dataset),
		KV("funcs", e.funcs),
	}
}

func (e *Evaluator) EvalBoolExpr(expr *EvalExpr) (bool, error) {
	if expr == nil {
		return true, nil
	}
	return EvalBoolExpr(expr, e.GetMap(true))
}

func (e *Evaluator) EvalStringExpr(expr *EvalExpr) (*string, error) {
	if expr == nil {
		return nil, nil
	}
	return EvalStringExpr(expr, e.GetMap(true))
}

func (e *Evaluator) EvalFileTemplate(inputPath string, libraryPaths []string, outputPath string) error {
	return EvalFileTemplate(inputPath, libraryPaths, outputPath, e.GetMap(false), e.funcs.ToTemplateFuncMap())
}

func (e *Evaluator) EvalStringTemplate(str string) (string, error) {
	return EvalStringTemplate(str, e.GetMap(false), e.funcs.ToTemplateFuncMap())
}
