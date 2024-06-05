package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
)

// region workspaceProfileSetting

type workspaceProfileSetting struct {
	File     string
	Optional bool
	Match    string
	match    *EvalExpr
}

type workspaceProfileSettingSet []*workspaceProfileSetting

func newWorkspaceProfileSetting(file string, optional bool, match string, matchObj *EvalExpr) *workspaceProfileSetting {
	return &workspaceProfileSetting{
		File:     file,
		Optional: optional,
		Match:    match,
		match:    matchObj,
	}
}

func (s workspaceProfileSettingSet) getFiles(evaluator *Evaluator) ([]string, error) {
	var files []string
	for i := 0; i < len(s); i++ {
		schema := s[i]
		if matched, err := evaluator.EvalBoolExpr(schema.match); err != nil {
			return nil, errW(err, "get workspace profile setting files error",
				reason("eval expr error"),
				kv("schema", schema),
			)
		} else if matched {
			rawFile, err := evaluator.EvalStringTemplate(schema.File)
			if err != nil {
				return nil, errW(err, "get workspace profile setting files error",
					reason("eval template error"),
					kv("schema", schema),
				)
			}
			file, err := filepath.Abs(rawFile)
			if err != nil {
				return nil, errW(err, "get workspace profile setting files error",
					reason("get abs-path error"),
					kv("schema", schema),
					kv("rawFile", rawFile),
				)
			}
			if dsh_utils.IsFileExists(file) {
				files = append(files, file)
			} else if !schema.Optional {
				return nil, errN("get workspace profile setting files error",
					reason("file not found"),
					kv("schema", schema),
					kv("rawFile", rawFile),
					kv("file", file),
				)
			}
		}
	}
	return files, nil
}

// endregion

// region workspaceProfileSettingModel

type workspaceProfileSettingModel struct {
	Items []*workspaceProfileItemSettingModel
}

func (m *workspaceProfileSettingModel) convert(ctx *modelConvertContext) (workspaceProfileSettingSet, error) {
	settings := workspaceProfileSettingSet{}
	for i := 0; i < len(m.Items); i++ {
		if model, err := m.Items[i].convert(ctx.ChildItem("items", i)); err != nil {
			return nil, err
		} else {
			settings = append(settings, model)
		}
	}

	return settings, nil
}

// endregion

// region workspaceProfileItemSettingModel

type workspaceProfileItemSettingModel struct {
	File     string
	Optional bool
	Match    string
}

func (m *workspaceProfileItemSettingModel) convert(ctx *modelConvertContext) (setting *workspaceProfileSetting, err error) {
	if m.File == "" {
		return nil, ctx.Child("file").NewValueEmptyError()
	}

	var matchObj *EvalExpr
	if m.Match != "" {
		matchObj, err = dsh_utils.CompileExpr(m.Match)
		if err != nil {
			return nil, ctx.Child("match").WrapValueInvalidError(err, m.Match)
		}
	}

	return newWorkspaceProfileSetting(m.File, m.Optional, m.Match, matchObj), nil
}

// endregion
