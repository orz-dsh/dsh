package dsh_core

// region project

type Project struct {
	Name    string
	Path    string
	context *appContext
	option  *projectOption
	script  *projectScript
	config  *projectConfig
}

func makeProject(context *appContext, manifest *ProjectManifest) (project *Project, err error) {
	context.logger.InfoDesc("load project", kv("name", manifest.projectName))
	option, err := makeProjectOption(context, manifest)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("make project option error"),
			kv("projectName", manifest.projectName),
			kv("projectPath", manifest.projectPath),
		)
	}
	script, err := makeProjectScript(context, manifest, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project script error"),
			kv("projectName", manifest.projectName),
			kv("projectPath", manifest.projectPath),
		)
	}
	config, err := makeProjectConfig(context, manifest, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project config error"),
			kv("projectName", manifest.projectName),
			kv("projectPath", manifest.projectPath),
		)
	}
	project = &Project{
		Name:    manifest.projectName,
		Path:    manifest.projectPath,
		context: context,
		option:  option,
		script:  script,
		config:  config,
	}
	return project, nil
}

func (p *Project) getImportContainer(scope projectImportScope) *projectImportContainer {
	if scope == projectImportScopeScript {
		return p.script.ImportContainer
	} else if scope == projectImportScopeConfig {
		return p.config.ImportContainer
	} else {
		impossible()
	}
	return nil
}

func (p *Project) loadImports(scope projectImportScope) error {
	return p.getImportContainer(scope).loadImports()
}

func (p *Project) loadConfigSources() error {
	return p.config.SourceContainer.loadSources()
}

func (p *Project) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool) ([]string, error) {
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
