package dsh_core

import (
	"dsh/dsh_utils"
	"slices"
)

type AppProfile struct {
	workspace       *Workspace
	projectManifest *projectManifest
	evalData        *appProfileEvalData
	evaluator       *appProfileEvaluator
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
		path, err := evaluator.evalString(paths[i])
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
		path, err := evaluator.evalMatchAndString(item.match, item.File)
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
		evaluator:       evaluator,
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
	manifest, err := MakeAppProfileManifest(nil, NewAppProfileManifestProject(NewAppProfileManifestProjectOption(values), nil, nil))
	if err != nil {
		return err
	}
	p.AddManifest(position, manifest)
	return nil
}

func (p *AppProfile) getOptionValues() (options map[string]string, err error) {
	options = make(map[string]string)
	for i := 0; i < len(p.Manifests); i++ {
		if err = p.Manifests[i].Project.Option.definitions.fillOptions(options, p.evaluator.newMatcher()); err != nil {
			return nil, err
		}
	}
	return options, nil
}

func (p *AppProfile) getWorkspaceShellDefinition(name string) (*workspaceShellDefinition, error) {
	definition := newWorkspaceShellDefinitionEmpty(name)
	for i := len(p.Manifests) - 1; i >= 0; i-- {
		manifest := p.Manifests[i]
		err := manifest.Workspace.Shell.definitions.fillDefinition(definition, p.evaluator.newMatcher())
		if err != nil {
			return nil, err
		}
		if definition.isCompleted() {
			return definition, nil
		}
	}
	err := p.workspace.manifest.Shell.definitions.fillDefinition(definition, p.evaluator.newMatcher())
	if err != nil {
		return nil, err
	}
	if err = definition.fillDefault(); err != nil {
		return nil, err
	}
	return definition, nil
}

func (p *AppProfile) getWorkspaceImportRegistryDefinition(name string) *workspaceImportRegistryDefinition {
	for i := len(p.Manifests) - 1; i >= 0; i-- {
		manifest := p.Manifests[i]
		if definition := manifest.Workspace.Import.registryDefinitions.getDefinition(name); definition != nil {
			return definition
		}
	}
	if definition := p.workspace.manifest.Import.registryDefinitions.getDefinition(name); definition != nil {
		return definition
	}
	if definition := getWorkspaceImportRegistryDefinitionDefault(name); definition != nil {
		return definition
	}
	return nil
}

func (p *AppProfile) getWorkspaceImportRedirectDefinition(resources []string) (*workspaceImportRedirectDefinition, string) {
	for i := len(p.Manifests) - 1; i >= 0; i-- {
		manifest := p.Manifests[i]
		definition, path := manifest.Workspace.Import.redirectDefinitions.getDefinition(resources)
		if definition != nil {
			return definition, path
		}
	}
	definition, path := p.workspace.manifest.Import.redirectDefinitions.getDefinition(resources)
	if definition != nil {
		return definition, path
	}
	return nil, ""
}

func (p *AppProfile) getProjectScriptSourceDefinitions() (definitions []*projectSourceDefinition) {
	for i := 0; i < len(p.Manifests); i++ {
		definitions = append(definitions, p.Manifests[i].Project.Script.sourceDefinitions...)
	}
	return definitions
}

func (p *AppProfile) getProjectScriptImportDefinitions() (definitions []*projectImportDefinition) {
	for i := 0; i < len(p.Manifests); i++ {
		definitions = append(definitions, p.Manifests[i].Project.Script.importDefinitions...)
	}
	return definitions
}

func (p *AppProfile) getProjectConfigSourceDefinitions() (definitions []*projectSourceDefinition) {
	for i := 0; i < len(p.Manifests); i++ {
		definitions = append(definitions, p.Manifests[i].Project.Config.sourceDefinitions...)
	}
	return definitions
}

func (p *AppProfile) getProjectConfigImportDefinitions() (definitions []*projectImportDefinition) {
	for i := 0; i < len(p.Manifests); i++ {
		definitions = append(definitions, p.Manifests[i].Project.Config.importDefinitions...)
	}
	return definitions
}
