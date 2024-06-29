package setting

import . "github.com/orz-dsh/dsh/utils"

// region ProfileSetting

type ProfileSetting struct {
	Argument *ArgumentSetting
	Addition *AdditionSetting
	Executor *ExecutorSetting
	Registry *RegistrySetting
	Redirect *RedirectSetting
}

func NewProfileSetting(argument *ArgumentSetting, addition *AdditionSetting, executor *ExecutorSetting, registry *RegistrySetting, redirect *RedirectSetting) *ProfileSetting {
	if argument == nil {
		argument = NewArgumentSetting(nil)
	}
	if addition == nil {
		addition = NewAdditionSetting(nil)
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
	return &ProfileSetting{
		Argument: argument,
		Addition: addition,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

func LoadProfileSetting(file string) (setting *ProfileSetting, error error) {
	model := &ProfileSettingModel{}
	metadata, err := DeserializeFromFile(file, "", model)
	if err != nil {
		return nil, ErrW(err, "load profile setting error",
			Reason("deserialize error"),
			KV("file", file),
		)
	}
	if setting, err = model.Convert(NewModelHelper("profile setting", metadata.File)); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region ProfileSettingModel

type ProfileSettingModel struct {
	Argument *ArgumentSettingModel `yaml:"argument,omitempty" toml:"argument,omitempty" json:"argument,omitempty"`
	Addition *AdditionSettingModel `yaml:"addition,omitempty" toml:"addition,omitempty" json:"addition,omitempty"`
	Executor *ExecutorSettingModel `yaml:"executor,omitempty" toml:"executor,omitempty" json:"executor,omitempty"`
	Registry *RegistrySettingModel `yaml:"registry,omitempty" toml:"registry,omitempty" json:"registry,omitempty"`
	Redirect *RedirectSettingModel `yaml:"redirect,omitempty" toml:"redirect,omitempty" json:"redirect,omitempty"`
}

func NewProfileSettingModel(argument *ArgumentSettingModel, addition *AdditionSettingModel, executor *ExecutorSettingModel, registry *RegistrySettingModel, redirect *RedirectSettingModel) *ProfileSettingModel {
	return &ProfileSettingModel{
		Argument: argument,
		Addition: addition,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

func (m *ProfileSettingModel) Convert(helper *ModelHelper) (_ *ProfileSetting, err error) {
	var argument *ArgumentSetting
	if m.Argument != nil {
		if argument, err = m.Argument.Convert(helper.Child("argument")); err != nil {
			return nil, err
		}
	}

	var addition *AdditionSetting
	if m.Addition != nil {
		if addition, err = m.Addition.Convert(helper.Child("addition")); err != nil {
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

	return NewProfileSetting(argument, addition, executor, registry, redirect), nil
}

func (m *ProfileSettingModel) GetSetting() (setting *ProfileSetting, err error) {
	if setting, err = m.Convert(NewModelHelper("profile setting", "")); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion
