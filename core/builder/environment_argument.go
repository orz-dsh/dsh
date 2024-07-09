package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region EnvironmentArgumentSettingModelBuilder

type EnvironmentArgumentSettingModelBuilder[R any] struct {
	commit func(*EnvironmentArgumentSettingModel) R
	items  []*EnvironmentArgumentItemSettingModel
}

func NewEnvironmentArgumentSettingModelBuilder[R any](commit func(*EnvironmentArgumentSettingModel) R) *EnvironmentArgumentSettingModelBuilder[R] {
	return &EnvironmentArgumentSettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *EnvironmentArgumentSettingModelBuilder[R]) SetItems(items []*EnvironmentArgumentItemSettingModel) *EnvironmentArgumentSettingModelBuilder[R] {
	b.items = items
	return b
}

func (b *EnvironmentArgumentSettingModelBuilder[R]) AddItem(name, value string) *EnvironmentArgumentSettingModelBuilder[R] {
	return b.AddItemModel(NewEnvironmentArgumentItemSettingModel(name, value))
}

func (b *EnvironmentArgumentSettingModelBuilder[R]) AddItemMap(m map[string]string) *EnvironmentArgumentSettingModelBuilder[R] {
	for k, v := range m {
		b.AddItem(k, v)
	}
	return b
}

func (b *EnvironmentArgumentSettingModelBuilder[R]) AddItemModel(item *EnvironmentArgumentItemSettingModel) *EnvironmentArgumentSettingModelBuilder[R] {
	b.items = append(b.items, item)
	return b
}

func (b *EnvironmentArgumentSettingModelBuilder[R]) CommitArgumentSetting() R {
	return b.commit(NewEnvironmentArgumentSettingModel(b.items))
}

// endregion
