package setting

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/utils"
)

// region ProjectResourceSetting

type ProjectResourceSetting struct {
	Items []*ProjectResourceItemSetting
}

func NewProjectResourceSetting(items []*ProjectResourceItemSetting) *ProjectResourceSetting {
	return &ProjectResourceSetting{
		Items: items,
	}
}

func (s *ProjectResourceSetting) Inspect() *ProjectResourceSettingInspection {
	var items []*ProjectResourceItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].Inspect())
	}
	return NewProjectResourceSettingInspection(items)
}

// endregion

// region ProjectResourceItemSetting

type ProjectResourceItemSetting struct {
	Dir      string
	Includes []string
	Excludes []string
	Match    string
	MatchObj *EvalExpr
}

func NewProjectResourceItemSetting(dir string, includes, excludes []string, match string, matchObj *EvalExpr) *ProjectResourceItemSetting {
	return &ProjectResourceItemSetting{
		Dir:      dir,
		Includes: includes,
		Excludes: excludes,
		Match:    match,
		MatchObj: matchObj,
	}
}

func (s *ProjectResourceItemSetting) Inspect() *ProjectResourceItemSettingInspection {
	return NewProjectResourceItemSettingInspection(s.Dir, s.Includes, s.Excludes, s.Match)
}

// endregion

// region ProjectResourceSettingModel

type ProjectResourceSettingModel struct {
	Items []*ProjectResourceItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewProjectResourceSettingModel(items []*ProjectResourceItemSettingModel) *ProjectResourceSettingModel {
	return &ProjectResourceSettingModel{
		Items: items,
	}
}

func (m *ProjectResourceSettingModel) Convert(helper *ModelHelper) (*ProjectResourceSetting, error) {
	items, err := ConvertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return NewProjectResourceSetting(items), nil
}

// endregion

// region ProjectResourceItemSettingModel

type ProjectResourceItemSettingModel struct {
	Dir      string   `yaml:"dir" toml:"dir" json:"dir"`
	Includes []string `yaml:"includes,omitempty" toml:"includes,omitempty" json:"includes,omitempty"`
	Excludes []string `yaml:"excludes,omitempty" toml:"excludes,omitempty" json:"excludes,omitempty"`
	Match    string   `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewProjectResourceItemSettingModel(dir string, includes, excludes []string, match string) *ProjectResourceItemSettingModel {
	return &ProjectResourceItemSettingModel{
		Dir:      dir,
		Includes: includes,
		Excludes: excludes,
		Match:    match,
	}
}

func (m *ProjectResourceItemSettingModel) Convert(helper *ModelHelper) (*ProjectResourceItemSetting, error) {
	if m.Dir == "" {
		return nil, helper.Child("dir").NewValueEmptyError()
	}

	if err := helper.CheckStringItemEmpty("includes", m.Includes); err != nil {
		return nil, err
	}

	if err := helper.CheckStringItemEmpty("excludes", m.Excludes); err != nil {
		return nil, err
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return NewProjectResourceItemSetting(m.Dir, m.Includes, m.Excludes, m.Match, matchObj), nil
}

// endregion
