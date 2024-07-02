package setting

import . "github.com/orz-dsh/dsh/utils"

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
		clean = workspaceCleanSettingDefault
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
	file := ""
	if metadata != nil {
		file = metadata.File
	}
	if setting, err = model.Convert(NewModelHelper(logger, "workspace setting", file)); err != nil {
		return nil, err
	}
	return setting, nil
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

func (s *WorkspaceSettingModel) Convert(helper *ModelHelper) (_ *WorkspaceSetting, err error) {
	var clean *WorkspaceCleanSetting
	if s.Clean != nil {
		if clean, err = s.Clean.Convert(helper.Child("clean")); err != nil {
			return nil, err
		}
	}

	var profile *WorkspaceProfileSetting
	if s.Profile != nil {
		if profile, err = s.Profile.Convert(helper.Child("profile")); err != nil {
			return nil, err
		}
	}

	var executor *ExecutorSetting
	if s.Executor != nil {
		if executor, err = s.Executor.Convert(helper.Child("executor")); err != nil {
			return nil, err
		}
	}

	var registry *RegistrySetting
	if s.Registry != nil {
		if registry, err = s.Registry.Convert(helper.Child("registry")); err != nil {
			return nil, err
		}
	}

	var redirect *RedirectSetting
	if s.Redirect != nil {
		if redirect, err = s.Redirect.Convert(helper.Child("redirect")); err != nil {
			return nil, err
		}
	}

	return NewWorkspaceSetting(clean, profile, executor, registry, redirect), nil
}

// endregion
