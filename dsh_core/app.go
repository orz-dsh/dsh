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

type AppMakeScriptsOptions struct {
	OutputPath      string
	ClearOutputPath bool
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

func (a *App) MakeScripts(options AppMakeScriptsOptions) (artifact *AppArtifact, err error) {
	configs, err := a.MakeConfigs()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	a.context.logger.Info("make scripts start")
	outputPath := options.OutputPath
	if outputPath == "" {
		outputPath, err = a.context.workspace.makeOutputDir(a.project.manifest.Name)
		if err != nil {
			return nil, errW(err, "make scripts error",
				reason("make output path error"),
			)
		}
	} else {
		absPath, err := filepath.Abs(outputPath)
		if err != nil {
			return nil, errW(err, "make scripts error",
				reason("get abs-path error"),
				kv("outputPath", outputPath),
			)
		}
		outputPath = absPath
		if options.ClearOutputPath {
			if err = dsh_utils.RemakeDir(outputPath); err != nil {
				return nil, errW(err, "make scripts error",
					reason("clear output path error"),
					kv("outputPath", outputPath),
				)
			}
		}
	}
	funcs := newTemplateFuncs()

	targetNames, err := a.scriptImportContainer.makeScripts(configs, funcs, outputPath)
	if err != nil {
		return nil, err
	}

	a.context.logger.InfoDesc("make scripts finish", kv("elapsed", time.Since(startTime)))
	return newAppArtifact(a, targetNames, outputPath), nil
}
