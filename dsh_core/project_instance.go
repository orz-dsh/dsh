package dsh_core

import (
	"path/filepath"
	"text/template"
)

type ProjectInstance struct {
	Context *Context
	Info    *ProjectInfo
	Script  *ProjectInstanceScript
	Config  *ProjectInstanceConfig
}

type ProjectInstanceSourceContainer interface {
	ScanSources(sourceDir string, includeFiles []string) error
}

func NewProjectInstance(context *Context, info *ProjectInfo) (instance *ProjectInstance, err error) {
	for i := 0; i < len(info.Manifest.Option.Items); i++ {
		// TODO: 遍历 options
	}

	script := NewProjectInstanceScript(context)
	config := NewProjectInstanceConfig(context)
	sources := [][]ProjectManifestSource{
		info.Manifest.Script.Sources,
		info.Manifest.Config.Sources,
	}
	sourceContainers := []ProjectInstanceSourceContainer{
		script.SourceContainer,
		config.SourceContainer,
	}
	imports := [][]ProjectManifestImport{
		info.Manifest.Script.Imports,
		info.Manifest.Config.Imports,
	}
	importContainers := []*ProjectInstanceImportShallowContainer{
		script.ImportContainer,
		config.ImportContainer,
	}
	for i := 0; i < len(sources); i++ {
		for j := 0; j < len(sources[i]); j++ {
			src := sources[i][j]
			if src.Dir != "" {
				// TODO: selector match
				if err = sourceContainers[i].ScanSources(filepath.Join(info.Path, src.Dir), src.Files); err != nil {
					return nil, err
				}
			}
		}
	}
	for i := 0; i < len(imports); i++ {
		for j := 0; j < len(imports[i]); j++ {
			imp := imports[i][j]
			if imp.Local != nil && imp.Local.Dir != "" {
				// TODO: selector match
				if err = importContainers[i].ImportLocal(context, imp.Local.Dir, info); err != nil {
					return nil, err
				}
			} else if imp.Git != nil && imp.Git.Url != "" && imp.Git.Ref != "" {
				// TODO: selector match
				if err = importContainers[i].ImportGit(context, info, imp.Git.Url, imp.Git.Ref); err != nil {
					return nil, err
				}
			}
		}
	}
	return &ProjectInstance{
		Context: context,
		Info:    info,
		Script:  script,
		Config:  config,
	}, nil
}

func (instance *ProjectInstance) GetImportContainer(scope ProjectInstanceImportScope) *ProjectInstanceImportShallowContainer {
	if scope == ProjectInstanceImportScopeScript {
		return instance.Script.ImportContainer
	} else if scope == ProjectInstanceImportScopeConfig {
		return instance.Config.ImportContainer
	}
	instance.Context.Logger.Panic("invalid import scope: scope=%s", scope)
	return nil
}

func (instance *ProjectInstance) LoadImports(scope ProjectInstanceImportScope) error {
	return instance.GetImportContainer(scope).LoadImports()
}

func (instance *ProjectInstance) BuildScriptSources(config map[string]any, funcs template.FuncMap, outputPath string) error {
	projectOutputPath := filepath.Join(outputPath, instance.Info.Name)
	return instance.Script.SourceContainer.BuildSources(config, funcs, projectOutputPath)
}

func (instance *ProjectInstance) LoadConfigSources() error {
	return instance.Config.SourceContainer.LoadSources()
}
