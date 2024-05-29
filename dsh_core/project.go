package dsh_core

// region project

type project struct {
	context  *appContext
	Manifest *projectManifest
	Option   *projectOption
	Script   *projectScript
	Config   *projectConfig
}

func loadProject(context *appContext, manifest *projectManifest) (p *project, err error) {
	context.logger.InfoDesc("load project", kv("name", manifest.Name))
	option, err := makeProjectOption(context, manifest)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("make project option error"),
			kv("projectName", manifest.Name),
			kv("projectPath", manifest.projectPath),
		)
	}
	script, err := loadProjectScript(context, manifest, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project script error"),
			kv("projectName", manifest.Name),
			kv("projectPath", manifest.projectPath),
		)
	}
	config, err := loadProjectConfig(context, manifest, option)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project config error"),
			kv("projectName", manifest.Name),
			kv("projectPath", manifest.projectPath),
		)
	}
	p = &project{
		context:  context,
		Manifest: manifest,
		Option:   option,
		Script:   script,
		Config:   config,
	}
	return p, nil
}

func (p *project) getImportContainer(scope projectImportScope) *projectImportContainer {
	if scope == projectImportScopeScript {
		return p.Script.ImportContainer
	} else if scope == projectImportScopeConfig {
		return p.Config.ImportContainer
	} else {
		impossible()
	}
	return nil
}

func (p *project) loadImports(scope projectImportScope) error {
	return p.getImportContainer(scope).loadImports()
}

func (p *project) loadConfigSources() error {
	return p.Config.SourceContainer.loadSources()
}

func (p *project) makeScripts(evaluator *Evaluator, outputPath string, useHardLink bool) ([]string, error) {
	evaluator = evaluator.SetData("options", p.Option.Items)
	targetNames, err := p.Script.SourceContainer.makeSources(evaluator, outputPath, useHardLink)
	if err != nil {
		return nil, errW(err, "make scripts error",
			reason("make sources error"),
			kv("projectName", p.Manifest.Name),
			kv("projectPath", p.Manifest.projectPath),
		)
	}
	return targetNames, nil
}

// endregion
