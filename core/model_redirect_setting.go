package core

import "regexp"

// region base

var redirectLinkCheckRegex = regexp.MustCompile("^(git|dir):.*$")

// endregion

// region redirectSettingModel

type redirectSettingModel struct {
	Items []*redirectItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newRedirectSettingModel(items []*redirectItemSettingModel) *redirectSettingModel {
	return &redirectSettingModel{
		Items: items,
	}
}

func (m *redirectSettingModel) convert(helper *modelHelper) (*redirectSetting, error) {
	items, err := convertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return newRedirectSetting(items), nil
}

// endregion

// region redirectItemSettingModel

type redirectItemSettingModel struct {
	Regex string `yaml:"regex" toml:"regex" json:"regex"`
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newRedirectItemSettingModel(regex, link, match string) *redirectItemSettingModel {
	return &redirectItemSettingModel{
		Regex: regex,
		Link:  link,
		Match: match,
	}
}

func (m *redirectItemSettingModel) convert(helper *modelHelper) (*redirectItemSetting, error) {
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

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return newRedirectItemSetting(m.Regex, m.Link, m.Match, regexObj, matchObj), nil
}

// endregion

// region RedirectSettingModelBuilder

type RedirectSettingModelBuilder[P any] struct {
	commit func(*redirectSettingModel) P
	items  []*redirectItemSettingModel
}

func newRedirectSettingModelBuilder[P any](commit func(*redirectSettingModel) P) *RedirectSettingModelBuilder[P] {
	return &RedirectSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *RedirectSettingModelBuilder[P]) AddItem(regex, link, match string) *RedirectSettingModelBuilder[P] {
	b.items = append(b.items, newRedirectItemSettingModel(regex, link, match))
	return b
}

func (b *RedirectSettingModelBuilder[P]) CommitRedirectSetting() P {
	return b.commit(newRedirectSettingModel(b.items))
}

// endregion
