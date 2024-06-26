package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
)

// region workspaceProfileSetting

type workspaceProfileSetting struct {
	Items []*workspaceProfileItemSetting
}

func newWorkspaceProfileSetting(items []*workspaceProfileItemSetting) *workspaceProfileSetting {
	return &workspaceProfileSetting{
		Items: items,
	}
}

func (s *workspaceProfileSetting) getFiles(evaluator *Evaluator) ([]string, error) {
	var files []string
	for i := 0; i < len(s.Items); i++ {
		item := s.Items[i]
		if matched, err := evaluator.EvalBoolExpr(item.match); err != nil {
			return nil, errW(err, "get workspace profile setting files error",
				reason("eval expr error"),
				kv("item", item),
			)
		} else if matched {
			rawFile, err := evaluator.EvalStringTemplate(item.File)
			if err != nil {
				return nil, errW(err, "get workspace profile setting files error",
					reason("eval template error"),
					kv("item", item),
				)
			}
			file, err := filepath.Abs(rawFile)
			if err != nil {
				return nil, errW(err, "get workspace profile setting files error",
					reason("get abs-path error"),
					kv("item", item),
					kv("rawFile", rawFile),
				)
			}
			if dsh_utils.IsFileExists(file) {
				files = append(files, file)
			} else if !item.Optional {
				return nil, errN("get workspace profile setting files error",
					reason("file not found"),
					kv("item", item),
					kv("rawFile", rawFile),
					kv("file", file),
				)
			}
		}
	}
	return files, nil
}

// endregion

// region workspaceProfileItemSetting

type workspaceProfileItemSetting struct {
	File     string
	Optional bool
	Match    string
	match    *EvalExpr
}

func newWorkspaceProfileItemSetting(file string, optional bool, match string, matchObj *EvalExpr) *workspaceProfileItemSetting {
	return &workspaceProfileItemSetting{
		File:     file,
		Optional: optional,
		Match:    match,
		match:    matchObj,
	}
}

// endregion

// region workspaceProfileSettingModel

type workspaceProfileSettingModel struct {
	Items []*workspaceProfileItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func (m *workspaceProfileSettingModel) convert(ctx *modelConvertContext) (*workspaceProfileSetting, error) {
	var items []*workspaceProfileItemSetting
	for i := 0; i < len(m.Items); i++ {
		item, err := m.Items[i].convert(ctx.ChildItem("items", i))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return newWorkspaceProfileSetting(items), nil
}

// endregion

// region workspaceProfileItemSettingModel

type workspaceProfileItemSettingModel struct {
	File     string `yaml:"file" toml:"file" json:"file"`
	Optional bool   `yaml:"optional" toml:"optional" json:"optional"`
	Match    string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func (m *workspaceProfileItemSettingModel) convert(ctx *modelConvertContext) (_ *workspaceProfileItemSetting, err error) {
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

	return newWorkspaceProfileItemSetting(m.File, m.Optional, m.Match, matchObj), nil
}

// endregion
