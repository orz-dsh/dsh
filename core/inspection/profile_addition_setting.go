package inspection

// region ProfileAdditionSettingInspection

type ProfileAdditionSettingInspection struct {
	Items []*ProfileAdditionItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewProfileAdditionSettingInspection(items []*ProfileAdditionItemSettingInspection) *ProfileAdditionSettingInspection {
	return &ProfileAdditionSettingInspection{
		Items: items,
	}
}

// endregion

// region ProfileAdditionItemSettingInspection

type ProfileAdditionItemSettingInspection struct {
	Name       string                              `yaml:"name" toml:"name" json:"name"`
	Dir        string                              `yaml:"dir" toml:"dir" json:"dir"`
	Match      string                              `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
	Dependency *ProjectDependencySettingInspection `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *ProjectResourceSettingInspection   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func NewProfileAdditionItemSettingInspection(name, dir, match string, dependency *ProjectDependencySettingInspection, resource *ProjectResourceSettingInspection) *ProfileAdditionItemSettingInspection {
	return &ProfileAdditionItemSettingInspection{
		Name:       name,
		Dir:        dir,
		Match:      match,
		Dependency: dependency,
		Resource:   resource,
	}
}

// endregion
