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

func (app *App) MakeConfigs() (map[string]any, error) {
	if app.configsMade {
		return app.configs, nil
	}

	startTime := time.Now()
	app.context.logger.Info("make configs start")

	configs, err := app.configImportContainer.makeConfigs()
	if err != nil {
		return nil, err
	}

	app.configs = configs
	app.configsMade = true

	app.context.logger.InfoDesc("make configs finish", kv("elapsed", time.Since(startTime)))
	return app.configs, nil
}

func (app *App) MakeScripts(outputPath string) (err error) {
	configs, err := app.MakeConfigs()
	if err != nil {
		return err
	}

	startTime := time.Now()
	app.context.logger.Info("make scripts start")
	if outputPath == "" {
		outputPath = filepath.Join(app.project.manifest.projectPath, "output")
		// TODO: build to workspace path
		//outputPath = filepath.Join(project.context.workspace.path, "output", project.manifest.Name)
	}
	funcs := newTemplateFuncs()

	if err = dsh_utils.RemakeDir(outputPath); err != nil {
		return err
	}

	if err = app.scriptImportContainer.makeScripts(configs, funcs, outputPath); err != nil {
		return err
	}

	app.context.logger.InfoDesc("make scripts finish", kv("elapsed", time.Since(startTime)))
	return nil
}
