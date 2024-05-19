package dsh_core

import (
	"text/template"
)

type project struct {
	context  *appContext
	Manifest *projectManifest
	Script   *projectScript
	Config   *projectConfig
}

func loadProject(context *appContext, manifest *projectManifest) (p *project, err error) {
	context.logger.InfoDesc("load project", kv("name", manifest.Name))
	err = context.Option.loadProjectOptions(manifest)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project options error"),
			kv("projectName", manifest.Name),
			kv("projectPath", manifest.projectPath),
		)
	}
	script, err := loadProjectScript(context, manifest)
	if err != nil {
		return nil, errW(err, "load project error",
			reason("load project script error"),
			kv("projectName", manifest.Name),
			kv("projectPath", manifest.projectPath),
		)
	}
	config, err := loadProjectConfig(context, manifest)
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

func (p *project) makeScripts(configs map[string]any, funcs template.FuncMap, outputPath string, useHardLink bool) ([]string, error) {
	data := map[string]any{
		"options": p.context.Option.getProjectOptions(p.Manifest),
		"configs": configs,
	}
	targetNames, err := p.Script.SourceContainer.makeSources(data, funcs, outputPath, useHardLink)
	if err != nil {
		return nil, errW(err, "make scripts error",
			reason("make sources error"),
			kv("projectName", p.Manifest.Name),
			kv("projectPath", p.Manifest.projectPath),
		)
	}
	return targetNames, nil
}
