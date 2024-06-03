package dsh_core

import "dsh/dsh_utils"

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

// endregion

// region workspaceImportRegistrySettingModel

type workspaceImportRegistrySettingModel struct {
	Items []*workspaceImportRegistryItemSettingModel
}

func (m *workspaceImportRegistrySettingModel) convert(ctx *ModelConvertContext) (workspaceImportRegistrySettingSet, error) {
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

func (m *workspaceImportRegistryItemSettingModel) convert(ctx *ModelConvertContext) (setting *workspaceImportRegistrySetting, err error) {
	if m.Name == "" {
		return nil, ctx.Child("name").NewValueEmptyError()
	}

	if m.Link == "" {
		return nil, ctx.Child("link").NewValueEmptyError()
	}
	// TODO: check link template

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