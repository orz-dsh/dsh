package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region WorkspaceCleanSettingModelBuilder

type WorkspaceCleanSettingModelBuilder[P any] struct {
	commit func(model *WorkspaceCleanSettingModel) P
	output *WorkspaceCleanOutputSettingModel
}

func NewWorkspaceCleanSettingModelBuilder[P any](commit func(model *WorkspaceCleanSettingModel) P) *WorkspaceCleanSettingModelBuilder[P] {
	return &WorkspaceCleanSettingModelBuilder[P]{
		commit: commit,
		output: NewWorkspaceCleanOutputSettingModel(nil, ""),
	}
}

func (b *WorkspaceCleanSettingModelBuilder[P]) SetOutputCount(count int) *WorkspaceCleanSettingModelBuilder[P] {
	b.output.Count = &count
	return b
}

func (b *WorkspaceCleanSettingModelBuilder[P]) SetOutputExpires(expires string) *WorkspaceCleanSettingModelBuilder[P] {
	b.output.Expires = expires
	return b
}

func (b *WorkspaceCleanSettingModelBuilder[P]) CommitCleanSetting() P {
	return b.commit(NewWorkspaceCleanSettingModel(b.output))
}

// endregion
