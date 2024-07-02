package setting

import (
	. "github.com/orz-dsh/dsh/utils"
	"regexp"
)

// region base

var projectNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9-]*[a-z0-9]$")

// endregion

// region ProjectSetting

type ProjectSetting struct {
	Name       string
	Dir        string
	Runtime    *ProjectRuntimeSetting
	Option     *ProjectOptionSetting
	Dependency *ProjectDependencySetting
	Resource   *ProjectResourceSetting
}

func NewProjectSetting(name, dir string, runtime *ProjectRuntimeSetting, option *ProjectOptionSetting, dependency *ProjectDependencySetting, resource *ProjectResourceSetting) *ProjectSetting {
	if runtime == nil {
		runtime = NewProjectRuntimeSetting("", "")
	}
	if option == nil {
		option = NewProjectOptionSetting(nil, nil)
	}
	if dependency == nil {
		dependency = NewProjectDependencySetting(nil)
	}
	if resource == nil {
		resource = NewProjectResourceSetting(nil)
	}
	return &ProjectSetting{
		Name:       name,
		Dir:        dir,
		Runtime:    runtime,
		Option:     option,
		Dependency: dependency,
		Resource:   resource,
	}
}

func LoadProjectSetting(logger *Logger, dir string) (setting *ProjectSetting, err error) {
	model := &ProjectSettingModel{}
	metadata, err := DeserializeDir(dir, []string{"project"}, model, true)
	if err != nil {
		return nil, ErrW(err, "load project setting error",
			Reason("deserialize error"),
			KV("dir", dir),
		)
	}
	if setting, err = model.convert(NewModelHelper(logger, "project setting", metadata.File), dir); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region ProjectSettingModel

type ProjectSettingModel struct {
	Name       string                         `yaml:"name" toml:"name" json:"name"`
	Runtime    *ProjectRuntimeSettingModel    `yaml:"runtime,omitempty" toml:"runtime,omitempty" json:"runtime,omitempty"`
	Option     *ProjectOptionSettingModel     `yaml:"option,omitempty" toml:"option,omitempty" json:"option,omitempty"`
	Dependency *ProjectDependencySettingModel `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *ProjectResourceSettingModel   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func (m *ProjectSettingModel) convert(helper *ModelHelper, dir string) (_ *ProjectSetting, err error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !projectNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}
	helper.AddVariable("projectName", m.Name)

	var runtime *ProjectRuntimeSetting
	if m.Runtime != nil {
		if runtime, err = m.Runtime.Convert(helper.Child("runtime")); err != nil {
			return nil, err
		}
	}

	var option *ProjectOptionSetting
	if m.Option != nil {
		if option, err = m.Option.Convert(helper.Child("option")); err != nil {
			return nil, err
		}
	}

	var dependency *ProjectDependencySetting
	if m.Dependency != nil {
		if dependency, err = m.Dependency.Convert(helper.Child("dependency")); err != nil {
			return nil, err
		}
	}

	var resource *ProjectResourceSetting
	if m.Resource != nil {
		if resource, err = m.Resource.Convert(helper.Child("resource")); err != nil {
			return nil, err
		}
	}

	return NewProjectSetting(m.Name, dir, runtime, option, dependency, resource), nil
}

// endregion
