package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
	"slices"
	"text/template"
)

type projectInstanceImport struct {
	context      *Context
	reference    *projectInfo
	importType   projectInstanceImportType
	unique       string
	projectPath  string
	gitRawUrl    string
	gitParsedUrl *url.URL
	gitRawRef    string
	gitParsedRef *gitRef
	instance     *projectInstance
}

type projectInstanceImportScope string

const (
	projectInstanceImportScopeScript projectInstanceImportScope = "Script"
	projectInstanceImportScopeConfig projectInstanceImportScope = "Config"
)

type projectInstanceImportType string

const (
	projectInstanceImportTypeLocal projectInstanceImportType = "Local"
	projectInstanceImportTypeGit   projectInstanceImportType = "Git"
)

type projectInstanceImportShallowContainer struct {
	context         *Context
	scope           projectInstanceImportScope
	importUniqueMap map[string]*projectInstanceImport
	imports         []*projectInstanceImport
	importsLoaded   bool
}

type projectInstanceImportDeepContainer struct {
	context       *Context
	instance      *projectInstance
	scope         projectInstanceImportScope
	imports       []*projectInstanceImport
	importsLoaded bool
}

func newProjectInstanceImportLocal(context *Context, reference *projectInfo, projectPath string) *projectInstanceImport {
	return &projectInstanceImport{
		context:     context,
		reference:   reference,
		importType:  projectInstanceImportTypeLocal,
		unique:      projectPath,
		projectPath: projectPath,
	}
}

func newProjectInstanceImportGit(context *Context, reference *projectInfo, projectPath string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) *projectInstanceImport {
	return &projectInstanceImport{
		context:      context,
		reference:    reference,
		importType:   projectInstanceImportTypeGit,
		unique:       projectPath,
		projectPath:  projectPath,
		gitRawUrl:    rawUrl,
		gitParsedUrl: parsedUrl,
		gitRawRef:    rawRef,
		gitParsedRef: parsedRef,
	}
}

func (imp *projectInstanceImport) load() error {
	if imp.instance == nil {
		workspace := imp.context.Workspace
		if imp.importType == projectInstanceImportTypeLocal {
			info, err := workspace.loadLocalProjectInfo(imp.projectPath)
			if err != nil {
				return dsh_utils.WrapError(err, "import local project info failed", map[string]any{
					"projectPath": imp.projectPath,
				})
			}
			instance, err := imp.context.newProjectInstance(info, nil)
			if err != nil {
				return dsh_utils.WrapError(err, "import local project instance failed", map[string]any{
					"projectPath": imp.projectPath,
				})
			}
			imp.instance = instance
		} else {
			info, err := workspace.loadGitProjectInfo(imp.projectPath, imp.gitRawUrl, imp.gitParsedUrl, imp.gitRawRef, imp.gitParsedRef)
			if err != nil {
				return dsh_utils.WrapError(err, "import git project info failed", map[string]any{
					"projectPath": imp.projectPath,
					"gitUrl":      imp.gitRawUrl,
					"gitRef":      imp.gitRawRef,
				})
			}
			instance, err := imp.context.newProjectInstance(info, nil)
			if err != nil {
				return dsh_utils.WrapError(err, "import git project instance failed", map[string]any{
					"projectPath": imp.projectPath,
					"gitUrl":      imp.gitRawUrl,
					"gitRef":      imp.gitRawRef,
				})
			}
			imp.instance = instance
		}
	}
	return nil
}

func newProjectInstanceImportShallowContainer(context *Context, scope projectInstanceImportScope) *projectInstanceImportShallowContainer {
	return &projectInstanceImportShallowContainer{
		context:         context,
		scope:           scope,
		importUniqueMap: make(map[string]*projectInstanceImport),
	}
}

func (container *projectInstanceImportShallowContainer) importLocal(context *Context, path string, reference *projectInfo) (err error) {
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
	if importProjectPath == reference.path {
		return nil
	}
	imp := newProjectInstanceImportLocal(context, reference, importProjectPath)
	if _, exist := container.importUniqueMap[imp.unique]; !exist {
		container.importUniqueMap[imp.unique] = imp
		container.imports = append(container.imports, imp)
	}
	return nil
}

