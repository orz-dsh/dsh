package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"regexp"
)

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

// region workspaceImportRedirectSetting

type workspaceImportRedirectSetting struct {
	Regex string
	Link  string
	Match string
	regex *regexp.Regexp
	match *EvalExpr
}

type workspaceImportRedirectSettingSet []*workspaceImportRedirectSetting

func newWorkspaceImportRedirectSetting(regexStr string, link string, match string, regexObj *regexp.Regexp, matchObj *EvalExpr) *workspaceImportRedirectSetting {
	return &workspaceImportRedirectSetting{
		Regex: regexStr,
		Link:  link,
		Match: match,
		regex: regexObj,
		match: matchObj,
	}
}

func (s workspaceImportRedirectSettingSet) getLink(originals []string, evaluator *Evaluator) (*projectLink, string, error) {
	for i := 0; i < len(originals); i++ {
		original := originals[i]
		for j := 0; j < len(s); j++ {
			model := s[j]
			matched, values := dsh_utils.RegexMatch(model.regex, original)
			if !matched {
				continue
			}
			matched, err := evaluator.EvalBoolExpr(model.match)
			if err != nil {
				return nil, "", errW(err, "get workspace import redirect setting link error",
					reason("eval expr error"),
					kv("model", model),
				)
			}
			if !matched {
				continue
			}
			evaluator2 := evaluator.SetData("regex", dsh_utils.MapStrStrToStrAny(values))
			rawLink, err := evaluator2.EvalStringTemplate(model.Link)
			if err != nil {
				return nil, "", errW(err, "get workspace import redirect setting link error",
					reason("eval template error"),
					kv("model", model),
				)
			}
			link, err := parseProjectLink(rawLink)
			if err != nil {
				return nil, "", errW(err, "get workspace import redirect setting link error",
					reason("parse link error"),
					kv("model", model),
					kv("rawLink", rawLink),
				)
			}
			return link, original, nil
		}
	}
	return nil, "", nil
}

// endregion

// region workspaceImportSettingModel

type workspaceImportSettingModel struct {
	Registry *workspaceImportRegistrySettingModel
	Redirect *workspaceImportRedirectSettingModel
}

func (m *workspaceImportSettingModel) convert(root *workspaceSettingModel) (registrySettings workspaceImportRegistrySettingSet, redirectSettings workspaceImportRedirectSettingSet, err error) {
	if registrySettings, err = m.Registry.convert(root); err != nil {
		return nil, nil, err
	}
	if redirectSettings, err = m.Redirect.convert(root); err != nil {
		return nil, nil, err
	}
	return registrySettings, redirectSettings, nil
}

// endregion

// region workspaceImportRegistrySettingModel

type workspaceImportRegistrySettingModel struct {
	Items []*workspaceImportRegistryItemSettingModel
}

func (m *workspaceImportRegistrySettingModel) convert(root *workspaceSettingModel) (workspaceImportRegistrySettingSet, error) {
	settings := workspaceImportRegistrySettingSet{}
	for i := 0; i < len(m.Items); i++ {
		item := m.Items[i]
		if setting, err := item.convert(root, i); err != nil {
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

func (m *workspaceImportRegistryItemSettingModel) convert(root *workspaceSettingModel, itemIndex int) (setting *workspaceImportRegistrySetting, err error) {
	if m.Name == "" {
		return nil, errN("workspace setting invalid",
			reason("value empty"),
			kv("path", root.path),
			kv("field", fmt.Sprintf("import.registry.items[%d].name", itemIndex)),
		)
	}

	if m.Link == "" {
		return nil, errN("workspace setting invalid",
			reason("value empty"),
			kv("path", root.path),
			kv("field", fmt.Sprintf("import.registry.items[%d].link", itemIndex)),
		)
	}
	// TODO: check link template

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, errW(err, "workspace setting invalid",
				reason("value invalid"),
				kv("path", root.path),
				kv("field", fmt.Sprintf("import.registry.items[%d].match", itemIndex)),
				kv("value", m.Match),
			)
		}
	}

	return newWorkspaceImportRegistrySetting(m.Name, m.Link, m.Match, matchObj), nil
}

// endregion

// region workspaceImportRedirectSettingModel

type workspaceImportRedirectSettingModel struct {
	Items []*workspaceImportRedirectItemSettingModel
}

func (m *workspaceImportRedirectSettingModel) convert(root *workspaceSettingModel) (workspaceImportRedirectSettingSet, error) {
	settings := workspaceImportRedirectSettingSet{}
	for i := 0; i < len(m.Items); i++ {
		if model, err := m.Items[i].convert(root, i); err != nil {
			return nil, err
		} else {
			settings = append(settings, model)
		}
	}
	return settings, nil
}

// endregion

// region workspaceImportRedirectItemSettingModel

type workspaceImportRedirectItemSettingModel struct {
	Regex string
	Link  string
	Match string
}

func (m *workspaceImportRedirectItemSettingModel) convert(root *workspaceSettingModel, itemIndex int) (setting *workspaceImportRedirectSetting, err error) {
	if m.Regex == "" {
		return nil, errN("workspace setting invalid",
			reason("value empty"),
			kv("path", root.path),
			kv("field", fmt.Sprintf("import.redirect.items[%d].regex", itemIndex)),
		)
	}
	regexObj, err := regexp.Compile(m.Regex)
	if err != nil {
		return nil, errW(err, "workspace setting invalid",
			reason("value invalid"),
			kv("path", root.path),
			kv("field", fmt.Sprintf("import.redirect.items[%d].regex", itemIndex)),
			kv("value", m.Regex),
		)
	}

	if m.Link == "" {
		return nil, errN("workspace setting invalid",
			reason("value empty"),
			kv("path", root.path),
			kv("field", fmt.Sprintf("import.redirect.items[%d].link", itemIndex)),
		)
	}
	// TODO: check link template

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, errW(err, "workspace setting invalid",
				reason("value invalid"),
				kv("path", root.path),
				kv("field", fmt.Sprintf("import.redirect.items[%d].match", itemIndex)),
				kv("value", m.Match),
			)
		}
	}

	return newWorkspaceImportRedirectSetting(m.Regex, m.Link, m.Match, regexObj, matchObj), nil
}

// endregion
