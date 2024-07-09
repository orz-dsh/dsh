package inspection

import "time"

// region WorkspaceCleanSettingInspection

type WorkspaceCleanSettingInspection struct {
	Output *WorkspaceCleanOutputSettingInspection `yaml:"output,omitempty" toml:"output,omitempty" json:"output,omitempty"`
}

func NewWorkspaceCleanSettingInspection(output *WorkspaceCleanOutputSettingInspection) *WorkspaceCleanSettingInspection {
	return &WorkspaceCleanSettingInspection{
		Output: output,
	}
}

// endregion

// region WorkspaceCleanOutputSettingInspection

type WorkspaceCleanOutputSettingInspection struct {
	Count   *int           `yaml:"count,omitempty" toml:"count,omitempty" json:"count,omitempty"`
	Expires *time.Duration `yaml:"expires,omitempty" toml:"expires,omitempty" json:"expires,omitempty"`
}

func NewWorkspaceCleanOutputSettingInspection(count *int, expires *time.Duration) *WorkspaceCleanOutputSettingInspection {
	return &WorkspaceCleanOutputSettingInspection{
		Count:   count,
		Expires: expires,
	}
}

// endregion
