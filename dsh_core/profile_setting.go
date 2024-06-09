package dsh_core

import (
	"dsh/dsh_utils"
)

// region profileSetting

type profileSetting struct {
	optionSettings                  profileOptionSettingSet
	projectSettings                 profileProjectSettingSet
	workspaceExecutorSettings       workspaceExecutorSettingSet
	workspaceImportRegistrySettings workspaceImportRegistrySettingSet
	workspaceImportRedirectSettings workspaceImportRedirectSettingSet
}

type profileSettingSet []*profileSetting

func newProfileSetting(optionSettings profileOptionSettingSet, projectSettings profileProjectSettingSet, workspaceExecutorSettings workspaceExecutorSettingSet, workspaceImportRegistrySettings workspaceImportRegistrySettingSet, workspaceImportRedirectSettings workspaceImportRedirectSettingSet) *profileSetting {
	if optionSettings == nil {
		optionSettings = profileOptionSettingSet{}
	}
	if projectSettings == nil {
		projectSettings = profileProjectSettingSet{}
	}
	if workspaceExecutorSettings == nil {
		workspaceExecutorSettings = workspaceExecutorSettingSet{}
	}
	if workspaceImportRegistrySettings == nil {
		workspaceImportRegistrySettings = workspaceImportRegistrySettingSet{}
	}
	if workspaceImportRedirectSettings == nil {
		workspaceImportRedirectSettings = workspaceImportRedirectSettingSet{}
	}
	return &profileSetting{
		optionSettings:                  optionSettings,
		projectSettings:                 projectSettings,
		workspaceExecutorSettings:       workspaceExecutorSettings,
		workspaceImportRegistrySettings: workspaceImportRegistrySettings,
		workspaceImportRedirectSettings: workspaceImportRedirectSettings,
	}
}

func loadProfileSetting(path string) (setting *profileSetting, error error) {
	model := &profileSettingModel{}

	metadata, err := dsh_utils.DeserializeFromFile(path, "", model)
	if err != nil {
		return nil, errW(err, "load profile setting error",
			reason("deserialize error"),
			kv("path", path),
		)
	}
	if setting, err = model.convert(newModelConvertContext("profile setting", metadata.Path)); err != nil {
		return nil, err
	}
	return setting, nil
}

func loadProfileSettingModel(model *profileSettingModel) (setting *profileSetting, err error) {
	if setting, err = model.convert(newModelConvertContext("profile setting", "")); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region profileSettingModel

type profileSettingModel struct {
	Option    *profileOptionSettingModel
	Project   *profileProjectSettingModel
	Workspace *profileWorkspaceSettingModel
}

func newProfileSettingModel(option *profileOptionSettingModel, project *profileProjectSettingModel, workspace *profileWorkspaceSettingModel) *profileSettingModel {
	return &profileSettingModel{
		Option:    option,
		Project:   project,
		Workspace: workspace,
	}
}

func (m *profileSettingModel) convert(ctx *modelConvertContext) (setting *profileSetting, err error) {
	var optionSettings profileOptionSettingSet
	if m.Option != nil {
		if optionSettings, err = m.Option.convert(ctx.Child("option")); err != nil {
			return nil, err
		}
	}

	var projectSettings profileProjectSettingSet
	if m.Project != nil {
		if projectSettings, err = m.Project.convert(ctx.Child("project")); err != nil {
			return nil, err
		}
	}

	var workspaceExecutorSettings workspaceExecutorSettingSet
	var workspaceImportRegistrySettings workspaceImportRegistrySettingSet
	var workspaceImportRedirectSettings workspaceImportRedirectSettingSet
	if m.Workspace != nil {
		if workspaceExecutorSettings, workspaceImportRegistrySettings, workspaceImportRedirectSettings, err = m.Workspace.convert(ctx.Child("workspace")); err != nil {
			return nil, err
		}
	}

	return newProfileSetting(optionSettings, projectSettings, workspaceExecutorSettings, workspaceImportRegistrySettings, workspaceImportRedirectSettings), nil
}

// endregion
