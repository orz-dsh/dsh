package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region ExecutorSettingModelBuilder

type ExecutorSettingModelBuilder[R any] struct {
	commit func(*ExecutorSettingModel) R
	items  []*ExecutorItemSettingModel
}

func NewExecutorSettingModelBuilder[R any](commit func(*ExecutorSettingModel) R) *ExecutorSettingModelBuilder[R] {
	return &ExecutorSettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *ExecutorSettingModelBuilder[R]) SetItems(items []*ExecutorItemSettingModel) *ExecutorSettingModelBuilder[R] {
	b.items = items
	return b
}

func (b *ExecutorSettingModelBuilder[R]) AddItem(name, path string, exts, args []string, match string) *ExecutorSettingModelBuilder[R] {
	return b.AddItemModel(NewExecutorItemSettingModel(name, path, exts, args, match))
}

func (b *ExecutorSettingModelBuilder[R]) AddItemModel(item *ExecutorItemSettingModel) *ExecutorSettingModelBuilder[R] {
	b.items = append(b.items, item)
	return b
}

func (b *ExecutorSettingModelBuilder[R]) CommitExecutorSetting() R {
	return b.commit(NewExecutorSettingModel(b.items))
}

// endregion
