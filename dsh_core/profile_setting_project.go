package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
)

// region profileProjectSetting

type profileProjectSetting struct {
	Items []*profileProjectItemSetting
}

func newProfileProjectSetting(items []*profileProjectItemSetting) *profileProjectSetting {
	return &profileProjectSetting{
		Items: items,
	}
}

func (s *profileProjectSetting) merge(setting *profileProjectSetting) {
	s.Items = append(s.Items, setting.Items...)
}

func (s *profileProjectSetting) getProjectSettings(evaluator *Evaluator) ([]*projectSetting, error) {
	var result []*projectSetting
	for i := len(s.Items) - 1; i >= 0; i-- {
		item := s.Items[i]
		matched, err := evaluator.EvalBoolExpr(item.match)
		if err != nil {
			return nil, errW(err, "get profile project settings error",
				reason("eval expr error"),
				kv("item", item),
			)
		}
		if !matched {
			continue
		}

		rawPath, err := evaluator.EvalStringTemplate(item.Path)
		if err != nil {
			return nil, errW(err, "get profile project settings error",
				reason("eval template error"),
				kv("item", item),
			)
		}
		path, err := filepath.Abs(rawPath)
		if err != nil {
			return nil, errW(err, "get profile project settings error",
				reason("get abs-path error"),
				kv("item", item),
				kv("rawPath", rawPath),
			)
		}

		result = append(result, newProjectSetting(item.Name, path, nil, nil, item.Dependency, item.Resource))
	}
	return result, nil
}

func (s *profileProjectSetting) inspect() *ProfileProjectSettingInspection {
	var items []*ProfileProjectItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newProfileProjectSettingInspection(items)
}

// endregion

// region profileProjectItemSetting

type profileProjectItemSetting struct {
	Name       string
	Path       string
	Match      string
	Dependency *projectDependencySetting
	Resource   *projectResourceSetting
	match      *EvalExpr
}

func newProfileProjectItemSetting(name, dir, match string, dependency *projectDependencySetting, resource *projectResourceSetting, matchObj *EvalExpr) *profileProjectItemSetting {
	if dependency == nil {
		dependency = newProjectDependencySetting(nil)
	}
	if resource == nil {
		resource = newProjectResourceSetting(nil)
	}
	return &profileProjectItemSetting{
		Name:       name,
		Path:       dir,
		Match:      match,
		Dependency: dependency,
		Resource:   resource,
		match:      matchObj,
	}
}

func (s *profileProjectItemSetting) inspect() *ProfileProjectItemSettingInspection {
	return newProfileProjectItemSettingInspection(s.Name, s.Path, s.Match, s.Dependency.inspect(), s.Resource.inspect())
}

// endregion

// region profileProjectSettingModel

type profileProjectSettingModel struct {
	Items []*profileProjectItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProfileProjectSettingModel(items []*profileProjectItemSettingModel) *profileProjectSettingModel {
	return &profileProjectSettingModel{
		Items: items,
	}
}

func (m *profileProjectSettingModel) convert(ctx *modelConvertContext) (*profileProjectSetting, error) {
	var items []*profileProjectItemSetting
	for i := 0; i < len(m.Items); i++ {
		item, err := m.Items[i].convert(ctx.ChildItem("items", i))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return newProfileProjectSetting(items), nil
}

// endregion

// region profileProjectItemSettingModel

type profileProjectItemSettingModel struct {
	Name       string                         `yaml:"name" toml:"name" json:"name"`
	Dir        string                         `yaml:"dir" toml:"dir" json:"dir"`
	Match      string                         `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
	Dependency *projectDependencySettingModel `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *projectResourceSettingModel   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func newProfileProjectItemSettingModel(name, dir, match string, dependency *projectDependencySettingModel, resource *projectResourceSettingModel) *profileProjectItemSettingModel {
	return &profileProjectItemSettingModel{
		Name:       name,
		Dir:        dir,
		Match:      match,
		Dependency: dependency,
		Resource:   resource,
	}
}

func (m *profileProjectItemSettingModel) convert(ctx *modelConvertContext) (_ *profileProjectItemSetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}
	if !projectNameCheckRegex.MatchString(m.Name) {
		return nil, ctx.Child("name").NewValueInvalidError(m.Name)
	}

	if m.Dir == "" {
		return nil, ctx.Child("dir").NewValueEmptyError()
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

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newProfileProjectItemSetting(m.Name, m.Dir, m.Match, dependency, resource, matchObj), nil
}

// endregion
