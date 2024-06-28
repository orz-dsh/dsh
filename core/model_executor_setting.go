package core

import "github.com/orz-dsh/dsh/utils"

// region executorSettingModel

type executorSettingModel struct {
	Items []*executorItemSettingModel `yaml:"items,omitempty" toml:"items,omitempty" json:"items,omitempty"`
}

func newExecutorSettingModel(items []*executorItemSettingModel) *executorSettingModel {
	return &executorSettingModel{
		Items: items,
	}
}

func (m *executorSettingModel) convert(helper *modelHelper) (*executorSetting, error) {
	items, err := convertChildModels(helper, "items", m.Items)
	if err != nil {
		return nil, err
	}
	return newExecutorSetting(items), nil
}

// endregion

// region executorItemSettingModel

type executorItemSettingModel struct {
	Name  string   `yaml:"name" toml:"name" json:"name"`
	File  string   `yaml:"file,omitempty" toml:"file,omitempty" json:"file,omitempty"`
	Exts  []string `yaml:"exts,omitempty" toml:"exts,omitempty" json:"exts,omitempty"`
	Args  []string `yaml:"args,omitempty" toml:"args,omitempty" json:"args,omitempty"`
	Match string   `yaml:"match,omitempty" toml:"match,omitempty" json:"match,omitempty"`
}

func newExecutorItemSettingModel(name, file string, exts, args []string, match string) *executorItemSettingModel {
	return &executorItemSettingModel{
		Name:  name,
		File:  file,
		Exts:  exts,
		Args:  args,
		Match: match,
	}
}

func (m *executorItemSettingModel) convert(helper *modelHelper) (*executorItemSetting, error) {
	if m.Name == "" {
		return nil, helper.Child("name").NewValueEmptyError()
	}

	if m.File != "" && !utils.IsFileExists(m.File) {
		return nil, helper.Child("file").NewValueInvalidError(m.File)
	}

	if err := helper.CheckStringItemEmpty("exts", m.Exts); err != nil {
		return nil, err
	}

	if err := helper.CheckStringItemEmpty("args", m.Args); err != nil {
		return nil, err
	}

	matchObj, err := helper.ConvertEvalExpr("match", m.Match)
	if err != nil {
		return nil, err
	}

	return newExecutorItemSetting(m.Name, m.File, m.Exts, m.Args, m.Match, matchObj), nil
}

// endregion

// region ExecutorSettingModelBuilder

type ExecutorSettingModelBuilder[P any] struct {
	commit func(*executorSettingModel) P
	items  []*executorItemSettingModel
}

func newExecutorSettingModelBuilder[P any](commit func(*executorSettingModel) P) *ExecutorSettingModelBuilder[P] {
	return &ExecutorSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *ExecutorSettingModelBuilder[P]) AddItem(name, path string, exts, args []string, match string) *ExecutorSettingModelBuilder[P] {
	b.items = append(b.items, newExecutorItemSettingModel(name, path, exts, args, match))
	return b
}

func (b *ExecutorSettingModelBuilder[P]) CommitExecutorSetting() P {
	return b.commit(newExecutorSettingModel(b.items))
}

// endregion
