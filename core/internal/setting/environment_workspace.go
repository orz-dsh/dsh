package setting

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
)

// region EnvironmentWorkspaceSetting

type EnvironmentWorkspaceSetting struct {
	Dir      string
	Clean    *WorkspaceCleanSetting
	Profile  *WorkspaceProfileSetting
	Executor *ExecutorSetting
	Registry *RegistrySetting
	Redirect *RedirectSetting
}

func NewEnvironmentWorkspaceSetting(dir string, clean *WorkspaceCleanSetting, profile *WorkspaceProfileSetting, executor *ExecutorSetting, registry *RegistrySetting, redirect *RedirectSetting) *EnvironmentWorkspaceSetting {
	if clean == nil {
		clean = NewWorkspaceCleanSetting(nil)
	}
	if profile == nil {
		profile = NewWorkspaceProfileSetting(nil)
	}
	if executor == nil {
		executor = NewExecutorSetting(nil)
	}
	if registry == nil {
		registry = NewRegistrySetting(nil)
	}
	if redirect == nil {
		redirect = NewRedirectSetting(nil)
	}
	return &EnvironmentWorkspaceSetting{
		Dir:      dir,
		Clean:    clean,
		Profile:  profile,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

func (s *EnvironmentWorkspaceSetting) GetWorkspaceSetting() *WorkspaceSetting {
	return NewWorkspaceSetting(s.Clean, s.Profile, s.Executor, s.Registry, s.Redirect)
}

func (s *EnvironmentWorkspaceSetting) Inspect() *EnvironmentWorkspaceSettingInspection {
	return NewEnvironmentWorkspaceSettingInspection(
		s.Dir,
		s.Clean.Inspect(),
		s.Profile.Inspect(),
		s.Executor.Inspect(),
		s.Registry.Inspect(),
		s.Redirect.Inspect(),
	)
}

// endregion

// region EnvironmentWorkspaceSettingModel

type EnvironmentWorkspaceSettingModel struct {
	Dir      string                        `yaml:"dir,omitempty" toml:"dir,omitempty" json:"dir,omitempty"`
	Clean    *WorkspaceCleanSettingModel   `yaml:"clean,omitempty" toml:"clean,omitempty" json:"clean,omitempty"`
	Profile  *WorkspaceProfileSettingModel `yaml:"profile,omitempty" toml:"profile,omitempty" json:"profile,omitempty"`
	Executor *ExecutorSettingModel         `yaml:"executor,omitempty" toml:"executor,omitempty" json:"executor,omitempty"`
	Registry *RegistrySettingModel         `yaml:"registry,omitempty" toml:"registry,omitempty" json:"registry,omitempty"`
	Redirect *RedirectSettingModel         `yaml:"redirect,omitempty" toml:"redirect,omitempty" json:"redirect,omitempty"`
}

func NewEnvironmentWorkspaceSettingModel(dir string, clean *WorkspaceCleanSettingModel, profile *WorkspaceProfileSettingModel, executor *ExecutorSettingModel, registry *RegistrySettingModel, redirect *RedirectSettingModel) *EnvironmentWorkspaceSettingModel {
	return &EnvironmentWorkspaceSettingModel{
		Dir:      dir,
		Clean:    clean,
		Profile:  profile,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

func (m *EnvironmentWorkspaceSettingModel) Convert(helper *ModelHelper) (_ *EnvironmentWorkspaceSetting, err error) {
	var clean *WorkspaceCleanSetting
	if m.Clean != nil {
		if clean, err = m.Clean.Convert(helper.Child("clean")); err != nil {
			return nil, err
		}
	}

	var profile *WorkspaceProfileSetting
	if m.Profile != nil {
		if profile, err = m.Profile.Convert(helper.Child("profile")); err != nil {
			return nil, err
		}
	}

	var executor *ExecutorSetting
	if m.Executor != nil {
		if executor, err = m.Executor.Convert(helper.Child("executor")); err != nil {
			return nil, err
		}
	}

	var registry *RegistrySetting
	if m.Registry != nil {
		if registry, err = m.Registry.Convert(helper.Child("registry")); err != nil {
			return nil, err
		}
	}

	var redirect *RedirectSetting
	if m.Redirect != nil {
		if redirect, err = m.Redirect.Convert(helper.Child("redirect")); err != nil {
			return nil, err
		}
	}

	return NewEnvironmentWorkspaceSetting(m.Dir, clean, profile, executor, registry, redirect), nil
}

// endregion
