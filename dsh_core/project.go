package dsh_core

import (
	"fmt"
	"path/filepath"
	"text/template"
)

type Project struct {
	Workspace *Workspace
	Path      string
	Name      string
	Script    *Script
	Config    *Config
}

func NewProject(workspace *Workspace, projectPath string) *Project {
	return &Project{
		Workspace: workspace,
		Path:      projectPath,
		Script: &Script{
			SourceContainer: NewScriptSourceContainer(),
			ImportContainer: NewShallowImportContainer(ImportScopeScript),
		},
		Config: &Config{
			SourceContainer: NewConfigSourceContainer(),
			ImportContainer: NewShallowImportContainer(ImportScopeConfig),
		},
	}
}

func (project *Project) ScanScriptSources(sourceDir string, includeFiles []string) error {
	return project.Script.SourceContainer.ScanSources(sourceDir, includeFiles)
}

func (project *Project) ScanConfigSources(sourceDir string, includeFiles []string) error {
	return project.Config.SourceContainer.ScanSources(sourceDir, includeFiles)
}

func (project *Project) GetImportContainer(scope ImportScope) *ShallowImportContainer {
	if scope == ImportScopeScript {
		return project.Script.ImportContainer
	} else if scope == ImportScopeConfig {
		return project.Config.ImportContainer
	}
	panic(fmt.Sprintf("invalid import scope [%s]", scope))
}

func (project *Project) AddLocalImport(scope ImportScope, path string) error {
	return project.GetImportContainer(scope).AddLocalImport(project, path)
}

func (project *Project) AddGitImport(scope ImportScope, rawUrl string, rawRef string) error {
	return project.GetImportContainer(scope).AddGitImport(project, rawUrl, rawRef)
}

func (project *Project) LoadImports(scope ImportScope) error {
	return project.GetImportContainer(scope).LoadImports(project.Workspace)
}

func (project *Project) BuildScriptSources(config map[string]interface{}, funcs template.FuncMap, outputPath string) error {
	projectOutputPath := filepath.Join(outputPath, project.Name)
	return project.Script.SourceContainer.BuildSources(config, funcs, projectOutputPath)
}

func (project *Project) LoadConfigSources() error {
	return project.Config.SourceContainer.LoadSources()
}

func (project *Project) NewBuilder() *Builder {
	return NewBuilder(project)
}
