package dsh_core

import (
	"fmt"
	"path/filepath"
	"text/template"
)

type projectInstance struct {
	context *projectContext
	info    *projectInfo
	option  *projectInstanceOption
	script  *projectInstanceScript
	config  *projectInstanceConfig
}

type projectInstanceSourceContainer interface {
	scanSources(sourceDir string, includeFiles []string) error
}

func newProjectInstance(context *projectContext, info *projectInfo, optionValues map[string]string) (instance *projectInstance, err error) {
	context.logger.Info("instance project: name=%s", info.name)

	option, err := newProjectInstanceOption(context, info, optionValues)
	if err != nil {
		return nil, err
	}
	script := newProjectInstanceScript(context)
	config := newProjectInstanceConfig(context)
	sources := [][]*projectManifestSource{
		info.manifest.Script.Sources,
		info.manifest.Config.Sources,
	}
	sourceContainers := []projectInstanceSourceContainer{
		script.sourceContainer,
		config.sourceContainer,
	}
	imports := [][]*projectManifestImport{
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
				if src.Match != "" {
					matched, err := option.match(src.match)
					if err != nil {
						return nil, err
					}
					if !matched {
						continue
					}
				}
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
				if imp.Match != "" {
					matched, err := option.match(imp.match)
					if err != nil {
						return nil, err
					}
					if !matched {
						continue
					}
				}
				if err = importContainers[i].importLocal(context, imp.Local.Dir, info); err != nil {
					return nil, err
				}
			} else if imp.Git != nil && imp.Git.Url != "" && imp.Git.Ref != "" {
				if imp.Match != "" {
					matched, err := option.match(imp.match)
					if err != nil {
						return nil, err
					}
					if !matched {
						continue
					}
				}
				if err = importContainers[i].importGit(context, info, imp.Git.Url, imp.Git.Ref); err != nil {
					return nil, err
				}
			}
		}
	}
	return &projectInstance{
		context: context,
		info:    info,
		option:  option,
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

func (instance *projectInstance) makeScript(config map[string]any, funcs template.FuncMap, outputPath string) error {
	projectOutputPath := filepath.Join(outputPath, instance.info.name)
	return instance.script.sourceContainer.make(config, funcs, projectOutputPath)
}

func (instance *projectInstance) loadConfigSources() error {
	return instance.config.sourceContainer.loadSources()
}
