package dsh_core

import (
	"dsh/dsh_utils"
	"runtime"
	"slices"
	"strings"
)

type AppFactory struct {
	workspace *Workspace
	evaluator *Evaluator
	manifests []*AppProfileManifest
}

func makeAppFactory(workspace *Workspace) (*AppFactory, error) {
	workingDir, err := dsh_utils.GetWorkingDir()
	if err != nil {
		return nil, err
	}
	evaluator := dsh_utils.NewEvaluator().SetData("local", map[string]any{
		"working_dir":          workingDir,
		"workspace_dir":        workspace.path,
		"runtime_version":      dsh_utils.GetRuntimeVersion(),
		"runtime_version_code": dsh_utils.GetRuntimeVersionCode(),
		"os":                   strings.ToLower(runtime.GOOS),
	})

	files, err := workspace.manifest.Profile.definitions.getFiles(evaluator)
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
		evaluator: evaluator,
		manifests: manifests,
	}
	return factory, nil
}

func (f *AppFactory) AddProfileManifest(position int, manifest *AppProfileManifest) {
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
	f.AddProfileManifest(position, manifest)
	return nil
}

func (f *AppFactory) AddProfileProjectOption(position int, values map[string]string) error {
	manifest, err := MakeAppProfileManifest(nil, NewAppProfileManifestProject(NewAppProfileManifestProjectOption(values), nil, nil))
	if err != nil {
		return err
	}
	f.AddProfileManifest(position, manifest)
	return nil
}

func (f *AppFactory) MakeApp(link string) (*App, error) {
	f.workspace.logger.InfoDesc("load app", kv("link", link))

	profile := newAppProfile(f.workspace, f.evaluator, f.manifests)

	resolvedLink, err := profile.resolveProjectRawLink(link)
	if err != nil {
		return nil, err
	}

	context, err := makeAppContext(f.workspace, f.evaluator, profile, resolvedLink)
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
