package setting

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"time"
)

// region default

var workspaceCleanOutputCountDefault = 3
var workspaceCleanOutputExpiresDefault = 24 * time.Hour

// endregion

// region WorkspaceCleanSetting

type WorkspaceCleanSetting struct {
	Output *WorkspaceCleanOutputSetting
}

func NewWorkspaceCleanSetting(output *WorkspaceCleanOutputSetting) *WorkspaceCleanSetting {
	if output == nil {
		output = NewWorkspaceCleanOutputSetting(nil, nil)
	}
	return &WorkspaceCleanSetting{
		Output: output,
	}
}

func (s *WorkspaceCleanSetting) Merge(other *WorkspaceCleanSetting) *WorkspaceCleanSetting {
	s.Output.Merge(other.Output)
	return s
}

func (s *WorkspaceCleanSetting) MergeDefault() *WorkspaceCleanSetting {
	s.Output.MergeDefault()
	return s
}

func (s *WorkspaceCleanSetting) Inspect() *WorkspaceCleanSettingInspection {
	return NewWorkspaceCleanSettingInspection(s.Output.Inspect())
}

// endregion

// region WorkspaceCleanOutputSetting

type WorkspaceCleanOutputSetting struct {
	Count   *int
	Expires *time.Duration
}

func NewWorkspaceCleanOutputSetting(count *int, expires *time.Duration) *WorkspaceCleanOutputSetting {
	return &WorkspaceCleanOutputSetting{
		Count:   count,
		Expires: expires,
	}
}

func (s *WorkspaceCleanOutputSetting) Merge(other *WorkspaceCleanOutputSetting) *WorkspaceCleanOutputSetting {
	if s.Count == nil {
		s.Count = other.Count
	}
	if s.Expires == nil {
		s.Expires = other.Expires
	}
	return s
}

func (s *WorkspaceCleanOutputSetting) MergeDefault() *WorkspaceCleanOutputSetting {
	if s.Count == nil {
		s.Count = &workspaceCleanOutputCountDefault
	}
	if s.Expires == nil {
		s.Expires = &workspaceCleanOutputExpiresDefault
	}
	return s
}

func (s *WorkspaceCleanOutputSetting) Inspect() *WorkspaceCleanOutputSettingInspection {
	return NewWorkspaceCleanOutputSettingInspection(s.Count, s.Expires)
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
	var output *WorkspaceCleanOutputSetting
	if m.Output != nil {
		output, err = m.Output.Convert(helper.Child("output"))
		if err != nil {
			return nil, err
		}
	}
	return NewWorkspaceCleanSetting(output), nil
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

func (m *WorkspaceCleanOutputSettingModel) Convert(helper *ModelHelper) (*WorkspaceCleanOutputSetting, error) {
	var count *int
	if m.Count != nil {
		value := *m.Count
		if value <= 0 {
			return nil, helper.Child("count").NewValueInvalidError(value)
		}
		count = &value
	}

	var expires *time.Duration
	if m.Expires != "" {
		value, err := time.ParseDuration(m.Expires)
		if err != nil {
			return nil, helper.Child("expires").WrapValueInvalidError(err, m.Expires)
		}
		expires = &value
	}

	return NewWorkspaceCleanOutputSetting(count, expires), nil
}

// endregion
