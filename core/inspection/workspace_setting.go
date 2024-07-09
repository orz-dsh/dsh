package inspection

// region WorkspaceSettingInspection

type WorkspaceSettingInspection struct {
	Clean    *WorkspaceCleanSettingInspection   `yaml:"clean,omitempty" toml:"clean,omitempty" json:"clean,omitempty"`
	Profile  *WorkspaceProfileSettingInspection `yaml:"profile,omitempty" toml:"profile,omitempty" json:"profile,omitempty"`
	Executor *ExecutorSettingInspection         `yaml:"executor,omitempty" toml:"executor,omitempty" json:"executor,omitempty"`
	Registry *RegistrySettingInspection         `yaml:"registry,omitempty" toml:"registry,omitempty" json:"registry,omitempty"`
	Redirect *RedirectSettingInspection         `yaml:"redirect,omitempty" toml:"redirect,omitempty" json:"redirect,omitempty"`
}

func NewWorkspaceSettingInspection(clean *WorkspaceCleanSettingInspection, profile *WorkspaceProfileSettingInspection, executor *ExecutorSettingInspection, registry *RegistrySettingInspection, redirect *RedirectSettingInspection) *WorkspaceSettingInspection {
	return &WorkspaceSettingInspection{
		Clean:    clean,
		Profile:  profile,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

// endregion
