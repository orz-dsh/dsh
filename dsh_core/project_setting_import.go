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

func (m *projectImportSettingModel) convert(ctx *ModelConvertContext) (setting *projectImportSetting, err error) {
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
