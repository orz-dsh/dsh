package inspection

// region ProjectDependencyInspection

type ProjectDependencyInspection struct {
	Items []*ProjectDependencyItemInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewProjectDependencyInspection(items []*ProjectDependencyItemInspection) *ProjectDependencyInspection {
	return &ProjectDependencyInspection{
		Items: items,
	}
}

// endregion

// region ProjectDependencyItemInspection

type ProjectDependencyItemInspection struct {
	Link   string `yaml:"link" toml:"link" json:"link"`
	Dir    string `yaml:"dir" toml:"dir" json:"dir"`
	GitUrl string `yaml:"gitUrl,omitempty" toml:"gitUrl,omitempty" json:"gitUrl,omitempty"`
	GitRef string `yaml:"gitRef,omitempty" toml:"gitRef,omitempty" json:"gitRef,omitempty"`
}

func NewProjectDependencyItemInspection(link, dir, gitUrl, gitRef string) *ProjectDependencyItemInspection {
	return &ProjectDependencyItemInspection{
		Link:   link,
		Dir:    dir,
		GitUrl: gitUrl,
		GitRef: gitRef,
	}
}

// endregion
