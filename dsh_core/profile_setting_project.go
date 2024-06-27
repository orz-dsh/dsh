package dsh_core

import (
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

func (m *profileProjectSettingModel) convert(helper *modelHelper) (*profileProjectSetting, error) {
	items, err := convertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
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

func (m *profileProjectItemSettingModel) convert(helper *modelHelper) (_ *profileProjectItemSetting, err error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !projectNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}

	if m.Dir == "" {
		return nil, helper.Child("dir").NewValueEmptyError()
	}

	var dependency *projectDependencySetting
	if m.Dependency != nil {
		if dependency, err = m.Dependency.convert(helper.Child("dependency")); err != nil {
			return nil, err
		}
	}

	var resource *projectResourceSetting
	if m.Resource != nil {
		if resource, err = m.Resource.convert(helper.Child("resource")); err != nil {
			return nil, err
		}
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return newProfileProjectItemSetting(m.Name, m.Dir, m.Match, dependency, resource, matchObj), nil
}

// endregion
