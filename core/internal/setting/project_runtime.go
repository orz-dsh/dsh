package setting

import . "github.com/orz-dsh/dsh/utils"

// region ProjectRuntimeSetting

type ProjectRuntimeSetting struct {
	MinVersion Version
	MaxVersion Version
}

func NewProjectRuntimeSetting(minVersion, maxVersion Version) *ProjectRuntimeSetting {
	return &ProjectRuntimeSetting{
		MinVersion: minVersion,
		MaxVersion: maxVersion,
	}
}

// endregion

// region ProjectRuntimeSettingModel

type ProjectRuntimeSettingModel struct {
	MinVersion Version `yaml:"minVersion,omitempty" toml:"minVersion,omitempty" json:"minVersion,omitempty"`
	MaxVersion Version `yaml:"maxVersion,omitempty" toml:"maxVersion,omitempty" json:"maxVersion,omitempty"`
}

func (m *ProjectRuntimeSettingModel) Convert(helper *ModelHelper) (*ProjectRuntimeSetting, error) {
	if err := CheckRuntimeVersion(m.MinVersion, m.MaxVersion); err != nil {
		return nil, helper.WrapError(err, "runtime incompatible",
			KV("minVersion", m.MinVersion),
			KV("maxVersion", m.MaxVersion),
			KV("runtimeVersion", GetRuntimeVersion()),
		)
	}
	return NewProjectRuntimeSetting(m.MinVersion, m.MaxVersion), nil
}

// endregion