func (container *projectInstanceImportShallowContainer) importGit(context *Context, reference *projectInfo, rawUrl string, rawRef string) error {
	parsedUrl, err := url.Parse(rawUrl)
	if err != nil {
		return dsh_utils.WrapError(err, "import git project url parse failed", map[string]any{
			"url": rawUrl,
		})
	}
	parsedRef := parseGitRef(rawRef)
	projectPath := context.Workspace.getGitProjectPath(parsedUrl, parsedRef)
	if projectPath == reference.path {
		return nil
	}
	imp := newProjectInstanceImportGit(context, reference, projectPath, rawUrl, parsedUrl, rawRef, parsedRef)
	if _, exist := container.importUniqueMap[imp.unique]; !exist {
		container.importUniqueMap[imp.unique] = imp
		container.imports = append(container.imports, imp)
	}
	return nil
}

func (container *projectInstanceImportShallowContainer) loadImports() (err error) {
	if container.importsLoaded {
		return nil
	}

	for i := 0; i < len(container.imports); i++ {
		if err = container.imports[i].load(); err != nil {
			return err
		}
	}

	container.importsLoaded = true
	return nil
}

func newProjectInstanceImportDeepContainer(instance *projectInstance, scope projectInstanceImportScope) *projectInstanceImportDeepContainer {
	return &projectInstanceImportDeepContainer{
		context:  instance.context,
		instance: instance,
		scope:    scope,
	}
}

func (container *projectInstanceImportDeepContainer) loadImports() (err error) {
	if container.importsLoaded {
		return nil
	}
	if err = container.instance.loadImports(container.scope); err != nil {
		return err
	}

	var deepImports []*projectInstanceImport
	var deepImportMap = make(map[string]*projectInstanceImport)

	shallowContainer := container.instance.getImportContainer(container.scope)
	for i := 0; i < len(shallowContainer.imports); i++ {
		imp := shallowContainer.imports[i]
		deepImports = append(deepImports, imp)
		deepImportMap[imp.unique] = imp
	}

	scanningImports := shallowContainer.imports
	for i := 0; i < len(scanningImports); i++ {
		imp1 := scanningImports[i]
		if err = imp1.instance.loadImports(shallowContainer.scope); err != nil {
			return err
		}
		sic1 := imp1.instance.getImportContainer(shallowContainer.scope)
		for j := 0; j < len(sic1.imports); j++ {
			imp2 := sic1.imports[j]
			if imp2.projectPath == container.instance.info.path {
				continue
			}
			if _, exist := deepImportMap[imp2.unique]; !exist {
				deepImports = append(deepImports, imp2)
				deepImportMap[imp2.unique] = imp2
				scanningImports = append(scanningImports, imp2)
			}
		}
	}

	container.imports = deepImports
	container.importsLoaded = true
	return nil
}

func (container *projectInstanceImportDeepContainer) makeScript(config map[string]any, funcs template.FuncMap, outputPath string) (err error) {
	if container.scope != projectInstanceImportScopeScript {
		panic("projectInstanceImportDeepContainer.makeScript() only support projectInstanceImportScopeScript")
	}
	if err = container.loadImports(); err != nil {
		return err
	}
	for i := 0; i < len(container.imports); i++ {
		if err = container.imports[i].instance.makeScript(config, funcs, outputPath); err != nil {
			return err
		}
	}
	if err = container.instance.makeScript(config, funcs, outputPath); err != nil {
		return err
	}
	return nil
}

func (container *projectInstanceImportDeepContainer) loadConfigSources() (sources []*projectInstanceConfigSource, err error) {
	if container.scope != projectInstanceImportScopeConfig {
		panic("projectInstanceImportDeepContainer.loadConfigSources() only support projectInstanceImportScopeConfig")
	}
	if err = container.loadImports(); err != nil {
		return nil, err
	}
	for i := 0; i < len(container.imports); i++ {
		if err = container.imports[i].instance.loadConfigSources(); err != nil {
			return nil, err
		}
	}
	if err = container.instance.loadConfigSources(); err != nil {
		return nil, err
	}

	for i := 0; i < len(container.imports); i++ {
		for j := 0; j < len(container.imports[i].instance.config.sourceContainer.sources); j++ {
			source := container.imports[i].instance.config.sourceContainer.sources[j]
			if source.content.match != nil {
				matched, err := container.imports[i].instance.option.match(source.content.match)
				if err != nil {
					return nil, err
				}
				if !matched {
					continue
				}
			}
			sources = append(sources, source)
		}
	}
	for i := 0; i < len(container.instance.config.sourceContainer.sources); i++ {
		source := container.instance.config.sourceContainer.sources[i]
		if source.content.match != nil {
			matched, err := container.instance.option.match(source.content.match)
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}
		}
		sources = append(sources, source)
	}

	slices.SortStableFunc(sources, func(a, b *projectInstanceConfigSource) int {
		rst := a.content.Order - b.content.Order
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
