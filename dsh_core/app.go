package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"time"
)

type App struct {
	context                *appContext
	mainProject            *appProject
	extraProjects          []*appProject
	scriptProjectContainer *appProjectContainer
	configProjectContainer *appProjectContainer
	configs                map[string]any
	configsMade            bool
}

type AppMakeScriptsSettings struct {
	OutputPath      string
	OutputPathClear bool
	UseHardLink     bool
}

func makeApp(context *appContext, mainProjectEntity *projectEntity, extraProjectEntities []*projectEntity) (*App, error) {
	mainProject, err := context.loadProject(mainProjectEntity)
	if err != nil {
		return nil, err
	}

	var extraProjects []*appProject
	for i := 0; i < len(extraProjectEntities); i++ {
		extraProject, err := makeAppProject(context, extraProjectEntities[i])
		if err != nil {
			return nil, err
		}
		extraProjects = append(extraProjects, extraProject)
	}

	app := &App{
		context:                context,
		mainProject:            mainProject,
		extraProjects:          extraProjects,
		scriptProjectContainer: newAppProjectContainer(mainProject, extraProjects, projectImportScopeScript),
		configProjectContainer: newAppProjectContainer(mainProject, extraProjects, projectImportScopeConfig),
	}
	return app, nil
}

func (a *App) DescExtraKeyValues() KVS {
	return KVS{
		kv("context", a.context),
		kv("mainProject", a.mainProject),
		kv("extraProjects", a.extraProjects),
		kv("scriptProjectContainer", a.scriptProjectContainer),
		kv("configProjectContainer", a.configProjectContainer),
	}
}

func (a *App) MakeConfigs() (map[string]any, error) {
	if a.configsMade {
		return a.configs, nil
	}

	startTime := time.Now()
	a.context.logger.Info("make configs start")

	configs, err := a.configProjectContainer.makeConfigs()
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
		outputPath, err = a.context.workspace.makeOutputDir(a.mainProject.Name)
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
	targetNames, err := a.scriptProjectContainer.makeScripts(evaluator, outputPath, settings.UseHardLink)
	if err != nil {
		return nil, err
	}

	a.context.logger.InfoDesc("make scripts finish", kv("elapsed", time.Since(startTime)))
	return newAppArtifact(a, targetNames, outputPath), nil
}
