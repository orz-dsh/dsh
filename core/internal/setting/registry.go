package setting

import (
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"regexp"
)

// region base

var registryNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9-]*[a-z0-9]$")

var registryLinkCheckRegex = regexp.MustCompile("^(git|dir):.*$")

// endregion

// region default

var registrySettingDefault = NewRegistrySetting([]*RegistryItemSetting{
	{
		Name: "github",
		Link: "git:https://github.com/{{.path}}.git#ref={{.ref}}",
	},
	{
		Name: "gitlab",
		Link: "git:https://gitlab.com/{{.path}}.git#ref={{.ref}}",
	},
	{
		Name: "gitee",
		Link: "git:https://gitee.com/{{.path}}.git#ref={{.ref}}",
	},
	{
		Name: "orz-dsh",
		Link: "git:https://github.com/orz-dsh/{{.path}}.git#ref={{.ref}}",
	},
	{
		Name: "orz-ops",
		Link: "git:https://github.com/orz-ops/{{.path}}.git#ref={{.ref}}",
	},
})

// endregion

// region RegistrySetting

type RegistrySetting struct {
	Items       []*RegistryItemSetting
	itemsByName map[string][]*RegistryItemSetting
}

func NewRegistrySetting(items []*RegistryItemSetting) *RegistrySetting {
	itemsByName := map[string][]*RegistryItemSetting{}
	for i := 0; i < len(items); i++ {
		item := items[i]
		itemsByName[item.Name] = append(itemsByName[item.Name], item)
	}
	return &RegistrySetting{
		Items:       items,
		itemsByName: itemsByName,
	}
}

func (s *RegistrySetting) Merge(other *RegistrySetting) {
	for i := 0; i < len(other.Items); i++ {
		item := other.Items[i]
		s.Items = append(s.Items, item)
		s.itemsByName[item.Name] = append(s.itemsByName[item.Name], item)
	}
}

func (s *RegistrySetting) MergeDefault() {
	s.Merge(registrySettingDefault)
}

func (s *RegistrySetting) GetLink(name string, evaluator *Evaluator) (*ProjectLink, error) {
	if items, exist := s.itemsByName[name]; exist {
		for i := 0; i < len(items); i++ {
			model := items[i]
			matched, err := evaluator.EvalBoolExpr(model.match)
			if err != nil {
				return nil, ErrW(err, "get workspace import registry setting link error",
					Reason("eval expr error"),
					KV("model", model),
				)
			}
			if matched {
				rawLink, err := evaluator.EvalStringTemplate(model.Link)
				if err != nil {
					return nil, ErrW(err, "get workspace import registry setting link error",
						Reason("eval template error"),
						KV("model", model),
					)
				}
				link, err := ParseProjectLink(rawLink)
				if err != nil {
					return nil, ErrW(err, "get workspace import registry setting link error",
						Reason("parse link error"),
						KV("model", model),
						KV("rawLink", rawLink),
					)
				}
				return link, nil
			}
		}
	}
	return nil, nil
}

func (s *RegistrySetting) Inspect() *RegistrySettingInspection {
	var items []*RegistryItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].Inspect())
	}
	return NewRegistrySettingInspection(items)
}

// endregion

// region RegistryItemSetting

type RegistryItemSetting struct {
	Name  string
	Link  string
	Match string
	match *EvalExpr
}

func NewRegistryItemSetting(name, link, match string, matchObj *EvalExpr) *RegistryItemSetting {
	return &RegistryItemSetting{
		Name:  name,
		Link:  link,
		Match: match,
		match: matchObj,
	}
}

func (s *RegistryItemSetting) Inspect() *RegistryItemSettingInspection {
	return NewRegistryItemSettingInspection(s.Name, s.Link, s.Match)
}

// endregion

// region RegistrySettingModel

type RegistrySettingModel struct {
	Items []*RegistryItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewRegistrySettingModel(items []*RegistryItemSettingModel) *RegistrySettingModel {
	return &RegistrySettingModel{
		Items: items,
	}
}

func (m *RegistrySettingModel) Convert(helper *ModelHelper) (*RegistrySetting, error) {
	items, err := ConvertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return NewRegistrySetting(items), nil
}

// endregion

// region RegistryItemSettingModel

type RegistryItemSettingModel struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewRegistryItemSettingModel(name, link, match string) *RegistryItemSettingModel {
	return &RegistryItemSettingModel{
		Name:  name,
		Link:  link,
		Match: match,
	}
}

func (m *RegistryItemSettingModel) Convert(helper *ModelHelper) (*RegistryItemSetting, error) {
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

	return NewRegistryItemSetting(m.Name, m.Link, m.Match, matchObj), nil
}

// endregion
