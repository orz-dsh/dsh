package setting

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"regexp"
)

// region base

var environmentArgumentNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9_]*[a-z0-9]$")

// endregion

// region EnvironmentArgumentSetting

type EnvironmentArgumentSetting struct {
	Items    []*EnvironmentArgumentItemSetting
	itemsMap map[string]string
}

func NewEnvironmentArgumentSetting(items []*EnvironmentArgumentItemSetting) *EnvironmentArgumentSetting {
	itemsMap := map[string]string{}
	for _, item := range items {
		itemsMap[item.Name] = item.Value
	}
	return &EnvironmentArgumentSetting{
		Items:    items,
		itemsMap: itemsMap,
	}
}

func (s *EnvironmentArgumentSetting) GetMap() map[string]string {
	return s.itemsMap
}

func (s *EnvironmentArgumentSetting) Inspect() *EnvironmentArgumentSettingInspection {
	items := make([]*EnvironmentArgumentItemSettingInspection, 0, len(s.Items))
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].Inspect())
	}
	return NewEnvironmentArgumentSettingInspection(items)
}

// endregion

// region EnvironmentArgumentItemSetting

type EnvironmentArgumentItemSetting struct {
	Name  string
	Value string
}

func NewEnvironmentArgumentItemSetting(name, value string) *EnvironmentArgumentItemSetting {
	return &EnvironmentArgumentItemSetting{
		Name:  name,
		Value: value,
	}
}

func (s *EnvironmentArgumentItemSetting) Inspect() *EnvironmentArgumentItemSettingInspection {
	return NewEnvironmentArgumentItemSettingInspection(s.Name, s.Value)
}

// endregion

// region EnvironmentArgumentSettingModel

type EnvironmentArgumentSettingModel struct {
	Items []*EnvironmentArgumentItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewEnvironmentArgumentSettingModel(items []*EnvironmentArgumentItemSettingModel) *EnvironmentArgumentSettingModel {
	return &EnvironmentArgumentSettingModel{
		Items: items,
	}
}

func (m *EnvironmentArgumentSettingModel) Convert(helper *ModelHelper) (*EnvironmentArgumentSetting, error) {
	items, err := ConvertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return NewEnvironmentArgumentSetting(items), nil
}

// endregion

// region EnvironmentArgumentItemSettingModel

type EnvironmentArgumentItemSettingModel struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Value string `yaml:"value" toml:"value" json:"value"`
}

func NewEnvironmentArgumentItemSettingModel(name, value string) *EnvironmentArgumentItemSettingModel {
	return &EnvironmentArgumentItemSettingModel{
		Name:  name,
		Value: value,
	}
}

func (m *EnvironmentArgumentItemSettingModel) Convert(helper *ModelHelper) (*EnvironmentArgumentItemSetting, error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}
	if !environmentArgumentNameCheckRegex.MatchString(m.Name) {
		return nil, helper.Child("name").NewValueInvalidError(m.Name)
	}
	return NewEnvironmentArgumentItemSetting(m.Name, m.Value), nil
}

// endregion
