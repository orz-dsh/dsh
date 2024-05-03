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

func (ic *appImportContainer) loadImports() (err error) {
	if ic.importsLoaded {
		return nil
	}
	if err = ic.project.loadImports(ic.scope); err != nil {
		return err
	}

	var imports []*projectImport
	var importsByUnique = make(map[string]*projectImport)

	pic := ic.project.getImportContainer(ic.scope)
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
			if imp2.projectPath == ic.project.manifest.projectPath {
				continue
			}
			if _, exist := importsByUnique[imp2.unique]; !exist {
				imports = append(imports, imp2)
				importsByUnique[imp2.unique] = imp2
				projectImports = append(projectImports, imp2)
			}
		}
	}

	ic.imports = imports
	ic.importsLoaded = true
	return nil
}

func (ic *appImportContainer) makeConfigs() (configs map[string]any, err error) {
	if ic.scope != projectImportScopeConfig {
		panic(desc("make configs only support scope config",
			kv("scope", ic.scope),
		))
	}
	if err = ic.loadImports(); err != nil {
		return nil, errW(err, "make configs error",
			reason("load imports error"),
			kv("projectName", ic.project.manifest.Name),
			kv("projectPath", ic.project.manifest.projectPath),
		)
	}
	for i := 0; i < len(ic.imports); i++ {
		if err = ic.imports[i].project.loadConfigSources(); err != nil {
			return nil, errW(err, "make configs error",
				reason("load config sources error"),
				kv("projectName", ic.imports[i].project.manifest.Name),
				kv("projectPath", ic.imports[i].project.manifest.projectPath),
			)
		}
	}
	if err = ic.project.loadConfigSources(); err != nil {
		return nil, errW(err, "make configs error",
			reason("load config sources error"),
			kv("projectName", ic.project.manifest.Name),
			kv("projectPath", ic.project.manifest.projectPath),
		)
	}

	var sources []*projectConfigSource
	for i := 0; i < len(ic.imports); i++ {
		for j := 0; j < len(ic.imports[i].project.config.sourceContainer.sources); j++ {
			source := ic.imports[i].project.config.sourceContainer.sources[j]
			sources = append(sources, source)
		}
	}
	for i := 0; i < len(ic.project.config.sourceContainer.sources); i++ {
		source := ic.project.config.sourceContainer.sources[i]
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

func (ic *appImportContainer) makeScripts(configs map[string]any, funcs template.FuncMap, outputPath string) (err error) {
	if ic.scope != projectImportScopeScript {
		panic(desc("make scripts only support scope script",
			kv("scope", ic.scope),
		))
	}
	if err = ic.loadImports(); err != nil {
		return errW(err, "make scripts error",
			reason("load imports error"),
			kv("projectName", ic.project.manifest.Name),
			kv("projectPath", ic.project.manifest.projectPath),
		)
	}
	for i := 0; i < len(ic.imports); i++ {
		if err = ic.imports[i].project.makeScripts(configs, funcs, outputPath); err != nil {
			return err
		}
	}
	if err = ic.project.makeScripts(configs, funcs, outputPath); err != nil {
		return err
	}
	return nil
}
