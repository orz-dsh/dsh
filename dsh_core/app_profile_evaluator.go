package dsh_core

import (
	"dsh/dsh_utils"
	"github.com/expr-lang/expr/vm"
	"maps"
)

type appProfileEvalData struct {
	WorkingPath   string
	WorkspacePath string
	ProjectPath   string
	ProjectName   string
}

func newAppProfileEvalData(workingPath, workspacePath, projectPath, projectName string) *appProfileEvalData {
	return &appProfileEvalData{
		WorkingPath:   workingPath,
		WorkspacePath: workspacePath,
		ProjectPath:   projectPath,
		ProjectName:   projectName,
	}
}

func (d *appProfileEvalData) newMap() map[string]any {
	return map[string]any{
		"workingPath":   d.WorkingPath,
		"workspacePath": d.WorkspacePath,
		"projectPath":   d.ProjectPath,
		"projectName":   d.ProjectName,
	}
}

func (d *appProfileEvalData) mergeMap(data map[string]any) map[string]any {
	result := d.newMap()
	maps.Copy(result, data)
	return result
}

type appProfileEvaluator struct {
	data *appProfileEvalData
}

func newAppProfileEvaluator(data *appProfileEvalData) *appProfileEvaluator {
	return &appProfileEvaluator{
		data: data,
	}
}

func (e *appProfileEvaluator) evalMatch(match *vm.Program) (bool, error) {
	return dsh_utils.EvalExprReturnBool(match, e.data.newMap())
}

func (e *appProfileEvaluator) evalPath(path string) (string, error) {
	return dsh_utils.EvalStringTemplate(path, e.data.newMap(), nil)
}

func (e *appProfileEvaluator) evalMatchAndPath(match *vm.Program, path string) (_ string, err error) {
	matched := match == nil
	if match != nil {
		matched, err = e.evalMatch(match)
		if err != nil {
			return "", err
		}
	}
	if matched {
		return e.evalPath(path)
	}
	return "", nil
}
