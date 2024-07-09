package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region RedirectSettingModelBuilder

type RedirectSettingModelBuilder[R any] struct {
	commit func(*RedirectSettingModel) R
	items  []*RedirectItemSettingModel
}

func NewRedirectSettingModelBuilder[R any](commit func(*RedirectSettingModel) R) *RedirectSettingModelBuilder[R] {
	return &RedirectSettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *RedirectSettingModelBuilder[R]) SetItems(items []*RedirectItemSettingModel) *RedirectSettingModelBuilder[R] {
	b.items = items
	return b
}

func (b *RedirectSettingModelBuilder[R]) AddItem(regex, link, match string) *RedirectSettingModelBuilder[R] {
	return b.AddItemModel(NewRedirectItemSettingModel(regex, link, match))
}

func (b *RedirectSettingModelBuilder[R]) AddItemModel(item *RedirectItemSettingModel) *RedirectSettingModelBuilder[R] {
	b.items = append(b.items, item)
	return b
}

func (b *RedirectSettingModelBuilder[R]) CommitRedirectSetting() R {
	return b.commit(NewRedirectSettingModel(b.items))
}

// endregion
