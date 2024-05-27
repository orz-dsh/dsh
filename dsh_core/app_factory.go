package dsh_core

import (
	"dsh/dsh_utils"
	"slices"
)

type AppFactory struct {
	workspace *Workspace
	evalData  *appEvalData
	manifests []*AppProfileManifest
}

func makeAppFactory(workspace *Workspace) (*AppFactory, error) {
	evalData, err := makeAppEvalData(workspace)
	if err != nil {
		return nil, err
	}
	matcher := dsh_utils.NewEvalMatcher(evalData)
	replacer := dsh_utils.NewEvalReplacer(evalData, nil)
	files, err := workspace.manifest.Profile.definitions.getFiles(matcher, replacer)
	if err != nil {
		return nil, err
	}
	var manifests []*AppProfileManifest
	for i := 0; i < len(files); i++ {
		manifest, err := loadAppProfileManifest(files[i])
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, manifest)
	}
	factory := &AppFactory{
		workspace: workspace,
		evalData:  evalData,
		manifests: manifests,
	}

	return factory, nil
}

func (f *AppFactory) AddManifest(position int, manifest *AppProfileManifest) {
	if position < 0 {
		f.manifests = append(f.manifests, manifest)
	} else {
		f.manifests = slices.Insert(f.manifests, position, manifest)
	}
}

func (f *AppFactory) AddManifestOptionValues(position int, values map[string]string) error {
	manifest, err := MakeAppProfileManifest(nil, NewAppProfileManifestProject(NewAppProfileManifestProjectOption(values), nil, nil))
	if err != nil {
		return err
	}
	f.AddManifest(position, manifest)
	return nil
}

func (f *AppFactory) MakeApp(rawLink string) (*App, error) {
	profile, err := makeAppProfile(f.workspace, f.evalData, f.manifests)
	if err != nil {
		return nil, err
	}
	link, err := ParseProjectLink(rawLink)
	if err != nil {
		return nil, err
	}
	resolvedLink, err := profile.resolveProjectLink(link)
	if err != nil {
		return nil, err
	}
	app, err := loadApp(f.workspace, resolvedLink, profile)
	if err != nil {
		return nil, err
	}
	return app, nil
}
