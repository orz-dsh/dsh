package core

// region WorkspaceExecutorSettingInspection

type WorkspaceExecutorSettingInspection struct {
	Items []*WorkspaceExecutorItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newWorkspaceExecutorSettingInspection(items []*WorkspaceExecutorItemSettingInspection) *WorkspaceExecutorSettingInspection {
	return &WorkspaceExecutorSettingInspection{
		Items: items,
	}
}

// endregion

// region WorkspaceExecutorItemSettingInspection

type WorkspaceExecutorItemSettingInspection struct {
	Name  string   `yaml:"name" toml:"name" json:"name"`
	File  string   `yaml:"file,omitempty" toml:"file,omitempty" json:"file,omitempty"`
	Exts  []string `yaml:"exts,omitempty" toml:"exts,omitempty" json:"exts,omitempty"`
	Args  []string `yaml:"args,omitempty" toml:"args,omitempty" json:"args,omitempty"`
	Match string   `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newWorkspaceExecutorItemSettingInspection(name, file string, exts, args []string, match string) *WorkspaceExecutorItemSettingInspection {
	return &WorkspaceExecutorItemSettingInspection{
		Name:  name,
		File:  file,
		Exts:  exts,
		Args:  args,
		Match: match,
	}
}

// endregion

// region WorkspaceRegistrySettingInspection

type WorkspaceRegistrySettingInspection struct {
	Items []*WorkspaceRegistryItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newWorkspaceRegistrySettingInspection(items []*WorkspaceRegistryItemSettingInspection) *WorkspaceRegistrySettingInspection {
	return &WorkspaceRegistrySettingInspection{
		Items: items,
	}
}

// endregion

// region WorkspaceRegistryItemSettingInspection

type WorkspaceRegistryItemSettingInspection struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newWorkspaceRegistryItemSettingInspection(name, link, match string) *WorkspaceRegistryItemSettingInspection {
	return &WorkspaceRegistryItemSettingInspection{
		Name:  name,
		Link:  link,
		Match: match,
	}
}

// endregion

// region WorkspaceRedirectSettingInspection

type WorkspaceRedirectSettingInspection struct {
	Items []*WorkspaceRedirectItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newWorkspaceRedirectSettingInspection(items []*WorkspaceRedirectItemSettingInspection) *WorkspaceRedirectSettingInspection {
	return &WorkspaceRedirectSettingInspection{
		Items: items,
	}
}

// endregion

// region WorkspaceRedirectItemSettingInspection

type WorkspaceRedirectItemSettingInspection struct {
	Regex string `yaml:"regex" toml:"regex" json:"regex"`
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newWorkspaceRedirectItemSettingInspection(regex, link, match string) *WorkspaceRedirectItemSettingInspection {
	return &WorkspaceRedirectItemSettingInspection{
		Regex: regex,
		Link:  link,
		Match: match,
	}
}

// endregion
