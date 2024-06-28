package core

import (
	"time"
)

// region workspaceSettingModel

type workspaceSettingModel struct {
	Clean    *workspaceCleanSettingModel   `yaml:"clean,omitempty" toml:"clean,omitempty" json:"clean,omitempty"`
	Profile  *workspaceProfileSettingModel `yaml:"profile,omitempty" toml:"profile,omitempty" json:"profile,omitempty"`
	Executor *executorSettingModel         `yaml:"executor,omitempty" toml:"executor,omitempty" json:"executor,omitempty"`
	Registry *registrySettingModel         `yaml:"registry,omitempty" toml:"registry,omitempty" json:"registry,omitempty"`
	Redirect *redirectSettingModel         `yaml:"redirect,omitempty" toml:"redirect,omitempty" json:"redirect,omitempty"`
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

	var executor *executorSetting
	if s.Executor != nil {
		if executor, err = s.Executor.convert(helper.Child("executor")); err != nil {
			return nil, err
		}
	}

	var registry *registrySetting
	if s.Registry != nil {
		if registry, err = s.Registry.convert(helper.Child("registry")); err != nil {
			return nil, err
		}
	}

	var redirect *redirectSetting
	if s.Redirect != nil {
		if redirect, err = s.Redirect.convert(helper.Child("redirect")); err != nil {
			return nil, err
		}
	}

	return newWorkspaceSetting(clean, profile, executor, registry, redirect), nil
}

// endregion

// region workspaceCleanSettingModel

type workspaceCleanSettingModel struct {
	Output *workspaceCleanOutputSettingModel `yaml:"output,omitempty" toml:"output,omitempty" json:"output,omitempty"`
}

func (m *workspaceCleanSettingModel) convert(helper *modelHelper) (*workspaceCleanSetting, error) {
	if m.Output != nil {
		outputCount, outputExpires, err := m.Output.convert(helper.Child("output"))
		if err != nil {
			return nil, err
		}
		return newWorkspaceCleanSetting(outputCount, outputExpires), nil
	}
	return workspaceCleanSettingDefault, nil
}

// endregion

// region workspaceCleanOutputSettingModel

type workspaceCleanOutputSettingModel struct {
	Count   *int   `yaml:"count,omitempty" toml:"count,omitempty" json:"count,omitempty"`
	Expires string `yaml:"expires,omitempty" toml:"expires,omitempty" json:"expires,omitempty"`
}

func (m *workspaceCleanOutputSettingModel) convert(helper *modelHelper) (int, time.Duration, error) {
	count := workspaceCleanSettingDefault.OutputCount
	if m.Count != nil {
		value := *m.Count
		if value <= 0 {
			return 0, 0, helper.Child("count").NewValueInvalidError(value)
		}
		count = value
	}

	expires := workspaceCleanSettingDefault.OutputExpires
	if m.Expires != "" {
		value, err := time.ParseDuration(m.Expires)
		if err != nil {
			return 0, 0, helper.Child("expires").WrapValueInvalidError(err, m.Expires)
		}
		expires = value
	}

	return count, expires, nil
}

// endregion

// region workspaceProfileSettingModel

type workspaceProfileSettingModel struct {
	Items []*workspaceProfileItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func (m *workspaceProfileSettingModel) convert(helper *modelHelper) (*workspaceProfileSetting, error) {
	items, err := convertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return newWorkspaceProfileSetting(items), nil
}

// endregion

// region workspaceProfileItemSettingModel

type workspaceProfileItemSettingModel struct {
	File     string `yaml:"file" toml:"file" json:"file"`
	Optional bool   `yaml:"optional,omitempty" toml:"optional,omitempty" json:"optional,omitempty"`
	Match    string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func (m *workspaceProfileItemSettingModel) convert(helper *modelHelper) (*workspaceProfileItemSetting, error) {
	if m.File == "" {
		return nil, helper.Child("file").NewValueEmptyError()
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return newWorkspaceProfileItemSetting(m.File, m.Optional, m.Match, matchObj), nil
}

// endregion
