package dsh_core

import (
	"dsh/dsh_utils"
)

// region workspaceSetting

type workspaceSetting struct {
	Clean    *workspaceCleanSetting
	Profile  *workspaceProfileSetting
	Executor *workspaceExecutorSetting
	Registry *workspaceRegistrySetting
	Redirect *workspaceRedirectSetting
}

func newWorkspaceSetting(clean *workspaceCleanSetting, profile *workspaceProfileSetting, executor *workspaceExecutorSetting, registry *workspaceRegistrySetting, redirect *workspaceRedirectSetting) *workspaceSetting {
	if clean == nil {
		clean = workspaceCleanSettingDefault
	}
	if profile == nil {
		profile = newWorkspaceProfileSetting(nil)
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
	return &workspaceSetting{
		Clean:    clean,
		Profile:  profile,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

func loadWorkspaceSetting(path string) (setting *workspaceSetting, err error) {
	model := &workspaceSettingModel{}
	metadata, err := dsh_utils.DeserializeFromDir(path, []string{"workspace"}, model, false)
	if err != nil {
		return nil, errW(err, "load workspace setting error",
			reason("deserialize error"),
			kv("path", path),
		)
	}
	file := ""
	if metadata != nil {
		file = metadata.File
	}
	if setting, err = model.convert(newModelHelper("workspace setting", file)); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region workspaceSettingModel

type workspaceSettingModel struct {
	Clean    *workspaceCleanSettingModel
	Profile  *workspaceProfileSettingModel
	Executor *workspaceExecutorSettingModel
	Registry *workspaceRegistrySettingModel
	Redirect *workspaceRedirectSettingModel
}

func (s *workspaceSettingModel) convert(helper *modelHelper) (_ *workspaceSetting, err error) {
	var clean *workspaceCleanSetting
	if s.Clean != nil {
		if clean, err = s.Clean.convert(helper.Child("clean")); err != nil {
			return nil, err
		}
	}

	var profile *workspaceProfileSetting
	if s.Profile != nil {
		if profile, err = s.Profile.convert(helper.Child("profile")); err != nil {
			return nil, err
		}
	}

	var executor *workspaceExecutorSetting
	if s.Executor != nil {
		if executor, err = s.Executor.convert(helper.Child("executor")); err != nil {
			return nil, err
		}
	}

	var registry *workspaceRegistrySetting
	if s.Registry != nil {
		if registry, err = s.Registry.convert(helper.Child("registry")); err != nil {
			return nil, err
		}
	}

	var redirect *workspaceRedirectSetting
	if s.Redirect != nil {
		if redirect, err = s.Redirect.convert(helper.Child("redirect")); err != nil {
			return nil, err
		}
	}

	return newWorkspaceSetting(clean, profile, executor, registry, redirect), nil
}

// endregion
