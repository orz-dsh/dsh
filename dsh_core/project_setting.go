package dsh_core

import (
	"dsh/dsh_utils"
	"regexp"
)

// region base

var projectNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9-]*[a-z0-9]$")

// endregion

// region projectSetting

type projectSetting struct {
	Name       string
	Dir        string
	Runtime    *projectRuntimeSetting
	Option     *projectOptionSetting
	Dependency *projectDependencySetting
	Resource   *projectResourceSetting
}

func newProjectSetting(name, dir string, runtime *projectRuntimeSetting, option *projectOptionSetting, dependency *projectDependencySetting, resource *projectResourceSetting) *projectSetting {
	if runtime == nil {
		runtime = newProjectRuntimeSetting("", "")
	}
	if option == nil {
		option = newProjectOptionSetting(nil, nil)
	}
	if dependency == nil {
		dependency = newProjectDependencySetting(nil)
	}
	if resource == nil {
		resource = newProjectResourceSetting(nil)
	}
	return &projectSetting{
		Name:       name,
		Dir:        dir,
		Runtime:    runtime,
		Option:     option,
		Dependency: dependency,
		Resource:   resource,
	}
}

func loadProjectSetting(dir string) (setting *projectSetting, err error) {
	model := &projectSettingModel{}
	metadata, err := dsh_utils.DeserializeFromDir(dir, []string{"project"}, model, true)
	if err != nil {
		return nil, errW(err, "load project setting error",
			reason("deserialize error"),
			kv("dir", dir),
		)
	}
	if setting, err = model.convert(newModelConvertContext("project setting", metadata.File), dir); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region projectSettingModel

type projectSettingModel struct {
	Name       string                         `yaml:"name" toml:"name" json:"name"`
	Runtime    *projectRuntimeSettingModel    `yaml:"runtime,omitempty" toml:"runtime,omitempty" json:"runtime,omitempty"`
	Option     *projectOptionSettingModel     `yaml:"option,omitempty" toml:"option,omitempty" json:"option,omitempty"`
	Dependency *projectDependencySettingModel `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *projectResourceSettingModel   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func (m *projectSettingModel) convert(ctx *modelConvertContext, dir string) (setting *projectSetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}
	if !projectNameCheckRegex.MatchString(m.Name) {
		return nil, ctx.Child("name").NewValueInvalidError(m.Name)
	}
	ctx.AddVariable("projectName", m.Name)

	var runtime *projectRuntimeSetting
	if m.Runtime != nil {
		if runtime, err = m.Runtime.convert(ctx.Child("runtime")); err != nil {
			return nil, err
		}
	}

	var option *projectOptionSetting
	if m.Option != nil {
		if option, err = m.Option.convert(ctx.Child("option")); err != nil {
			return nil, err
		}
	}

	var dependency *projectDependencySetting
	if m.Dependency != nil {
		if dependency, err = m.Dependency.convert(ctx.Child("dependency")); err != nil {
			return nil, err
		}
	}

	var resource *projectResourceSetting
	if m.Resource != nil {
		if resource, err = m.Resource.convert(ctx.Child("resource")); err != nil {
			return nil, err
		}
	}

	return newProjectSetting(m.Name, dir, runtime, option, dependency, resource), nil
}

// endregion
