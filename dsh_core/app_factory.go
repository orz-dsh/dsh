package dsh_core

import (
	"slices"
)

type AppFactory struct {
	workspace *Workspace
	manifests []*AppProfileManifest
}

func newAppFactory(workspace *Workspace) *AppFactory {
	factory := &AppFactory{
		workspace: workspace,
		manifests: []*AppProfileManifest{},
	}
	for i := 0; i < len(workspace.profileManifests); i++ {
		factory.AddManifest(-1, workspace.profileManifests[i])
	}
	return factory
}

func (f *AppFactory) AddManifest(position int, manifest *AppProfileManifest) {
	if position < 0 {
		f.manifests = append(f.manifests, manifest)
	} else {
		f.manifests = slices.Insert(f.manifests, position, manifest)
	}
}

func (f *AppFactory) AddProfile(position int, file string) error {
	manifest, err := loadAppProfileManifest(file)
	if err != nil {
		return err
	}
	f.AddManifest(position, manifest)
	return nil
}

func (f *AppFactory) AddProjectOptionItems(position int, items map[string]string) error {
	manifest, err := MakeAppProfileManifest(nil, NewAppProfileManifestProject(NewAppProfileManifestProjectOption(items), nil, nil))
	if err != nil {
		return err
	}
	f.AddManifest(position, manifest)
	return nil
}

func (f *AppFactory) MakeApp(link string) (*App, error) {
	f.workspace.logger.InfoDesc("load app", kv("link", link))

	profile := newAppProfile(f.workspace, f.manifests)

	resolvedLink, err := profile.resolveProjectRawLink(link)
	if err != nil {
		return nil, err
	}

	context, err := makeAppContext(f.workspace, profile, resolvedLink)
	if err != nil {
		return nil, err
	}

	proj, err := context.loadMainProject()
	if err != nil {
		return nil, err
	}

	app, err := newApp(context, proj)
	if err != nil {
		return nil, err
	}
	return app, nil
}
