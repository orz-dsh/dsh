package dsh_core

import (
	"dsh/dsh_utils"
	"regexp"
)

// region base

var workspaceImportRedirectLinkCheckRegex = regexp.MustCompile("^(git|dir):.*$")

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

// region workspaceImportRedirectSettingModel

type workspaceImportRedirectSettingModel struct {
	Items []*workspaceImportRedirectItemSettingModel
}

func (m *workspaceImportRedirectSettingModel) convert(ctx *modelConvertContext) (workspaceImportRedirectSettingSet, error) {
	settings := workspaceImportRedirectSettingSet{}
	for i := 0; i < len(m.Items); i++ {
		if model, err := m.Items[i].convert(ctx.ChildItem("items", i)); err != nil {
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

func newWorkspaceImportRedirectItemSettingModel(regex, link, match string) *workspaceImportRedirectItemSettingModel {
	return &workspaceImportRedirectItemSettingModel{
		Regex: regex,
		Link:  link,
		Match: match,
	}
}

func (m *workspaceImportRedirectItemSettingModel) convert(ctx *modelConvertContext) (setting *workspaceImportRedirectSetting, err error) {
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
	if !workspaceImportRedirectLinkCheckRegex.MatchString(m.Link) {
		return nil, ctx.Child("link").NewValueInvalidError(m.Link)
	}

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newWorkspaceImportRedirectSetting(m.Regex, m.Link, m.Match, regexObj, matchObj), nil
}

// endregion
