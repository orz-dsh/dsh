package dsh_core

// region ProfileOptionSettingInspection

type ProfileOptionSettingInspection struct {
	Items []*ProfileOptionItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProfileOptionSettingInspection(items []*ProfileOptionItemSettingInspection) *ProfileOptionSettingInspection {
	return &ProfileOptionSettingInspection{
		Items: items,
	}
}

// endregion

// region ProfileOptionItemSettingInspection

type ProfileOptionItemSettingInspection struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Value string `yaml:"value" toml:"value" json:"value"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newProfileOptionItemSettingInspection(name, value, match string) *ProfileOptionItemSettingInspection {
	return &ProfileOptionItemSettingInspection{
		Name:  name,
		Value: value,
		Match: match,
	}
}

// endregion

// region ProfileProjectSettingInspection

type ProfileProjectSettingInspection struct {
	Items []*ProfileProjectItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProfileProjectSettingInspection(items []*ProfileProjectItemSettingInspection) *ProfileProjectSettingInspection {
	return &ProfileProjectSettingInspection{
		Items: items,
	}
}

// endregion

// region ProfileProjectItemSettingInspection

type ProfileProjectItemSettingInspection struct {
	Name       string                              `yaml:"name" toml:"name" json:"name"`
	Path       string                              `yaml:"path" toml:"path" json:"path"`
	Match      string                              `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
	Dependency *ProjectDependencySettingInspection `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *ProjectResourceSettingInspection   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func newProfileProjectItemSettingInspection(name string, path string, match string, dependency *ProjectDependencySettingInspection, resource *ProjectResourceSettingInspection) *ProfileProjectItemSettingInspection {
	return &ProfileProjectItemSettingInspection{
		Name:       name,
		Path:       path,
		Match:      match,
		Dependency: dependency,
		Resource:   resource,
	}
}

// endregion
