package core

import (
	"github.com/orz-dsh/dsh/utils"
	"regexp"
)

// region redirectSetting

type redirectSetting struct {
	Items []*redirectItemSetting
}

func newRedirectSetting(items []*redirectItemSetting) *redirectSetting {
	return &redirectSetting{
		Items: items,
	}
}

func (s *redirectSetting) merge(setting *redirectSetting) {
	s.Items = append(s.Items, setting.Items...)
}

func (s *redirectSetting) getLink(originals []string, evaluator *Evaluator) (*projectLink, string, error) {
	for i := 0; i < len(originals); i++ {
		original := originals[i]
		for j := 0; j < len(s.Items); j++ {
			item := s.Items[j]
			matched, values := utils.RegexMatch(item.regex, original)
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
			evaluator2 := evaluator.SetData("regex", utils.MapStrStrToStrAny(values))
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

func (s *redirectSetting) inspect() *WorkspaceRedirectSettingInspection {
	var items []*WorkspaceRedirectItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newWorkspaceRedirectSettingInspection(items)
}

// endregion

// region redirectItemSetting

type redirectItemSetting struct {
	Regex string
	Link  string
	Match string
	regex *regexp.Regexp
	match *EvalExpr
}

func newRedirectItemSetting(regexStr string, link string, match string, regexObj *regexp.Regexp, matchObj *EvalExpr) *redirectItemSetting {
	return &redirectItemSetting{
		Regex: regexStr,
		Link:  link,
		Match: match,
		regex: regexObj,
		match: matchObj,
	}
}

func (s *redirectItemSetting) inspect() *WorkspaceRedirectItemSettingInspection {
	return newWorkspaceRedirectItemSettingInspection(s.Regex, s.Link, s.Match)
}

// endregion
