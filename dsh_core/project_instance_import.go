package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
	"slices"
	"text/template"
)

type ProjectInstanceImport struct {
	Context      *Context
	Reference    *ProjectInfo
	Type         ProjectInstanceImportType
	Unique       string
	ProjectPath  string
	GitRawUrl    string
	GitParsedUrl *url.URL
	GitRawRef    string
	GitParsedRef *GitRef
	Instance     *ProjectInstance
}

type ProjectInstanceImportScope string

const (
	ProjectInstanceImportScopeScript ProjectInstanceImportScope = "Script"
	ProjectInstanceImportScopeConfig ProjectInstanceImportScope = "Config"
)

type ProjectInstanceImportType string

const (
	ProjectInstanceImportTypeLocal ProjectInstanceImportType = "Local"
	ProjectInstanceImportTypeGit   ProjectInstanceImportType = "Git"
)

type ProjectInstanceImportShallowContainer struct {
	Context         *Context
	Scope           ProjectInstanceImportScope
	ImportUniqueMap map[string]*ProjectInstanceImport
	Imports         []*ProjectInstanceImport
	ImportsLoaded   bool
}

type ProjectInstanceImportDeepContainer struct {
	Context       *Context
	Instance      *ProjectInstance
	Scope         ProjectInstanceImportScope
	Imports       []*ProjectInstanceImport
	ImportsLoaded bool
}

func NewProjectInstanceImportLocal(context *Context, reference *ProjectInfo, projectPath string) *ProjectInstanceImport {
	return &ProjectInstanceImport{
		Context:     context,
		Reference:   reference,
		Type:        ProjectInstanceImportTypeLocal,
		Unique:      projectPath,
		ProjectPath: projectPath,
	}
}

func NewProjectInstanceImportGit(context *Context, reference *ProjectInfo, projectPath string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *GitRef) *ProjectInstanceImport {
	return &ProjectInstanceImport{
		Context:      context,
		Reference:    reference,
		Type:         ProjectInstanceImportTypeGit,
		Unique:       projectPath,
		ProjectPath:  projectPath,
		GitRawUrl:    rawUrl,
		GitParsedUrl: parsedUrl,
		GitRawRef:    rawRef,
		GitParsedRef: parsedRef,
	}
}

func (imp *ProjectInstanceImport) LoadProject() error {
	if imp.Instance == nil {
		workspace := imp.Context.Workspace
		if imp.Type == ProjectInstanceImportTypeLocal {
			project, err := workspace.LoadLocalProjectInfo(imp.ProjectPath)
			if err != nil {
				return dsh_utils.WrapError(err, "import local project load failed", map[string]any{
					"projectPath": imp.ProjectPath,
				})
			}
			instance, err := NewProjectInstance(imp.Context, project)
			if err != nil {
				return dsh_utils.WrapError(err, "import local project open failed", map[string]any{
					"projectPath": imp.ProjectPath,
				})
			}
			imp.Instance = instance
		} else {
			project, err := workspace.LoadGitProjectInfo(imp.ProjectPath, imp.GitRawUrl, imp.GitParsedUrl, imp.GitRawRef, imp.GitParsedRef)
			if err != nil {
				return dsh_utils.WrapError(err, "import git project load failed", map[string]any{
					"projectPath": imp.ProjectPath,
					"gitUrl":      imp.GitRawUrl,
					"gitRef":      imp.GitRawRef,
				})
			}
			instance, err := NewProjectInstance(imp.Context, project)
			if err != nil {
				return dsh_utils.WrapError(err, "import git project open failed", map[string]any{
					"projectPath": imp.ProjectPath,
					"gitUrl":      imp.GitRawUrl,
					"gitRef":      imp.GitRawRef,
				})
			}
			imp.Instance = instance
		}
	}
	return nil
}

func NewShallowImportContainer(context *Context, scope ProjectInstanceImportScope) *ProjectInstanceImportShallowContainer {
	return &ProjectInstanceImportShallowContainer{
		Context:         context,
		Scope:           scope,
		ImportUniqueMap: make(map[string]*ProjectInstanceImport),
	}
}

func (container *ProjectInstanceImportShallowContainer) ImportLocal(context *Context, path string, reference *ProjectInfo) (err error) {
	if !dsh_utils.IsDirExists(path) {
		return dsh_utils.NewError("import local project dir not found", map[string]any{
			"path": path,
		})
	}
	importProjectPath, err := filepath.Abs(path)
	if err != nil {
		return dsh_utils.WrapError(err, "import load project abs-path get failed", map[string]any{
			"path": path,
		})
	}
	if importProjectPath == reference.Path {
		return nil
	}
	imp := NewProjectInstanceImportLocal(context, reference, importProjectPath)
	if _, exist := container.ImportUniqueMap[imp.Unique]; !exist {
		container.ImportUniqueMap[imp.Unique] = imp
		container.Imports = append(container.Imports, imp)
	}
	return nil
}

