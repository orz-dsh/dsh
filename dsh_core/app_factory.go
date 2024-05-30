package dsh_core

import (
	"slices"
)

type AppFactory struct {
	workspace *Workspace
	manifests []*ProfileManifest
}

func newAppFactory(workspace *Workspace) *AppFactory {
	factory := &AppFactory{
		workspace: workspace,
		manifests: []*ProfileManifest{},
	}
	for i := 0; i < len(workspace.profileManifests); i++ {
		factory.AddManifest(-1, workspace.profileManifests[i])
	}
	return factory
}

func (f *AppFactory) AddManifest(position int, manifest *ProfileManifest) {
	if position < 0 {
		f.manifests = append(f.manifests, manifest)
	} else {
		f.manifests = slices.Insert(f.manifests, position, manifest)
	}
}

func (f *AppFactory) AddProfile(position int, file string) error {
	manifest, err := loadProfileManifest(file)
	if err != nil {
		return err
	}
	f.AddManifest(position, manifest)
	return nil
}

func (f *AppFactory) AddProjectOptionItems(position int, items map[string]string) error {
	manifest, err := MakeProfileManifest(nil, NewProfileManifestProject(NewProfileManifestProjectOption(items), nil, nil))
	if err != nil {
		return err
	}
	f.AddManifest(position, manifest)
	return nil
}

func (f *AppFactory) MakeApp(link string) (*App, error) {
	f.workspace.logger.InfoDesc("load app", kv("link", link))

	profile := newAppProfile(f.workspace, f.manifests)

	manifest, err := profile.getProjectManifestByRawLink(link)
	if err != nil {
		return nil, err
	}

	option, err := profile.makeAppOption(manifest.Name)
	if err != nil {
		return nil, err
	}

	context := newAppContext(f.workspace, profile, manifest, option)

	project, err := context.loadMainProject()
	if err != nil {
		return nil, err
	}

	app, err := newApp(context, project)
	if err != nil {
		return nil, err
	}
	return app, nil
}
