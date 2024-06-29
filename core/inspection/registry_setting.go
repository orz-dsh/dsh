package inspection

// region RegistrySettingInspection

type RegistrySettingInspection struct {
	Items []*RegistryItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewRegistrySettingInspection(items []*RegistryItemSettingInspection) *RegistrySettingInspection {
	return &RegistrySettingInspection{
		Items: items,
	}
}

// endregion

// region RegistryItemSettingInspection

type RegistryItemSettingInspection struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewRegistryItemSettingInspection(name, link, match string) *RegistryItemSettingInspection {
	return &RegistryItemSettingInspection{
		Name:  name,
		Link:  link,
		Match: match,
	}
}

// endregion
