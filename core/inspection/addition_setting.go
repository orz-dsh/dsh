package inspection

// region AdditionSettingInspection

type AdditionSettingInspection struct {
	Items []*AdditionItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewAdditionSettingInspection(items []*AdditionItemSettingInspection) *AdditionSettingInspection {
	return &AdditionSettingInspection{
		Items: items,
	}
}

// endregion

// region AdditionItemSettingInspection

type AdditionItemSettingInspection struct {
	Name       string                              `yaml:"name" toml:"name" json:"name"`
	Dir        string                              `yaml:"dir" toml:"dir" json:"dir"`
	Match      string                              `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
	Dependency *ProjectDependencySettingInspection `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *ProjectResourceSettingInspection   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func NewAdditionItemSettingInspection(name string, dir string, match string, dependency *ProjectDependencySettingInspection, resource *ProjectResourceSettingInspection) *AdditionItemSettingInspection {
	return &AdditionItemSettingInspection{
		Name:       name,
		Dir:        dir,
		Match:      match,
		Dependency: dependency,
		Resource:   resource,
	}
}

// endregion
