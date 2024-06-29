package inspection

// region ProjectResourceSettingInspection

type ProjectResourceSettingInspection struct {
	Items []*ProjectResourceItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewProjectResourceSettingInspection(items []*ProjectResourceItemSettingInspection) *ProjectResourceSettingInspection {
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

func NewProjectResourceItemSettingInspection(dir string, includes, excludes []string, match string) *ProjectResourceItemSettingInspection {
	return &ProjectResourceItemSettingInspection{
		Dir:      dir,
		Includes: includes,
		Excludes: excludes,
		Match:    match,
	}
}

// endregion
