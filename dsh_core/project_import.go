package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
	"slices"
	"text/template"
)

type Import struct {
	Type            ImportType
	Unique          string
	FromProjectName string
	ProjectPath     string
	GitRawUrl       string
	GitParsedUrl    *url.URL
	GitRawRef       string
	GitParsedRef    *GitRef
	Project         *Project
}

type ImportScope string

const (
	ImportScopeScript ImportScope = "Script"
	ImportScopeConfig ImportScope = "Config"
)

type ImportType string

const (
	ImportTypeLocal ImportType = "Local"
	ImportTypeGit   ImportType = "Git"
)

type ShallowImportContainer struct {
	Scope           ImportScope
	ImportUniqueMap map[string]*Import
	Imports         []*Import
	ImportsLoaded   bool
}

type DeepImportContainer struct {
	Scope         ImportScope
	Project       *Project
	Imports       []*Import
	ImportsLoaded bool
}

func NewLocalImport(fromProjectName string, projectPath string) *Import {
	return &Import{
		Type:            ImportTypeLocal,
		Unique:          projectPath,
		FromProjectName: fromProjectName,
		ProjectPath:     projectPath,
	}
}

func NewGitImport(fromProjectName string, projectPath string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *GitRef) *Import {
	return &Import{
		Type:            ImportTypeGit,
		Unique:          projectPath,
		FromProjectName: fromProjectName,
		ProjectPath:     projectPath,
		GitRawUrl:       rawUrl,
		GitParsedUrl:    parsedUrl,
		GitRawRef:       rawRef,
		GitParsedRef:    parsedRef,
	}
}

func (imp *Import) LoadProject(workspace *Workspace) error {
	if imp.Project == nil {
		if imp.Type == ImportTypeLocal {
			project, err := workspace.LoadLocalProject(imp.ProjectPath)
			if err != nil {
				return dsh_utils.WrapError(err, "import local project load failed", map[string]interface{}{
					"projectPath": imp.ProjectPath,
				})
			}
			imp.Project = project
		} else {
			project, err := workspace.LoadGitProject(imp.ProjectPath, imp.GitRawUrl, imp.GitParsedUrl, imp.GitRawRef, imp.GitParsedRef)
			if err != nil {
				return dsh_utils.WrapError(err, "import git project load failed", map[string]interface{}{
					"projectPath": imp.ProjectPath,
					"gitUrl":      imp.GitRawUrl,
					"gitRef":      imp.GitRawRef,
				})
			}
			imp.Project = project
		}
	}
	return nil
}

func NewShallowImportContainer(scope ImportScope) *ShallowImportContainer {
	return &ShallowImportContainer{
		Scope:           scope,
		ImportUniqueMap: make(map[string]*Import),
	}
}

func (sic *ShallowImportContainer) AddLocalImport(project *Project, path string) (err error) {
	if !dsh_utils.IsDirExists(path) {
		return dsh_utils.NewError("import local project dir not found", map[string]interface{}{
			"path": path,
		})
	}
	importProjectPath, err := filepath.Abs(path)
	if err != nil {
		return dsh_utils.WrapError(err, "import load project abs-path get failed", map[string]interface{}{
			"path": path,
		})
	}
	if importProjectPath == project.Path {
		return nil
	}
	imp := NewLocalImport(project.Name, importProjectPath)
	if _, exist := sic.ImportUniqueMap[imp.Unique]; !exist {
		sic.ImportUniqueMap[imp.Unique] = imp
		sic.Imports = append(sic.Imports, imp)
	}
	return nil
}

func (sic *ShallowImportContainer) AddGitImport(project *Project, rawUrl string, rawRef string) error {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return dsh_utils.WrapError(err, "import git project url parse failed", map[string]interface{}{
			"url": rawUrl,
		})
	}
	parsedRef := ParseGitRef(rawRef)
	projectPath := project.Workspace.GetGitProjectPath(parsedUrl, parsedRef)
	if projectPath == project.Path {
		return nil
	}
	imp := NewGitImport(project.Name, projectPath, rawUrl, parsedUrl, rawRef, parsedRef)
	if _, exist := sic.ImportUniqueMap[imp.Unique]; !exist {
		sic.ImportUniqueMap[imp.Unique] = imp
		sic.Imports = append(sic.Imports, imp)
	}
	return nil
}

