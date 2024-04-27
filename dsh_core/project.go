package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
	"time"
)

type Project struct {
	context               *Context
	info                  *projectInfo
	instance              *projectInstance
	scriptImportContainer *projectInstanceImportDeepContainer
	configImportContainer *projectInstanceImportDeepContainer
	config                map[string]any
	configMade            bool
}

func openProject(context *Context, info *projectInfo) (*Project, error) {
	context.Logger.Info("open project: name=%s", info.name)
	if context.Project != nil {
		return nil, dsh_utils.NewError("context already open project", map[string]any{
			"projectPath": context.Project.info.path,
		})
	}
	instance, err := context.newProjectInstance(info)
	if err != nil {
		return nil, err
	}
	project := &Project{
		context:               context,
		info:                  info,
		instance:              instance,
		scriptImportContainer: newProjectInstanceImportDeepContainer(instance, projectInstanceImportScopeScript),
		configImportContainer: newProjectInstanceImportDeepContainer(instance, projectInstanceImportScopeConfig),
	}
	context.Project = project
	return project, nil
}

func (project *Project) getImportContainer(scope projectInstanceImportScope) *projectInstanceImportDeepContainer {
	if scope == projectInstanceImportScopeScript {
		return project.scriptImportContainer
	} else if scope == projectInstanceImportScopeConfig {
		return project.configImportContainer
	}
	panic(fmt.Sprintf("invalid import scope: scope=%s", scope))
	return nil
}

func (project *Project) loadImports(scope projectInstanceImportScope) (err error) {
	return project.getImportContainer(scope).loadImports()
}

func (project *Project) MakeConfig() (map[string]any, error) {
	if project.configMade {
		return project.config, nil
	}

	startTime := time.Now()
	project.context.Logger.Info("make config start")

	sources, err := project.configImportContainer.loadConfigSources()
	if err != nil {
		return nil, err
	}

	config := make(map[string]any)

	for i := 0; i < len(sources); i++ {
		source := sources[i]
		source.content.merge(config)
	}

	project.config = config
	project.configMade = true
	project.context.Logger.Info("make config finish: elapsed=%s", time.Since(startTime))
	return project.config, nil
}

func (project *Project) MakeScript(outputPath string) (err error) {
	startTime := time.Now()
	project.context.Logger.Info("make script start")
	if outputPath == "" {
		outputPath = filepath.Join(project.instance.info.path, "output")
		// TODO: build to workspace path
		// outputPath = filepath.Join(project.ProjectInfo.Workspace.path, "output", project.ProjectInfo.name)
	}

	config, err := project.MakeConfig()
	if err != nil {
		return err
	}
	funcs := newTemplateFuncs()

	if err = dsh_utils.RemakeDir(outputPath); err != nil {
		return err
	}

	if err = project.scriptImportContainer.makeScript(config, funcs, outputPath); err != nil {
		return err
	}

	project.context.Logger.Info("make script finish: elapsed=%s", time.Since(startTime))
	return nil
}