func (container *ProjectInstanceImportShallowContainer) ImportGit(context *Context, reference *ProjectInfo, rawUrl string, rawRef string) error {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return dsh_utils.WrapError(err, "import git project url parse failed", map[string]any{
			"url": rawUrl,
		})
	}
	parsedRef := ParseGitRef(rawRef)
	projectPath := context.Workspace.GetGitProjectPath(parsedUrl, parsedRef)
	if projectPath == reference.Path {
		return nil
	}
	imp := NewProjectInstanceImportGit(context, reference, projectPath, rawUrl, parsedUrl, rawRef, parsedRef)
	if _, exist := container.ImportUniqueMap[imp.Unique]; !exist {
		container.ImportUniqueMap[imp.Unique] = imp
		container.Imports = append(container.Imports, imp)
	}
	return nil
}

func (container *ProjectInstanceImportShallowContainer) LoadImports() (err error) {
	if container.ImportsLoaded {
		return nil
	}

	for i := 0; i < len(container.Imports); i++ {
		if err = container.Imports[i].LoadProject(); err != nil {
			return err
		}
	}

	container.ImportsLoaded = true
	return nil
}

func NewDeepImportContainer(instance *ProjectInstance, scope ProjectInstanceImportScope) *ProjectInstanceImportDeepContainer {
	return &ProjectInstanceImportDeepContainer{
		Context:  instance.Context,
		Instance: instance,
		Scope:    scope,
	}
}

func (container *ProjectInstanceImportDeepContainer) LoadImports() (err error) {
	if container.ImportsLoaded {
		return nil
	}
	if err = container.Instance.LoadImports(container.Scope); err != nil {
		return err
	}

	var deepImports []*ProjectInstanceImport
	var deepImportMap = make(map[string]*ProjectInstanceImport)

	shallowContainer := container.Instance.GetImportContainer(container.Scope)
	for i := 0; i < len(shallowContainer.Imports); i++ {
		imp := shallowContainer.Imports[i]
		deepImports = append(deepImports, imp)
		deepImportMap[imp.Unique] = imp
	}

	scanningImports := shallowContainer.Imports
	for i := 0; i < len(scanningImports); i++ {
		imp1 := scanningImports[i]
		if err = imp1.Instance.LoadImports(shallowContainer.Scope); err != nil {
			return err
		}
		sic1 := imp1.Instance.GetImportContainer(shallowContainer.Scope)
		for j := 0; j < len(sic1.Imports); j++ {
			imp2 := sic1.Imports[j]
			if imp2.ProjectPath == container.Instance.Info.Path {
				continue
			}
			if _, exist := deepImportMap[imp2.Unique]; !exist {
				deepImports = append(deepImports, imp2)
				deepImportMap[imp2.Unique] = imp2
				scanningImports = append(scanningImports, imp2)
			}
		}
	}

	container.Imports = deepImports
	container.ImportsLoaded = true
	return nil
}

func (container *ProjectInstanceImportDeepContainer) BuildScriptSources(config map[string]any, funcs template.FuncMap, outputPath string) (err error) {
	if container.Scope != ProjectInstanceImportScopeScript {
		panic("ProjectInstanceImportDeepContainer.BuildScriptSources() only support ProjectInstanceImportScopeScript")
	}
	if err = container.LoadImports(); err != nil {
		return err
	}
	for i := 0; i < len(container.Imports); i++ {
		if err = container.Imports[i].Instance.BuildScriptSources(config, funcs, outputPath); err != nil {
			return err
		}
	}
	if err = container.Instance.BuildScriptSources(config, funcs, outputPath); err != nil {
		return err
	}
	return nil
}

func (container *ProjectInstanceImportDeepContainer) LoadConfigSources() (sources []*ProjectInstanceConfigSource, err error) {
	if container.Scope != ProjectInstanceImportScopeConfig {
		panic("ProjectInstanceImportDeepContainer.LoadConfigSources() only support ProjectInstanceImportScopeConfig")
	}
	if err = container.LoadImports(); err != nil {
		return nil, err
	}
	for i := 0; i < len(container.Imports); i++ {
		if err = container.Imports[i].Instance.LoadConfigSources(); err != nil {
			return nil, err
		}
	}
	if err = container.Instance.LoadConfigSources(); err != nil {
		return nil, err
	}

	for i := 0; i < len(container.Imports); i++ {
		for j := 0; j < len(container.Imports[i].Instance.Config.SourceContainer.YamlSources); j++ {
			source := container.Imports[i].Instance.Config.SourceContainer.YamlSources[j]
			sources = append(sources, source)
		}
	}
	for i := 0; i < len(container.Instance.Config.SourceContainer.YamlSources); i++ {
		source := container.Instance.Config.SourceContainer.YamlSources[i]
		sources = append(sources, source)
	}

	slices.SortStableFunc(sources, func(a, b *ProjectInstanceConfigSource) int {
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
