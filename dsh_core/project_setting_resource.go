package dsh_core

// region projectResourceSetting

type projectResourceSetting struct {
	Items []*projectResourceItemSetting
}

func newProjectResourceSetting(items []*projectResourceItemSetting) *projectResourceSetting {
	return &projectResourceSetting{
		Items: items,
	}
}

func (s *projectResourceSetting) inspect() *ProjectResourceSettingInspection {
	var items []*ProjectResourceItemSettingInspection
	for i := 0; i < len(s.Items); i++ {
		items = append(items, s.Items[i].inspect())
	}
	return newProjectResourceSettingInspection(items)
}

// endregion

// region projectResourceItemSetting

type projectResourceItemSetting struct {
	Dir      string
	Includes []string
	Excludes []string
	Match    string
	match    *EvalExpr
}

func newProjectResourceItemSetting(dir string, includes, excludes []string, match string, matchObj *EvalExpr) *projectResourceItemSetting {
	return &projectResourceItemSetting{
		Dir:      dir,
		Includes: includes,
		Excludes: excludes,
		Match:    match,
		match:    matchObj,
	}
}

func (s *projectResourceItemSetting) inspect() *ProjectResourceItemSettingInspection {
	return newProjectResourceItemSettingInspection(s.Dir, s.Includes, s.Excludes, s.Match)
}

// endregion

// region projectResourceSettingModel

type projectResourceSettingModel struct {
	Items []*projectResourceItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newProjectResourceSettingModel(items []*projectResourceItemSettingModel) *projectResourceSettingModel {
	return &projectResourceSettingModel{
		Items: items,
	}
}

func (m *projectResourceSettingModel) convert(helper *modelHelper) (*projectResourceSetting, error) {
	items, err := convertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return newProjectResourceSetting(items), nil
}

// endregion

// region projectResourceItemSettingModel

type projectResourceItemSettingModel struct {
	Dir      string
	Includes []string
	Excludes []string
	Match    string
}

func newProjectResourceItemSettingModel(dir string, includes, excludes []string, match string) *projectResourceItemSettingModel {
	return &projectResourceItemSettingModel{
		Dir:      dir,
		Includes: includes,
		Excludes: excludes,
		Match:    match,
	}
}

func (m *projectResourceItemSettingModel) convert(helper *modelHelper) (*projectResourceItemSetting, error) {
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

	return newProjectResourceItemSetting(m.Dir, m.Includes, m.Excludes, m.Match, matchObj), nil
}

// endregion
