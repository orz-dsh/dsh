package dsh_core

import (
	"path/filepath"
	"text/template"
)

type project struct {
	context  *appContext
	manifest *projectManifest
	script   *projectScript
	config   *projectConfig
}

func loadProject(context *appContext, manifest *projectManifest) (p *project, err error) {
	context.logger.InfoDesc("load project", kv("name", manifest.Name))
	err = context.option.loadProjectOptions(manifest)
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
		manifest: manifest,
		script:   script,
		config:   config,
	}
	return p, nil
}

func (p *project) getImportContainer(scope projectImportScope) *projectImportContainer {
	if scope == projectImportScopeScript {
		return p.script.importContainer
	} else if scope == projectImportScopeConfig {
		return p.config.importContainer
	}
	panic(desc("invalid import scope", kv("scope", scope)))
}

func (p *project) loadImports(scope projectImportScope) error {
	return p.getImportContainer(scope).loadImports()
}

func (p *project) loadConfigSources() error {
	return p.config.sourceContainer.loadSources()
}

func (p *project) makeScripts(configs map[string]any, funcs template.FuncMap, outputPath string) error {
	env := map[string]any{
		"options": p.context.option.getProjectOptions(p.manifest),
		"configs": configs,
	}
	err := p.script.sourceContainer.makeSources(env, funcs, filepath.Join(outputPath, p.manifest.Name))
	if err != nil {
		return errW(err, "make scripts error",
			reason("make sources error"),
			kv("projectName", p.manifest.Name),
			kv("projectPath", p.manifest.projectPath),
		)
	}
	return nil
}
