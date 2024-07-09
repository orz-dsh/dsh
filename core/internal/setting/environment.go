package setting

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
)

// region EnvironmentSetting

type EnvironmentSetting struct {
	Argument  *EnvironmentArgumentSetting
	Workspace *EnvironmentWorkspaceSetting
}

func NewEnvironmentSetting(argument *EnvironmentArgumentSetting, workspace *EnvironmentWorkspaceSetting) *EnvironmentSetting {
	if argument == nil {
		argument = NewEnvironmentArgumentSetting(nil)
	}
	if workspace == nil {
		workspace = NewEnvironmentWorkspaceSetting("", nil, nil, nil, nil, nil)
	}
	return &EnvironmentSetting{
		Argument:  argument,
		Workspace: workspace,
	}
}

func (s *EnvironmentSetting) Inspect() *EnvironmentSettingInspection {
	return NewEnvironmentSettingInspection(s.Argument.Inspect(), s.Workspace.Inspect())
}

// endregion

// region EnvironmentSettingModel

type EnvironmentSettingModel struct {
	Argument  *EnvironmentArgumentSettingModel  `yaml:"argument,omitempty" toml:"argument,omitempty" json:"argument,omitempty"`
	Workspace *EnvironmentWorkspaceSettingModel `yaml:"workspace,omitempty" toml:"workspace,omitempty" json:"workspace,omitempty"`
}

func NewEnvironmentSettingModel(argument *EnvironmentArgumentSettingModel, workspace *EnvironmentWorkspaceSettingModel) *EnvironmentSettingModel {
	return &EnvironmentSettingModel{
		Argument:  argument,
		Workspace: workspace,
	}
}

func (m *EnvironmentSettingModel) Convert(helper *ModelHelper) (_ *EnvironmentSetting, err error) {
	var argument *EnvironmentArgumentSetting
	if m.Argument != nil {
		if argument, err = m.Argument.Convert(helper.Child("argument")); err != nil {
			return nil, err
		}
	}

	var workspace *EnvironmentWorkspaceSetting
	if m.Workspace != nil {
		if workspace, err = m.Workspace.Convert(helper.Child("workspace")); err != nil {
			return nil, err
		}
	}

	return NewEnvironmentSetting(argument, workspace), nil
}

// endregion
