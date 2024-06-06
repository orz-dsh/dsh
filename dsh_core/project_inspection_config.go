package dsh_core

// region ProjectConfigInspection

type ProjectConfigInspection struct {
	Sources []*ProjectConfigSourceInspection `yaml:"sources,omitempty" toml:"sources,omitempty" json:"sources,omitempty"`
	Imports []*ProjectImportInspection       `yaml:"imports,omitempty" toml:"imports,omitempty" json:"imports,omitempty"`
}

func newProjectConfigInspection(sources []*ProjectConfigSourceInspection, imports []*ProjectImportInspection) *ProjectConfigInspection {
	return &ProjectConfigInspection{
		Sources: sources,
		Imports: imports,
	}
}

// endregion

// region ProjectConfigInspectionResult

type ProjectConfigSourceInspection struct {
	SourcePath string `yaml:"sourcePath" toml:"sourcePath" json:"sourcePath"`
}

func newProjectConfigSourceInspection(sourcePath string) *ProjectConfigSourceInspection {
	return &ProjectConfigSourceInspection{
		SourcePath: sourcePath,
	}
}

// endregion
