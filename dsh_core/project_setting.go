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
	Name                string
	Path                string
	RuntimeSetting      *projectRuntimeSetting
	OptionSettings      projectOptionSettingSet
	OptionCheckSettings projectOptionCheckSettingSet
	SourceSettings      projectSourceSettingSet
	ImportSettings      projectImportSettingSet
}

type projectSettingSet []*projectSetting

func newProjectSetting(name string, path string, runtimeSetting *projectRuntimeSetting, optionSettings projectOptionSettingSet, optionCheckSettings projectOptionCheckSettingSet, importSettings projectImportSettingSet, sourceSettings projectSourceSettingSet) *projectSetting {
	if runtimeSetting == nil {
		runtimeSetting = newProjectRuntimeSetting("", "")
	}
	if optionSettings == nil {
		optionSettings = projectOptionSettingSet{}
	}
	if optionCheckSettings == nil {
		optionCheckSettings = projectOptionCheckSettingSet{}
	}
	if importSettings == nil {
		importSettings = projectImportSettingSet{}
	}
	if sourceSettings == nil {
		sourceSettings = projectSourceSettingSet{}
	}
	return &projectSetting{
		Name:                name,
		Path:                path,
		RuntimeSetting:      runtimeSetting,
		OptionSettings:      optionSettings,
		OptionCheckSettings: optionCheckSettings,
		ImportSettings:      importSettings,
		SourceSettings:      sourceSettings,
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
	Imports projectImportSettingModelSet
	Sources projectSourceSettingModelSet
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
	var optionCheckSettings projectOptionCheckSettingSet
	if m.Option != nil {
		if optionSettings, optionCheckSettings, err = m.Option.convert(ctx.Child("option")); err != nil {
			return nil, err
		}
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

	return newProjectSetting(m.Name, projectPath, runtimeSetting, optionSettings, optionCheckSettings, importSettings, sourceSettings), nil
}

// endregion
