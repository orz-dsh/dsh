package dsh_core

import (
	"dsh/dsh_utils"
	"maps"
	"slices"
)

type AppProfile struct {
	workspace       *Workspace
	projectManifest *projectManifest
	evalData        *appProfileEvalData
	Manifests       []*AppProfileManifest
}

func loadAppProfile(workspace *Workspace, projectManifest *projectManifest, paths []string) (*AppProfile, error) {
	workingPath, err := dsh_utils.GetWorkingDir()
	if err != nil {
		return nil, err
	}
	evalData := newAppProfileEvalData(workingPath, workspace.path, projectManifest.projectPath, projectManifest.Name)
	evaluator := newAppProfileEvaluator(evalData)

	var manifests []*AppProfileManifest
	var allPaths []string
	pathsDict := make(map[string]bool)

	for i := len(paths) - 1; i >= 0; i-- {
		path, err := evaluator.evalPath(paths[i])
		if err != nil {
			return nil, err
		}
		if path != "" && !pathsDict[path] {
			allPaths = append(allPaths, path)
			pathsDict[path] = true
		}
	}

	for i := len(workspace.manifest.Profile.Items) - 1; i >= 0; i-- {
		item := workspace.manifest.Profile.Items[i]
		path, err := evaluator.evalMatchAndPath(item.match, item.File)
		if err != nil {
			return nil, err
		}
		if path != "" && !pathsDict[path] {
			if dsh_utils.IsFileExists(path) || !item.Optional {
				allPaths = append(allPaths, path)
				pathsDict[path] = true
			}
		}
	}

	for i := len(allPaths) - 1; i >= 0; i-- {
		path := allPaths[i]
		manifest, err := loadAppProfileManifest(path)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, manifest)
	}

	profile := &AppProfile{
		workspace:       workspace,
		projectManifest: projectManifest,
		evalData:        evalData,
		Manifests:       manifests,
	}
	return profile, nil
}

func (p *AppProfile) MakeApp() (*App, error) {
	return loadApp(p.workspace, p.projectManifest, p)
}

func (p *AppProfile) AddManifest(position int, manifest *AppProfileManifest) {
	if position < 0 {
		p.Manifests = append(p.Manifests, manifest)
	} else {
		p.Manifests = slices.Insert(p.Manifests, position, manifest)
	}
}

func (p *AppProfile) AddManifestOptionValues(position int, values map[string]string) error {
	manifest, err := MakeAppProfileManifest(values, nil, nil, nil)
	if err != nil {
		return err
	}
	p.AddManifest(position, manifest)
	return nil
}

func (p *AppProfile) getOptionValues() map[string]string {
	options := make(map[string]string)
	for i := 0; i < len(p.Manifests); i++ {
		manifest := p.Manifests[i]
		maps.Copy(options, manifest.Option.Values)
	}
	return options
}

func (p *AppProfile) getShellPath(shell string) string {
	for i := len(p.Manifests) - 1; i >= 0; i-- {
		manifest := p.Manifests[i]
		path := manifest.Shell.getPath(shell)
		if path != "" {
			return path
		}
	}
	return p.workspace.manifest.Shell.getPath(shell)
}

func (p *AppProfile) getShellExts(shell string) []string {
	for i := len(p.Manifests) - 1; i >= 0; i-- {
		manifest := p.Manifests[i]
		exts := manifest.Shell.getExts(shell)
		if exts != nil {
			return exts
		}
	}
	return p.workspace.manifest.Shell.getExts(shell)
}

func (p *AppProfile) getShellArgs(shell string) []string {
	for i := len(p.Manifests) - 1; i >= 0; i-- {
		manifest := p.Manifests[i]
		args := manifest.Shell.getArgs(shell)
		if args != nil {
			return args
		}
	}
	return p.workspace.manifest.Shell.getArgs(shell)
}

func (p *AppProfile) getImportRegistryDefinition(name string) *importRegistryDefinition {
	for i := len(p.Manifests) - 1; i >= 0; i-- {
		manifest := p.Manifests[i]
		definition := manifest.Import.getRegistryDefinition(name)
		if definition != nil {
			return definition
		}
	}
	definition := p.workspace.manifest.Import.getRegistryDefinition(name)
	if definition != nil {
		return definition
	}
	return nil
}

func (p *AppProfile) getImportRedirectDefinition(resources []string) (*importRedirectDefinition, string) {
	for i := len(p.Manifests) - 1; i >= 0; i-- {
		manifest := p.Manifests[i]
		definition, path := manifest.Import.getRedirectDefinition(resources)
		if definition != nil {
			return definition, path
		}
	}
	definition, path := p.workspace.manifest.Import.getRedirectDefinition(resources)
	if definition != nil {
		return definition, path
	}
	return nil, ""
}
