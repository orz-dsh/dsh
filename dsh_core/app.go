package dsh_core

import (
	"dsh/dsh_utils"
	"os"
	"path/filepath"
	"time"
)

type App struct {
	context                *appContext
	mainProject            *projectInstance
	extraProjects          []*projectInstance
	scriptProjectContainer *projectInstanceContainer
	configProjectContainer *projectInstanceContainer
	configs                map[string]any
	configTraces           map[string]any
	configsMade            bool
}

type AppMakeScriptsSettings struct {
	OutputPath      string
	OutputPathClear bool
	UseHardLink     bool
	Inspection      bool
}

func makeApp(context *appContext, mainProjectEntity *projectSetting, extraProjectEntities []*projectSetting) (*App, error) {
	mainProject, err := context.loadProject(mainProjectEntity)
	if err != nil {
		return nil, err
	}

	var extraProjects []*projectInstance
	for i := 0; i < len(extraProjectEntities); i++ {
		extraProject, err := newProjectInstance(context, extraProjectEntities[i])
		if err != nil {
			return nil, err
		}
		extraProjects = append(extraProjects, extraProject)
	}

	app := &App{
		context:                context,
		mainProject:            mainProject,
		extraProjects:          extraProjects,
		scriptProjectContainer: newProjectInstanceContainer(mainProject, extraProjects, projectImportScopeScript),
		configProjectContainer: newProjectInstanceContainer(mainProject, extraProjects, projectImportScopeConfig),
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

func (a *App) MakeConfigs() (map[string]any, map[string]any, error) {
	if a.configsMade {
		return a.configs, a.configTraces, nil
	}

	startTime := time.Now()
	a.context.logger.Info("make configs start")

	configs, configTraces, err := a.configProjectContainer.makeConfigs()
	if err != nil {
		return nil, nil, err
	}

	a.configs = configs
	a.configTraces = configTraces
	a.configsMade = true

	a.context.logger.InfoDesc("make configs finish", kv("elapsed", time.Since(startTime)))
	return a.configs, a.configTraces, nil
}

func (a *App) MakeScripts(settings AppMakeScriptsSettings) (artifact *AppArtifact, err error) {
	configs, configTraces, err := a.MakeConfigs()
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

	if settings.Inspection {
		inspectionPath := filepath.Join(outputPath, "@inspection")
		if err = os.MkdirAll(inspectionPath, os.ModePerm); err != nil {
			return nil, errW(err, "make scripts error",
				reason("make inspection dir error"),
				kv("path", inspectionPath),
			)
		}
		configsInspectionPath := filepath.Join(inspectionPath, "configs.yml")
		if err = dsh_utils.WriteYamlFile(configsInspectionPath, configs); err != nil {
			return nil, errW(err, "make scripts error",
				reason("write configs inspection file error"),
				kv("path", configsInspectionPath),
			)
		}
		configTracesInspectionPath := filepath.Join(inspectionPath, "config-traces.yml")
		if err = dsh_utils.WriteYamlFile(configTracesInspectionPath, configTraces); err != nil {
			return nil, errW(err, "make scripts error",
				reason("write config traces inspection file error"),
				kv("path", configTracesInspectionPath),
			)
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
