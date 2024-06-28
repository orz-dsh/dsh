package core

// region ProjectEntityInspection

type ProjectEntityInspection struct {
	Name       string                             `yaml:"name" toml:"name" json:"name"`
	Dir        string                             `yaml:"dir" toml:"dir" json:"dir"`
	Option     *ProjectOptionEntityInspection     `yaml:"option,omitempty" toml:"option,omitempty" json:"option,omitempty"`
	Dependency *ProjectDependencyEntityInspection `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *ProjectResourceEntityInspection   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func newProjectEntityInspection(name, dir string, option *ProjectOptionEntityInspection, dependency *ProjectDependencyEntityInspection, resource *ProjectResourceEntityInspection) *ProjectEntityInspection {
	return &ProjectEntityInspection{
		Name:       name,
		Dir:        dir,
		Option:     option,
		Dependency: dependency,
		Resource:   resource,
	}
}

// endregion

// region ProjectOptionEntityInspection

type ProjectOptionEntityInspection struct {
	Items map[string]any `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProjectOptionEntityInspection(items map[string]any) *ProjectOptionEntityInspection {
	return &ProjectOptionEntityInspection{
		Items: items,
	}
}

// endregion

// region ProjectDependencyEntityInspection

type ProjectDependencyEntityInspection struct {
	Items []*ProjectDependencyItemEntityInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProjectDependencyEntityInspection(items []*ProjectDependencyItemEntityInspection) *ProjectDependencyEntityInspection {
	return &ProjectDependencyEntityInspection{
		Items: items,
	}
}

// endregion

// region ProjectDependencyItemEntityInspection

type ProjectDependencyItemEntityInspection struct {
	Link   string `yaml:"link" toml:"link" json:"link"`
	Dir    string `yaml:"dir" toml:"dir" json:"dir"`
	GitUrl string `yaml:"gitUrl,omitempty" toml:"gitUrl,omitempty" json:"gitUrl,omitempty"`
	GitRef string `yaml:"gitRef,omitempty" toml:"gitRef,omitempty" json:"gitRef,omitempty"`
}

func newProjectDependencyItemEntityInspection(link, dir, gitUrl, gitRef string) *ProjectDependencyItemEntityInspection {
	return &ProjectDependencyItemEntityInspection{
		Link:   link,
		Dir:    dir,
		GitUrl: gitUrl,
		GitRef: gitRef,
	}
}

// endregion

// region ProjectResourceEntityInspection

type ProjectResourceEntityInspection struct {
	ConfigItems      []*ProjectResourceConfigItemEntityInspection      `yaml:"configItems,omitempty" toml:"configItems,omitempty" json:"configItems,omitempty"`
	TemplateItems    []*ProjectResourceTemplateItemEntityInspection    `yaml:"templateItems,omitempty" toml:"templateItems,omitempty" json:"templateItems,omitempty"`
	TemplateLibItems []*ProjectResourceTemplateLibItemEntityInspection `yaml:"templateLibItems,omitempty" toml:"templateLibItems,omitempty" json:"templateLibItems,omitempty"`
	PlainItems       []*ProjectResourcePlainItemEntityInspection       `yaml:"plainItems,omitempty" toml:"plainItems,omitempty" json:"plainItems,omitempty"`
}

func newProjectResourceEntityInspection(configItems []*ProjectResourceConfigItemEntityInspection, templateItems []*ProjectResourceTemplateItemEntityInspection, templateLibItems []*ProjectResourceTemplateLibItemEntityInspection, plainItems []*ProjectResourcePlainItemEntityInspection) *ProjectResourceEntityInspection {
	return &ProjectResourceEntityInspection{
		ConfigItems:      configItems,
		TemplateItems:    templateItems,
		TemplateLibItems: templateLibItems,
		PlainItems:       plainItems,
	}
}

// endregion

// region ProjectResourceConfigItemEntityInspection

type ProjectResourceConfigItemEntityInspection struct {
	File   string `yaml:"file" toml:"file" json:"file"`
	Format string `yaml:"format" toml:"format" json:"format"`
}

func newProjectResourceConfigItemEntityInspection(file, format string) *ProjectResourceConfigItemEntityInspection {
	return &ProjectResourceConfigItemEntityInspection{
		File:   file,
		Format: format,
	}
}

// endregion

// region ProjectResourceTemplateItemEntityInspection

type ProjectResourceTemplateItemEntityInspection struct {
	File   string `yaml:"file" toml:"file" json:"file"`
	Target string `yaml:"target" toml:"target" json:"target"`
}

func newProjectResourceTemplateItemEntityInspection(file, target string) *ProjectResourceTemplateItemEntityInspection {
	return &ProjectResourceTemplateItemEntityInspection{
		File:   file,
		Target: target,
	}
}

// endregion

// region ProjectResourceTemplateLibItemEntityInspection

type ProjectResourceTemplateLibItemEntityInspection struct {
	File string `yaml:"file" toml:"file" json:"file"`
}

func newProjectResourceTemplateLibItemEntityInspection(file string) *ProjectResourceTemplateLibItemEntityInspection {
	return &ProjectResourceTemplateLibItemEntityInspection{
		File: file,
	}
}

// endregion

// region ProjectResourcePlainItemEntityInspection

type ProjectResourcePlainItemEntityInspection struct {
	File   string `yaml:"file" toml:"file" json:"file"`
	Target string `yaml:"target" toml:"target" json:"target"`
}

func newProjectResourcePlainItemEntityInspection(file, target string) *ProjectResourcePlainItemEntityInspection {
	return &ProjectResourcePlainItemEntityInspection{
		File:   file,
		Target: target,
	}
}

// endregion
