package dsh_core

import (
	"slices"
)

// region project

type appProject struct {
	Name    string
	Path    string
	context *appContext
	option  *projectOption
	script  *projectScript
	config  *projectConfig
}

func makeAppProject(context *appContext, entity *projectSetting) (project *appProject, err error) {
	context.logger.InfoDesc("load project", kv("name", entity.Name))
	option, err := makeProjectOption(context, entity)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("make project option error"),
			kv("projectName", entity.Name),
			kv("projectPath", entity.Path),
		)
	}
	script, err := makeProjectScript(context, entity, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project script error"),
			kv("projectName", entity.Name),
			kv("projectPath", entity.Path),
		)
	}
	config, err := makeProjectConfig(context, entity, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project config error"),
			kv("projectName", entity.Name),
			kv("projectPath", entity.Path),
		)
	}
	project = &appProject{
		Name:    entity.Name,
		Path:    entity.Path,
		context: context,
		option:  option,
		script:  script,
		config:  config,
	}
	return project, nil
}

func (p *appProject) getImportContainer(scope projectImportScope) *projectImportContainer {
	if scope == projectImportScopeScript {
		return p.script.ImportContainer
	} else if scope == projectImportScopeConfig {
		return p.config.ImportContainer
	} else {
		impossible()
	}
	return nil
}

func (p *appProject) loadImports(scope projectImportScope) error {
	return p.getImportContainer(scope).loadImports()
}

func (p *appProject) loadConfigSources() error {
	return p.config.SourceContainer.loadSources()
}

func (p *appProject) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool) ([]string, error) {
	evaluator = evaluator.SetData("options", p.option.Items)
	targetNames, err := p.script.SourceContainer.makeSources(evaluator, outputPath, useHardLink)
	if err != nil {
		return nil, errW(err, "make scripts error",
			reason("make sources error"),
			kv("project", p),
		)
	}
	return targetNames, nil
}

// endregion

// region container

type appProjectContainer struct {
	context       *appContext
	mainProject   *appProject
	extraProjects []*appProject
	projects      []*appProject
	scope         projectImportScope
	Imports       []*projectEntityImport
	importsLoaded bool
}

func newAppProjectContainer(mainProject *appProject, extraProjects []*appProject, scope projectImportScope) *appProjectContainer {
	return &appProjectContainer{
		context:       mainProject.context,
		mainProject:   mainProject,
		extraProjects: extraProjects,
		projects:      append([]*appProject{mainProject}, extraProjects...),
		scope:         scope,
	}
}

func (c *appProjectContainer) loadImports() (err error) {
	if c.importsLoaded {
		return nil
	}
	for i := 0; i < len(c.projects); i++ {
		if err = c.projects[i].loadImports(c.scope); err != nil {
			return err
		}
	}

	var imports []*projectEntityImport
	var importsByPath = map[string]*projectEntityImport{}

	for i := 0; i < len(c.projects); i++ {
		pic := c.projects[i].getImportContainer(c.scope)
		for j := 0; j < len(pic.Imports); j++ {
			imp := pic.Imports[j]
			imports = append(imports, imp)
			importsByPath[imp.Target.Path] = imp
		}

		projectImports := pic.Imports
		for j := 0; j < len(projectImports); j++ {
			imp1 := projectImports[j]
			if err = imp1.project.loadImports(pic.scope); err != nil {
				return err
			}
			pic1 := imp1.project.getImportContainer(pic.scope)
			for k := 0; k < len(pic1.Imports); k++ {
				imp2 := pic1.Imports[k]
				if imp2.Target.Path == c.mainProject.Path {
					// TODO: import extra project ?
					continue
				}
				if _, exist := importsByPath[imp2.Target.Path]; !exist {
					imports = append(imports, imp2)
					importsByPath[imp2.Target.Path] = imp2
					projectImports = append(projectImports, imp2)
				}
			}
		}
	}

	c.Imports = imports
	c.importsLoaded = true
	return nil
}

func (c *appProjectContainer) makeConfigs() (configs map[string]any, err error) {
	if c.scope != projectImportScopeConfig {
		panic(desc("make configs only support scope config",
			kv("scope", c.scope),
		))
	}
	if err = c.loadImports(); err != nil {
		return nil, errW(err, "make configs error",
			reason("load imports error"),
			// TODO: error
			kv("project", c.mainProject),
		)
	}
	for i := 0; i < len(c.Imports); i++ {
		if err = c.Imports[i].project.loadConfigSources(); err != nil {
			return nil, errW(err, "make configs error",
				reason("load config sources error"),
				kv("project", c.Imports[i].project),
			)
		}
	}
	for i := 0; i < len(c.projects); i++ {
		if err = c.projects[i].loadConfigSources(); err != nil {
			return nil, errW(err, "make configs error",
				reason("load config sources error"),
				kv("project", c.projects[i]),
			)
		}
	}

	var sources []*projectConfigSource
	for i := 0; i < len(c.Imports); i++ {
		for j := 0; j < len(c.Imports[i].project.config.SourceContainer.Sources); j++ {
			source := c.Imports[i].project.config.SourceContainer.Sources[j]
			sources = append(sources, source)
		}
	}
	for i := 0; i < len(c.projects); i++ {
		for j := 0; j < len(c.projects[i].config.SourceContainer.Sources); j++ {
			source := c.projects[i].config.SourceContainer.Sources[j]
			sources = append(sources, source)
		}
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

	configs = map[string]any{}
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

func (c *appProjectContainer) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool) ([]string, error) {
	if c.scope != projectImportScopeScript {
		panic(desc("make scripts only support scope script",
			kv("scope", c.scope),
		))
	}
	if err := c.loadImports(); err != nil {
		return nil, errW(err, "make scripts error",
			reason("load imports error"),
			// TODO: error
			kv("project", c.mainProject),
		)
	}
	var targetNames []string
	for i := 0; i < len(c.Imports); i++ {
		iTargetNames, err := c.Imports[i].project.makeScripts(evaluator, outputPath, useHardLink)
		if err != nil {
			return nil, err
		}
		targetNames = append(targetNames, iTargetNames...)
	}
	for i := 0; i < len(c.projects); i++ {
		pTargetNames, err := c.projects[i].makeScripts(evaluator, outputPath, useHardLink)
		if err != nil {
			return nil, err
		}
		// TODO: duplicate target names
		targetNames = append(targetNames, pTargetNames...)
	}
	return targetNames, nil
}

// endregion
