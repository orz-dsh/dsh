package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region ProfileSettingModelBuilder

type ProfileSettingModelBuilder[R any] struct {
	commit   func(*ProfileSettingModel) R
	argument *ProfileArgumentSettingModel
	addition *ProfileAdditionSettingModel
	executor *ExecutorSettingModel
	registry *RegistrySettingModel
	redirect *RedirectSettingModel
}

func NewProfileSettingModelBuilder[R any](commit func(*ProfileSettingModel) R) *ProfileSettingModelBuilder[R] {
	return &ProfileSettingModelBuilder[R]{
		commit: commit,
	}
}

func (b *ProfileSettingModelBuilder[R]) SetArgumentSetting() *ProfileArgumentSettingModelBuilder[*ProfileSettingModelBuilder[R]] {
	return NewProfileArgumentSettingModelBuilder(b.setArgumentSettingModel)
}

func (b *ProfileSettingModelBuilder[R]) SetAdditionSetting() *ProfileAdditionSettingModelBuilder[*ProfileSettingModelBuilder[R]] {
	return NewProfileAdditionSettingModelBuilder(b.setAdditionSettingModel)
}

func (b *ProfileSettingModelBuilder[R]) SetExecutorSetting() *ExecutorSettingModelBuilder[*ProfileSettingModelBuilder[R]] {
	return NewExecutorSettingModelBuilder(b.setExecutorSettingModel)
}

func (b *ProfileSettingModelBuilder[R]) SetRegistrySetting() *RegistrySettingModelBuilder[*ProfileSettingModelBuilder[R]] {
	return NewProfileRegistrySettingBuilder(b.setRegistrySettingModel)
}

func (b *ProfileSettingModelBuilder[R]) SetRedirectSetting() *RedirectSettingModelBuilder[*ProfileSettingModelBuilder[R]] {
	return NewRedirectSettingModelBuilder(b.setRedirectSettingModel)
}

func (b *ProfileSettingModelBuilder[R]) CommitProfileSetting() R {
	return b.commit(NewProfileSettingModel(b.argument, b.addition, b.executor, b.registry, b.redirect))
}

func (b *ProfileSettingModelBuilder[R]) setArgumentSettingModel(argument *ProfileArgumentSettingModel) *ProfileSettingModelBuilder[R] {
	b.argument = argument
	return b
}

func (b *ProfileSettingModelBuilder[R]) setAdditionSettingModel(addition *ProfileAdditionSettingModel) *ProfileSettingModelBuilder[R] {
	b.addition = addition
	return b
}

func (b *ProfileSettingModelBuilder[R]) setExecutorSettingModel(executor *ExecutorSettingModel) *ProfileSettingModelBuilder[R] {
	b.executor = executor
	return b
}

func (b *ProfileSettingModelBuilder[R]) setRegistrySettingModel(registry *RegistrySettingModel) *ProfileSettingModelBuilder[R] {
	b.registry = registry
	return b
}

func (b *ProfileSettingModelBuilder[R]) setRedirectSettingModel(redirect *RedirectSettingModel) *ProfileSettingModelBuilder[R] {
	b.redirect = redirect
	return b
}

// endregion
