package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"time"
)

type App struct {
	context               *appContext
	project               *project
	scriptImportContainer *appImportContainer
	configImportContainer *appImportContainer
	configs               map[string]any
	configsMade           bool
}

func loadApp(workspace *Workspace, manifest *projectManifest, options map[string]string) (app *App, err error) {
	workspace.logger.InfoDesc("load app", kv("name", manifest.Name))
	option, err := loadAppOption(manifest, options)
	if err != nil {
		return nil, errW(err, "load app error",
			reason("load app option error"),
			kv("projectName", manifest.Name),
			kv("projectPath", manifest.projectPath),
			kv("options", options),
		)
	}
	c := newAppContext(workspace, option)
	p, err := c.loadProject(manifest)
	if err != nil {
		return nil, errW(err, "load app error",
			reason("load project error"),
			kv("projectName", manifest.Name),
			kv("projectPath", manifest.projectPath),
		)
	}
	app = &App{
		context:               c,
		project:               p,
		scriptImportContainer: newAppImportContainer(p, projectImportScopeScript),
		configImportContainer: newAppImportContainer(p, projectImportScopeConfig),
	}
	return app, nil
}

func (a *App) MakeConfigs() (map[string]any, error) {
	if a.configsMade {
		return a.configs, nil
	}

	startTime := time.Now()
	a.context.logger.Info("make configs start")

	configs, err := a.configImportContainer.makeConfigs()
	if err != nil {
		return nil, err
	}

	a.configs = configs
	a.configsMade = true

	a.context.logger.InfoDesc("make configs finish", kv("elapsed", time.Since(startTime)))
	return a.configs, nil
}

func (a *App) MakeScripts(outputPath string) (artifact *AppArtifact, err error) {
	configs, err := a.MakeConfigs()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	a.context.logger.Info("make scripts start")
	if outputPath == "" {
		outputPath = filepath.Join(a.project.manifest.projectPath, "output")
		// TODO: build to workspace path
		//outputPath = filepath.Join(project.context.workspace.path, "output", project.manifest.Name)
	}
	funcs := newTemplateFuncs()

	if err = dsh_utils.RemakeDir(outputPath); err != nil {
		return nil, err
	}

	targetNames, err := a.scriptImportContainer.makeScripts(configs, funcs, outputPath)
	if err != nil {
		return nil, err
	}

	a.context.logger.InfoDesc("make scripts finish", kv("elapsed", time.Since(startTime)))
	return newAppArtifact(a, targetNames, outputPath), nil
}
