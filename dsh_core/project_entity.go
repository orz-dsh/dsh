package dsh_core

// region project

type projectEntity struct {
	Name    string
	Path    string
	context *appContext
	option  *projectOption
	script  *projectScript
	config  *projectConfig
}

func createProjectEntity(context *appContext, schema *projectSetting) (project *projectEntity, err error) {
	context.logger.InfoDesc("create project instance", kv("name", schema.Name))
	option, err := makeProjectOption(context, schema)
	if err != nil {
		return nil, err
	}
	script, err := makeProjectScript(context, schema, option)
	if err != nil {
		return nil, err
	}
	config, err := makeProjectConfig(context, schema, option)
	if err != nil {
		return nil, err
	}
	project = &projectEntity{
		Name:    schema.Name,
		Path:    schema.Path,
		context: context,
		option:  option,
		script:  script,
		config:  config,
	}
	return project, nil
}

func (p *projectEntity) getImportContainer(scope projectImportScope) *projectImportContainer {
	if scope == projectImportScopeScript {
		return p.script.ImportContainer
	} else if scope == projectImportScopeConfig {
		return p.config.ImportContainer
	} else {
		impossible()
	}
	return nil
}

func (p *projectEntity) loadImports(scope projectImportScope) error {
	return p.getImportContainer(scope).loadImports()
}

func (p *projectEntity) loadConfigSources() error {
	return p.config.SourceContainer.loadSources()
}

func (p *projectEntity) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool) ([]string, error) {
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
