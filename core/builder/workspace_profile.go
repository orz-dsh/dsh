package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region WorkspaceProfileSettingModelBuilder

type WorkspaceProfileSettingModelBuilder[R any] struct {
	commit func(*WorkspaceProfileSettingModel) R
	items  []*WorkspaceProfileItemSettingModel
}

func NewWorkspaceProfileSettingModelBuilder[R any](commit func(*WorkspaceProfileSettingModel) R) *WorkspaceProfileSettingModelBuilder[R] {
	return &WorkspaceProfileSettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *WorkspaceProfileSettingModelBuilder[R]) SetItems(items []*WorkspaceProfileItemSettingModel) *WorkspaceProfileSettingModelBuilder[R] {
	b.items = items
	return b
}

func (b *WorkspaceProfileSettingModelBuilder[R]) AddItem(file string, optional bool, match string) *WorkspaceProfileSettingModelBuilder[R] {
	return b.AddItemModel(NewWorkspaceProfileItemSettingModel(file, optional, match))
}

func (b *WorkspaceProfileSettingModelBuilder[R]) AddItemModel(item *WorkspaceProfileItemSettingModel) *WorkspaceProfileSettingModelBuilder[R] {
	b.items = append(b.items, item)
	return b
}

func (b *WorkspaceProfileSettingModelBuilder[R]) CommitProfileSetting() R {
	return b.commit(NewWorkspaceProfileSettingModel(b.items))
}

// endregion
