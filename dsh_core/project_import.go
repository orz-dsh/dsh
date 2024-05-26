package dsh_core

// region import

type projectImport struct {
	context    *appContext
	manifest   *projectManifest
	Definition *projectImportDefinition
	Link       *projectResolvedLink
	target     *project
}

type projectImportScope string

const (
	projectImportScopeScript projectImportScope = "script"
	projectImportScopeConfig projectImportScope = "config"
)

func newProjectImport(context *appContext, manifest *projectManifest, definition *projectImportDefinition, link *projectResolvedLink) *projectImport {
	return &projectImport{
		context:    context,
		manifest:   manifest,
		Definition: definition,
		Link:       link,
	}
}

func (i *projectImport) loadTarget() error {
	if i.target == nil {
		m, err := i.context.workspace.loadProjectManifest(i.Link)
		if err != nil {
			return errW(err, "load import target error",
				kv("reason", "load project manifest error"),
				kv("link", i.Link),
			)
		}
		p, err := i.context.loadProject(m)
		if err != nil {
			return errW(err, "load import target error",
				kv("reason", "load project error"),
				kv("link", i.Link),
			)
		}
		i.target = p
	}
	return nil
}

// endregion

// region container

type projectImportContainer struct {
	context       *appContext
	manifest      *projectManifest
	scope         projectImportScope
	evaluator     *projectImportEvaluator
	Imports       []*projectImport
	importsByPath map[string]*projectImport
	importsLoaded bool
}

func makeProjectImportContainer(context *appContext, manifest *projectManifest, scope projectImportScope) (container *projectImportContainer, err error) {
	var definitions []*projectImportDefinition
	if scope == projectImportScopeScript {
		definitions = manifest.Script.importDefinitions
		if context.isMainProject(manifest) {
			definitions = append(definitions, context.Profile.getProjectScriptImportDefinitions()...)
		}
	} else if scope == projectImportScopeConfig {
		definitions = manifest.Config.importDefinitions
		if context.isMainProject(manifest) {
			definitions = append(definitions, context.Profile.getProjectConfigImportDefinitions()...)
		}
	} else {
		impossible()
	}
	container = &projectImportContainer{
		context:       context,
		manifest:      manifest,
		scope:         scope,
		evaluator:     newProjectImportEvaluator(context.Profile.evalData),
		importsByPath: make(map[string]*projectImport),
	}
	for i := 0; i < len(definitions); i++ {
		if err = container.addImport(definitions[i]); err != nil {
			return nil, err
		}
	}
	return container, nil
}

func (c *projectImportContainer) addImport(definition *projectImportDefinition) (err error) {
	if definition.match != nil {
		matched, err := c.context.Option.evalMatch(c.manifest, definition.match)
		if err != nil {
			return errW(err, "add import error",
				reason("eval match error"),
				kv("scope", c.scope),
				kv("definition", definition),
			)
		}
		if !matched {
			return nil
		}
	}
	resolved, err := c.context.Profile.resolveProjectLink(definition.Link)
	if err != nil {
		return errW(err, "add import error",
			reason("resolve project link error"),
			kv("scope", c.scope),
			kv("definition", definition),
		)
	}
	if resolved.Path == c.manifest.projectPath {
		return nil
	}
	imp := newProjectImport(c.context, c.manifest, definition, resolved)
	if _, exist := c.importsByPath[resolved.Path]; !exist {
		c.Imports = append(c.Imports, imp)
		c.importsByPath[resolved.Path] = imp
	}
	return nil
}

