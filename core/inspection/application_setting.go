package inspection

// region ApplicationSettingInspection

type ApplicationSettingInspection struct {
	Argument *ProfileArgumentSettingInspection `yaml:"argument,omitempty" toml:"argument,omitempty" json:"argument,omitempty"`
	Addition *ProfileAdditionSettingInspection `yaml:"addition,omitempty" toml:"addition,omitempty" json:"addition,omitempty"`
	Executor *ExecutorSettingInspection        `yaml:"executor,omitempty" toml:"executor,omitempty" json:"executor,omitempty"`
	Registry *RegistrySettingInspection        `yaml:"registry,omitempty" toml:"registry,omitempty" json:"registry,omitempty"`
	Redirect *RedirectSettingInspection        `yaml:"redirect,omitempty" toml:"redirect,omitempty" json:"redirect,omitempty"`
}

func NewApplicationSettingInspection(argument *ProfileArgumentSettingInspection, addition *ProfileAdditionSettingInspection, executor *ExecutorSettingInspection, registry *RegistrySettingInspection, redirect *RedirectSettingInspection) *ApplicationSettingInspection {
	return &ApplicationSettingInspection{
		Argument: argument,
		Addition: addition,
		Executor: executor,
		Registry: registry,
		Redirect: redirect,
	}
}

// endregion
