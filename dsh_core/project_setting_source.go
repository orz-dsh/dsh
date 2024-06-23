package dsh_core

import (
	"dsh/dsh_utils"
)

// region projectSourceSetting

type projectSourceSetting struct {
	Dir   string
	Files []string
	Match string
	match *EvalExpr
}

type projectSourceSettingSet []*projectSourceSetting

func newProjectSourceSetting(dir string, files []string, match string, matchObj *EvalExpr) *projectSourceSetting {
	return &projectSourceSetting{
		Dir:   dir,
		Files: files,
		Match: match,
		match: matchObj,
	}
}

func (s *projectSourceSetting) inspect() *ProjectSourceSettingInspection {
	return newProjectSourceSettingInspection(s.Dir, s.Files, s.Match)
}

func (s projectSourceSettingSet) inspect() []*ProjectSourceSettingInspection {
	var inspections []*ProjectSourceSettingInspection
	for i := 0; i < len(s); i++ {
		inspections = append(inspections, s[i].inspect())
	}
	return inspections
}

// endregion

// region projectSourceSettingModel

type projectSourceSettingModel struct {
	Dir   string
	Files []string
	Match string
}

func newProjectSourceSettingModel(dir string, files []string, match string) *projectSourceSettingModel {
	return &projectSourceSettingModel{
		Dir:   dir,
		Files: files,
		Match: match,
	}
}

func (m *projectSourceSettingModel) convert(ctx *modelConvertContext) (setting *projectSourceSetting, err error) {
	if m.Dir == "" {
		return nil, ctx.Child("dir").NewValueEmptyError()
	}

	for i := 0; i < len(m.Files); i++ {
		if m.Files[i] == "" {
			return nil, ctx.ChildItem("files", i).NewValueEmptyError()
		}
	}

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newProjectSourceSetting(m.Dir, m.Files, m.Match, matchObj), nil
}

// endregion

// region projectImportSettingModelSet

type projectSourceSettingModelSet []*projectSourceSettingModel

func (s projectSourceSettingModelSet) convert(ctx *modelConvertContext) (settings projectSourceSettingSet, _ error) {
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

// region ProjectSourceSettingInspection

type ProjectSourceSettingInspection struct {
	Dir   string   `yaml:"dir" toml:"dir" json:"dir"`
	Files []string `yaml:"files,omitempty" toml:"files,omitempty" json:"files,omitempty"`
	Match string   `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newProjectSourceSettingInspection(dir string, files []string, match string) *ProjectSourceSettingInspection {
	return &ProjectSourceSettingInspection{
		Dir:   dir,
		Files: files,
		Match: match,
	}
}

// endregion
