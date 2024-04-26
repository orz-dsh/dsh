package dsh_core

import (
	"dsh/dsh_utils"
	"fmt"
	"path/filepath"
	"time"
)

type Builder struct {
	Project               *Project
	ScriptImportContainer *DeepImportContainer
	ConfigImportContainer *DeepImportContainer
	Config                map[string]interface{}
	ConfigMade            bool
}

func NewBuilder(project *Project) *Builder {
	return &Builder{
		Project:               project,
		ScriptImportContainer: NewDeepImportContainer(ImportScopeScript, project),
		ConfigImportContainer: NewDeepImportContainer(ImportScopeConfig, project),
	}
}

func (builder *Builder) GetImportContainer(scope ImportScope) *DeepImportContainer {
	if scope == ImportScopeScript {
		return builder.ScriptImportContainer
	} else if scope == ImportScopeConfig {
		return builder.ConfigImportContainer
	}
	panic(fmt.Sprintf("invalid import scope [%s]", scope))
}

func (builder *Builder) LoadImports(scope ImportScope) (err error) {
	return builder.GetImportContainer(scope).LoadImports()
}

func (builder *Builder) MakeConfig() (map[string]interface{}, error) {
	if builder.ConfigMade {
		return builder.Config, nil
	}

	startTime := time.Now()
	builder.Project.Workspace.Logger.Info("make config start")

	sources, err := builder.ConfigImportContainer.LoadConfigSources()
	if err != nil {
		return nil, err
	}

	config := make(map[string]interface{})

	for i := 0; i < len(sources); i++ {
		source := sources[i]
		source.Content.Merge(config)
	}

	builder.Config = config
	builder.ConfigMade = true
	builder.Project.Workspace.Logger.Info("make config finish: elapsed=%s", time.Since(startTime))
	return builder.Config, nil
}

func (builder *Builder) Build(outputPath string) (err error) {
	startTime := time.Now()
	builder.Project.Workspace.Logger.Info("build start")
	if outputPath == "" {
		outputPath = filepath.Join(builder.Project.Path, "output")
		// TODO: build to workspace path
		// outputPath = filepath.Join(builder.Project.Workspace.Path, "output", builder.Project.Name)
	}

	config, err := builder.MakeConfig()
	if err != nil {
		return err
	}
	funcs := NewTemplateFuncs()

	if err = dsh_utils.RemakeDir(outputPath); err != nil {
		return err
	}

	if err = builder.ScriptImportContainer.BuildScriptSources(config, funcs, outputPath); err != nil {
		return err
	}

	builder.Project.Workspace.Logger.Info("build finish: elapsed=%s", time.Since(startTime))
	return nil
}
