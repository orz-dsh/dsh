package dsh_core

import (
	"path/filepath"
	"slices"
)

type AppMaker struct {
	workspace       *Workspace
	profileSettings profileSettingSet
}

func newAppMaker(workspace *Workspace) *AppMaker {
	factory := &AppMaker{
		workspace:       workspace,
		profileSettings: profileSettingSet{},
	}
	for i := 0; i < len(workspace.profileSettings); i++ {
		factory.addProfileSetting(-1, workspace.profileSettings[i])
	}
	return factory
}

func (f *AppMaker) addProfileSetting(position int, setting *profileSetting) {
	if position < 0 {
		f.profileSettings = append(f.profileSettings, setting)
	} else {
		f.profileSettings = slices.Insert(f.profileSettings, position, setting)
	}
}

func (f *AppMaker) AddProfile(position int, file string) error {
	absPath, err := filepath.Abs(file)
	if err != nil {
		return errW(err, "add profile error",
			reason("get abs-path error"),
			kv("file", file),
		)
	}
	manifest, err := loadProfileSetting(absPath)
	if err != nil {
		return err
	}
	f.addProfileSetting(position, manifest)
	return nil
}

func (f *AppMaker) AddProfileSettingBuilder(position int, builder *ProfileSettingBuilder) error {
	setting, err := loadProfileSettingBuilder(builder)
	if err != nil {
		return err
	}
	f.addProfileSetting(position, setting)
	return nil
}

func (f *AppMaker) Build(link string) (*App, error) {
	f.workspace.logger.InfoDesc("load app", kv("link", link))

	profile := newAppProfile(f.workspace, f.profileSettings)

	entity, err := profile.getProjectEntityByRawLink(link)
	if err != nil {
		return nil, err
	}

	evaluator := f.workspace.evaluator.SetData("main_project", map[string]any{
		"name": entity.Name,
		"path": entity.Path,
	})

	option, err := profile.getAppOption(entity, evaluator)
	if err != nil {
		return nil, err
	}

	extraProjectEntities, err := profile.getExtraProjectEntities(evaluator)
	if err != nil {
		return nil, err
	}

	context := newAppContext(f.workspace, evaluator, profile, option)

	app, err := makeApp(context, entity, extraProjectEntities)
	if err != nil {
		return nil, err
	}
	return app, nil
}
