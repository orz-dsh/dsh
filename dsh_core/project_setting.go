package dsh_core

import (
	"dsh/dsh_utils"
	"regexp"
)

// region base

var projectNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9-]*[a-z0-9]$")

// endregion

// region projectSetting

type projectSetting struct {
	Name                 string
	Path                 string
	RuntimeSetting       *projectRuntimeSetting
	OptionSettings       projectOptionSettingSet
	OptionVerifySettings projectOptionVerifySettingSet
	ScriptSourceSettings projectSourceSettingSet
	ScriptImportSettings projectImportSettingSet
	ConfigSourceSettings projectSourceSettingSet
	ConfigImportSettings projectImportSettingSet
}

type projectSettingSet []*projectSetting

func newProjectSetting(name string, path string, runtimeSetting *projectRuntimeSetting, optionSettings projectOptionSettingSet, optionVerifySettings projectOptionVerifySettingSet, scriptSourceSettings projectSourceSettingSet, scriptImportSettings projectImportSettingSet, configSourceSettings projectSourceSettingSet, configImportSettings projectImportSettingSet) *projectSetting {
	if runtimeSetting == nil {
		runtimeSetting = newProjectRuntimeSetting("", "")
	}
	if optionSettings == nil {
		optionSettings = projectOptionSettingSet{}
	}
	if optionVerifySettings == nil {
		optionVerifySettings = projectOptionVerifySettingSet{}
	}
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
	return &projectSetting{
		Name:                 name,
		Path:                 path,
		RuntimeSetting:       runtimeSetting,
		OptionSettings:       optionSettings,
		OptionVerifySettings: optionVerifySettings,
		ScriptSourceSettings: scriptSourceSettings,
		ScriptImportSettings: scriptImportSettings,
		ConfigSourceSettings: configSourceSettings,
		ConfigImportSettings: configImportSettings,
	}
}

func loadProjectSetting(path string) (setting *projectSetting, err error) {
	model := &projectSettingModel{}
	metadata, err := dsh_utils.DeserializeFromDir(path, []string{"project"}, model, true)
	if err != nil {
		return nil, errW(err, "load project setting error",
			reason("deserialize error"),
			kv("path", path),
		)
	}
	if setting, err = model.convert(newModelConvertContext("project setting", metadata.Path), path); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region projectSettingModel

type projectSettingModel struct {
	Name    string
	Runtime *projectRuntimeSettingModel
	Option  *projectOptionSettingModel
	Script  *projectScriptSettingModel
	Config  *projectConfigSettingModel
}

func (m *projectSettingModel) convert(ctx *modelConvertContext, projectPath string) (setting *projectSetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}
	if !projectNameCheckRegex.MatchString(m.Name) {
		return nil, ctx.Child("name").NewValueInvalidError(m.Name)
	}
	ctx.AddVariable("projectName", m.Name)

	var runtimeSetting *projectRuntimeSetting
	if m.Runtime != nil {
		if runtimeSetting, err = m.Runtime.convert(ctx.Child("runtime")); err != nil {
			return nil, err
		}
	}

	var optionSettings projectOptionSettingSet
	var optionVerifySettings projectOptionVerifySettingSet
	if m.Option != nil {
		if optionSettings, optionVerifySettings, err = m.Option.convert(ctx.Child("option")); err != nil {
			return nil, err
		}
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

	return newProjectSetting(m.Name, projectPath, runtimeSetting, optionSettings, optionVerifySettings, scriptSourceSettings, scriptImportSettings, configSourceSettings, configImportSettings), nil
}

// endregion
