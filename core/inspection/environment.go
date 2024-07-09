package inspection

// region EnvironmentInspection

type EnvironmentInspection struct {
	System   *EnvironmentSystemInspection   `yaml:"system,omitempty" toml:"system,omitempty" json:"system,omitempty"`
	Variable *EnvironmentVariableInspection `yaml:"variable,omitempty" toml:"variable,omitempty" json:"variable,omitempty"`
	Setting  *EnvironmentSettingInspection  `yaml:"setting,omitempty" toml:"setting,omitempty" json:"setting,omitempty"`
}

func NewEnvironmentInspection(system *EnvironmentSystemInspection, variable *EnvironmentVariableInspection, setting *EnvironmentSettingInspection) *EnvironmentInspection {
	return &EnvironmentInspection{
		System:   system,
		Variable: variable,
		Setting:  setting,
	}
}

// endregion
