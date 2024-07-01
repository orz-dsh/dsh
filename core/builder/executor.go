package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region ExecutorSettingModelBuilder

type ExecutorSettingModelBuilder[P any] struct {
	commit func(*ExecutorSettingModel) P
	items  []*ExecutorItemSettingModel
}

func NewExecutorSettingModelBuilder[P any](commit func(*ExecutorSettingModel) P) *ExecutorSettingModelBuilder[P] {
	return &ExecutorSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *ExecutorSettingModelBuilder[P]) AddItem(name, path string, exts, args []string, match string) *ExecutorSettingModelBuilder[P] {
	b.items = append(b.items, NewExecutorItemSettingModel(name, path, exts, args, match))
	return b
}

func (b *ExecutorSettingModelBuilder[P]) CommitExecutorSetting() P {
	return b.commit(NewExecutorSettingModel(b.items))
}

// endregion
