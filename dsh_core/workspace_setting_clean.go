package dsh_core

import "time"

// region default

var workspaceCleanSettingDefault = newWorkspaceCleanSetting(3, 24*time.Hour)

// endregion

// region workspaceCleanSetting

type workspaceCleanSetting struct {
	OutputCount   int
	OutputExpires time.Duration
}

func newWorkspaceCleanSetting(outputCount int, outputExpires time.Duration) *workspaceCleanSetting {
	return &workspaceCleanSetting{
		OutputCount:   outputCount,
		OutputExpires: outputExpires,
	}
}

// endregion

// region workspaceCleanSettingModel

type workspaceCleanSettingModel struct {
	Output *workspaceCleanOutputSettingModel
}

func (m *workspaceCleanSettingModel) convert(helper *modelHelper) (*workspaceCleanSetting, error) {
	if m.Output != nil {
		if outputCount, outputExpires, err := m.Output.convert(helper.Child("output")); err != nil {
			return nil, err
		} else {
			return newWorkspaceCleanSetting(outputCount, outputExpires), nil
		}
	} else {
		return workspaceCleanSettingDefault, nil
	}
}

// endregion

// region workspaceCleanOutputSettingModel

type workspaceCleanOutputSettingModel struct {
	Count   *int
	Expires string
}

func (m *workspaceCleanOutputSettingModel) convert(helper *modelHelper) (int, time.Duration, error) {
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
