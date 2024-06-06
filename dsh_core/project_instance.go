package dsh_core

import "slices"

// region projectInstance

type projectInstance struct {
	Name    string
	Path    string
	context *appContext
	option  *projectOptionInstance
	script  *projectScriptInstance
	config  *projectConfigInstance
}

func newProjectInstance(context *appContext, setting *projectSetting) (instance *projectInstance, err error) {
	context.logger.InfoDesc("load project", kv("name", setting.Name))
	option, err := makeProjectOption(context, setting)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("make project option error"),
			kv("projectName", setting.Name),
			kv("projectPath", setting.Path),
		)
	}
	script, err := newProjectScriptInstance(context, setting, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project script error"),
			kv("projectName", setting.Name),
			kv("projectPath", setting.Path),
		)
	}
	config, err := newProjectConfigInstance(context, setting, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project config error"),
			kv("projectName", setting.Name),
			kv("projectPath", setting.Path),
		)
	}
	instance = &projectInstance{
		Name:    setting.Name,
		Path:    setting.Path,
		context: context,
		option:  option,
		script:  script,
		config:  config,
	}
	return instance, nil
}

func (p *projectInstance) getImportContainer(scope projectImportScope) *projectImportInstanceContainer {
	if scope == projectImportScopeScript {
		return p.script.ImportContainer
	} else if scope == projectImportScopeConfig {
		return p.config.ImportContainer
	} else {
		impossible()
	}
	return nil
}

func (p *projectInstance) loadImports(scope projectImportScope) error {
	return p.getImportContainer(scope).loadImports()
}

func (p *projectInstance) loadConfigContents() ([]*projectConfigContentInstance, error) {
	return p.config.SourceContainer.loadContents()
}

func (p *projectInstance) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool) ([]string, error) {
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

// region projectInstanceContainer

type projectInstanceContainer struct {
	context       *appContext
	mainProject   *projectInstance
	extraProjects []*projectInstance
	projects      []*projectInstance
	scope         projectImportScope
	Imports       []*projectImportInstance
	importsLoaded bool
}

func newProjectInstanceContainer(mainProject *projectInstance, extraProjects []*projectInstance, scope projectImportScope) *projectInstanceContainer {
	return &projectInstanceContainer{
		context:       mainProject.context,
		mainProject:   mainProject,
		extraProjects: extraProjects,
		projects:      append([]*projectInstance{mainProject}, extraProjects...),
		scope:         scope,
	}
}

func (c *projectInstanceContainer) loadImports() (err error) {
	if c.importsLoaded {
		return nil
	}
	for i := 0; i < len(c.projects); i++ {
		if err = c.projects[i].loadImports(c.scope); err != nil {
			return err
		}
	}

	var imports []*projectImportInstance
	var importsByPath = map[string]*projectImportInstance{}

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

func (c *projectInstanceContainer) makeConfigs() (configs map[string]any, configTraces map[string]any, err error) {
	if c.scope != projectImportScopeConfig {
		panic(desc("make configs only support scope config",
			kv("scope", c.scope),
		))
	}
	if err = c.loadImports(); err != nil {
		return nil, nil, errW(err, "make configs error",
			reason("load imports error"),
			// TODO: error
			kv("project", c.mainProject),
		)
	}

	var contents []*projectConfigContentInstance
	for i := 0; i < len(c.Imports); i++ {
		iContents, err := c.Imports[i].project.loadConfigContents()
		if err != nil {
			return nil, nil, errW(err, "make configs error",
				reason("load config contents error"),
				kv("project", c.Imports[i].project),
			)
		}
		contents = append(contents, iContents...)
	}
	for i := 0; i < len(c.projects); i++ {
		pContents, err := c.projects[i].loadConfigContents()
		if err != nil {
			return nil, nil, errW(err, "make configs error",
				reason("load config contents error"),
				kv("project", c.projects[i]),
			)
		}
		contents = append(contents, pContents...)
	}

	slices.SortStableFunc(contents, func(l, r *projectConfigContentInstance) int {
		n := l.Order - r.Order
		if n < 0 {
			return 1
		} else if n > 0 {
			return -1
		} else {
			return 0
		}
	})

	configs = map[string]any{}
	configTraces = map[string]any{}
	for i := 0; i < len(contents); i++ {
		content := contents[i]
		if err = content.merge(configs, configTraces); err != nil {
			return nil, nil, errW(err, "make configs error",
				reason("merge configs error"),
				kv("sourcePath", content.sourcePath),
			)
		}
	}
	return configs, configTraces, nil
}

func (c *projectInstanceContainer) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool) ([]string, error) {
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
