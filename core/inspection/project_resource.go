package inspection

// region ProjectResourceInspection

type ProjectResourceInspection struct {
	ConfigItems      []*ProjectResourceConfigItemInspection      `yaml:"configItems,omitempty" toml:"configItems,omitempty" json:"configItems,omitempty"`
	TemplateItems    []*ProjectResourceTemplateItemInspection    `yaml:"templateItems,omitempty" toml:"templateItems,omitempty" json:"templateItems,omitempty"`
	TemplateLibItems []*ProjectResourceTemplateLibItemInspection `yaml:"templateLibItems,omitempty" toml:"templateLibItems,omitempty" json:"templateLibItems,omitempty"`
	PlainItems       []*ProjectResourcePlainItemInspection       `yaml:"plainItems,omitempty" toml:"plainItems,omitempty" json:"plainItems,omitempty"`
}

func NewProjectResourceInspection(configItems []*ProjectResourceConfigItemInspection, templateItems []*ProjectResourceTemplateItemInspection, templateLibItems []*ProjectResourceTemplateLibItemInspection, plainItems []*ProjectResourcePlainItemInspection) *ProjectResourceInspection {
	return &ProjectResourceInspection{
		ConfigItems:      configItems,
		TemplateItems:    templateItems,
		TemplateLibItems: templateLibItems,
		PlainItems:       plainItems,
	}
}

// endregion

// region ProjectResourceConfigItemInspection

type ProjectResourceConfigItemInspection struct {
	File   string `yaml:"file" toml:"file" json:"file"`
	Format string `yaml:"format" toml:"format" json:"format"`
}

func NewProjectResourceConfigItemInspection(file, format string) *ProjectResourceConfigItemInspection {
	return &ProjectResourceConfigItemInspection{
		File:   file,
		Format: format,
	}
}

// endregion

// region ProjectResourceTemplateItemInspection

type ProjectResourceTemplateItemInspection struct {
	File   string `yaml:"file" toml:"file" json:"file"`
	Target string `yaml:"target" toml:"target" json:"target"`
}

func NewProjectResourceTemplateItemInspection(file, target string) *ProjectResourceTemplateItemInspection {
	return &ProjectResourceTemplateItemInspection{
		File:   file,
		Target: target,
	}
}

// endregion

// region ProjectResourceTemplateLibItemInspection

type ProjectResourceTemplateLibItemInspection struct {
	File string `yaml:"file" toml:"file" json:"file"`
}

func NewProjectResourceTemplateLibItemInspection(file string) *ProjectResourceTemplateLibItemInspection {
	return &ProjectResourceTemplateLibItemInspection{
		File: file,
	}
}

// endregion

// region ProjectResourcePlainItemInspection

type ProjectResourcePlainItemInspection struct {
	File   string `yaml:"file" toml:"file" json:"file"`
	Target string `yaml:"target" toml:"target" json:"target"`
}

func NewProjectResourcePlainItemInspection(file, target string) *ProjectResourcePlainItemInspection {
	return &ProjectResourcePlainItemInspection{
		File:   file,
		Target: target,
	}
}

// endregion
