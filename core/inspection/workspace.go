package inspection

// region WorkspaceInspection

type WorkspaceInspection struct {
	Dir     string                      `yaml:"dir" toml:"dir" json:"dir"`
	Setting *WorkspaceSettingInspection `yaml:"setting,omitempty" toml:"setting,omitempty" json:"setting,omitempty"`
}

func NewWorkspaceInspection(dir string, setting *WorkspaceSettingInspection) *WorkspaceInspection {
	return &WorkspaceInspection{
		Dir:     dir,
		Setting: setting,
	}
}

// endregion
