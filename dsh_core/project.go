package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
	"time"
)

type Project struct {
	context               *projectContext
	manifest              *projectManifest
	instance              *projectInstance
	scriptImportContainer *projectInstanceImportDeepContainer
	configImportContainer *projectInstanceImportDeepContainer
	config                map[string]any
	configMade            bool
}

func openProject(workspace *Workspace, manifest *projectManifest, optionValues map[string]string) (*Project, error) {
	context := newProjectContext(workspace, workspace.logger)
	context.logger.Info("open project: name=%s", manifest.Name)
	instance, err := context.newProjectInstance(manifest, optionValues)
	if err != nil {
		return nil, err
	}
	project := &Project{
		context:               context,
		manifest:              manifest,
		instance:              instance,
		scriptImportContainer: newProjectInstanceImportDeepContainer(instance, projectInstanceImportScopeScript),
		configImportContainer: newProjectInstanceImportDeepContainer(instance, projectInstanceImportScopeConfig),
	}
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
	project.context.logger.Info("make config start")

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
	project.context.logger.Info("make config finish: elapsed=%s", time.Since(startTime))
	return project.config, nil
}

func (project *Project) MakeScript(outputPath string) (err error) {
	startTime := time.Now()
	project.context.logger.Info("make script start")
	if outputPath == "" {
		outputPath = filepath.Join(project.instance.manifest.projectPath, "output")
		// TODO: build to workspace path
		// outputPath = filepath.Join(project.ProjectInfo.workspace.path, "output", project.ProjectInfo.name)
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

	project.context.logger.Info("make script finish: elapsed=%s", time.Since(startTime))
	return nil
}
