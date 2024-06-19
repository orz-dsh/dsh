package dsh_core

import (
	"dsh/dsh_utils"
	"os"
	"path/filepath"
	"time"
)

type App struct {
	context          *appContext
	mainProjectName  string
	projectContainer *projectInstanceContainer
	configs          map[string]any
	configsTraces    map[string]any
	configsMade      bool
}

type AppMakeScriptsSettings struct {
	OutputPath      string
	OutputPathClear bool
	UseHardLink     bool
	Inspection      bool
}

func newApp(context *appContext, mainSetting *projectSetting, extraSettings []*projectSetting) *App {
	return &App{
		context:          context,
		mainProjectName:  mainSetting.Name,
		projectContainer: newProjectInstanceContainerTest(context, mainSetting, extraSettings),
	}
}

func (a *App) DescExtraKeyValues() KVS {
	return KVS{
		kv("context", a.context),
		kv("mainProjectName", a.mainProjectName),
		kv("projectContainer", a.projectContainer),
	}
}

func (a *App) MakeConfigs() (map[string]any, map[string]any, error) {
	if a.configsMade {
		return a.configs, a.configsTraces, nil
	}

	startTime := time.Now()
	a.context.logger.Info("make configs start")

	configs, configsTraces, err := a.projectContainer.makeConfigs()
	if err != nil {
		return nil, nil, err
	}

	a.configs = configs
	a.configsTraces = configsTraces
	a.configsMade = true

	a.context.logger.InfoDesc("make configs finish", kv("elapsed", time.Since(startTime)))
	return a.configs, a.configsTraces, nil
}

func (a *App) MakeScripts(settings AppMakeScriptsSettings) (artifact *AppArtifact, err error) {
	configs, configsTraces, err := a.MakeConfigs()
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	a.context.logger.Info("make scripts start")
	outputPath := settings.OutputPath
	if outputPath == "" {
		outputPath, err = a.context.workspace.makeOutputDir(a.mainProjectName)
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

	inspectionPath := ""
	if settings.Inspection {
		inspectionPath = filepath.Join(outputPath, "@inspection")
		if err = os.MkdirAll(inspectionPath, os.ModePerm); err != nil {
			return nil, errW(err, "make scripts error",
				reason("make inspection dir error"),
				kv("path", inspectionPath),
			)
		}
		configsTracesInspectionPath := filepath.Join(inspectionPath, "app.configs-traces.yml")
		if err = dsh_utils.WriteYamlFile(configsTracesInspectionPath, configsTraces); err != nil {
			return nil, errW(err, "make scripts error",
				reason("write configs traces inspection file error"),
				kv("path", configsTracesInspectionPath),
			)
		}
		dataInspectionPath := filepath.Join(inspectionPath, "app.environment.yml")
		if err = dsh_utils.WriteYamlFile(dataInspectionPath, evaluator.ToMap(false)); err != nil {
			return nil, errW(err, "make scripts error",
				reason("write data inspection file error"),
				kv("path", dataInspectionPath),
			)
		}
		optionInspectionPath := filepath.Join(inspectionPath, "app.option.yml")
		if err = dsh_utils.WriteYamlFile(optionInspectionPath, a.context.option.inspect()); err != nil {
			return nil, errW(err, "make scripts error",
				reason("write option inspection file error"),
				kv("path", optionInspectionPath),
			)
		}
		profileInspectionPath := filepath.Join(inspectionPath, "app.profile.yml")
		if err = dsh_utils.WriteYamlFile(profileInspectionPath, a.context.profile.inspect()); err != nil {
			return nil, errW(err, "make scripts error",
				reason("write profile inspection file error"),
				kv("path", profileInspectionPath),
			)
		}
	}

	targetNames, err := a.projectContainer.makeScripts(evaluator, outputPath, settings.UseHardLink, inspectionPath)
	if err != nil {
		return nil, err
	}

	a.context.logger.InfoDesc("make scripts finish", kv("elapsed", time.Since(startTime)))
	return newAppArtifact(a, targetNames, outputPath), nil
}
