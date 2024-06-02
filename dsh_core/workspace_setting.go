package dsh_core

import (
	"dsh/dsh_utils"
)

// region workspaceSetting

type workspaceSetting struct {
	Profile          workspaceProfileSettingSet
	Clean            *workspaceCleanSetting
	Shell            workspaceShellSettingSet
	ImportRegistries workspaceImportRegistrySettingSet
	ImportRedirects  workspaceImportRedirectSettingSet
}

func newWorkspaceSetting(profileSettings workspaceProfileSettingSet, cleanSetting *workspaceCleanSetting, shellSettings workspaceShellSettingSet, importRegistrySettings workspaceImportRegistrySettingSet, importRedirectSettings workspaceImportRedirectSettingSet) *workspaceSetting {
	return &workspaceSetting{
		Profile:          profileSettings,
		Clean:            cleanSetting,
		Shell:            shellSettings,
		ImportRegistries: importRegistrySettings,
		ImportRedirects:  importRedirectSettings,
	}
}

func loadWorkspaceSetting(path string) (setting *workspaceSetting, err error) {
	model := &workspaceSettingModel{
		Profile: &workspaceProfileSettingModel{},
		Clean: &workspaceCleanSettingModel{
			Output: &workspaceCleanOutputSettingModel{},
		},
		Shell: &workspaceShellSettingModel{},
		Import: &workspaceImportSettingModel{
			Registry: &workspaceImportRegistrySettingModel{},
			Redirect: &workspaceImportRedirectSettingModel{},
		},
	}
	metadata, err := dsh_utils.DeserializeFromDir(path, []string{"workspace"}, model, false)
	if err != nil {
		return nil, errW(err, "load workspace setting error",
			reason("deserialize error"),
			kv("path", path),
		)
	}
	if metadata != nil {
		model.path = metadata.Path
	}
	if setting, err = model.convert(); err != nil {
		return nil, err
	}
	return setting, nil
}

// endregion

// region workspaceSettingModel

type workspaceSettingModel struct {
	Profile *workspaceProfileSettingModel
	Clean   *workspaceCleanSettingModel
	Shell   *workspaceShellSettingModel
	Import  *workspaceImportSettingModel
	path    string
}

func (s *workspaceSettingModel) DescExtraKeyValues() KVS {
	return KVS{
		kv("path", s.path),
	}
}

func (s *workspaceSettingModel) convert() (setting *workspaceSetting, err error) {
	var profileSettings workspaceProfileSettingSet
	if profileSettings, err = s.Profile.convert(s); err != nil {
		return nil, err
	}

	var cleanSetting *workspaceCleanSetting
	if cleanSetting, err = s.Clean.convert(s); err != nil {
		return nil, err
	}

	var shellSettings workspaceShellSettingSet
	if shellSettings, err = s.Shell.convert(s); err != nil {
		return nil, err
	}

	var importRegistrySettings workspaceImportRegistrySettingSet
	var importRedirectSettings workspaceImportRedirectSettingSet
	if importRegistrySettings, importRedirectSettings, err = s.Import.convert(s); err != nil {
		return nil, err
	}

	return newWorkspaceSetting(profileSettings, cleanSetting, shellSettings, importRegistrySettings, importRedirectSettings), nil
}

// endregion
