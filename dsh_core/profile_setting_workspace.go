package dsh_core

// region ProfileWorkspaceSettingModel

type ProfileWorkspaceSettingModel struct {
	Shell  *workspaceShellSettingModel
	Import *workspaceImportSettingModel
}

func (m *ProfileWorkspaceSettingModel) convert(ctx *ModelConvertContext) (shellSettings workspaceShellSettingSet, importRegistrySettings workspaceImportRegistrySettingSet, importRedirectSettings workspaceImportRedirectSettingSet, err error) {
	if m.Shell != nil {
		if shellSettings, err = m.Shell.convert(ctx.Child("shell")); err != nil {
			return nil, nil, nil, err
		}
	}

	if m.Import != nil {
		if importRegistrySettings, importRedirectSettings, err = m.Import.convert(ctx.Child("import")); err != nil {
			return nil, nil, nil, err
		}
	}

	return shellSettings, importRegistrySettings, importRedirectSettings, nil
}

// endregion
