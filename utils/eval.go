package utils

import (
	"github.com/expr-lang/expr"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func EvalExpr(str string, data map[string]any, cast CastType) (any, error) {
	program, err := expr.Compile(str, expr.Env(data))
	if err != nil {
		return nil, ErrW(err, "eval expr error",
			Reason("compile expr error"),
			KV("str", str),
			KV("data", data),
		)
	}
	result, err := expr.Run(program, data)
	if err != nil {
		return nil, ErrW(err, "eval expr error",
			Reason("eval expr error"),
			KV("str", str),
			KV("data", data),
		)
	}
	castResult, err := Cast(result, cast)
	if err != nil {
		return nil, ErrW(err, "eval expr error",
			Reason("cast result error"),
			KV("result", result),
			KV("cast", cast),
		)
	}
	return castResult, nil
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

func (ds EvalDataset) MergeData(name string, data map[string]any) EvalDataset {
	newData := map[string]any{}
	maps.Copy(newData, ds[name])
	maps.Copy(newData, data)
	return ds.SetData(name, newData)
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

func (e *Evaluator) GetData(name string) map[string]any {
	return e.dataset[name]
}

func (e *Evaluator) SetData(name string, data map[string]any) *Evaluator {
	return &Evaluator{
		root:    e.root,
		dataset: e.dataset.SetData(name, data),
		funcs:   e.funcs,
	}
}

func (e *Evaluator) MergeData(name string, data map[string]any) *Evaluator {
	return &Evaluator{
		root:    e.root,
		dataset: e.dataset.MergeData(name, data),
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

func (e *Evaluator) DescExtraKeyValues() KVS {
	return KVS{
		KV("root", e.root),
		KV("dataset", e.dataset),
		KV("funcs", e.funcs),
	}
}

func (e *Evaluator) EvalExpr(expr string, cast CastType) (any, error) {
	if expr == "" {
		return nil, nil
	}
	return EvalExpr(expr, e.GetMap(true), cast)
}

func (e *Evaluator) EvalBoolExpr(expr string) (bool, error) {
	if expr == "" {
		return true, nil
	}
	if result, err := EvalExpr(expr, e.GetMap(true), CastTypeBool); err != nil {
		return false, err
	} else if result == nil {
		return false, nil
	} else {
		return result.(bool), nil
	}
}

func (e *Evaluator) EvalFileTemplate(inputPath string, libraryPaths []string, outputPath string) error {
	return EvalFileTemplate(inputPath, libraryPaths, outputPath, e.GetMap(false), e.funcs.ToTemplateFuncMap())
}

func (e *Evaluator) EvalStringTemplate(str string) (string, error) {
	return EvalStringTemplate(str, e.GetMap(false), e.funcs.ToTemplateFuncMap())
}
