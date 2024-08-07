package setting

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
)

// region WorkspaceSetting

type WorkspaceSetting struct {
	Clean    *WorkspaceCleanSetting
	Profile  *WorkspaceProfileSetting
	Executor *ExecutorSetting
	Registry *RegistrySetting
	Redirect *RedirectSetting
}

func NewWorkspaceSetting(clean *WorkspaceCleanSetting, profile *WorkspaceProfileSetting, executor *ExecutorSetting, registry *RegistrySetting, redirect *RedirectSetting) *WorkspaceSetting {
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
	return &WorkspaceSetting{
		Clean:    clean,
		Profile:  profile,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

func LoadWorkspaceSetting(logger *Logger, dir string) (setting *WorkspaceSetting, err error) {
	model := &WorkspaceSettingModel{}
	metadata, err := DeserializeDir(dir, []string{"workspace"}, model, false)
	if err != nil {
		return nil, ErrW(err, "load workspace setting error",
			Reason("deserialize error"),
			KV("dir", dir),
		)
	}
	source := "default"
	if metadata != nil {
		source = metadata.File
	}
	if setting, err = model.Convert(NewModelHelper(logger, "workspace setting", source)); err != nil {
		return nil, err
	}
	return setting, nil
}

func (s *WorkspaceSetting) Merge(other *WorkspaceSetting) {
	s.Clean.Merge(other.Clean)
	s.Profile.Merge(other.Profile)
	s.Executor.Merge(other.Executor)
	s.Registry.Merge(other.Registry)
	s.Redirect.Merge(other.Redirect)
}

func (s *WorkspaceSetting) MergeDefault() {
	s.Clean.MergeDefault()
	s.Executor.MergeDefault()
	s.Registry.MergeDefault()
}

func (s *WorkspaceSetting) Inspect() *WorkspaceSettingInspection {
	return NewWorkspaceSettingInspection(
		s.Clean.Inspect(),
		s.Profile.Inspect(),
		s.Executor.Inspect(),
		s.Registry.Inspect(),
		s.Redirect.Inspect(),
	)
}

// endregion

// region WorkspaceSettingModel

type WorkspaceSettingModel struct {
	Clean    *WorkspaceCleanSettingModel   `yaml:"clean,omitempty" toml:"clean,omitempty" json:"clean,omitempty"`
	Profile  *WorkspaceProfileSettingModel `yaml:"profile,omitempty" toml:"profile,omitempty" json:"profile,omitempty"`
	Executor *ExecutorSettingModel         `yaml:"executor,omitempty" toml:"executor,omitempty" json:"executor,omitempty"`
	Registry *RegistrySettingModel         `yaml:"registry,omitempty" toml:"registry,omitempty" json:"registry,omitempty"`
	Redirect *RedirectSettingModel         `yaml:"redirect,omitempty" toml:"redirect,omitempty" json:"redirect,omitempty"`
}

func NewWorkspaceSettingModel(clean *WorkspaceCleanSettingModel, profile *WorkspaceProfileSettingModel, executor *ExecutorSettingModel, registry *RegistrySettingModel, redirect *RedirectSettingModel) *WorkspaceSettingModel {
	return &WorkspaceSettingModel{
		Clean:    clean,
		Profile:  profile,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

func (m *WorkspaceSettingModel) Convert(helper *ModelHelper) (_ *WorkspaceSetting, err error) {
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

	return NewWorkspaceSetting(clean, profile, executor, registry, redirect), nil
}

// endregion
