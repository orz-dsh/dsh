package builder

import . "github.com/orz-dsh/dsh/core/internal/setting"

// region ProfileSettingModelBuilder

type ProfileSettingModelBuilder[P any] struct {
	commit   func(*ProfileSettingModel) P
	argument *ArgumentSettingModel
	addition *AdditionSettingModel
	executor *ExecutorSettingModel
	registry *RegistrySettingModel
	redirect *RedirectSettingModel
}

func NewProfileSettingModelBuilder[P any](commit func(*ProfileSettingModel) P) *ProfileSettingModelBuilder[P] {
	return &ProfileSettingModelBuilder[P]{
		commit: commit,
	}
}

func (b *ProfileSettingModelBuilder[P]) SetArgumentSetting() *ArgumentSettingModelBuilder[*ProfileSettingModelBuilder[P]] {
	return NewArgumentSettingModelBuilder(b.setArgumentSettingModel)
}

func (b *ProfileSettingModelBuilder[P]) SetAdditionSetting() *AdditionSettingModelBuilder[*ProfileSettingModelBuilder[P]] {
	return NewAdditionSettingModelBuilder(b.setAdditionSettingModel)
}

func (b *ProfileSettingModelBuilder[P]) SetExecutorSetting() *ExecutorSettingModelBuilder[*ProfileSettingModelBuilder[P]] {
	return NewExecutorSettingModelBuilder(b.setExecutorSettingModel)
}

func (b *ProfileSettingModelBuilder[P]) SetRegistrySetting() *RegistrySettingModelBuilder[*ProfileSettingModelBuilder[P]] {
	return NewProfileRegistrySettingBuilder(b.setRegistrySettingModel)
}

func (b *ProfileSettingModelBuilder[P]) SetRedirectSetting() *RedirectSettingModelBuilder[*ProfileSettingModelBuilder[P]] {
	return NewRedirectSettingModelBuilder(b.setRedirectSettingModel)
}

func (b *ProfileSettingModelBuilder[P]) CommitProfileSetting() P {
	return b.commit(NewProfileSettingModel(b.argument, b.addition, b.executor, b.registry, b.redirect))
}

func (b *ProfileSettingModelBuilder[P]) setArgumentSettingModel(argument *ArgumentSettingModel) *ProfileSettingModelBuilder[P] {
	b.argument = argument
	return b
}

func (b *ProfileSettingModelBuilder[P]) setAdditionSettingModel(addition *AdditionSettingModel) *ProfileSettingModelBuilder[P] {
	b.addition = addition
	return b
}

func (b *ProfileSettingModelBuilder[P]) setExecutorSettingModel(executor *ExecutorSettingModel) *ProfileSettingModelBuilder[P] {
	b.executor = executor
	return b
}

func (b *ProfileSettingModelBuilder[P]) setRegistrySettingModel(registry *RegistrySettingModel) *ProfileSettingModelBuilder[P] {
	b.registry = registry
	return b
}

func (b *ProfileSettingModelBuilder[P]) setRedirectSettingModel(redirect *RedirectSettingModel) *ProfileSettingModelBuilder[P] {
	b.redirect = redirect
	return b
}

// endregion
