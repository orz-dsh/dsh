package inspection

// region ApplicationInspection

type ApplicationInspection struct {
	Environment        *EnvironmentInspection         `yaml:"environment,omitempty" toml:"environment,omitempty" json:"environment,omitempty"`
	Workspace          *WorkspaceInspection           `yaml:"workspace,omitempty" toml:"workspace,omitempty" json:"workspace,omitempty"`
	Variable           *ApplicationVariableInspection `yaml:"variable,omitempty" toml:"variable,omitempty" json:"variable,omitempty"`
	Setting            *ApplicationSettingInspection  `yaml:"setting,omitempty" toml:"setting,omitempty" json:"setting,omitempty"`
	Option             *ApplicationOptionInspection   `yaml:"option,omitempty" toml:"option,omitempty" json:"option,omitempty"`
	Config             *ApplicationConfigInspection   `yaml:"config,omitempty" toml:"config,omitempty" json:"config,omitempty"`
	MainProject        *ProjectInspection             `yaml:"mainProject,omitempty" toml:"mainProject,omitempty" json:"mainProject,omitempty"`
	AdditionProjects   []*ProjectInspection           `yaml:"additionProjects,omitempty" toml:"additionProjects,omitempty" json:"additionProjects,omitempty"`
	DependencyProjects []*ProjectInspection           `yaml:"dependencyProjects,omitempty" toml:"dependencyProjects,omitempty" json:"dependencyProjects,omitempty"`
}

func NewApplicationInspection(environment *EnvironmentInspection, workspace *WorkspaceInspection, variable *ApplicationVariableInspection, setting *ApplicationSettingInspection, option *ApplicationOptionInspection, config *ApplicationConfigInspection, mainProject *ProjectInspection, additionProjects []*ProjectInspection, dependencyProjects []*ProjectInspection) *ApplicationInspection {
	return &ApplicationInspection{
		Environment:        environment,
		Workspace:          workspace,
		Variable:           variable,
		Setting:            setting,
		Option:             option,
		Config:             config,
		MainProject:        mainProject,
		AdditionProjects:   additionProjects,
		DependencyProjects: dependencyProjects,
	}
}

// endregion
