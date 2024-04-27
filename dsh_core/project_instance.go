package dsh_core

import (
	"fmt"
	"path/filepath"
	"text/template"
)

type projectInstance struct {
	context *Context
	info    *projectInfo
	script  *projectInstanceScript
	config  *projectInstanceConfig
}

type projectInstanceSourceContainer interface {
	scanSources(sourceDir string, includeFiles []string) error
}

func newProjectInstance(context *Context, info *projectInfo) (instance *projectInstance, err error) {
	context.Logger.Info("instance project: name=%s", info.name)

	for i := 0; i < len(info.manifest.Option.Items); i++ {
		// TODO: 遍历 options
	}

	script := newProjectInstanceScript(context)
	config := newProjectInstanceConfig(context)
	sources := [][]projectManifestSource{
		info.manifest.Script.Sources,
		info.manifest.Config.Sources,
	}
	sourceContainers := []projectInstanceSourceContainer{
		script.sourceContainer,
		config.sourceContainer,
	}
	imports := [][]projectManifestImport{
		info.manifest.Script.Imports,
		info.manifest.Config.Imports,
	}
	importContainers := []*projectInstanceImportShallowContainer{
		script.importContainer,
		config.importContainer,
	}
	for i := 0; i < len(sources); i++ {
		for j := 0; j < len(sources[i]); j++ {
			src := sources[i][j]
			if src.Dir != "" {
				// TODO: selector match
				if err = sourceContainers[i].scanSources(filepath.Join(info.path, src.Dir), src.Files); err != nil {
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
				if err = importContainers[i].importLocal(context, imp.Local.Dir, info); err != nil {
					return nil, err
				}
			} else if imp.Git != nil && imp.Git.Url != "" && imp.Git.Ref != "" {
				// TODO: selector match
				if err = importContainers[i].importGit(context, info, imp.Git.Url, imp.Git.Ref); err != nil {
					return nil, err
				}
			}
		}
	}
	return &projectInstance{
		context: context,
		info:    info,
		script:  script,
		config:  config,
	}, nil
}

func (instance *projectInstance) getImportContainer(scope projectInstanceImportScope) *projectInstanceImportShallowContainer {
	if scope == projectInstanceImportScopeScript {
		return instance.script.importContainer
	} else if scope == projectInstanceImportScopeConfig {
		return instance.config.importContainer
	}
	panic(fmt.Sprintf("invalid import scope: scope=%s", scope))
	return nil
}

func (instance *projectInstance) loadImports(scope projectInstanceImportScope) error {
	return instance.getImportContainer(scope).loadImports()
}

func (instance *projectInstance) buildScriptSources(config map[string]any, funcs template.FuncMap, outputPath string) error {
	projectOutputPath := filepath.Join(outputPath, instance.info.name)
	return instance.script.sourceContainer.buildSources(config, funcs, projectOutputPath)
}

func (instance *projectInstance) loadConfigSources() error {
	return instance.config.sourceContainer.loadSources()
}