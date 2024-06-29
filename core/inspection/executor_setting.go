package inspection

// region ExecutorSettingInspection

type ExecutorSettingInspection struct {
	Items []*ExecutorItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewExecutorSettingInspection(items []*ExecutorItemSettingInspection) *ExecutorSettingInspection {
	return &ExecutorSettingInspection{
		Items: items,
	}
}

// endregion

// region ExecutorItemSettingInspection

type ExecutorItemSettingInspection struct {
	Name  string   `yaml:"name" toml:"name" json:"name"`
	File  string   `yaml:"file,omitempty" toml:"file,omitempty" json:"file,omitempty"`
	Exts  []string `yaml:"exts,omitempty" toml:"exts,omitempty" json:"exts,omitempty"`
	Args  []string `yaml:"args,omitempty" toml:"args,omitempty" json:"args,omitempty"`
	Match string   `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewExecutorItemSettingInspection(name, file string, exts, args []string, match string) *ExecutorItemSettingInspection {
	return &ExecutorItemSettingInspection{
		Name:  name,
		File:  file,
		Exts:  exts,
		Args:  args,
		Match: match,
	}
}

// endregion
