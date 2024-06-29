package inspection

// region ProjectInspection

type ProjectInspection struct {
	Name       string                       `yaml:"name" toml:"name" json:"name"`
	Dir        string                       `yaml:"dir" toml:"dir" json:"dir"`
	Option     *ProjectOptionInspection     `yaml:"option,omitempty" toml:"option,omitempty" json:"option,omitempty"`
	Dependency *ProjectDependencyInspection `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *ProjectResourceInspection   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func NewProjectInspection(name, dir string, option *ProjectOptionInspection, dependency *ProjectDependencyInspection, resource *ProjectResourceInspection) *ProjectInspection {
	return &ProjectInspection{
		Name:       name,
		Dir:        dir,
		Option:     option,
		Dependency: dependency,
		Resource:   resource,
	}
}

// endregion
