package dsh_core

// region project

type projectInstance struct {
	Name    string
	Path    string
	context *appContext
	option  *projectOption
	script  *projectScript
	config  *projectConfigInstance
}

func createProjectInstance(context *appContext, setting *projectSetting) (instance *projectInstance, err error) {
	context.logger.InfoDesc("create project instance", kv("name", setting.Name))
	option, err := makeProjectOption(context, setting)
	if err != nil {
		return nil, err
	}
	script, err := makeProjectScript(context, setting, option)
	if err != nil {
		return nil, err
	}
	config, err := newProjectConfigInstance(context, setting, option)
	if err != nil {
		return nil, err
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
