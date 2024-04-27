package dsh_core

import (
	"dsh/dsh_utils"
	"path/filepath"
	"time"
)

type Project struct {
	Context               *Context
	Info                  *ProjectInfo
	Instance              *ProjectInstance
	ScriptImportContainer *ProjectInstanceImportDeepContainer
	ConfigImportContainer *ProjectInstanceImportDeepContainer
	Config                map[string]any
	ConfigMade            bool
}

func OpenProject(context *Context, info *ProjectInfo) (*Project, error) {
	if context.Project != nil {
		return nil, dsh_utils.NewError("context already open project", map[string]any{
			"projectPath": context.Project.Info.Path,
		})
	}
	instance, err := NewProjectInstance(context, info)
	if err != nil {
		return nil, err
	}
	return &Project{
		Context:               context,
		Info:                  info,
		Instance:              instance,
		ScriptImportContainer: NewDeepImportContainer(instance, ProjectInstanceImportScopeScript),
		ConfigImportContainer: NewDeepImportContainer(instance, ProjectInstanceImportScopeConfig),
	}, nil
}

func (project *Project) GetImportContainer(scope ProjectInstanceImportScope) *ProjectInstanceImportDeepContainer {
	if scope == ProjectInstanceImportScopeScript {
		return project.ScriptImportContainer
	} else if scope == ProjectInstanceImportScopeConfig {
		return project.ConfigImportContainer
	}
	project.Context.Logger.Panic("invalid import scope: scope=%s", scope)
	return nil
}

func (project *Project) LoadImports(scope ProjectInstanceImportScope) (err error) {
	return project.GetImportContainer(scope).LoadImports()
}

func (project *Project) MakeConfig() (map[string]any, error) {
	if project.ConfigMade {
		return project.Config, nil
	}

	startTime := time.Now()
	project.Context.Logger.Info("make config start")

	sources, err := project.ConfigImportContainer.LoadConfigSources()
	if err != nil {
		return nil, err
	}

	config := make(map[string]any)

	for i := 0; i < len(sources); i++ {
		source := sources[i]
		source.Content.Merge(config)
	}

	project.Config = config
	project.ConfigMade = true
	project.Context.Logger.Info("make config finish: elapsed=%s", time.Since(startTime))
	return project.Config, nil
}

func (project *Project) Build(outputPath string) (err error) {
	startTime := time.Now()
	project.Context.Logger.Info("build start")
	if outputPath == "" {
		outputPath = filepath.Join(project.Instance.Info.Path, "output")
		// TODO: build to workspace path
		// outputPath = filepath.Join(project.ProjectInfo.Workspace.Path, "output", project.ProjectInfo.Name)
	}

	config, err := project.MakeConfig()
	if err != nil {
		return err
	}
	funcs := NewTemplateFuncs()

	if err = dsh_utils.RemakeDir(outputPath); err != nil {
		return err
	}

	if err = project.ScriptImportContainer.BuildScriptSources(config, funcs, outputPath); err != nil {
		return err
	}

	project.Context.Logger.Info("build finish: elapsed=%s", time.Since(startTime))
	return nil
}
