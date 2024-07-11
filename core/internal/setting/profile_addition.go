package setting

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"path/filepath"
)

// region ProfileAdditionSetting

type ProfileAdditionSetting struct {
	Items []*ProfileAdditionItemSetting
}

func NewProfileAdditionSetting(items []*ProfileAdditionItemSetting) *ProfileAdditionSetting {
	return &ProfileAdditionSetting{
		Items: items,
	}
}

func (s *ProfileAdditionSetting) Merge(other *ProfileAdditionSetting) {
	s.Items = append(s.Items, other.Items...)
}

func (s *ProfileAdditionSetting) GetProjectSettings(evaluator *Evaluator) ([]*ProjectSetting, error) {
	var result []*ProjectSetting
	for i := len(s.Items) - 1; i >= 0; i-- {
		item := s.Items[i]
		matched, err := evaluator.EvalBoolExpr(item.Match)
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

func (s *ProfileAdditionSetting) Inspect() *ProfileAdditionSettingInspection {
	var items []*ProfileAdditionItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].Inspect())
	}
	return NewProfileAdditionSettingInspection(items)
}

// endregion

// region ProfileAdditionItemSetting

type ProfileAdditionItemSetting struct {
	Name       string
	Dir        string
	Match      string
	Dependency *ProjectDependencySetting
	Resource   *ProjectResourceSetting
}

func NewProfileAdditionItemSetting(name, dir, match string, dependency *ProjectDependencySetting, resource *ProjectResourceSetting) *ProfileAdditionItemSetting {
	if dependency == nil {
		dependency = NewProjectDependencySetting(nil)
	}
	if resource == nil {
		resource = NewProjectResourceSetting(nil)
	}
	return &ProfileAdditionItemSetting{
		Name:       name,
		Dir:        dir,
		Match:      match,
		Dependency: dependency,
		Resource:   resource,
	}
}

func (s *ProfileAdditionItemSetting) Inspect() *ProfileAdditionItemSettingInspection {
	return NewProfileAdditionItemSettingInspection(s.Name, s.Dir, s.Match, s.Dependency.Inspect(), s.Resource.Inspect())
}

// endregion

// region ProfileAdditionSettingModel

type ProfileAdditionSettingModel struct {
	Items []*ProfileAdditionItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewProfileAdditionSettingModel(items []*ProfileAdditionItemSettingModel) *ProfileAdditionSettingModel {
	return &ProfileAdditionSettingModel{
		Items: items,
	}
}

func (m *ProfileAdditionSettingModel) Convert(helper *ModelHelper) (*ProfileAdditionSetting, error) {
	items, err := ConvertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return NewProfileAdditionSetting(items), nil
}

// endregion

// region ProfileAdditionItemSettingModel

type ProfileAdditionItemSettingModel struct {
	Name       string                         `yaml:"name" toml:"name" json:"name"`
	Dir        string                         `yaml:"dir" toml:"dir" json:"dir"`
	Match      string                         `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
	Dependency *ProjectDependencySettingModel `yaml:"dependency,omitempty" toml:"dependency,omitempty" json:"dependency,omitempty"`
	Resource   *ProjectResourceSettingModel   `yaml:"resource,omitempty" toml:"resource,omitempty" json:"resource,omitempty"`
}

func NewProfileAdditionItemSettingModel(name, dir, match string, dependency *ProjectDependencySettingModel, resource *ProjectResourceSettingModel) *ProfileAdditionItemSettingModel {
	return &ProfileAdditionItemSettingModel{
		Name:       name,
		Dir:        dir,
		Match:      match,
		Dependency: dependency,
		Resource:   resource,
	}
}

func (m *ProfileAdditionItemSettingModel) Convert(helper *ModelHelper) (_ *ProfileAdditionItemSetting, err error) {
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

	return NewProfileAdditionItemSetting(m.Name, m.Dir, m.Match, dependency, resource), nil
}

// endregion
