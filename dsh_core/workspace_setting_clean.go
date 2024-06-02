package dsh_core

import "time"

// region default

const workspaceCleanOutputSettingCountDefault = 3
const workspaceCleanOutputSettingExpiresDefault = 24 * time.Hour

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

func (m *workspaceCleanSettingModel) convert(root *workspaceSettingModel) (setting *workspaceCleanSetting, err error) {
	outputCount := workspaceCleanOutputSettingCountDefault
	if m.Output.Count != nil {
		value := *m.Output.Count
		if value <= 0 {
			return nil, errN("workspace setting invalid",
				reason("value invalid"),
				kv("path", root.path),
				kv("field", "clean.output.count"),
				kv("value", value),
			)
		}
		outputCount = value
	}

	outputExpires := workspaceCleanOutputSettingExpiresDefault
	if m.Output.Expires != "" {
		outputExpires, err = time.ParseDuration(m.Output.Expires)
		if err != nil {
			return nil, errN("workspace setting invalid",
				reason("value invalid"),
				kv("path", root.path),
				kv("field", "clean.output.expires"),
				kv("value", m.Output.Expires),
			)
		}
	}

	return newWorkspaceCleanSetting(outputCount, outputExpires), nil
}

// endregion

// region workspaceCleanOutputSettingModel

type workspaceCleanOutputSettingModel struct {
	Count   *int
	Expires string
}

// endregion
