package dsh_core

import (
	"slices"
	"text/template"
)

type appImportContainer struct {
	context       *appContext
	project       *project
	scope         projectImportScope
	imports       []*projectImport
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
	var importsByUnique = make(map[string]*projectImport)

	pic := c.project.getImportContainer(c.scope)
	for i := 0; i < len(pic.imports); i++ {
		imp := pic.imports[i]
		imports = append(imports, imp)
		importsByUnique[imp.unique] = imp
	}

	projectImports := pic.imports
	for i := 0; i < len(projectImports); i++ {
		imp1 := projectImports[i]
		if err = imp1.project.loadImports(pic.scope); err != nil {
			return err
		}
		pic1 := imp1.project.getImportContainer(pic.scope)
		for j := 0; j < len(pic1.imports); j++ {
			imp2 := pic1.imports[j]
			if imp2.projectPath == c.project.manifest.projectPath {
				continue
			}
			if _, exist := importsByUnique[imp2.unique]; !exist {
				imports = append(imports, imp2)
				importsByUnique[imp2.unique] = imp2
				projectImports = append(projectImports, imp2)
			}
		}
	}

	c.imports = imports
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
			kv("projectName", c.project.manifest.Name),
			kv("projectPath", c.project.manifest.projectPath),
		)
	}
	for i := 0; i < len(c.imports); i++ {
		if err = c.imports[i].project.loadConfigSources(); err != nil {
			return nil, errW(err, "make configs error",
				reason("load config sources error"),
				kv("projectName", c.imports[i].project.manifest.Name),
				kv("projectPath", c.imports[i].project.manifest.projectPath),
			)
		}
	}
	if err = c.project.loadConfigSources(); err != nil {
		return nil, errW(err, "make configs error",
			reason("load config sources error"),
			kv("projectName", c.project.manifest.Name),
			kv("projectPath", c.project.manifest.projectPath),
		)
	}

	var sources []*projectConfigSource
	for i := 0; i < len(c.imports); i++ {
		for j := 0; j < len(c.imports[i].project.config.sourceContainer.sources); j++ {
			source := c.imports[i].project.config.sourceContainer.sources[j]
			sources = append(sources, source)
		}
	}
	for i := 0; i < len(c.project.config.sourceContainer.sources); i++ {
		source := c.project.config.sourceContainer.sources[i]
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
				kv("sourcePath", source.sourcePath),
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
			kv("projectName", c.project.manifest.Name),
			kv("projectPath", c.project.manifest.projectPath),
		)
	}
	var targetNames []string
	for i := 0; i < len(c.imports); i++ {
		iTargetNames, err := c.imports[i].project.makeScripts(configs, funcs, outputPath, useHardLink)
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
