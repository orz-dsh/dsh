package dsh_core

// region workspaceImportSettingModel

type workspaceImportSettingModel struct {
	Registry *workspaceImportRegistrySettingModel
	Redirect *workspaceImportRedirectSettingModel
}

func newWorkspaceImportSettingModel(registry *workspaceImportRegistrySettingModel, redirect *workspaceImportRedirectSettingModel) *workspaceImportSettingModel {
	return &workspaceImportSettingModel{
		Registry: registry,
		Redirect: redirect,
	}
}

func (m *workspaceImportSettingModel) convert(ctx *ModelConvertContext) (registrySettings workspaceImportRegistrySettingSet, redirectSettings workspaceImportRedirectSettingSet, err error) {
	if m.Registry != nil {
		if registrySettings, err = m.Registry.convert(ctx.Child("registry")); err != nil {
			return nil, nil, err
		}
	}
	if m.Redirect != nil {
		if redirectSettings, err = m.Redirect.convert(ctx.Child("redirect")); err != nil {
			return nil, nil, err
		}
	}
	return registrySettings, redirectSettings, nil
}

// endregion
