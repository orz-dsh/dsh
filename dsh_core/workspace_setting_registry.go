package dsh_core

import (
	"dsh/dsh_utils"
	"regexp"
)

// region base

var workspaceRegistryNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9-]*[a-z0-9]$")
var workspaceRegistryLinkCheckRegex = regexp.MustCompile("^(git|dir):.*$")

// endregion

// region default

var workspaceRegistrySettingDefault = newWorkspaceRegistrySetting([]*workspaceRegistryItemSetting{
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

// region workspaceRegistrySetting

type workspaceRegistrySetting struct {
	Items       []*workspaceRegistryItemSetting
	itemsByName map[string][]*workspaceRegistryItemSetting
}

func newWorkspaceRegistrySetting(items []*workspaceRegistryItemSetting) *workspaceRegistrySetting {
	itemsByName := map[string][]*workspaceRegistryItemSetting{}
	for i := 0; i < len(items); i++ {
		item := items[i]
		itemsByName[item.Name] = append(itemsByName[item.Name], item)
	}
	return &workspaceRegistrySetting{
		Items:       items,
		itemsByName: itemsByName,
	}
}

func (s *workspaceRegistrySetting) merge(setting *workspaceRegistrySetting) {
	for i := 0; i < len(setting.Items); i++ {
		item := setting.Items[i]
		s.Items = append(s.Items, item)
		s.itemsByName[item.Name] = append(s.itemsByName[item.Name], item)
	}
}

func (s *workspaceRegistrySetting) mergeDefault() {
	s.merge(workspaceRegistrySettingDefault)
}

func (s *workspaceRegistrySetting) getLink(name string, evaluator *Evaluator) (*projectLink, error) {
	if items, exist := s.itemsByName[name]; exist {
		for i := 0; i < len(items); i++ {
			model := items[i]
			matched, err := evaluator.EvalBoolExpr(model.match)
			if err != nil {
				return nil, errW(err, "get workspace import registry setting link error",
					reason("eval expr error"),
					kv("model", model),
				)
			}
			if matched {
				rawLink, err := evaluator.EvalStringTemplate(model.Link)
				if err != nil {
					return nil, errW(err, "get workspace import registry setting link error",
						reason("eval template error"),
						kv("model", model),
					)
				}
				link, err := parseProjectLink(rawLink)
				if err != nil {
					return nil, errW(err, "get workspace import registry setting link error",
						reason("parse link error"),
						kv("model", model),
						kv("rawLink", rawLink),
					)
				}
				return link, nil
			}
		}
	}
	return nil, nil
}

func (s *workspaceRegistrySetting) inspect() *WorkspaceRegistrySettingInspection {
	var items []*WorkspaceRegistryItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newWorkspaceRegistrySettingInspection(items)
}

// endregion

// region workspaceRegistryItemSetting

type workspaceRegistryItemSetting struct {
	Name  string
	Link  string
	Match string
	match *EvalExpr
}

func newWorkspaceRegistryItemSetting(name, link, match string, matchObj *EvalExpr) *workspaceRegistryItemSetting {
	return &workspaceRegistryItemSetting{
		Name:  name,
		Link:  link,
		Match: match,
		match: matchObj,
	}
}

func (s *workspaceRegistryItemSetting) inspect() *WorkspaceRegistryItemSettingInspection {
	return newWorkspaceRegistryItemSettingInspection(s.Name, s.Link, s.Match)
}

// endregion

// region workspaceRegistrySettingModel

type workspaceRegistrySettingModel struct {
	Items []*workspaceRegistryItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newWorkspaceRegistrySettingModel(items []*workspaceRegistryItemSettingModel) *workspaceRegistrySettingModel {
	return &workspaceRegistrySettingModel{
		Items: items,
	}
}

func (m *workspaceRegistrySettingModel) convert(ctx *modelConvertContext) (*workspaceRegistrySetting, error) {
	var items []*workspaceRegistryItemSetting
	for i := 0; i < len(m.Items); i++ {
		item, err := m.Items[i].convert(ctx.ChildItem("items", i))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return newWorkspaceRegistrySetting(items), nil
}

// endregion

// region workspaceRegistryItemSettingModel

type workspaceRegistryItemSettingModel struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newWorkspaceRegistryItemSettingModel(name, link, match string) *workspaceRegistryItemSettingModel {
	return &workspaceRegistryItemSettingModel{
		Name:  name,
		Link:  link,
		Match: match,
	}
}

func (m *workspaceRegistryItemSettingModel) convert(ctx *modelConvertContext) (_ *workspaceRegistryItemSetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}
	if !workspaceRegistryNameCheckRegex.MatchString(m.Name) {
		return nil, ctx.Child("name").NewValueInvalidError(m.Name)
	}

	if m.Link == "" {
		return nil, ctx.Child("link").NewValueEmptyError()
	}
	if !workspaceRegistryLinkCheckRegex.MatchString(m.Link) {
		return nil, ctx.Child("link").NewValueInvalidError(m.Link)
	}

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newWorkspaceRegistryItemSetting(m.Name, m.Link, m.Match, matchObj), nil
}

// endregion
