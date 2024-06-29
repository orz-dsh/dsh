package setting

import (
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
)

// region ProjectDependencySetting

type ProjectDependencySetting struct {
	Items []*ProjectDependencyItemSetting
}

func NewProjectDependencySetting(items []*ProjectDependencyItemSetting) *ProjectDependencySetting {
	return &ProjectDependencySetting{
		Items: items,
	}
}

func (s *ProjectDependencySetting) Inspect() *ProjectDependencySettingInspection {
	var items []*ProjectDependencyItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].Inspect())
	}
	return NewProjectDependencySettingInspection(items)
}

// endregion

// region ProjectDependencyItemSetting

type ProjectDependencyItemSetting struct {
	Link     string
	Match    string
	LinkObj  *ProjectLink
	MatchObj *EvalExpr
}

func NewProjectDependencyItemSetting(link string, match string, linkObj *ProjectLink, matchObj *EvalExpr) *ProjectDependencyItemSetting {
	return &ProjectDependencyItemSetting{
		Link:     link,
		Match:    match,
		LinkObj:  linkObj,
		MatchObj: matchObj,
	}
}

func (s *ProjectDependencyItemSetting) Inspect() *ProjectDependencyItemSettingInspection {
	return NewProjectDependencyItemSettingInspection(s.Link, s.Match)
}

// endregion

// region ProjectDependencySettingModel

type ProjectDependencySettingModel struct {
	Items []*ProjectDependencyItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewProjectDependencySettingModel(items []*ProjectDependencyItemSettingModel) *ProjectDependencySettingModel {
	return &ProjectDependencySettingModel{
		Items: items,
	}
}

func (m *ProjectDependencySettingModel) Convert(helper *ModelHelper) (*ProjectDependencySetting, error) {
	items, err := ConvertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return NewProjectDependencySetting(items), nil
}

// endregion

// region ProjectDependencyItemSettingModel

type ProjectDependencyItemSettingModel struct {
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewProjectDependencyItemSettingModel(link, match string) *ProjectDependencyItemSettingModel {
	return &ProjectDependencyItemSettingModel{
		Link:  link,
		Match: match,
	}
}

func (m *ProjectDependencyItemSettingModel) Convert(helper *ModelHelper) (*ProjectDependencyItemSetting, error) {
	if m.Link == "" {
		return nil, helper.Child("link").NewValueEmptyError()
	}
	linkObj, err := ParseProjectLink(m.Link)
	if err != nil {
		return nil, helper.Child("link").WrapValueInvalidError(err, m.Link)
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return NewProjectDependencyItemSetting(m.Link, m.Match, linkObj, matchObj), nil
}

// endregion
