package setting

import (
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
	"regexp"
)

// region base

var redirectLinkCheckRegex = regexp.MustCompile("^(git|dir):.*$")

// endregion

// region RedirectSetting

type RedirectSetting struct {
	Items []*RedirectItemSetting
}

func NewRedirectSetting(items []*RedirectItemSetting) *RedirectSetting {
	return &RedirectSetting{
		Items: items,
	}
}

func (s *RedirectSetting) Merge(other *RedirectSetting) {
	s.Items = append(s.Items, other.Items...)
}

func (s *RedirectSetting) GetLink(originals []string, evaluator *Evaluator) (*ProjectLink, string, error) {
	for i := 0; i < len(originals); i++ {
		original := originals[i]
		for j := 0; j < len(s.Items); j++ {
			item := s.Items[j]
			matched, values := RegexMatch(item.RegexObj, original)
			if !matched {
				continue
			}
			matched, err := evaluator.EvalBoolExpr(item.Match)
			if err != nil {
				return nil, "", ErrW(err, "get workspace import redirect setting link error",
					Reason("eval expr error"),
					KV("item", item),
				)
			}
			if !matched {
				continue
			}
			evaluator2 := evaluator.SetData("regex", MapAnyByStr(values))
			rawLink, err := evaluator2.EvalStringTemplate(item.Link)
			if err != nil {
				return nil, "", ErrW(err, "get workspace import redirect setting link error",
					Reason("eval template error"),
					KV("item", item),
				)
			}
			link, err := ParseProjectLink(rawLink)
			if err != nil {
				return nil, "", ErrW(err, "get workspace import redirect setting link error",
					Reason("parse link error"),
					KV("item", item),
					KV("rawLink", rawLink),
				)
			}
			return link, original, nil
		}
	}
	return nil, "", nil
}

func (s *RedirectSetting) Inspect() *RedirectSettingInspection {
	var items []*RedirectItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].Inspect())
	}
	return NewRedirectSettingInspection(items)
}

// endregion

// region RedirectItemSetting

type RedirectItemSetting struct {
	Regex    string
	Link     string
	Match    string
	RegexObj *regexp.Regexp
}

func NewRedirectItemSetting(regex, link, match string, regexObj *regexp.Regexp) *RedirectItemSetting {
	return &RedirectItemSetting{
		Regex:    regex,
		Link:     link,
		Match:    match,
		RegexObj: regexObj,
	}
}

func (s *RedirectItemSetting) Inspect() *RedirectItemSettingInspection {
	return NewRedirectItemSettingInspection(s.Regex, s.Link, s.Match)
}

// endregion

// region RedirectSettingModel

type RedirectSettingModel struct {
	Items []*RedirectItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewRedirectSettingModel(items []*RedirectItemSettingModel) *RedirectSettingModel {
	return &RedirectSettingModel{
		Items: items,
	}
}

func (m *RedirectSettingModel) Convert(helper *ModelHelper) (*RedirectSetting, error) {
	items, err := ConvertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return NewRedirectSetting(items), nil
}

// endregion

// region RedirectItemSettingModel

type RedirectItemSettingModel struct {
	Regex string `yaml:"regex" toml:"regex" json:"regex"`
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewRedirectItemSettingModel(regex, link, match string) *RedirectItemSettingModel {
	return &RedirectItemSettingModel{
		Regex: regex,
		Link:  link,
		Match: match,
	}
}

func (m *RedirectItemSettingModel) Convert(helper *ModelHelper) (*RedirectItemSetting, error) {
	if m.Regex == "" {
		return nil, helper.Child("regex").NewValueEmptyError()
	}
	regexObj, err := regexp.Compile(m.Regex)
	if err != nil {
		return nil, helper.Child("regex").WrapValueInvalidError(err, m.Regex)
	}

	if m.Link == "" {
		return nil, helper.Child("link").NewValueEmptyError()
	}
	if !redirectLinkCheckRegex.MatchString(m.Link) {
		return nil, helper.Child("link").NewValueInvalidError(m.Link)
	}

	return NewRedirectItemSetting(m.Regex, m.Link, m.Match, regexObj), nil
}

// endregion
