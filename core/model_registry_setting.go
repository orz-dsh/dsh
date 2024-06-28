package core

import "regexp"

// region base

var registryNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9-]*[a-z0-9]$")

var registryLinkCheckRegex = regexp.MustCompile("^(git|dir):.*$")

// endregion

// region registrySettingModel

type registrySettingModel struct {
	Items []*registryItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newRegistrySettingModel(items []*registryItemSettingModel) *registrySettingModel {
	return &registrySettingModel{
		Items: items,
	}
}

func (m *registrySettingModel) convert(helper *modelHelper) (*registrySetting, error) {
	items, err := convertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return newRegistrySetting(items), nil
}

// endregion

// region registryItemSettingModel

type registryItemSettingModel struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newRegistryItemSettingModel(name, link, match string) *registryItemSettingModel {
	return &registryItemSettingModel{
		Name:  name,
		Link:  link,
		Match: match,
	}
}

func (m *registryItemSettingModel) convert(helper *modelHelper) (*registryItemSetting, error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !registryNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}

	if m.Link == "" {
		return nil, helper.Child("link").NewValueEmptyError()
	}
	if !registryLinkCheckRegex.MatchString(m.Link) {
		return nil, helper.Child("link").NewValueInvalidError(m.Link)
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return newRegistryItemSetting(m.Name, m.Link, m.Match, matchObj), nil
}

// endregion

// region RegistrySettingModelBuilder

type RegistrySettingModelBuilder[P any] struct {
	commit func(*registrySettingModel) P
	items  []*registryItemSettingModel
}

func newProfileRegistrySettingBuilder[P any](commit func(*registrySettingModel) P) *RegistrySettingModelBuilder[P] {
	return &RegistrySettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *RegistrySettingModelBuilder[P]) AddItem(name, link, match string) *RegistrySettingModelBuilder[P] {
	b.items = append(b.items, newRegistryItemSettingModel(name, link, match))
	return b
}

func (b *RegistrySettingModelBuilder[P]) CommitRegistrySetting() P {
	return b.commit(newRegistrySettingModel(b.items))
}

// endregion
