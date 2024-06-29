package inspection

// region ProjectDependencySettingInspection

type ProjectDependencySettingInspection struct {
	Items []*ProjectDependencyItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewProjectDependencySettingInspection(items []*ProjectDependencyItemSettingInspection) *ProjectDependencySettingInspection {
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

func NewProjectDependencyItemSettingInspection(link, match string) *ProjectDependencyItemSettingInspection {
	return &ProjectDependencyItemSettingInspection{
		Link:  link,
		Match: match,
	}
}

// endregion
