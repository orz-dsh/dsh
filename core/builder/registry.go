package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region RegistrySettingModelBuilder

type RegistrySettingModelBuilder[R any] struct {
	commit func(*RegistrySettingModel) R
	items  []*RegistryItemSettingModel
}

func NewProfileRegistrySettingBuilder[R any](commit func(*RegistrySettingModel) R) *RegistrySettingModelBuilder[R] {
	return &RegistrySettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *RegistrySettingModelBuilder[R]) SetItems(items []*RegistryItemSettingModel) *RegistrySettingModelBuilder[R] {
	b.items = items
	return b
}

func (b *RegistrySettingModelBuilder[R]) AddItem(name, link, match string) *RegistrySettingModelBuilder[R] {
	return b.AddItemModel(NewRegistryItemSettingModel(name, link, match))
}

func (b *RegistrySettingModelBuilder[R]) AddItemModel(item *RegistryItemSettingModel) *RegistrySettingModelBuilder[R] {
	b.items = append(b.items, item)
	return b
}

func (b *RegistrySettingModelBuilder[R]) CommitRegistrySetting() R {
	return b.commit(NewRegistrySettingModel(b.items))
}

// endregion
