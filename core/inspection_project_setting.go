package core

// region ProjectDependencySettingInspection

type ProjectDependencySettingInspection struct {
	Items []*ProjectDependencyItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProjectDependencySettingInspection(items []*ProjectDependencyItemSettingInspection) *ProjectDependencySettingInspection {
	return &ProjectDependencySettingInspection{
		Items: items,
	}
}

// endregion

// region ProjectDependencyItemSettingInspection

type ProjectDependencyItemSettingInspection struct {
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newProjectDependencyItemSettingInspection(link, match string) *ProjectDependencyItemSettingInspection {
	return &ProjectDependencyItemSettingInspection{
		Link:  link,
		Match: match,
	}
}

// endregion

// region ProjectResourceSettingInspection

type ProjectResourceSettingInspection struct {
	Items []*ProjectResourceItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProjectResourceSettingInspection(items []*ProjectResourceItemSettingInspection) *ProjectResourceSettingInspection {
	return &ProjectResourceSettingInspection{
		Items: items,
	}
}

// endregion

// region ProjectResourceItemSettingInspection

type ProjectResourceItemSettingInspection struct {
	Dir      string   `yaml:"dir" toml:"dir" json:"dir"`
	Includes []string `yaml:"includes,omitempty" toml:"includes,omitempty" json:"includes,omitempty"`
	Excludes []string `yaml:"excludes,omitempty" toml:"excludes,omitempty" json:"excludes,omitempty"`
	Match    string   `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newProjectResourceItemSettingInspection(dir string, includes, excludes []string, match string) *ProjectResourceItemSettingInspection {
	return &ProjectResourceItemSettingInspection{
		Dir:      dir,
		Includes: includes,
		Excludes: excludes,
		Match:    match,
	}
}

// endregion
