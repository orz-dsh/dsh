package dsh_core

import (
	"maps"
	"slices"
)

type AppProfile struct {
	workspace *Workspace
	Manifests []*AppProfileManifest
}

func loadAppProfile(workspace *Workspace, paths []string) (*AppProfile, error) {
	var manifests []*AppProfileManifest
	for i := 0; i < len(paths); i++ {
		path := paths[i]
		manifest, err := loadAppProfileManifest(path)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, manifest)
	}

	profile := &AppProfile{
		workspace: workspace,
		Manifests: manifests,
	}
	return profile, nil
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

func (p *AppProfile) getImportRegistry(name string) *importRegistryDefinition {
	for i := len(p.Manifests) - 1; i >= 0; i-- {
		manifest := p.Manifests[i]
		registry := manifest.Import.getRegistry(name)
		if registry != nil {
			return registry.definition
		}
	}
	registry := p.workspace.manifest.Import.getRegistry(name)
	if registry != nil {
		return registry.definition
	}
	return nil
}
