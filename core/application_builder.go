package core

import (
	. "github.com/orz-dsh/dsh/core/builder"
	. "github.com/orz-dsh/dsh/core/internal"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"path/filepath"
	"slices"
)

type ApplicationBuilder struct {
	workspace       *WorkspaceCore
	profileSettings []*ProfileSetting
	err             error
}

func newAppBuilder(workspace *WorkspaceCore) *ApplicationBuilder {
	var profileSettings []*ProfileSetting
	for i := 0; i < len(workspace.ProfileSettings); i++ {
		profileSettings = append(profileSettings, workspace.ProfileSettings[i])
	}
	return &ApplicationBuilder{
		workspace:       workspace,
		profileSettings: profileSettings,
	}
}

func (b *ApplicationBuilder) AddProfileSetting(source string, position int) *ProfileSettingModelBuilder[*ApplicationBuilder] {
	return NewProfileSettingModelBuilder(source, func(model *ProfileSettingModel) *ApplicationBuilder {
		return b.addProfileSettingModel(position, model)
	})
}

func (b *ApplicationBuilder) AddProfileSettingFile(position int, file string) *ApplicationBuilder {
	path, err := filepath.Abs(file)
	if err != nil {
		return b.addProfileSetting(position, nil, err)
	}
	setting, err := LoadProfileSetting(b.workspace.Logger, path)
	if err != nil {
		return b.addProfileSetting(position, nil, err)
	}
	return b.addProfileSetting(position, setting, nil)
}

func (b *ApplicationBuilder) Error() error {
	return b.err
}

func (b *ApplicationBuilder) Build(link string) (*Application, error) {
	b.workspace.Logger.InfoDesc("load app", KV("link", link))

	if b.err != nil {
		return nil, b.err
	}

	setting := NewApplicationSetting(b.workspace, b.profileSettings)
	core, err := NewApplicationCore(b.workspace, setting, link)
	if err != nil {
		return nil, err
	}
	return newApplication(core), nil
}

func (b *ApplicationBuilder) addProfileSettingModel(position int, model *ProfileSettingModel) *ApplicationBuilder {
	if b.err != nil {
		return b
	}
	setting, err := model.GetSetting(b.workspace.Logger)
	return b.addProfileSetting(position, setting, err)
}

func (b *ApplicationBuilder) addProfileSetting(position int, setting *ProfileSetting, err error) *ApplicationBuilder {
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
