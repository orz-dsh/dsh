package dsh_core

import (
	"dsh/dsh_utils"
)

// region profileSetting

type profileSetting struct {
	Option   *profileOptionSetting
	Project  *profileProjectSetting
	Executor *workspaceExecutorSetting
	Registry *workspaceRegistrySetting
	Redirect *workspaceRedirectSetting
}

func newProfileSetting(option *profileOptionSetting, project *profileProjectSetting, executor *workspaceExecutorSetting, registry *workspaceRegistrySetting, redirect *workspaceRedirectSetting) *profileSetting {
	if option == nil {
		option = newProfileOptionSetting(nil)
	}
	if project == nil {
		project = newProfileProjectSetting(nil)
	}
	if executor == nil {
		executor = newWorkspaceExecutorSetting(nil)
	}
	if registry == nil {
		registry = newWorkspaceRegistrySetting(nil)
	}
	if redirect == nil {
		redirect = newWorkspaceRedirectSetting(nil)
	}
	return &profileSetting{
		Option:   option,
		Project:  project,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
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
	if setting, err = model.convert(newModelHelper("profile setting", metadata.File)); err != nil {
		return nil, err
	}
	return setting, nil
}

func loadProfileSettingModel(model *profileSettingModel) (setting *profileSetting, err error) {
	if setting, err = model.convert(newModelHelper("profile setting", "")); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region profileSettingModel

type profileSettingModel struct {
	Option   *profileOptionSettingModel
	Project  *profileProjectSettingModel
	Executor *workspaceExecutorSettingModel
	Registry *workspaceRegistrySettingModel
	Redirect *workspaceRedirectSettingModel
}

func newProfileSettingModel(option *profileOptionSettingModel, project *profileProjectSettingModel, executor *workspaceExecutorSettingModel, registry *workspaceRegistrySettingModel, redirect *workspaceRedirectSettingModel) *profileSettingModel {
	return &profileSettingModel{
		Option:   option,
		Project:  project,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

func (m *profileSettingModel) convert(helper *modelHelper) (_ *profileSetting, err error) {
	var option *profileOptionSetting
	if m.Option != nil {
		if option, err = m.Option.convert(helper.Child("option")); err != nil {
			return nil, err
		}
	}

	var project *profileProjectSetting
	if m.Project != nil {
		if project, err = m.Project.convert(helper.Child("project")); err != nil {
			return nil, err
		}
	}

	var executor *workspaceExecutorSetting
	if m.Executor != nil {
		if executor, err = m.Executor.convert(helper.Child("executor")); err != nil {
			return nil, err
		}
	}

	var registry *workspaceRegistrySetting
	if m.Registry != nil {
		if registry, err = m.Registry.convert(helper.Child("registry")); err != nil {
			return nil, err
		}
	}

	var redirect *workspaceRedirectSetting
	if m.Redirect != nil {
		if redirect, err = m.Redirect.convert(helper.Child("redirect")); err != nil {
			return nil, err
		}
	}

	return newProfileSetting(option, project, executor, registry, redirect), nil
}

// endregion