//func (c *projectImportContainer) makeRegistryImport(projectDefinition *projectImportRegistryDefinition) (*projectImport, error) {
//	workspaceDefinition, err := c.context.Profile.getWorkspaceImportRegistryLink(projectDefinition.Name)
//	if err != nil {
//		return nil, errW(err, "make registry import error",
//			reason("get registry definition error"),
//			kv("scope", c.scope),
//			kv("import", projectDefinition),
//		)
//	}
//	// TODO: error info
//	if workspaceDefinition == nil {
//		return nil, errN("make registry import error",
//			reason("registry not found"),
//			kv("scope", c.scope),
//			kv("import", projectDefinition),
//		)
//	}
//	if workspaceDefinition.Local != nil {
//		localRawDir, err := c.evaluator.evalRegistry(workspaceDefinition.Local.Dir, projectDefinition.Path, projectDefinition.Ref)
//		if err != nil {
//			return nil, errW(err, "add registry import error",
//				reason("eval local dir template error"),
//				kv("scope", c.scope),
//				kv("name", projectDefinition.Name),
//				kv("path", projectDefinition.Path),
//				kv("ref", projectDefinition.Ref),
//			)
//		}
//		return c.makeDirImport(nil, &projectImportRegistry{
//			Name: projectDefinition.Name,
//			Path: projectDefinition.Path,
//			Ref:  projectDefinition.Ref,
//		}, localRawDir)
//	} else if workspaceDefinition.Git != nil {
//		gitRawUrl, err := c.evaluator.evalRegistry(workspaceDefinition.Git.Url, projectDefinition.Path, projectDefinition.Ref)
//		if err != nil {
//			return nil, errW(err, "new registry import error",
//				reason("execute git url template error"),
//				kv("scope", c.scope),
//				kv("import", projectDefinition),
//				kv("registry", workspaceDefinition),
//			)
//		}
//		gitRawRef := t(projectDefinition.Ref != "", projectDefinition.Ref, workspaceDefinition.Git.Ref)
//		return c.makeGitImport(nil, &projectImportRegistry{
//			Name: projectDefinition.Name,
//			Path: projectDefinition.Path,
//			Ref:  projectDefinition.Ref,
//		}, gitRawUrl, nil, gitRawRef, nil)
//	} else {
//		impossible()
//	}
//	return nil, nil
//}
//
//func (c *projectImportContainer) makeDirImport(original *projectImport, registry *projectImportRegistry, rawDir string) (imp *projectImport, err error) {
//	path := rawDir
//	if !dsh_utils.IsDirExists(path) {
//		return nil, errN("new local import error",
//			reason("dir not exists"),
//			kv("scope", c.scope),
//			kv("path", path),
//		)
//	}
//	absPath, err := filepath.Abs(path)
//	if err != nil {
//		return nil, errW(err, "new local import error",
//			reason("get abs-path error"),
//			kv("scope", c.scope),
//			kv("path", path),
//		)
//	}
//	path = absPath
//	if path == c.manifest.projectPath {
//		return nil, nil
//	}
//	local := &projectImportLocal{
//		RawDir: rawDir,
//	}
//	imp = newProjectImport(c.context, c.manifest, original, registry, path, local, nil)
//	return imp, nil
//}
//
//func (c *projectImportContainer) makeGitImport(original *projectImport, registry *projectImportRegistry, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) (imp *projectImport, err error) {
//	if parsedUrl == nil {
//		if parsedUrl, err = url.Parse(rawUrl); err != nil {
//			return nil, errW(err, "new git import error",
//				reason("parse git url error"),
//				kv("scope", c.scope),
//				kv("url", rawUrl),
//			)
//		}
//	}
//	if rawRef == "" {
//		rawRef = "main"
//		parsedRef = parseGitRef(rawRef)
//	}
//	if parsedRef == nil {
//		parsedRef = parseGitRef(rawRef)
//	}
//	path := c.context.workspace.getGitProjectPath(parsedUrl, parsedRef)
//	if path == c.manifest.projectPath {
//		return nil, nil
//	}
//	git := &projectImportGit{
//		RawUrl:    rawUrl,
//		parsedUrl: parsedUrl,
//		RawRef:    rawRef,
//		parsedRef: parsedRef,
//	}
//	imp = newProjectImport(c.context, c.manifest, original, registry, path, nil, git)
//	return imp, nil
//}
//
//func (c *projectImportContainer) redirectImport(original *projectImport) (_ *projectImport, err error) {
//	var resources []string
//	if original.Local != nil {
//		resources = []string{original.Local.RawDir, original.Path}
//	} else if original.Git != nil {
//		resources = []string{original.Git.RawUrl, original.Path}
//	} else {
//		impossible()
//	}
//	definition, path, err := c.context.Profile.getWorkspaceImportRedirectDefinition(resources)
//	if err != nil {
//		return nil, errW(err, "redirect import error",
//			reason("get redirect definition error"),
//			kv("scope", c.scope),
//			kv("resources", resources),
//		)
//	}
//	if definition != nil {
//		if definition.Local != nil {
//			localRawDir, err := c.evaluator.evalRedirect(definition.Local.Dir, path, original)
//			if err != nil {
//				return nil, errW(err, "redirect import error",
//					reason("eval local dir template error"),
//					kv("scope", c.scope),
//					kv("definition", definition),
//				)
//			}
//			return c.makeDirImport(original, original.Registry, localRawDir)
//		} else if definition.Git != nil {
//			gitRawUrl, err := c.evaluator.evalRedirect(definition.Git.Url, path, original)
//			if err != nil {
//				return nil, errW(err, "redirect import error",
//					reason("eval git url template error"),
//					kv("scope", c.scope),
//					kv("definition", definition),
//				)
//			}
//			gitRawRef := t(original.Git != nil, original.Git.RawRef, definition.Git.Ref)
//			gitRawRef = t(gitRawRef != "", gitRawRef, "main")
//			return c.makeGitImport(original, original.Registry, gitRawUrl, nil, gitRawRef, nil)
//		} else {
//			impossible()
//		}
//	}
//	return original, nil
//}

func (c *projectImportContainer) loadImports() (err error) {
	if c.importsLoaded {
		return nil
	}
	for i := 0; i < len(c.Imports); i++ {
		if err = c.Imports[i].loadTarget(); err != nil {
			return errW(err, "load imports error",
				reason("load import target error"),
				kv("scope", c.scope),
			)
		}
	}
	c.importsLoaded = true
	return nil
}

// endregion
