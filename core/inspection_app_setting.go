package core

// region AppSettingInspection

type AppSettingInspection struct {
	Option   *ProfileOptionSettingInspection     `yaml:"option,omitempty" toml:"option,omitempty" json:"option,omitempty"`
	Project  *ProfileProjectSettingInspection    `yaml:"project,omitempty" toml:"project,omitempty" json:"project,omitempty"`
	Executor *WorkspaceExecutorSettingInspection `yaml:"executor,omitempty" toml:"executor,omitempty" json:"executor,omitempty"`
	Registry *WorkspaceRegistrySettingInspection `yaml:"registry,omitempty" toml:"registry,omitempty" json:"registry,omitempty"`
	Redirect *WorkspaceRedirectSettingInspection `yaml:"redirect,omitempty" toml:"redirect,omitempty" json:"redirect,omitempty"`
}

func newAppSettingInspection(option *ProfileOptionSettingInspection, project *ProfileProjectSettingInspection, executor *WorkspaceExecutorSettingInspection, registry *WorkspaceRegistrySettingInspection, redirect *WorkspaceRedirectSettingInspection) *AppSettingInspection {
	return &AppSettingInspection{
		Option:   option,
		Project:  project,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

// endregion
