package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region EnvironmentWorkspaceSettingModelBuilder

type EnvironmentWorkspaceSettingModelBuilder[R any] struct {
	commit   func(*EnvironmentWorkspaceSettingModel) R
	dir      string
	clean    *WorkspaceCleanSettingModel
	profile  *WorkspaceProfileSettingModel
	executor *ExecutorSettingModel
	registry *RegistrySettingModel
	redirect *RedirectSettingModel
}

func NewEnvironmentWorkspaceSettingModelBuilder[R any](commit func(*EnvironmentWorkspaceSettingModel) R) *EnvironmentWorkspaceSettingModelBuilder[R] {
	return &EnvironmentWorkspaceSettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) SetDir(dir string) *EnvironmentWorkspaceSettingModelBuilder[R] {
	b.dir = dir
	return b
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) SetCleanSetting() *WorkspaceCleanSettingModelBuilder[*EnvironmentWorkspaceSettingModelBuilder[R]] {
	return NewWorkspaceCleanSettingModelBuilder(b.setCleanSettingModel)
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) SetProfileSetting() *WorkspaceProfileSettingModelBuilder[*EnvironmentWorkspaceSettingModelBuilder[R]] {
	return NewWorkspaceProfileSettingModelBuilder(b.setProfileSettingModel)
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) SetExecutorSetting() *ExecutorSettingModelBuilder[*EnvironmentWorkspaceSettingModelBuilder[R]] {
	return NewExecutorSettingModelBuilder(b.setExecutorSettingModel)
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) SetRegistrySetting() *RegistrySettingModelBuilder[*EnvironmentWorkspaceSettingModelBuilder[R]] {
	return NewProfileRegistrySettingBuilder(b.setRegistrySettingModel)
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) SetRedirectSetting() *RedirectSettingModelBuilder[*EnvironmentWorkspaceSettingModelBuilder[R]] {
	return NewRedirectSettingModelBuilder(b.setRedirectSettingModel)
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) CommitWorkspaceSetting() R {
	return b.commit(NewEnvironmentWorkspaceSettingModel(b.dir, b.clean, b.profile, b.executor, b.registry, b.redirect))
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) setCleanSettingModel(clean *WorkspaceCleanSettingModel) *EnvironmentWorkspaceSettingModelBuilder[R] {
	b.clean = clean
	return b
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) setProfileSettingModel(profile *WorkspaceProfileSettingModel) *EnvironmentWorkspaceSettingModelBuilder[R] {
	b.profile = profile
	return b
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) setExecutorSettingModel(executor *ExecutorSettingModel) *EnvironmentWorkspaceSettingModelBuilder[R] {
	b.executor = executor
	return b
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) setRegistrySettingModel(registry *RegistrySettingModel) *EnvironmentWorkspaceSettingModelBuilder[R] {
	b.registry = registry
	return b
}

func (b *EnvironmentWorkspaceSettingModelBuilder[R]) setRedirectSettingModel(redirect *RedirectSettingModel) *EnvironmentWorkspaceSettingModelBuilder[R] {
	b.redirect = redirect
	return b
}

// endregion
