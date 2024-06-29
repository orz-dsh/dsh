package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region RedirectSettingModelBuilder

type RedirectSettingModelBuilder[P any] struct {
	commit func(*RedirectSettingModel) P
	items  []*RedirectItemSettingModel
}

func NewRedirectSettingModelBuilder[P any](commit func(*RedirectSettingModel) P) *RedirectSettingModelBuilder[P] {
	return &RedirectSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *RedirectSettingModelBuilder[P]) AddItem(regex, link, match string) *RedirectSettingModelBuilder[P] {
	b.items = append(b.items, NewRedirectItemSettingModel(regex, link, match))
	return b
}

func (b *RedirectSettingModelBuilder[P]) CommitRedirectSetting() P {
	return b.commit(NewRedirectSettingModel(b.items))
}

// endregion
