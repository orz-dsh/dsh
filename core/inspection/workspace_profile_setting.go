package inspection

// region WorkspaceProfileSettingInspection

type WorkspaceProfileSettingInspection struct {
	Items []*WorkspaceProfileItemSettingInspection `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewWorkspaceProfileSettingInspection(items []*WorkspaceProfileItemSettingInspection) *WorkspaceProfileSettingInspection {
	return &WorkspaceProfileSettingInspection{
		Items: items,
	}
}

// endregion

// region WorkspaceProfileItemSettingInspection

type WorkspaceProfileItemSettingInspection struct {
	File     string `yaml:"file" toml:"file" json:"file"`
	Optional bool   `yaml:"optional,omitempty" toml:"optional,omitempty" json:"optional,omitempty"`
	Match    string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewWorkspaceProfileItemSettingInspection(file string, optional bool, match string) *WorkspaceProfileItemSettingInspection {
	return &WorkspaceProfileItemSettingInspection{
		File:     file,
		Optional: optional,
		Match:    match,
	}
}

// endregion