func (sic *ShallowImportContainer) LoadImports(workspace *Workspace) (err error) {
	if sic.ImportsLoaded {
		return nil
	}

	for i := 0; i < len(sic.Imports); i++ {
		if err = sic.Imports[i].LoadProject(workspace); err != nil {
			return err
		}
	}

	sic.ImportsLoaded = true
	return nil
}

func NewDeepImportContainer(scope ImportScope, project *Project) *DeepImportContainer {
	return &DeepImportContainer{
		Scope:   scope,
		Project: project,
	}
}

func (dic *DeepImportContainer) LoadImports() (err error) {
	if dic.ImportsLoaded {
		return nil
	}
	if err = dic.Project.LoadImports(dic.Scope); err != nil {
		return err
	}

	var deepImports []*Import
	var deepImportMap = make(map[string]*Import)

	sic := dic.Project.GetImportContainer(dic.Scope)
	for i := 0; i < len(sic.Imports); i++ {
		imp := sic.Imports[i]
		deepImports = append(deepImports, imp)
		deepImportMap[imp.Unique] = imp
	}

	scanningImports := sic.Imports
	for i := 0; i < len(scanningImports); i++ {
		imp1 := scanningImports[i]
		if err = imp1.Project.LoadImports(sic.Scope); err != nil {
			return err
		}
		sic1 := imp1.Project.GetImportContainer(sic.Scope)
		for j := 0; j < len(sic1.Imports); j++ {
			imp2 := sic1.Imports[j]
			if imp2.ProjectPath == dic.Project.Path {
				continue
			}
			if _, exist := deepImportMap[imp2.Unique]; !exist {
				deepImports = append(deepImports, imp2)
				deepImportMap[imp2.Unique] = imp2
				scanningImports = append(scanningImports, imp2)
			}
		}
	}

	dic.Imports = deepImports
	dic.ImportsLoaded = true
	return nil
}

func (dic *DeepImportContainer) BuildScriptSources(config map[string]interface{}, funcs template.FuncMap, outputPath string) (err error) {
	if dic.Scope != ImportScopeScript {
		panic("DeepImportContainer.BuildScriptSources() only support ImportScopeScript")
	}
	if err = dic.LoadImports(); err != nil {
		return err
	}
	for i := 0; i < len(dic.Imports); i++ {
		if err = dic.Imports[i].Project.BuildScriptSources(config, funcs, outputPath); err != nil {
			return err
		}
	}
	if err = dic.Project.BuildScriptSources(config, funcs, outputPath); err != nil {
		return err
	}
	return nil
}

func (dic *DeepImportContainer) LoadConfigSources() (sources []*ConfigSource, err error) {
	if dic.Scope != ImportScopeConfig {
		panic("DeepImportContainer.LoadConfigSources() only support ImportScopeConfig")
	}
	if err = dic.LoadImports(); err != nil {
		return nil, err
	}
	for i := 0; i < len(dic.Imports); i++ {
		if err = dic.Imports[i].Project.LoadConfigSources(); err != nil {
			return nil, err
		}
	}
	if err = dic.Project.LoadConfigSources(); err != nil {
		return nil, err
	}

	for i := 0; i < len(dic.Imports); i++ {
		for j := 0; j < len(dic.Imports[i].Project.Config.SourceContainer.YamlSources); j++ {
			source := dic.Imports[i].Project.Config.SourceContainer.YamlSources[j]
			sources = append(sources, source)
		}
	}
	for i := 0; i < len(dic.Project.Config.SourceContainer.YamlSources); i++ {
		source := dic.Project.Config.SourceContainer.YamlSources[i]
		sources = append(sources, source)
	}

	slices.SortStableFunc(sources, func(a, b *ConfigSource) int {
		rst := a.Content.Order - b.Content.Order
		if rst < 0 {
			return 1
		} else if rst > 0 {
			return -1
		} else {
			return 0
		}
	})

	return sources, nil
}
