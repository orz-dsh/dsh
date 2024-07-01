package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region RegistrySettingModelBuilder

type RegistrySettingModelBuilder[P any] struct {
	commit func(*RegistrySettingModel) P
	items  []*RegistryItemSettingModel
}

func NewProfileRegistrySettingBuilder[P any](commit func(*RegistrySettingModel) P) *RegistrySettingModelBuilder[P] {
	return &RegistrySettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *RegistrySettingModelBuilder[P]) AddItem(name, link, match string) *RegistrySettingModelBuilder[P] {
	b.items = append(b.items, NewRegistryItemSettingModel(name, link, match))
	return b
}

func (b *RegistrySettingModelBuilder[P]) CommitRegistrySetting() P {
	return b.commit(NewRegistrySettingModel(b.items))
}

// endregion
