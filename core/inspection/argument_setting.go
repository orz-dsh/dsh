package inspection

// region ArgumentSettingInspection

type ArgumentSettingInspection struct {
	Items []*ArgumentItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewArgumentSettingInspection(items []*ArgumentItemSettingInspection) *ArgumentSettingInspection {
	return &ArgumentSettingInspection{
		Items: items,
	}
}

// endregion

// region ArgumentItemSettingInspection

type ArgumentItemSettingInspection struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Value string `yaml:"value" toml:"value" json:"value"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewArgumentItemSettingInspection(name, value, match string) *ArgumentItemSettingInspection {
	return &ArgumentItemSettingInspection{
		Name:  name,
		Value: value,
		Match: match,
	}
}

// endregion
