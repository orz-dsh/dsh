package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region WorkspaceSettingModelBuilder

type WorkspaceSettingModelBuilder[R any] struct {
	commit   func(*WorkspaceSettingModel) R
	clean    *WorkspaceCleanSettingModel
	profile  *WorkspaceProfileSettingModel
	executor *ExecutorSettingModel
	registry *RegistrySettingModel
	redirect *RedirectSettingModel
}

func NewWorkspaceSettingModelBuilder[R any](commit func(*WorkspaceSettingModel) R) *WorkspaceSettingModelBuilder[R] {
	return &WorkspaceSettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *WorkspaceSettingModelBuilder[R]) SetCleanSetting() *WorkspaceCleanSettingModelBuilder[*WorkspaceSettingModelBuilder[R]] {
	return NewWorkspaceCleanSettingModelBuilder(b.setCleanSettingModel)
}

func (b *WorkspaceSettingModelBuilder[R]) SetProfileSetting() *WorkspaceProfileSettingModelBuilder[*WorkspaceSettingModelBuilder[R]] {
	return NewWorkspaceProfileSettingModelBuilder(b.setProfileSettingModel)
}

func (b *WorkspaceSettingModelBuilder[R]) SetExecutorSetting() *ExecutorSettingModelBuilder[*WorkspaceSettingModelBuilder[R]] {
	return NewExecutorSettingModelBuilder(b.setExecutorSettingModel)
}

func (b *WorkspaceSettingModelBuilder[R]) SetRegistrySetting() *RegistrySettingModelBuilder[*WorkspaceSettingModelBuilder[R]] {
	return NewProfileRegistrySettingBuilder(b.setRegistrySettingModel)
}

func (b *WorkspaceSettingModelBuilder[R]) SetRedirectSetting() *RedirectSettingModelBuilder[*WorkspaceSettingModelBuilder[R]] {
	return NewRedirectSettingModelBuilder(b.setRedirectSettingModel)
}

func (b *WorkspaceSettingModelBuilder[R]) CommitWorkspaceSetting() R {
	return b.commit(NewWorkspaceSettingModel(b.clean, b.profile, b.executor, b.registry, b.redirect))
}

func (b *WorkspaceSettingModelBuilder[R]) setCleanSettingModel(clean *WorkspaceCleanSettingModel) *WorkspaceSettingModelBuilder[R] {
	b.clean = clean
	return b
}

func (b *WorkspaceSettingModelBuilder[R]) setProfileSettingModel(profile *WorkspaceProfileSettingModel) *WorkspaceSettingModelBuilder[R] {
	b.profile = profile
	return b
}

func (b *WorkspaceSettingModelBuilder[R]) setExecutorSettingModel(executor *ExecutorSettingModel) *WorkspaceSettingModelBuilder[R] {
	b.executor = executor
	return b
}

func (b *WorkspaceSettingModelBuilder[R]) setRegistrySettingModel(registry *RegistrySettingModel) *WorkspaceSettingModelBuilder[R] {
	b.registry = registry
	return b
}

func (b *WorkspaceSettingModelBuilder[R]) setRedirectSettingModel(redirect *RedirectSettingModel) *WorkspaceSettingModelBuilder[R] {
	b.redirect = redirect
	return b
}

// endregion
