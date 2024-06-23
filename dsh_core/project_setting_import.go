package dsh_core

import (
	"dsh/dsh_utils"
)

// region projectImportSetting

type projectImportSetting struct {
	Link  string
	Match string
	link  *projectLink
	match *EvalExpr
}

type projectImportSettingSet []*projectImportSetting

func newProjectImportSetting(link string, match string, linkObj *projectLink, matchObj *EvalExpr) *projectImportSetting {
	return &projectImportSetting{
		Link:  link,
		Match: match,
		link:  linkObj,
		match: matchObj,
	}
}

func (s *projectImportSetting) inspect() *ProjectImportSettingInspection {
	return newProjectImportSettingInspection(s.Link, s.Match)
}

func (s projectImportSettingSet) inspect() []*ProjectImportSettingInspection {
	var inspections []*ProjectImportSettingInspection
	for i := 0; i < len(s); i++ {
		inspections = append(inspections, s[i].inspect())
	}
	return inspections
}

// endregion

// region projectImportSettingModel

type projectImportSettingModel struct {
	Link  string
	Match string
}

func newProjectImportSettingModel(link, match string) *projectImportSettingModel {
	return &projectImportSettingModel{
		Link:  link,
		Match: match,
	}
}

func (m *projectImportSettingModel) convert(ctx *modelConvertContext) (setting *projectImportSetting, err error) {
	if m.Link == "" {
		return nil, ctx.Child("link").NewValueEmptyError()
	}
	linkObj, err := parseProjectLink(m.Link)
	if err != nil {
		return nil, ctx.Child("link").WrapValueInvalidError(err, m.Link)
	}

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newProjectImportSetting(m.Link, m.Match, linkObj, matchObj), nil
}

// endregion

// region projectImportSettingModelSet

type projectImportSettingModelSet []*projectImportSettingModel

func (s projectImportSettingModelSet) convert(ctx *modelConvertContext) (settings projectImportSettingSet, _ error) {
	for i := 0; i < len(s); i++ {
		setting, err := s[i].convert(ctx.Item(i))
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}
	return settings, nil
}

// endregion

// region ProjectImportSettingInspection

type ProjectImportSettingInspection struct {
	Link  string `yaml:"link" toml:"link" json:"link"`
	Match string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newProjectImportSettingInspection(link, match string) *ProjectImportSettingInspection {
	return &ProjectImportSettingInspection{
		Link:  link,
		Match: match,
	}
}

// endregion
