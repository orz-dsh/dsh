package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region ProfileArgumentSettingModelBuilder

type ProfileArgumentSettingModelBuilder[R any] struct {
	commit func(*ProfileArgumentSettingModel) R
	items  []*ProfileArgumentItemSettingModel
}

func NewProfileArgumentSettingModelBuilder[R any](commit func(*ProfileArgumentSettingModel) R) *ProfileArgumentSettingModelBuilder[R] {
	return &ProfileArgumentSettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *ProfileArgumentSettingModelBuilder[R]) AddItem(name, value, match string) *ProfileArgumentSettingModelBuilder[R] {
	b.items = append(b.items, NewProfileArgumentItemSettingModel(name, value, match))
	return b
}

func (b *ProfileArgumentSettingModelBuilder[R]) AddItemMap(items map[string]string) *ProfileArgumentSettingModelBuilder[R] {
	for name, value := range items {
		b.AddItem(name, value, "")
	}
	return b
}

func (b *ProfileArgumentSettingModelBuilder[R]) CommitArgumentSetting() R {
	return b.commit(NewProfileArgumentSettingModel(b.items))
}

// endregion
