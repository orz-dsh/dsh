package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region WorkspaceSettingModelBuilder

type WorkspaceSettingModelBuilder[P any] struct {
	commit   func(*WorkspaceSettingModel) P
	clean    *WorkspaceCleanSettingModel
	profile  *WorkspaceProfileSettingModel
	executor *ExecutorSettingModel
	registry *RegistrySettingModel
	redirect *RedirectSettingModel
}

func NewWorkspaceSettingModelBuilder[P any](commit func(*WorkspaceSettingModel) P) *WorkspaceSettingModelBuilder[P] {
	return &WorkspaceSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *WorkspaceSettingModelBuilder[P]) SetCleanSetting() *WorkspaceCleanSettingModelBuilder[*WorkspaceSettingModelBuilder[P]] {
	return NewWorkspaceCleanSettingModelBuilder(b.setCleanSettingModel)
}

func (b *WorkspaceSettingModelBuilder[P]) SetProfileSetting() *WorkspaceProfileSettingModelBuilder[*WorkspaceSettingModelBuilder[P]] {
	return NewWorkspaceProfileSettingModelBuilder(b.setProfileSettingModel)
}

func (b *WorkspaceSettingModelBuilder[P]) SetExecutorSetting() *ExecutorSettingModelBuilder[*WorkspaceSettingModelBuilder[P]] {
	return NewExecutorSettingModelBuilder(b.setExecutorSettingModel)
}

func (b *WorkspaceSettingModelBuilder[P]) SetRegistrySetting() *RegistrySettingModelBuilder[*WorkspaceSettingModelBuilder[P]] {
	return NewProfileRegistrySettingBuilder(b.setRegistrySettingModel)
}

func (b *WorkspaceSettingModelBuilder[P]) SetRedirectSetting() *RedirectSettingModelBuilder[*WorkspaceSettingModelBuilder[P]] {
	return NewRedirectSettingModelBuilder(b.setRedirectSettingModel)
}

func (b *WorkspaceSettingModelBuilder[P]) CommitWorkspaceSetting() P {
	return b.commit(NewWorkspaceSettingModel(b.clean, b.profile, b.executor, b.registry, b.redirect))
}

func (b *WorkspaceSettingModelBuilder[P]) setCleanSettingModel(clean *WorkspaceCleanSettingModel) *WorkspaceSettingModelBuilder[P] {
	b.clean = clean
	return b
}

func (b *WorkspaceSettingModelBuilder[P]) setProfileSettingModel(profile *WorkspaceProfileSettingModel) *WorkspaceSettingModelBuilder[P] {
	b.profile = profile
	return b
}

func (b *WorkspaceSettingModelBuilder[P]) setExecutorSettingModel(executor *ExecutorSettingModel) *WorkspaceSettingModelBuilder[P] {
	b.executor = executor
	return b
}

func (b *WorkspaceSettingModelBuilder[P]) setRegistrySettingModel(registry *RegistrySettingModel) *WorkspaceSettingModelBuilder[P] {
	b.registry = registry
	return b
}

func (b *WorkspaceSettingModelBuilder[P]) setRedirectSettingModel(redirect *RedirectSettingModel) *WorkspaceSettingModelBuilder[P] {
	b.redirect = redirect
	return b
}

// endregion
