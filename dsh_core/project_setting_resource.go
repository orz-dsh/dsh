package dsh_core

import (
	"dsh/dsh_utils"
)

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

func (m *projectResourceSettingModel) convert(ctx *modelConvertContext) (*projectResourceSetting, error) {
	var items []*projectResourceItemSetting
	for i := 0; i < len(m.Items); i++ {
		item, err := m.Items[i].convert(ctx.ChildItem("items", i))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
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

func (m *projectResourceItemSettingModel) convert(ctx *modelConvertContext) (setting *projectResourceItemSetting, err error) {
	if m.Dir == "" {
		return nil, ctx.Child("dir").NewValueEmptyError()
	}

	for i := 0; i < len(m.Includes); i++ {
		if m.Includes[i] == "" {
			return nil, ctx.ChildItem("includes", i).NewValueEmptyError()
		}
	}

	for i := 0; i < len(m.Excludes); i++ {
		if m.Excludes[i] == "" {
			return nil, ctx.ChildItem("excludes", i).NewValueEmptyError()
		}
	}

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newProjectResourceItemSetting(m.Dir, m.Includes, m.Excludes, m.Match, matchObj), nil
}

// endregion
