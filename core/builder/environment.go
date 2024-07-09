package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region EnvironmentSettingModelBuilder

type EnvironmentSettingModelBuilder[R any] struct {
	commit    func(*EnvironmentSettingModel) R
	argument  *EnvironmentArgumentSettingModel
	workspace *EnvironmentWorkspaceSettingModel
}

func NewEnvironmentSettingModelBuilder[R any](commit func(*EnvironmentSettingModel) R) *EnvironmentSettingModelBuilder[R] {
	return &EnvironmentSettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *EnvironmentSettingModelBuilder[R]) SetArgumentSetting() *EnvironmentArgumentSettingModelBuilder[*EnvironmentSettingModelBuilder[R]] {
	return NewEnvironmentArgumentSettingModelBuilder(b.setArgumentSettingModel)
}

func (b *EnvironmentSettingModelBuilder[R]) SetWorkspaceSetting() *EnvironmentWorkspaceSettingModelBuilder[*EnvironmentSettingModelBuilder[R]] {
	return NewEnvironmentWorkspaceSettingModelBuilder(b.setWorkspaceSettingModel)
}

func (b *EnvironmentSettingModelBuilder[R]) CommitEnvironmentSetting() R {
	return b.commit(NewEnvironmentSettingModel(b.argument, b.workspace))
}

func (b *EnvironmentSettingModelBuilder[R]) setArgumentSettingModel(argument *EnvironmentArgumentSettingModel) *EnvironmentSettingModelBuilder[R] {
	b.argument = argument
	return b
}

func (b *EnvironmentSettingModelBuilder[R]) setWorkspaceSettingModel(workspace *EnvironmentWorkspaceSettingModel) *EnvironmentSettingModelBuilder[R] {
	b.workspace = workspace
	return b
}

// endregion
