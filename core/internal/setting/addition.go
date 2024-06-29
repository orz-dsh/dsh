package setting

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"path/filepath"
)

// region AdditionSetting

type AdditionSetting struct {
	Items []*AdditionItemSetting
}

func NewAdditionSetting(items []*AdditionItemSetting) *AdditionSetting {
	return &AdditionSetting{
		Items: items,
	}
}

func (s *AdditionSetting) Merge(other *AdditionSetting) {
	s.Items = append(s.Items, other.Items...)
}

func (s *AdditionSetting) GetProjectSettings(evaluator *Evaluator) ([]*ProjectSetting, error) {
	var result []*ProjectSetting
	for i := len(s.Items) - 1; i >= 0; i-- {
		item := s.Items[i]
		matched, err := evaluator.EvalBoolExpr(item.match)
		if err != nil {
			return nil, ErrW(err, "get profile project settings error",
				Reason("eval expr error"),
				KV("item", item),
			)
		}
		if !matched {
			continue
		}

		rawPath, err := evaluator.EvalStringTemplate(item.Dir)
		if err != nil {
			return nil, ErrW(err, "get profile project settings error",
				Reason("eval template error"),
				KV("item", item),
			)
		}
		path, err := filepath.Abs(rawPath)
		if err != nil {
			return nil, ErrW(err, "get profile project settings error",
				Reason("get abs-path error"),
				KV("item", item),
				KV("rawPath", rawPath),
			)
		}

		result = append(result, NewProjectSetting(item.Name, path, nil, nil, item.Dependency, item.Resource))
	}
	return result, nil
}

func (s *AdditionSetting) Inspect() *AdditionSettingInspection {
	var items []*AdditionItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].Inspect())
	}
	return NewAdditionSettingInspection(items)
}

// endregion

// region AdditionItemSetting

type AdditionItemSetting struct {
	Name       string
	Dir        string
	Match      string
	Dependency *ProjectDependencySetting
	Resource   *ProjectResourceSetting
	match      *EvalExpr
}

func NewAdditionItemSetting(name, dir, match string, dependency *ProjectDependencySetting, resource *ProjectResourceSetting, matchObj *EvalExpr) *AdditionItemSetting {
	if dependency == nil {
		dependency = NewProjectDependencySetting(nil)
	}
	if resource == nil {
		resource = NewProjectResourceSetting(nil)
	}
	return &AdditionItemSetting{
		Name:       name,
		Dir:        dir,
		Match:      match,
		Dependency: dependency,
		Resource:   resource,
		match:      matchObj,
	}
}

func (s *AdditionItemSetting) Inspect() *AdditionItemSettingInspection {
	return NewAdditionItemSettingInspection(s.Name, s.Dir, s.Match, s.Dependency.Inspect(), s.Resource.Inspect())
}

// endregion

// region AdditionSettingModel

type AdditionSettingModel struct {
	Items []*AdditionItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewAdditionSettingModel(items []*AdditionItemSettingModel) *AdditionSettingModel {
	return &AdditionSettingModel{
		Items: items,
	}
}

func (m *AdditionSettingModel) Convert(helper *ModelHelper) (*AdditionSetting, error) {
	items, err := ConvertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return NewAdditionSetting(items), nil
}

// endregion

// region AdditionItemSettingModel

type AdditionItemSettingModel struct {
	Name       string                         `yaml:"name" toml:"name" json:"name"`
	Dir        string                         `yaml:"dir" toml:"dir" json:"dir"`
	Match      string                         `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
	Dependency *ProjectDependencySettingModel `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *ProjectResourceSettingModel   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func NewAdditionItemSettingModel(name, dir, match string, dependency *ProjectDependencySettingModel, resource *ProjectResourceSettingModel) *AdditionItemSettingModel {
	return &AdditionItemSettingModel{
		Name:       name,
		Dir:        dir,
		Match:      match,
		Dependency: dependency,
		Resource:   resource,
	}
}

func (m *AdditionItemSettingModel) Convert(helper *ModelHelper) (_ *AdditionItemSetting, err error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !projectNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}

	if m.Dir == "" {
		return nil, helper.Child("dir").NewValueEmptyError()
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

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return NewAdditionItemSetting(m.Name, m.Dir, m.Match, dependency, resource, matchObj), nil
}

// endregion
