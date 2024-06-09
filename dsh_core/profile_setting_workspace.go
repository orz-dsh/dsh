package dsh_core

// region profileWorkspaceSettingModel

type profileWorkspaceSettingModel struct {
	Executor *workspaceExecutorSettingModel
	Import   *workspaceImportSettingModel
}

func newProfileWorkspaceSettingModel(executor *workspaceExecutorSettingModel, import_ *workspaceImportSettingModel) *profileWorkspaceSettingModel {
	return &profileWorkspaceSettingModel{
		Executor: executor,
		Import:   import_,
	}
}

func (m *profileWorkspaceSettingModel) convert(ctx *modelConvertContext) (executorSettings workspaceExecutorSettingSet, importRegistrySettings workspaceImportRegistrySettingSet, importRedirectSettings workspaceImportRedirectSettingSet, err error) {
	if m.Executor != nil {
		if executorSettings, err = m.Executor.convert(ctx.Child("executor")); err != nil {
			return nil, nil, nil, err
		}
	}

	if m.Import != nil {
		if importRegistrySettings, importRedirectSettings, err = m.Import.convert(ctx.Child("import")); err != nil {
			return nil, nil, nil, err
		}
	}

	return executorSettings, importRegistrySettings, importRedirectSettings, nil
}

// endregion
