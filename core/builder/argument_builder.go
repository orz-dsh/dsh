package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region ArgumentSettingModelBuilder

type ArgumentSettingModelBuilder[P any] struct {
	commit func(*ArgumentSettingModel) P
	items  []*ArgumentItemSettingModel
}

func NewArgumentSettingModelBuilder[P any](commit func(*ArgumentSettingModel) P) *ArgumentSettingModelBuilder[P] {
	return &ArgumentSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *ArgumentSettingModelBuilder[P]) AddItem(name, value, match string) *ArgumentSettingModelBuilder[P] {
	b.items = append(b.items, NewArgumentItemSettingModel(name, value, match))
	return b
}

func (b *ArgumentSettingModelBuilder[P]) AddItemMap(items map[string]string) *ArgumentSettingModelBuilder[P] {
	for name, value := range items {
		b.AddItem(name, value, "")
	}
	return b
}

func (b *ArgumentSettingModelBuilder[P]) CommitArgumentSetting() P {
	return b.commit(NewArgumentSettingModel(b.items))
}

// endregion
