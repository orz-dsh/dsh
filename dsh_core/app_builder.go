package dsh_core

import (
	"path/filepath"
	"slices"
)

type AppMaker struct {
	workspace *Workspace
	manifests []*ProfilePref
}

func newAppMaker(workspace *Workspace) *AppMaker {
	factory := &AppMaker{
		workspace: workspace,
		manifests: []*ProfilePref{},
	}
	for i := 0; i < len(workspace.profileManifests); i++ {
		factory.AddManifest(-1, workspace.profileManifests[i])
	}
	return factory
}

func (f *AppMaker) AddManifest(position int, manifest *ProfilePref) {
	if position < 0 {
		f.manifests = append(f.manifests, manifest)
	} else {
		f.manifests = slices.Insert(f.manifests, position, manifest)
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
	manifest, err := loadProfilePref(absPath)
	if err != nil {
		return err
	}
	f.AddManifest(position, manifest)
	return nil
}

func (f *AppMaker) AddOptionSpecifyItems(position int, items map[string]string) error {
	var prefItems []*ProfileOptionItemPref
	for name, value := range items {
		prefItems = append(prefItems, NewProfileOptionItemPref(name, value, ""))
	}
	pref, err := MakeProfilePref(NewProfileOptionPref(prefItems), nil, nil)
	if err != nil {
		return err
	}
	f.AddManifest(position, pref)
	return nil
}

func (f *AppMaker) Build(link string) (*App, error) {
	f.workspace.logger.InfoDesc("load app", kv("link", link))

	profile := newAppProfile(f.workspace, f.manifests)

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
