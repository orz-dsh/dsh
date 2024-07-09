package inspection

// region EnvironmentSettingInspection

type EnvironmentSettingInspection struct {
	Argument  *EnvironmentArgumentSettingInspection  `yaml:"argument,omitempty" toml:"argument,omitempty" json:"argument,omitempty"`
	Workspace *EnvironmentWorkspaceSettingInspection `yaml:"workspace,omitempty" toml:"workspace,omitempty" json:"workspace,omitempty"`
}

func NewEnvironmentSettingInspection(argument *EnvironmentArgumentSettingInspection, workspace *EnvironmentWorkspaceSettingInspection) *EnvironmentSettingInspection {
	return &EnvironmentSettingInspection{
		Argument:  argument,
		Workspace: workspace,
	}
}

// endregion
