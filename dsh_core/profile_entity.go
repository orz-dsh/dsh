package dsh_core

import (
	"path/filepath"
	"regexp"
)

// region option

type profileOptionSpecifyEntity struct {
	Name  string
	Value string
	Match string
	match *EvalExpr
}

type profileOptionSpecifyEntitySet []*profileOptionSpecifyEntity

var profileOptionNameCheckRegex = regexp.MustCompile("^_?[a-z][a-z0-9_]*$")

func newProfileOptionSpecifyEntity(name string, value string, match string, matchObj *EvalExpr) *profileOptionSpecifyEntity {
	return &profileOptionSpecifyEntity{
		Name:  name,
		Value: value,
		Match: match,
		match: matchObj,
	}
}

func (s profileOptionSpecifyEntitySet) getItems(evaluator *Evaluator) (map[string]string, error) {
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

type profileProjectEntity struct {
	Name          string
	Path          string
	Match         string
	ScriptSources projectSourceEntitySet
	ScriptImports projectImportEntitySet
	ConfigSources projectSourceEntitySet
	ConfigImports projectImportEntitySet
	match         *EvalExpr
}

type profileProjectEntitySet []*profileProjectEntity

func newProfileProjectEntity(name string, path string, match string, scriptSources projectSourceEntitySet, scriptImports projectImportEntitySet, configSources projectSourceEntitySet, configImports projectImportEntitySet, matchObj *EvalExpr) *profileProjectEntity {
	return &profileProjectEntity{
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

func (s profileProjectEntitySet) getProjectEntities(evaluator *Evaluator) (projectEntitySet, error) {
	result := projectEntitySet{}
	for i := 0; i < len(s); i++ {
		entity := s[i]
		matched, err := evaluator.EvalBoolExpr(entity.match)
		if err != nil {
			return nil, errW(err, "get profile project entities error",
				reason("eval expr error"),
				kv("entity", entity),
			)
		}
		if !matched {
			continue
		}

		rawPath, err := evaluator.EvalStringTemplate(entity.Path)
		if err != nil {
			return nil, errW(err, "get profile project entities error",
				reason("eval template error"),
				kv("entity", entity),
			)
		}
		path, err := filepath.Abs(rawPath)
		if err != nil {
			return nil, errW(err, "get profile project entities error",
				reason("get abs-path error"),
				kv("entity", entity),
				kv("rawPath", rawPath),
			)
		}

		result = append(result, newProjectEntity(entity.Name, path, nil, nil, entity.ScriptSources, entity.ScriptImports, entity.ConfigSources, entity.ConfigImports))
	}
	return result, nil
}

// endregion
