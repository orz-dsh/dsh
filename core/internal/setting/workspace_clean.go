package setting

import (
	. "github.com/orz-dsh/dsh/utils"
	"time"
)

// region default

var workspaceCleanOutputCountDefault = 3
var workspaceCleanOutputExpiresDefault = 24 * time.Hour

// endregion

// region WorkspaceCleanSetting

type WorkspaceCleanSetting struct {
	OutputCount   *int
	OutputExpires *time.Duration
}

func NewWorkspaceCleanSetting(outputCount *int, outputExpires *time.Duration) *WorkspaceCleanSetting {
	return &WorkspaceCleanSetting{
		OutputCount:   outputCount,
		OutputExpires: outputExpires,
	}
}

func (s *WorkspaceCleanSetting) Merge(other *WorkspaceCleanSetting) *WorkspaceCleanSetting {
	if s.OutputCount == nil {
		s.OutputCount = other.OutputCount
	}
	if s.OutputExpires == nil {
		s.OutputExpires = other.OutputExpires
	}
	return s
}

func (s *WorkspaceCleanSetting) MergeDefault() *WorkspaceCleanSetting {
	if s.OutputCount == nil {
		s.OutputCount = &workspaceCleanOutputCountDefault
	}
	if s.OutputExpires == nil {
		s.OutputExpires = &workspaceCleanOutputExpiresDefault
	}
	return s
}

// endregion

// region WorkspaceCleanSettingModel

type WorkspaceCleanSettingModel struct {
	Output *WorkspaceCleanOutputSettingModel `yaml:"output,omitempty" toml:"output,omitempty" json:"output,omitempty"`
}

func NewWorkspaceCleanSettingModel(output *WorkspaceCleanOutputSettingModel) *WorkspaceCleanSettingModel {
	return &WorkspaceCleanSettingModel{
		Output: output,
	}
}

func (m *WorkspaceCleanSettingModel) Convert(helper *ModelHelper) (_ *WorkspaceCleanSetting, err error) {
	var outputCount *int
	var outputExpires *time.Duration
	if m.Output != nil {
		outputCount, outputExpires, err = m.Output.Convert(helper.Child("output"))
		if err != nil {
			return nil, err
		}
	}
	return NewWorkspaceCleanSetting(outputCount, outputExpires), nil
}

// endregion

// region WorkspaceCleanOutputSettingModel

type WorkspaceCleanOutputSettingModel struct {
	Count   *int   `yaml:"count,omitempty" toml:"count,omitempty" json:"count,omitempty"`
	Expires string `yaml:"expires,omitempty" toml:"expires,omitempty" json:"expires,omitempty"`
}

func NewWorkspaceCleanOutputSettingModel(count *int, expires string) *WorkspaceCleanOutputSettingModel {
	return &WorkspaceCleanOutputSettingModel{
		Count:   count,
		Expires: expires,
	}
}

func (m *WorkspaceCleanOutputSettingModel) Convert(helper *ModelHelper) (*int, *time.Duration, error) {
	var count *int
	if m.Count != nil {
		value := *m.Count
		if value <= 0 {
			return nil, nil, helper.Child("count").NewValueInvalidError(value)
		}
		count = &value
	}

	var expires *time.Duration
	if m.Expires != "" {
		value, err := time.ParseDuration(m.Expires)
		if err != nil {
			return nil, nil, helper.Child("expires").WrapValueInvalidError(err, m.Expires)
		}
		expires = &value
	}

	return count, expires, nil
}

// endregion
