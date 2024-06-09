package dsh_core

import (
	"dsh/dsh_utils"
)

// region workspaceSetting

type workspaceSetting struct {
	CleanSetting           *workspaceCleanSetting
	ProfileSettings        workspaceProfileSettingSet
	ExecutorSettings       workspaceExecutorSettingSet
	ImportRegistrySettings workspaceImportRegistrySettingSet
	ImportRedirectSettings workspaceImportRedirectSettingSet
}

func newWorkspaceSetting(cleanSetting *workspaceCleanSetting, profileSettings workspaceProfileSettingSet, executorSettings workspaceExecutorSettingSet, importRegistrySettings workspaceImportRegistrySettingSet, importRedirectSettings workspaceImportRedirectSettingSet) *workspaceSetting {
	if cleanSetting == nil {
		cleanSetting = workspaceCleanSettingDefault
	}
	if profileSettings == nil {
		profileSettings = workspaceProfileSettingSet{}
	}
	if executorSettings == nil {
		executorSettings = workspaceExecutorSettingSet{}
	}
	if importRegistrySettings == nil {
		importRegistrySettings = workspaceImportRegistrySettingSet{}
	}
	if importRedirectSettings == nil {
		importRedirectSettings = workspaceImportRedirectSettingSet{}
	}
	return &workspaceSetting{
		CleanSetting:           cleanSetting,
		ProfileSettings:        profileSettings,
		ExecutorSettings:       executorSettings,
		ImportRegistrySettings: importRegistrySettings,
		ImportRedirectSettings: importRedirectSettings,
	}
}

func loadWorkspaceSetting(path string) (setting *workspaceSetting, err error) {
	model := &workspaceSettingModel{}
	metadata, err := dsh_utils.DeserializeFromDir(path, []string{"workspace"}, model, false)
	if err != nil {
		return nil, errW(err, "load workspace setting error",
			reason("deserialize error"),
			kv("path", path),
		)
	}
	file := ""
	if metadata != nil {
		file = metadata.Path
	}
	if setting, err = model.convert(newModelConvertContext("workspace setting", file)); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region workspaceSettingModel

type workspaceSettingModel struct {
	Clean    *workspaceCleanSettingModel
	Profile  *workspaceProfileSettingModel
	Executor *workspaceExecutorSettingModel
	Import   *workspaceImportSettingModel
}

func (s *workspaceSettingModel) convert(ctx *modelConvertContext) (setting *workspaceSetting, err error) {
	var cleanSetting *workspaceCleanSetting
	if s.Clean != nil {
		if cleanSetting, err = s.Clean.convert(ctx.Child("clean")); err != nil {
			return nil, err
		}
	}

	var profileSettings workspaceProfileSettingSet
	if s.Profile != nil {
		if profileSettings, err = s.Profile.convert(ctx.Child("profile")); err != nil {
			return nil, err
		}
	}

	var executorSettings workspaceExecutorSettingSet
	if s.Executor != nil {
		if executorSettings, err = s.Executor.convert(ctx.Child("executor")); err != nil {
			return nil, err
		}
	}

	var importRegistrySettings workspaceImportRegistrySettingSet
	var importRedirectSettings workspaceImportRedirectSettingSet
	if s.Import != nil {
		if importRegistrySettings, importRedirectSettings, err = s.Import.convert(ctx.Child("import")); err != nil {
			return nil, err
		}
	}

	return newWorkspaceSetting(cleanSetting, profileSettings, executorSettings, importRegistrySettings, importRedirectSettings), nil
}

// endregion
