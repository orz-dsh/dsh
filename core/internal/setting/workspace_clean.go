package setting

import (
	. "github.com/orz-dsh/dsh/utils"
	"time"
)

// region default

var workspaceCleanSettingDefault = NewWorkspaceCleanSetting(3, 24*time.Hour)

// endregion

// region WorkspaceCleanSetting

type WorkspaceCleanSetting struct {
	OutputCount   int
	OutputExpires time.Duration
}

func NewWorkspaceCleanSetting(outputCount int, outputExpires time.Duration) *WorkspaceCleanSetting {
	return &WorkspaceCleanSetting{
		OutputCount:   outputCount,
		OutputExpires: outputExpires,
	}
}

// endregion

// region WorkspaceCleanSettingModel

type WorkspaceCleanSettingModel struct {
	Output *WorkspaceCleanOutputSettingModel `yaml:"output,omitempty" toml:"output,omitempty" json:"output,omitempty"`
}

func (m *WorkspaceCleanSettingModel) Convert(helper *ModelHelper) (*WorkspaceCleanSetting, error) {
	if m.Output != nil {
		outputCount, outputExpires, err := m.Output.Convert(helper.Child("output"))
		if err != nil {
			return nil, err
		}
		return NewWorkspaceCleanSetting(outputCount, outputExpires), nil
	}
	return workspaceCleanSettingDefault, nil
}

// endregion

// region WorkspaceCleanOutputSettingModel

type WorkspaceCleanOutputSettingModel struct {
	Count   *int   `yaml:"count,omitempty" toml:"count,omitempty" json:"count,omitempty"`
	Expires string `yaml:"expires,omitempty" toml:"expires,omitempty" json:"expires,omitempty"`
}

func (m *WorkspaceCleanOutputSettingModel) Convert(helper *ModelHelper) (int, time.Duration, error) {
	count := workspaceCleanSettingDefault.OutputCount
	if m.Count != nil {
		value := *m.Count
		if value <= 0 {
			return 0, 0, helper.Child("count").NewValueInvalidError(value)
		}
		count = value
	}

	expires := workspaceCleanSettingDefault.OutputExpires
	if m.Expires != "" {
		value, err := time.ParseDuration(m.Expires)
		if err != nil {
			return 0, 0, helper.Child("expires").WrapValueInvalidError(err, m.Expires)
		}
		expires = value
	}

	return count, expires, nil
}

// endregion
