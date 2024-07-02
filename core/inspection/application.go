package inspection

// region ApplicationInspection

type ApplicationInspection struct {
	Variables          map[string]any                `yaml:"variables,omitempty" toml:"variables,omitempty" json:"variables,omitempty"`
	Setting            *ApplicationSettingInspection `yaml:"setting,omitempty" toml:"setting,omitempty" json:"setting,omitempty"`
	Option             *ApplicationOptionInspection  `yaml:"option,omitempty" toml:"option,omitempty" json:"option,omitempty"`
	Config             *ApplicationConfigInspection  `yaml:"config,omitempty" toml:"config,omitempty" json:"config,omitempty"`
	MainProject        *ProjectInspection            `yaml:"mainProject,omitempty" toml:"mainProject,omitempty" json:"mainProject,omitempty"`
	AdditionProjects   []*ProjectInspection          `yaml:"additionProjects,omitempty" toml:"additionProjects,omitempty" json:"additionProjects,omitempty"`
	DependencyProjects []*ProjectInspection          `yaml:"dependencyProjects,omitempty" toml:"dependencyProjects,omitempty" json:"dependencyProjects,omitempty"`
}

func NewApplicationInspection(variables map[string]any, setting *ApplicationSettingInspection, option *ApplicationOptionInspection, config *ApplicationConfigInspection, mainProject *ProjectInspection, additionProjects []*ProjectInspection, dependencyProjects []*ProjectInspection) *ApplicationInspection {
	return &ApplicationInspection{
		Variables:          variables,
		Setting:            setting,
		Option:             option,
		Config:             config,
		MainProject:        mainProject,
		AdditionProjects:   additionProjects,
		DependencyProjects: dependencyProjects,
	}
}

// endregion
