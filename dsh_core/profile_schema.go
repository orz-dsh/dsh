package dsh_core

import (
	"path/filepath"
	"regexp"
)

// region option

type profileOptionSchema struct {
	Name  string
	Value string
	Match string
	match *EvalExpr
}

type profileOptionSchemaSet []*profileOptionSchema

var profileOptionNameCheckRegex = regexp.MustCompile("^_?[a-z][a-z0-9_]*$")

func newProfileOptionSchema(name string, value string, match string, matchObj *EvalExpr) *profileOptionSchema {
	return &profileOptionSchema{
		Name:  name,
		Value: value,
		Match: match,
		match: matchObj,
	}
}

func (s profileOptionSchemaSet) getItems(evaluator *Evaluator) (map[string]string, error) {
	items := map[string]string{}
	for i := 0; i < len(s); i++ {
		entity := s[i]
		if _, exist := items[entity.Name]; exist {
			continue
		}
		matched, err := evaluator.EvalBoolExpr(entity.match)
		if err != nil {
			return nil, errW(err, "get profile option specify items error",
				reason("eval expr error"),
				kv("entity", entity),
			)
		}
		if matched {
			items[entity.Name] = entity.Value
		}
	}
	return items, nil
}

// endregion

// region project

type profileProjectSchema struct {
	Name          string
	Path          string
	Match         string
	ScriptSources projectSchemaSourceSet
	ScriptImports projectSchemaImportSet
	ConfigSources projectSchemaSourceSet
	ConfigImports projectSchemaImportSet
	match         *EvalExpr
}

type profileProjectSchemaSet []*profileProjectSchema

func newProfileProjectSchema(name string, path string, match string, scriptSources projectSchemaSourceSet, scriptImports projectSchemaImportSet, configSources projectSchemaSourceSet, configImports projectSchemaImportSet, matchObj *EvalExpr) *profileProjectSchema {
	return &profileProjectSchema{
		Name:          name,
		Path:          path,
		Match:         match,
		ScriptSources: scriptSources,
		ScriptImports: scriptImports,
		ConfigSources: configSources,
		ConfigImports: configImports,
		match:         matchObj,
	}
}

func (s profileProjectSchemaSet) getProjectEntities(evaluator *Evaluator) (projectSchemaSet, error) {
	result := projectSchemaSet{}
	for i := len(s) - 1; i >= 0; i-- {
		entity := s[i]
		matched, err := evaluator.EvalBoolExpr(entity.match)
		if err != nil {
			return nil, errW(err, "get profile project models error",
				reason("eval expr error"),
				kv("entity", entity),
			)
		}
		if !matched {
			continue
		}

		rawPath, err := evaluator.EvalStringTemplate(entity.Path)
		if err != nil {
			return nil, errW(err, "get profile project models error",
				reason("eval template error"),
				kv("entity", entity),
			)
		}
		path, err := filepath.Abs(rawPath)
		if err != nil {
			return nil, errW(err, "get profile project models error",
				reason("get abs-path error"),
				kv("entity", entity),
				kv("rawPath", rawPath),
			)
		}

		result = append(result, newProjectSchema(entity.Name, path, nil, nil, entity.ScriptSources, entity.ScriptImports, entity.ConfigSources, entity.ConfigImports))
	}
	return result, nil
}

// endregion
