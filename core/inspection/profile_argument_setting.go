package inspection

// region ProfileArgumentSettingInspection

type ProfileArgumentSettingInspection struct {
	Items []*ProfileArgumentItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewProfileArgumentSettingInspection(items []*ProfileArgumentItemSettingInspection) *ProfileArgumentSettingInspection {
	return &ProfileArgumentSettingInspection{
		Items: items,
	}
}

// endregion

// region ProfileArgumentItemSettingInspection

type ProfileArgumentItemSettingInspection struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Value string `yaml:"value" toml:"value" json:"value"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewProfileArgumentItemSettingInspection(name, value, match string) *ProfileArgumentItemSettingInspection {
	return &ProfileArgumentItemSettingInspection{
		Name:  name,
		Value: value,
		Match: match,
	}
}

// endregion
