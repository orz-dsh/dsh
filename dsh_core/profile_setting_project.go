package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
)

// region profileProjectSetting

type profileProjectSetting struct {
	Name                 string
	Path                 string
	Match                string
	ScriptSourceSettings projectSourceSettingSet
	ScriptImportSettings projectImportSettingSet
	ConfigSourceSettings projectSourceSettingSet
	ConfigImportSettings projectImportSettingSet
	match                *EvalExpr
}

type profileProjectSettingSet []*profileProjectSetting

func newProfileProjectSetting(name string, path string, match string, scriptSourceSettings projectSourceSettingSet, scriptImportSettings projectImportSettingSet, configSourceSettings projectSourceSettingSet, configImportSettings projectImportSettingSet, matchObj *EvalExpr) *profileProjectSetting {
	if scriptSourceSettings == nil {
		scriptSourceSettings = projectSourceSettingSet{}
	}
	if scriptImportSettings == nil {
		scriptImportSettings = projectImportSettingSet{}
	}
	if configSourceSettings == nil {
		configSourceSettings = projectSourceSettingSet{}
	}
	if configImportSettings == nil {
		configImportSettings = projectImportSettingSet{}
	}
	return &profileProjectSetting{
		Name:                 name,
		Path:                 path,
		Match:                match,
		ScriptSourceSettings: scriptSourceSettings,
		ScriptImportSettings: scriptImportSettings,
		ConfigSourceSettings: configSourceSettings,
		ConfigImportSettings: configImportSettings,
		match:                matchObj,
	}
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

		result = append(result, newProjectSetting(setting.Name, path, nil, nil, nil, setting.ScriptSourceSettings, setting.ScriptImportSettings, setting.ConfigSourceSettings, setting.ConfigImportSettings))
	}
	return result, nil
}

// endregion

// region ProfileProjectSettingModel

type ProfileProjectSettingModel struct {
	Items []*ProfileProjectItemSettingModel
}

func NewProfileProjectSettingModel(items []*ProfileProjectItemSettingModel) *ProfileProjectSettingModel {
	return &ProfileProjectSettingModel{
		Items: items,
	}
}

func (m *ProfileProjectSettingModel) convert(ctx *ModelConvertContext) (profileProjectSettingSet, error) {
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

// region ProfileProjectItemSettingModel

type ProfileProjectItemSettingModel struct {
	Name   string
	Path   string
	Match  string
	Script *projectScriptSettingModel
	Config *projectConfigSettingModel
}

func (m *ProfileProjectItemSettingModel) convert(ctx *ModelConvertContext) (setting *profileProjectSetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}
	if checked := projectNameCheckRegex.MatchString(m.Name); !checked {
		return nil, ctx.Child("name").NewValueInvalidError(m.Name)
	}

	if m.Path == "" {
		return nil, ctx.Child("path").NewValueEmptyError()
	}

	var scriptSourceSettings projectSourceSettingSet
	var scriptImportSettings projectImportSettingSet
	if m.Script != nil {
		if scriptSourceSettings, scriptImportSettings, err = m.Script.convert(ctx.Child("script")); err != nil {
			return nil, err
		}
	}

	var configSourceSettings projectSourceSettingSet
	var configImportSettings projectImportSettingSet
	if m.Config != nil {
		if configSourceSettings, configImportSettings, err = m.Config.convert(ctx.Child("config")); err != nil {
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

	return newProfileProjectSetting(m.Name, m.Path, m.Match, scriptSourceSettings, scriptImportSettings, configSourceSettings, configImportSettings, matchObj), nil
}

// endregion