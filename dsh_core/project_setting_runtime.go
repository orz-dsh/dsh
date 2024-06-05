package dsh_core

import "dsh/dsh_utils"

// region projectRuntimeSetting

type projectRuntimeSetting struct {
	MinVersion dsh_utils.Version
	MaxVersion dsh_utils.Version
	minVersion int32
}

func newProjectRuntimeSetting(minVersion dsh_utils.Version, maxVersion dsh_utils.Version) *projectRuntimeSetting {
	return &projectRuntimeSetting{
		MinVersion: minVersion,
		MaxVersion: maxVersion,
	}
}

// endregion

// region projectRuntimeSettingModel

type projectRuntimeSettingModel struct {
	MinVersion dsh_utils.Version `yaml:"minVersion" toml:"minVersion" json:"minVersion"`
	MaxVersion dsh_utils.Version `yaml:"maxVersion" toml:"maxVersion" json:"maxVersion"`
}

func (m *projectRuntimeSettingModel) convert(ctx *modelConvertContext) (setting *projectRuntimeSetting, err error) {
	if err = dsh_utils.CheckRuntimeVersion(m.MinVersion, m.MaxVersion); err != nil {
		return nil, ctx.WrapError(err, "runtime incompatible",
			kv("minVersion", m.MinVersion),
			kv("maxVersion", m.MaxVersion),
			kv("runtimeVersion", dsh_utils.GetRuntimeVersion()),
		)
	}
	return newProjectRuntimeSetting(m.MinVersion, m.MaxVersion), nil
}

// endregion
