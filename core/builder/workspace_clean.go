package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region WorkspaceCleanSettingModelBuilder

type WorkspaceCleanSettingModelBuilder[R any] struct {
	commit func(model *WorkspaceCleanSettingModel) R
	output *WorkspaceCleanOutputSettingModel
}

func NewWorkspaceCleanSettingModelBuilder[R any](commit func(model *WorkspaceCleanSettingModel) R) *WorkspaceCleanSettingModelBuilder[R] {
	return &WorkspaceCleanSettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *WorkspaceCleanSettingModelBuilder[R]) SetOutputCount(count int) *WorkspaceCleanSettingModelBuilder[R] {
	if b.output == nil {
		b.output = NewWorkspaceCleanOutputSettingModel(&count, "")
	} else {
		b.output.Count = &count
	}
	return b
}

func (b *WorkspaceCleanSettingModelBuilder[R]) SetOutputExpires(expires string) *WorkspaceCleanSettingModelBuilder[R] {
	if b.output == nil {
		b.output = NewWorkspaceCleanOutputSettingModel(nil, expires)
	} else {
		b.output.Expires = expires
	}
	return b
}

func (b *WorkspaceCleanSettingModelBuilder[R]) SetOutputModel(output *WorkspaceCleanOutputSettingModel) *WorkspaceCleanSettingModelBuilder[R] {
	b.output = output
	return b
}

func (b *WorkspaceCleanSettingModelBuilder[R]) CommitCleanSetting() R {
	return b.commit(NewWorkspaceCleanSettingModel(b.output))
}

// endregion
