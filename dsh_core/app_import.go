package dsh_core

import (
	"slices"
	"text/template"
)

type appImportContainer struct {
	context       *appContext
	project       *project
	scope         projectImportScope
	Imports       []*projectImport
	importsLoaded bool
}

func newAppImportContainer(project *project, scope projectImportScope) *appImportContainer {
	return &appImportContainer{
		context: project.context,
		project: project,
		scope:   scope,
	}
}

func (c *appImportContainer) loadImports() (err error) {
	if c.importsLoaded {
		return nil
	}
	if err = c.project.loadImports(c.scope); err != nil {
		return err
	}

	var imports []*projectImport
	var importsByPath = make(map[string]*projectImport)

	pic := c.project.getImportContainer(c.scope)
	for i := 0; i < len(pic.Imports); i++ {
		imp := pic.Imports[i]
		imports = append(imports, imp)
		importsByPath[imp.Path] = imp
	}

	projectImports := pic.Imports
	for i := 0; i < len(projectImports); i++ {
		imp1 := projectImports[i]
		if err = imp1.target.loadImports(pic.scope); err != nil {
			return err
		}
		pic1 := imp1.target.getImportContainer(pic.scope)
		for j := 0; j < len(pic1.Imports); j++ {
			imp2 := pic1.Imports[j]
			if imp2.Path == c.project.Manifest.projectPath {
				continue
			}
			if _, exist := importsByPath[imp2.Path]; !exist {
				imports = append(imports, imp2)
				importsByPath[imp2.Path] = imp2
				projectImports = append(projectImports, imp2)
			}
		}
	}

	c.Imports = imports
	c.importsLoaded = true
	return nil
}

func (c *appImportContainer) makeConfigs() (configs map[string]any, err error) {
	if c.scope != projectImportScopeConfig {
		panic(desc("make configs only support scope config",
			kv("scope", c.scope),
		))
	}
	if err = c.loadImports(); err != nil {
		return nil, errW(err, "make configs error",
			reason("load imports error"),
			kv("projectName", c.project.Manifest.Name),
			kv("projectPath", c.project.Manifest.projectPath),
		)
	}
	for i := 0; i < len(c.Imports); i++ {
		if err = c.Imports[i].target.loadConfigSources(); err != nil {
			return nil, errW(err, "make configs error",
				reason("load config sources error"),
				kv("projectName", c.Imports[i].target.Manifest.Name),
				kv("projectPath", c.Imports[i].target.Manifest.projectPath),
			)
		}
	}
	if err = c.project.loadConfigSources(); err != nil {
		return nil, errW(err, "make configs error",
			reason("load config sources error"),
			kv("projectName", c.project.Manifest.Name),
			kv("projectPath", c.project.Manifest.projectPath),
		)
	}

	var sources []*projectConfigSource
	for i := 0; i < len(c.Imports); i++ {
		for j := 0; j < len(c.Imports[i].target.Config.SourceContainer.Sources); j++ {
			source := c.Imports[i].target.Config.SourceContainer.Sources[j]
			sources = append(sources, source)
		}
	}
	for i := 0; i < len(c.project.Config.SourceContainer.Sources); i++ {
		source := c.project.Config.SourceContainer.Sources[i]
		sources = append(sources, source)
	}

	slices.SortStableFunc(sources, func(l, r *projectConfigSource) int {
		n := l.content.Order - r.content.Order
		if n < 0 {
			return 1
		} else if n > 0 {
			return -1
		} else {
			return 0
		}
	})

	configs = make(map[string]any)
	for i := 0; i < len(sources); i++ {
		source := sources[i]
		if err = source.mergeConfigs(configs); err != nil {
			return nil, errW(err, "make configs error",
				reason("merge configs error"),
				kv("sourcePath", source.SourcePath),
			)
		}
	}
	return configs, nil
}

func (c *appImportContainer) makeScripts(configs map[string]any, funcs template.FuncMap, outputPath string, useHardLink bool) ([]string, error) {
	if c.scope != projectImportScopeScript {
		panic(desc("make scripts only support scope script",
			kv("scope", c.scope),
		))
	}
	if err := c.loadImports(); err != nil {
		return nil, errW(err, "make scripts error",
			reason("load imports error"),
			kv("projectName", c.project.Manifest.Name),
			kv("projectPath", c.project.Manifest.projectPath),
		)
	}
	var targetNames []string
	for i := 0; i < len(c.Imports); i++ {
		iTargetNames, err := c.Imports[i].target.makeScripts(configs, funcs, outputPath, useHardLink)
		if err != nil {
			return nil, err
		}
		targetNames = append(targetNames, iTargetNames...)
	}
	pTargetNames, err := c.project.makeScripts(configs, funcs, outputPath, useHardLink)
	if err != nil {
		return nil, err
	}
	targetNames = append(targetNames, pTargetNames...)
	return targetNames, nil
}
