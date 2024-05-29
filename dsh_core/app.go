package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"time"
)

type App struct {
	context               *appContext
	project               *Project
	scriptImportContainer *appImportContainer
	configImportContainer *appImportContainer
	configs               map[string]any
	configsMade           bool
}

type AppMakeScriptsSettings struct {
	OutputPath      string
	OutputPathClear bool
	UseHardLink     bool
}

func newApp(context *appContext, project *Project) (app *App, err error) {
	app = &App{
		context:               context,
		project:               project,
		scriptImportContainer: newAppImportContainer(project, projectImportScopeScript),
		configImportContainer: newAppImportContainer(project, projectImportScopeConfig),
	}
	return app, nil
}

func (a *App) DescExtraKeyValues() KVS {
	return KVS{
		kv("context", a.context),
		kv("project", a.project),
		kv("scriptImportContainer", a.scriptImportContainer),
		kv("configImportContainer", a.configImportContainer),
	}
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

func (a *App) MakeScripts(settings AppMakeScriptsSettings) (artifact *AppArtifact, err error) {
	configs, err := a.MakeConfigs()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	a.context.logger.Info("make scripts start")
	outputPath := settings.OutputPath
	if outputPath == "" {
		outputPath, err = a.context.workspace.makeOutputDir(a.project.Name)
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
				kv("path", outputPath),
			)
		}
		outputPath = absPath
		if settings.OutputPathClear {
			if err = dsh_utils.ClearDir(outputPath); err != nil {
				return nil, errW(err, "make scripts error",
					reason("clear output dir error"),
					kv("path", outputPath),
				)
			}
		}
	}

	evaluator := a.context.evaluator.SetData("configs", configs).MergeFuncs(newProjectScriptTemplateFuncs())
	targetNames, err := a.scriptImportContainer.makeScripts(evaluator, outputPath, settings.UseHardLink)
	if err != nil {
		return nil, err
	}

	a.context.logger.InfoDesc("make scripts finish", kv("elapsed", time.Since(startTime)))
	return newAppArtifact(a, targetNames, outputPath), nil
}
