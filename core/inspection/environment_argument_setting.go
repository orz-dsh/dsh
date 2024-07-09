package inspection

// region EnvironmentArgumentSettingInspection

type EnvironmentArgumentSettingInspection struct {
	Items []*EnvironmentArgumentItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewEnvironmentArgumentSettingInspection(items []*EnvironmentArgumentItemSettingInspection) *EnvironmentArgumentSettingInspection {
	return &EnvironmentArgumentSettingInspection{
		Items: items,
	}
}

// endregion

// region EnvironmentArgumentItemSettingInspection

type EnvironmentArgumentItemSettingInspection struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Value string `yaml:"value" toml:"value" json:"value"`
}

func NewEnvironmentArgumentItemSettingInspection(name, value string) *EnvironmentArgumentItemSettingInspection {
	return &EnvironmentArgumentItemSettingInspection{
		Name:  name,
		Value: value,
	}
}

// endregion
