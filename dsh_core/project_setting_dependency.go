package dsh_core

// region projectDependencySetting

type projectDependencySetting struct {
	Items []*projectDependencyItemSetting
}

func newProjectDependencySetting(items []*projectDependencyItemSetting) *projectDependencySetting {
	return &projectDependencySetting{
		Items: items,
	}
}

func (s *projectDependencySetting) inspect() *ProjectDependencySettingInspection {
	var items []*ProjectDependencyItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newProjectDependencySettingInspection(items)
}

// endregion

// region projectDependencyItemSetting

type projectDependencyItemSetting struct {
	Link  string
	Match string
	link  *projectLink
	match *EvalExpr
}

func newProjectDependencyItemSetting(link string, match string, linkObj *projectLink, matchObj *EvalExpr) *projectDependencyItemSetting {
	return &projectDependencyItemSetting{
		Link:  link,
		Match: match,
		link:  linkObj,
		match: matchObj,
	}
}

func (s *projectDependencyItemSetting) inspect() *ProjectDependencyItemSettingInspection {
	return newProjectDependencyItemSettingInspection(s.Link, s.Match)
}

// endregion

// region projectDependencySettingModel

type projectDependencySettingModel struct {
	Items []*projectDependencyItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProjectDependencySettingModel(items []*projectDependencyItemSettingModel) *projectDependencySettingModel {
	return &projectDependencySettingModel{
		Items: items,
	}
}

func (m *projectDependencySettingModel) convert(helper *modelHelper) (*projectDependencySetting, error) {
	items, err := convertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return newProjectDependencySetting(items), nil
}

// endregion

// region projectDependencyItemSettingModel

type projectDependencyItemSettingModel struct {
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newProjectDependencyItemSettingModel(link, match string) *projectDependencyItemSettingModel {
	return &projectDependencyItemSettingModel{
		Link:  link,
		Match: match,
	}
}

func (m *projectDependencyItemSettingModel) convert(helper *modelHelper) (*projectDependencyItemSetting, error) {
	if m.Link == "" {
		return nil, helper.Child("link").NewValueEmptyError()
	}
	linkObj, err := parseProjectLink(m.Link)
	if err != nil {
		return nil, helper.Child("link").WrapValueInvalidError(err, m.Link)
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return newProjectDependencyItemSetting(m.Link, m.Match, linkObj, matchObj), nil
}

// endregion
