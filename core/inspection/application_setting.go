package inspection

// region ApplicationSettingInspection

type ApplicationSettingInspection struct {
	Argument *ArgumentSettingInspection `yaml:"argument,omitempty" toml:"argument,omitempty" json:"argument,omitempty"`
	Addition *AdditionSettingInspection `yaml:"addition,omitempty" toml:"addition,omitempty" json:"addition,omitempty"`
	Executor *ExecutorSettingInspection `yaml:"executor,omitempty" toml:"executor,omitempty" json:"executor,omitempty"`
	Registry *RegistrySettingInspection `yaml:"registry,omitempty" toml:"registry,omitempty" json:"registry,omitempty"`
	Redirect *RedirectSettingInspection `yaml:"redirect,omitempty" toml:"redirect,omitempty" json:"redirect,omitempty"`
}

func NewApplicationSettingInspection(argument *ArgumentSettingInspection, addition *AdditionSettingInspection, executor *ExecutorSettingInspection, registry *RegistrySettingInspection, redirect *RedirectSettingInspection) *ApplicationSettingInspection {
	return &ApplicationSettingInspection{
		Argument: argument,
		Addition: addition,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

// endregion
