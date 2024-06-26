package dsh_core

import (
	"dsh/dsh_utils"
	"regexp"
)

// region base

var workspaceRedirectLinkCheckRegex = regexp.MustCompile("^(git|dir):.*$")

// endregion

// region workspaceRedirectSetting

type workspaceRedirectSetting struct {
	Items []*workspaceRedirectItemSetting
}

func newWorkspaceRedirectSetting(items []*workspaceRedirectItemSetting) *workspaceRedirectSetting {
	return &workspaceRedirectSetting{
		Items: items,
	}
}

func (s *workspaceRedirectSetting) merge(setting *workspaceRedirectSetting) {
	s.Items = append(s.Items, setting.Items...)
}

func (s *workspaceRedirectSetting) getLink(originals []string, evaluator *Evaluator) (*projectLink, string, error) {
	for i := 0; i < len(originals); i++ {
		original := originals[i]
		for j := 0; j < len(s.Items); j++ {
			item := s.Items[j]
			matched, values := dsh_utils.RegexMatch(item.regex, original)
			if !matched {
				continue
			}
			matched, err := evaluator.EvalBoolExpr(item.match)
			if err != nil {
				return nil, "", errW(err, "get workspace import redirect setting link error",
					reason("eval expr error"),
					kv("item", item),
				)
			}
			if !matched {
				continue
			}
			evaluator2 := evaluator.SetData("regex", dsh_utils.MapStrStrToStrAny(values))
			rawLink, err := evaluator2.EvalStringTemplate(item.Link)
			if err != nil {
				return nil, "", errW(err, "get workspace import redirect setting link error",
					reason("eval template error"),
					kv("item", item),
				)
			}
			link, err := parseProjectLink(rawLink)
			if err != nil {
				return nil, "", errW(err, "get workspace import redirect setting link error",
					reason("parse link error"),
					kv("item", item),
					kv("rawLink", rawLink),
				)
			}
			return link, original, nil
		}
	}
	return nil, "", nil
}

func (s *workspaceRedirectSetting) inspect() *WorkspaceRedirectSettingInspection {
	var items []*WorkspaceRedirectItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newWorkspaceRedirectSettingInspection(items)
}

// endregion

// region workspaceRedirectItemSetting

type workspaceRedirectItemSetting struct {
	Regex string
	Link  string
	Match string
	regex *regexp.Regexp
	match *EvalExpr
}

func newWorkspaceRedirectItemSetting(regexStr string, link string, match string, regexObj *regexp.Regexp, matchObj *EvalExpr) *workspaceRedirectItemSetting {
	return &workspaceRedirectItemSetting{
		Regex: regexStr,
		Link:  link,
		Match: match,
		regex: regexObj,
		match: matchObj,
	}
}

func (s *workspaceRedirectItemSetting) inspect() *WorkspaceRedirectItemSettingInspection {
	return newWorkspaceRedirectItemSettingInspection(s.Regex, s.Link, s.Match)
}

// endregion

// region workspaceRedirectSettingModel

type workspaceRedirectSettingModel struct {
	Items []*workspaceRedirectItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newWorkspaceRedirectSettingModel(items []*workspaceRedirectItemSettingModel) *workspaceRedirectSettingModel {
	return &workspaceRedirectSettingModel{
		Items: items,
	}
}

func (m *workspaceRedirectSettingModel) convert(ctx *modelConvertContext) (*workspaceRedirectSetting, error) {
	var items []*workspaceRedirectItemSetting
	for i := 0; i < len(m.Items); i++ {
		item, err := m.Items[i].convert(ctx.ChildItem("items", i))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return newWorkspaceRedirectSetting(items), nil
}

// endregion

// region workspaceRedirectItemSettingModel

type workspaceRedirectItemSettingModel struct {
	Regex string `yaml:"regex" toml:"regex" json:"regex"`
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newWorkspaceRedirectItemSettingModel(regex, link, match string) *workspaceRedirectItemSettingModel {
	return &workspaceRedirectItemSettingModel{
		Regex: regex,
		Link:  link,
		Match: match,
	}
}

func (m *workspaceRedirectItemSettingModel) convert(ctx *modelConvertContext) (_ *workspaceRedirectItemSetting, err error) {
	if m.Regex == "" {
		return nil, ctx.Child("regex").NewValueEmptyError()
	}
	regexObj, err := regexp.Compile(m.Regex)
	if err != nil {
		return nil, ctx.Child("regex").WrapValueInvalidError(err, m.Regex)
	}

	if m.Link == "" {
		return nil, ctx.Child("link").NewValueEmptyError()
	}
	if !workspaceRedirectLinkCheckRegex.MatchString(m.Link) {
		return nil, ctx.Child("link").NewValueInvalidError(m.Link)
	}

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newWorkspaceRedirectItemSetting(m.Regex, m.Link, m.Match, regexObj, matchObj), nil
}

// endregion
