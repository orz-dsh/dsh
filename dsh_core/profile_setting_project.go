package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
)

// region profileProjectSetting

type profileProjectSetting struct {
	Name           string
	Path           string
	Match          string
	ImportSettings projectImportSettingSet
	SourceSettings projectSourceSettingSet
	match          *EvalExpr
}

type profileProjectSettingSet []*profileProjectSetting

func newProfileProjectSetting(name string, path string, match string, importSettings projectImportSettingSet, sourceSettings projectSourceSettingSet, matchObj *EvalExpr) *profileProjectSetting {
	if importSettings == nil {
		importSettings = projectImportSettingSet{}
	}
	if sourceSettings == nil {
		sourceSettings = projectSourceSettingSet{}
	}
	return &profileProjectSetting{
		Name:           name,
		Path:           path,
		Match:          match,
		ImportSettings: importSettings,
		SourceSettings: sourceSettings,
		match:          matchObj,
	}
}

func (s *profileProjectSetting) inspect() *ProfileProjectSettingInspection {
	return newProfileProjectSettingInspection(s.Name, s.Path, s.Match, s.ImportSettings.inspect(), s.SourceSettings.inspect())
}

func (s profileProjectSettingSet) getProjectSettings(evaluator *Evaluator) (projectSettingSet, error) {
	result := projectSettingSet{}
	for i := len(s) - 1; i >= 0; i-- {
		setting := s[i]
		matched, err := evaluator.EvalBoolExpr(setting.match)
		if err != nil {
			return nil, errW(err, "get profile project settings error",
				reason("eval expr error"),
				kv("setting", setting),
			)
		}
		if !matched {
			continue
		}

		rawPath, err := evaluator.EvalStringTemplate(setting.Path)
		if err != nil {
			return nil, errW(err, "get profile project settings error",
				reason("eval template error"),
				kv("setting", setting),
			)
		}
		path, err := filepath.Abs(rawPath)
		if err != nil {
			return nil, errW(err, "get profile project settings error",
				reason("get abs-path error"),
				kv("setting", setting),
				kv("rawPath", rawPath),
			)
		}

		result = append(result, newProjectSetting(setting.Name, path, nil, nil, nil, setting.ImportSettings, setting.SourceSettings))
	}
	return result, nil
}

func (s profileProjectSettingSet) inspect() []*ProfileProjectSettingInspection {
	var inspections []*ProfileProjectSettingInspection
	for i := 0; i < len(s); i++ {
		inspections = append(inspections, s[i].inspect())
	}
	return inspections
}

// endregion

// region profileProjectSettingModel

type profileProjectSettingModel struct {
	Items []*profileProjectItemSettingModel
}

func newProfileProjectSettingModel(items []*profileProjectItemSettingModel) *profileProjectSettingModel {
	return &profileProjectSettingModel{
		Items: items,
	}
}

func (m *profileProjectSettingModel) convert(ctx *modelConvertContext) (profileProjectSettingSet, error) {
	settings := profileProjectSettingSet{}
	for i := 0; i < len(m.Items); i++ {
		if setting, err := m.Items[i].convert(ctx.ChildItem("items", i)); err != nil {
			return nil, err
		} else {
			settings = append(settings, setting)
		}
	}
	return settings, nil
}

// endregion

// region profileProjectItemSettingModel

type profileProjectItemSettingModel struct {
	Name    string
	Path    string
	Match   string
	Imports projectImportSettingModelSet
	Sources projectSourceSettingModelSet
}

func newProfileProjectItemSettingModel(name, path, match string, imports projectImportSettingModelSet, sources projectSourceSettingModelSet) *profileProjectItemSettingModel {
	return &profileProjectItemSettingModel{
		Name:    name,
		Path:    path,
		Match:   match,
		Imports: imports,
		Sources: sources,
	}
}

func (m *profileProjectItemSettingModel) convert(ctx *modelConvertContext) (setting *profileProjectSetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}
	if !projectNameCheckRegex.MatchString(m.Name) {
		return nil, ctx.Child("name").NewValueInvalidError(m.Name)
	}

	if m.Path == "" {
		return nil, ctx.Child("path").NewValueEmptyError()
	}

	var importSettings projectImportSettingSet
	if m.Imports != nil {
		if importSettings, err = m.Imports.convert(ctx.Child("imports")); err != nil {
			return nil, err
		}
	}

	var sourceSettings projectSourceSettingSet
	if m.Sources != nil {
		if sourceSettings, err = m.Sources.convert(ctx.Child("sources")); err != nil {
			return nil, err
		}
	}

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newProfileProjectSetting(m.Name, m.Path, m.Match, importSettings, sourceSettings, matchObj), nil
}

// endregion

// region ProfileProjectSettingInspection

type ProfileProjectSettingInspection struct {
	Name    string                            `yaml:"name" toml:"name" json:"name"`
	Path    string                            `yaml:"path" toml:"path" json:"path"`
	Match   string                            `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
	Imports []*ProjectImportSettingInspection `yaml:"imports,omitempty" toml:"imports,omitempty" json:"imports,omitempty"`
	Sources []*ProjectSourceSettingInspection `yaml:"sources,omitempty" toml:"sources,omitempty" json:"sources,omitempty"`
}

func newProfileProjectSettingInspection(name string, path string, match string, imports []*ProjectImportSettingInspection, sources []*ProjectSourceSettingInspection) *ProfileProjectSettingInspection {
	return &ProfileProjectSettingInspection{
		Name:    name,
		Path:    path,
		Match:   match,
		Imports: imports,
		Sources: sources,
	}
}

// endregion
