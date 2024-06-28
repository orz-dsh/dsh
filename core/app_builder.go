package core

import (
	"path/filepath"
	"slices"
)

type AppBuilder struct {
	workspace       *Workspace
	profileSettings []*profileSetting
	err             error
}

func newAppBuilder(workspace *Workspace) *AppBuilder {
	var profileSettings []*profileSetting
	for i := 0; i < len(workspace.profileSettings); i++ {
		profileSettings = append(profileSettings, workspace.profileSettings[i])
	}
	return &AppBuilder{
		workspace:       workspace,
		profileSettings: profileSettings,
	}
}

func (b *AppBuilder) AddProfileSetting(position int) *ProfileSettingModelBuilder[*AppBuilder] {
	return newProfileSettingModelBuilder(func(model *profileSettingModel) *AppBuilder {
		return b.addProfileSettingModel(position, model)
	})
}

func (b *AppBuilder) AddProfileSettingFile(position int, file string) *AppBuilder {
	path, err := filepath.Abs(file)
	if err != nil {
		return b.addProfileSetting(position, nil, err)
	}
	setting, err := loadProfileSetting(path)
	if err != nil {
		return b.addProfileSetting(position, nil, err)
	}
	return b.addProfileSetting(position, setting, nil)
}

func (b *AppBuilder) Error() error {
	return b.err
}

func (b *AppBuilder) Build(link string) (*App, error) {
	b.workspace.logger.InfoDesc("load app", kv("link", link))

	if b.err != nil {
		return nil, b.err
	}

	profile := newAppSetting(b.workspace, b.profileSettings)

	mainProjectSetting, err := profile.getProjectEntityByRawLink(link)
	if err != nil {
		return nil, err
	}

	evaluator := b.workspace.evaluator.SetData("main_project", map[string]any{
		"name": mainProjectSetting.Name,
		"path": mainProjectSetting.Dir,
	})

	option, err := profile.getAppOption(mainProjectSetting, evaluator)
	if err != nil {
		return nil, err
	}

	extraProjectSettings, err := profile.getExtraProjectSettings(evaluator)
	if err != nil {
		return nil, err
	}

	context := newAppContext(b.workspace, evaluator, profile, option)

	return newApp(context, mainProjectSetting, extraProjectSettings), nil
}

func (b *AppBuilder) addProfileSettingModel(position int, model *profileSettingModel) *AppBuilder {
	if b.err != nil {
		return b
	}
	setting, err := loadProfileSettingModel(model)
	return b.addProfileSetting(position, setting, err)
}

func (b *AppBuilder) addProfileSetting(position int, setting *profileSetting, err error) *AppBuilder {
	if b.err != nil {
		return b
	}
	if err != nil {
		b.err = err
		return b
	}
	if position < 0 {
		b.profileSettings = append(b.profileSettings, setting)
	} else {
		b.profileSettings = slices.Insert(b.profileSettings, position, setting)
	}
	return b
}
