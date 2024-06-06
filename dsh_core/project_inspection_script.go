package dsh_core

// region ProjectScriptInspection

type ProjectScriptInspection struct {
	PlainSources       []*ProjectScriptSourceInspection `yaml:"plainSources,omitempty" toml:"plainSources,omitempty" json:"plainSources,omitempty"`
	TemplateSources    []*ProjectScriptSourceInspection `yaml:"templateSources,omitempty" toml:"templateSources,omitempty" json:"templateSources,omitempty"`
	TemplateLibSources []*ProjectScriptSourceInspection `yaml:"templateLibSources,omitempty" toml:"templateLibSources,omitempty" json:"templateLibSources,omitempty"`
	Imports            []*ProjectImportInspection       `yaml:"imports,omitempty" toml:"imports,omitempty" json:"imports,omitempty"`
}

func newProjectScriptInspection(plainSources []*ProjectScriptSourceInspection, templateSources []*ProjectScriptSourceInspection, templateLibSources []*ProjectScriptSourceInspection, imports []*ProjectImportInspection) *ProjectScriptInspection {
	return &ProjectScriptInspection{
		PlainSources:       plainSources,
		TemplateSources:    templateSources,
		TemplateLibSources: templateLibSources,
		Imports:            imports,
	}
}

// endregion

// region ProjectScriptSourceInspection

type ProjectScriptSourceInspection struct {
	SourcePath string `yaml:"sourcePath" toml:"sourcePath" json:"sourcePath"`
	SourceName string `yaml:"sourceName" toml:"sourceName" json:"sourceName"`
}

func newProjectScriptSourceInspection(sourcePath string, sourceName string) *ProjectScriptSourceInspection {
	return &ProjectScriptSourceInspection{
		SourcePath: sourcePath,
		SourceName: sourceName,
	}
}

// endregion
