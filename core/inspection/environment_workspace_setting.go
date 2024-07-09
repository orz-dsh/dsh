package inspection

// region EnvironmentWorkspaceSettingInspection

type EnvironmentWorkspaceSettingInspection struct {
	Dir      string                             `yaml:"dir,omitempty" toml:"dir,omitempty" json:"dir,omitempty"`
	Clean    *WorkspaceCleanSettingInspection   `yaml:"clean,omitempty" toml:"clean,omitempty" json:"clean,omitempty"`
	Profile  *WorkspaceProfileSettingInspection `yaml:"profile,omitempty" toml:"profile,omitempty" json:"profile,omitempty"`
	Executor *ExecutorSettingInspection         `yaml:"executor,omitempty" toml:"executor,omitempty" json:"executor,omitempty"`
	Registry *RegistrySettingInspection         `yaml:"registry,omitempty" toml:"registry,omitempty" json:"registry,omitempty"`
	Redirect *RedirectSettingInspection         `yaml:"redirect,omitempty" toml:"redirect,omitempty" json:"redirect,omitempty"`
}

func NewEnvironmentWorkspaceSettingInspection(dir string, clean *WorkspaceCleanSettingInspection, profile *WorkspaceProfileSettingInspection, executor *ExecutorSettingInspection, registry *RegistrySettingInspection, redirect *RedirectSettingInspection) *EnvironmentWorkspaceSettingInspection {
	return &EnvironmentWorkspaceSettingInspection{
		Dir:      dir,
		Clean:    clean,
		Profile:  profile,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

// endregion
