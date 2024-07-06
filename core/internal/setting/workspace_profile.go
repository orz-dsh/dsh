package setting

import (
	. "github.com/orz-dsh/dsh/utils"
	"path/filepath"
)

// region WorkspaceProfileSetting

type WorkspaceProfileSetting struct {
	Items []*WorkspaceProfileItemSetting
}

func NewWorkspaceProfileSetting(items []*WorkspaceProfileItemSetting) *WorkspaceProfileSetting {
	return &WorkspaceProfileSetting{
		Items: items,
	}
}

func (s *WorkspaceProfileSetting) Merge(other *WorkspaceProfileSetting) {
	s.Items = append(s.Items, other.Items...)
}

func (s *WorkspaceProfileSetting) GetFiles(evaluator *Evaluator) ([]string, error) {
	var files []string
	for i := 0; i < len(s.Items); i++ {
		item := s.Items[i]
		if matched, err := evaluator.EvalBoolExpr(item.match); err != nil {
			return nil, ErrW(err, "get workspace profile setting files error",
				Reason("eval expr error"),
				KV("item", item),
			)
		} else if matched {
			rawFile, err := evaluator.EvalStringTemplate(item.File)
			if err != nil {
				return nil, ErrW(err, "get workspace profile setting files error",
					Reason("eval template error"),
					KV("item", item),
				)
			}
			file, err := filepath.Abs(rawFile)
			if err != nil {
				return nil, ErrW(err, "get workspace profile setting files error",
					Reason("get abs-path error"),
					KV("item", item),
					KV("rawFile", rawFile),
				)
			}
			if IsFileExists(file) {
				files = append(files, file)
			} else if !item.Optional {
				return nil, ErrN("get workspace profile setting files error",
					Reason("file not found"),
					KV("item", item),
					KV("rawFile", rawFile),
					KV("file", file),
				)
			}
		}
	}
	return files, nil
}

// endregion

// region WorkspaceProfileItemSetting

type WorkspaceProfileItemSetting struct {
	File     string
	Optional bool
	Match    string
	match    *EvalExpr
}

func NewWorkspaceProfileItemSetting(file string, optional bool, match string, matchObj *EvalExpr) *WorkspaceProfileItemSetting {
	return &WorkspaceProfileItemSetting{
		File:     file,
		Optional: optional,
		Match:    match,
		match:    matchObj,
	}
}

// endregion

// region WorkspaceProfileSettingModel

type WorkspaceProfileSettingModel struct {
	Items []*WorkspaceProfileItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func NewWorkspaceProfileSettingModel(items []*WorkspaceProfileItemSettingModel) *WorkspaceProfileSettingModel {
	return &WorkspaceProfileSettingModel{
		Items: items,
	}
}

func (m *WorkspaceProfileSettingModel) Convert(helper *ModelHelper) (*WorkspaceProfileSetting, error) {
	items, err := ConvertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return NewWorkspaceProfileSetting(items), nil
}

// endregion

// region WorkspaceProfileItemSettingModel

type WorkspaceProfileItemSettingModel struct {
	File     string `yaml:"file" toml:"file" json:"file"`
	Optional bool   `yaml:"optional,omitempty" toml:"optional,omitempty" json:"optional,omitempty"`
	Match    string `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func NewWorkspaceProfileItemSettingModel(file string, optional bool, match string) *WorkspaceProfileItemSettingModel {
	return &WorkspaceProfileItemSettingModel{
		File:     file,
		Optional: optional,
		Match:    match,
	}
}

func (m *WorkspaceProfileItemSettingModel) Convert(helper *ModelHelper) (*WorkspaceProfileItemSetting, error) {
	if m.File == "" {
		return nil, helper.Child("file").NewValueEmptyError()
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return NewWorkspaceProfileItemSetting(m.File, m.Optional, m.Match, matchObj), nil
}

// endregion
