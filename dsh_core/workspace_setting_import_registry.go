package dsh_core

import (
	"dsh/dsh_utils"
	"regexp"
)

// region base

var workspaceImportRegistryNameCheckRegex = regexp.MustCompile("^[a-z][a-z0-9-]*[a-z0-9]$")
var workspaceImportRegistryLinkCheckRegex = regexp.MustCompile("^(git|dir):.*$")

// endregion

// region default

var workspaceImportRegistrySettingsDefault = workspaceImportRegistrySettingSet{
	"orz-dsh": {{
		Name: "orz-dsh",
		Link: "git:https://github.com/orz-dsh/{{.path}}.git#ref={{.ref}}",
	}},
	"orz-ops": {{
		Name: "orz-ops",
		Link: "git:https://github.com/orz-ops/{{.path}}.git#ref={{.ref}}",
	}},
}

// endregion

// region workspaceImportRegistrySetting

type workspaceImportRegistrySetting struct {
	Name  string
	Link  string
	Match string
	match *EvalExpr
}

type workspaceImportRegistrySettingSet map[string][]*workspaceImportRegistrySetting

func newWorkspaceImportRegistrySetting(name string, link string, match string, matchObj *EvalExpr) *workspaceImportRegistrySetting {
	return &workspaceImportRegistrySetting{
		Name:  name,
		Link:  link,
		Match: match,
		match: matchObj,
	}
}

func (s *workspaceImportRegistrySetting) inspect() *WorkspaceImportRegistrySettingInspection {
	return newWorkspaceImportRegistrySettingInspection(s.Name, s.Link, s.Match)
}

func (s workspaceImportRegistrySettingSet) merge(models workspaceImportRegistrySettingSet) {
	for name, list := range models {
		s[name] = append(s[name], list...)
	}
}

func (s workspaceImportRegistrySettingSet) mergeDefault() {
	s.merge(workspaceImportRegistrySettingsDefault)
}

func (s workspaceImportRegistrySettingSet) getLink(name string, evaluator *Evaluator) (*projectLink, error) {
	if models, exist := s[name]; exist {
		for i := 0; i < len(models); i++ {
			model := models[i]
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

func (s workspaceImportRegistrySettingSet) inspect() []*WorkspaceImportRegistrySettingInspection {
	var inspections []*WorkspaceImportRegistrySettingInspection
	for _, list := range s {
		for i := 0; i < len(list); i++ {
			inspections = append(inspections, list[i].inspect())
		}
	}
	return inspections
}

// endregion

// region workspaceImportRegistrySettingModel

type workspaceImportRegistrySettingModel struct {
	Items []*workspaceImportRegistryItemSettingModel
}

func (m *workspaceImportRegistrySettingModel) convert(ctx *modelConvertContext) (workspaceImportRegistrySettingSet, error) {
	settings := workspaceImportRegistrySettingSet{}
	for i := 0; i < len(m.Items); i++ {
		item := m.Items[i]
		if setting, err := item.convert(ctx.ChildItem("items", i)); err != nil {
			return nil, err
		} else {
			settings[item.Name] = append(settings[item.Name], setting)
		}
	}
	return settings, nil
}

// endregion

// region workspaceImportRegistryItemSettingModel

type workspaceImportRegistryItemSettingModel struct {
	Name  string
	Link  string
	Match string
}

func newWorkspaceImportRegistryItemSettingModel(name, link, match string) *workspaceImportRegistryItemSettingModel {
	return &workspaceImportRegistryItemSettingModel{
		Name:  name,
		Link:  link,
		Match: match,
	}
}

func (m *workspaceImportRegistryItemSettingModel) convert(ctx *modelConvertContext) (setting *workspaceImportRegistrySetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}
	if !workspaceImportRegistryNameCheckRegex.MatchString(m.Name) {
		return nil, ctx.Child("name").NewValueInvalidError(m.Name)
	}

	if m.Link == "" {
		return nil, ctx.Child("link").NewValueEmptyError()
	}
	if !workspaceImportRegistryLinkCheckRegex.MatchString(m.Link) {
		return nil, ctx.Child("link").NewValueInvalidError(m.Link)
	}

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newWorkspaceImportRegistrySetting(m.Name, m.Link, m.Match, matchObj), nil
}

// endregion

// region WorkspaceImportRegistrySettingInspection

type WorkspaceImportRegistrySettingInspection struct {
	Name  string `yaml:"name" toml:"name" json:"name"`
	Link  string `yaml:"link,omitempty" toml:"link,omitempty" json:"link,omitempty"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newWorkspaceImportRegistrySettingInspection(name, link, match string) *WorkspaceImportRegistrySettingInspection {
	return &WorkspaceImportRegistrySettingInspection{
		Name:  name,
		Link:  link,
		Match: match,
	}
}

// endregion
