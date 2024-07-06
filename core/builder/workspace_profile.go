package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region WorkspaceProfileSettingModelBuilder

type WorkspaceProfileSettingModelBuilder[P any] struct {
	commit func(*WorkspaceProfileSettingModel) P
	items  []*WorkspaceProfileItemSettingModel
}

func NewWorkspaceProfileSettingModelBuilder[P any](commit func(*WorkspaceProfileSettingModel) P) *WorkspaceProfileSettingModelBuilder[P] {
	return &WorkspaceProfileSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *WorkspaceProfileSettingModelBuilder[P]) AddItem(file string, optional bool, match string) *WorkspaceProfileSettingModelBuilder[P] {
	b.items = append(b.items, NewWorkspaceProfileItemSettingModel(file, optional, match))
	return b
}

func (b *WorkspaceProfileSettingModelBuilder[P]) CommitProfileSetting() P {
	return b.commit(NewWorkspaceProfileSettingModel(b.items))
}

// endregion
