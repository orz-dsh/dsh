package dsh_core

// region ProjectImportInspection

type ProjectImportInspection struct {
	Link   string `yaml:"link" toml:"link" json:"link"`
	Path   string `yaml:"path" toml:"path" json:"path"`
	GitUrl string `yaml:"gitUrl,omitempty" toml:"gitUrl,omitempty" json:"gitUrl,omitempty"`
	GitRef string `yaml:"gitRef,omitempty" toml:"gitRef,omitempty" json:"gitRef,omitempty"`
}

func newProjectImportInspection(link string, path string, gitUrl string, gitRef string) *ProjectImportInspection {
	return &ProjectImportInspection{
		Link:   link,
		Path:   path,
		GitUrl: gitUrl,
		GitRef: gitRef,
	}
}

// endregion
