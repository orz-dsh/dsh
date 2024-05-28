package dsh_utils

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

func CompileExpr(content string) (*vm.Program, error) {
	program, err := expr.Compile(content)
	if err != nil {
		return nil, errW(err, "compile expr error", kv("content", content))
	}
	return program, nil
}

func EvalBoolExpr(program *vm.Program, data map[string]any) (bool, error) {
	result, err := expr.Run(program, data)
	if err != nil {
		return false, errW(err, "eval bool expr error",
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
			return false, errN("eval bool expr error",
				reason("unsupported result type"),
				kv("result", result),
				kv("resultType", reflect.TypeOf(result)),
			)
		}
	}
	return false, nil
}

func EvalStringExpr(program *vm.Program, data map[string]any) (*string, error) {
	result, err := expr.Run(program, data)
	if err != nil {
		return nil, errW(err, "eval string expr error",
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
				return nil, errW(err, "eval string expr error",
					reason("array result marshal json error"),
					kv("result", result),
				)
			} else {
				str = string(bytes)
			}
		case map[string]any:
			if bytes, err := json.Marshal(result.(map[string]any)); err != nil {
				return nil, errW(err, "eval string expr error",
					reason("map result marshal json error"),
					kv("result", result),
				)
			} else {
				str = string(bytes)
			}
		default:
			return nil, errN("eval string expr error",
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

func (e *Evaluator) ToMap(includeFuncs bool) map[string]any {
	return e.dataset.ToMap(e.root, t(includeFuncs, e.funcs, nil))
}

func (e *Evaluator) DescExtraKeyValues() KVS {
	return KVS{
		kv("root", e.root),
		kv("dataset", e.dataset),
		kv("funcs", e.funcs),
	}
}

func (e *Evaluator) EvalBoolExpr(expr *vm.Program) (bool, error) {
	if expr == nil {
		return true, nil
	}
	return EvalBoolExpr(expr, e.ToMap(true))
}

func (e *Evaluator) EvalStringExpr(expr *vm.Program) (*string, error) {
	if expr == nil {
		return nil, nil
	}
	return EvalStringExpr(expr, e.ToMap(true))
}

func (e *Evaluator) EvalFileTemplate(inputPath string, libraryPaths []string, outputPath string) error {
	return EvalFileTemplate(inputPath, libraryPaths, outputPath, e.ToMap(false), e.funcs.ToTemplateFuncMap())
}

func (e *Evaluator) EvalStringTemplate(str string) (string, error) {
	return EvalStringTemplate(str, e.ToMap(false), e.funcs.ToTemplateFuncMap())
}
