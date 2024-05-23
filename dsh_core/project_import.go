package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
)

// region import

type projectImport struct {
	context  *appContext
	manifest *projectManifest
	Original *projectImport
	Registry *projectImportRegistry
	Path     string
	Local    *projectImportLocal
	Git      *projectImportGit
	target   *project
}

type projectImportRegistry struct {
	Name string
	Path string
	Ref  string
}

type projectImportLocal struct {
	RawDir string
}

type projectImportGit struct {
	RawUrl    string
	parsedUrl *url.URL
	RawRef    string
	parsedRef *gitRef
}

type projectImportScope string

const (
	projectImportScopeScript projectImportScope = "script"
	projectImportScopeConfig projectImportScope = "config"
)

func newProjectImport(context *appContext, manifest *projectManifest, original *projectImport, registry *projectImportRegistry, path string, local *projectImportLocal, git *projectImportGit) *projectImport {
	importModeCount := 0
	if local != nil {
		importModeCount++
	}
	if git != nil {
		importModeCount++
	}
	if importModeCount != 1 {
		panic(desc("invalid import",
			kv("path", path),
			kv("local", local),
			kv("git", git),
		))
	}
	return &projectImport{
		context:  context,
		manifest: manifest,
		Original: original,
		Registry: registry,
		Path:     path,
		Local:    local,
		Git:      git,
	}
}

func (i *projectImport) loadTarget() error {
	if i.target == nil {
		w := i.context.workspace
		if i.Local != nil {
			m, err := w.loadProjectManifest(i.Path)
			if err != nil {
				return errW(err, "load import target error",
					kv("reason", "load project manifest error"),
					kv("path", i.Path),
				)
			}
			p, err := i.context.loadProject(m)
			if err != nil {
				return errW(err, "load import target error",
					kv("reason", "load project error"),
					kv("path", i.Path),
				)
			}
			i.target = p
		} else if i.Git != nil {
			m, err := w.loadGitProjectManifest(i.Path, i.Git.RawUrl, i.Git.parsedUrl, i.Git.RawRef, i.Git.parsedRef)
			if err != nil {
				return errW(err, "load import target error",
					kv("reason", "load git project manifest error"),
					kv("path", i.target),
					kv("gitUrl", i.Git.RawUrl),
					kv("gitRef", i.Git.RawRef),
				)
			}
			p, err := i.context.loadProject(m)
			if err != nil {
				return errW(err, "load import target error",
					kv("reason", "load project error"),
					kv("path", i.target),
					kv("gitUrl", i.Git.RawUrl),
					kv("gitRef", i.Git.RawRef),
				)
			}
			i.target = p
		} else {
			impossible()
		}
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
				kv("match", definition.Match),
			)
		}
		if !matched {
			return nil
		}
	}
	var imp *projectImport
	if definition.Registry != nil {
		if imp, err = c.makeRegistryImport(definition.Registry); err != nil {
			return err
		}
	} else if definition.Local != nil {
		if imp, err = c.makeLocalImport(nil, nil, definition.Local.Dir); err != nil {
			return err
		}
	} else if definition.Git != nil {
		if imp, err = c.makeGitImport(nil, nil, definition.Git.Url, definition.Git.url, definition.Git.Ref, definition.Git.ref); err != nil {
			return err
		}
	}
	if imp != nil {
		imp, err = c.redirectImport(imp)
		if err != nil {
			return err
		}
		if _, exist := c.importsByPath[imp.Path]; !exist {
			c.Imports = append(c.Imports, imp)
			c.importsByPath[imp.Path] = imp
		}
	}
	return nil
}

func (c *projectImportContainer) makeRegistryImport(projectDefinition *projectImportRegistryDefinition) (*projectImport, error) {
	workspaceDefinition := c.context.Profile.getWorkspaceImportRegistryDefinition(projectDefinition.Name)
	// TODO: error info
	if workspaceDefinition == nil {
		return nil, errN("make registry import error",
			reason("registry not found"),
			kv("scope", c.scope),
			kv("import", projectDefinition),
		)
	}
	if workspaceDefinition.Local != nil {
		localRawDir, err := c.evaluator.evalRegistry(workspaceDefinition.Local.Dir, projectDefinition.Path, projectDefinition.Ref)
		if err != nil {
			return nil, errW(err, "add registry import error",
				reason("eval local dir template error"),
				kv("scope", c.scope),
				kv("name", projectDefinition.Name),
				kv("path", projectDefinition.Path),
				kv("ref", projectDefinition.Ref),
			)
		}
		return c.makeLocalImport(nil, &projectImportRegistry{
			Name: projectDefinition.Name,
			Path: projectDefinition.Path,
			Ref:  projectDefinition.Ref,
		}, localRawDir)
	} else if workspaceDefinition.Git != nil {
		gitRawUrl, err := c.evaluator.evalRegistry(workspaceDefinition.Git.Url, projectDefinition.Path, projectDefinition.Ref)
		if err != nil {
			return nil, errW(err, "new registry import error",
				reason("execute git url template error"),
				kv("scope", c.scope),
				kv("import", projectDefinition),
				kv("registry", workspaceDefinition),
			)
		}
		gitRawRef := t(projectDefinition.Ref != "", projectDefinition.Ref, workspaceDefinition.Git.Ref)
		return c.makeGitImport(nil, &projectImportRegistry{
			Name: projectDefinition.Name,
			Path: projectDefinition.Path,
			Ref:  projectDefinition.Ref,
		}, gitRawUrl, nil, gitRawRef, nil)
	} else {
		impossible()
	}
	return nil, nil
}

func (c *projectImportContainer) makeLocalImport(original *projectImport, registry *projectImportRegistry, rawDir string) (imp *projectImport, err error) {
	path := rawDir
	if !dsh_utils.IsDirExists(path) {
		return nil, errN("new local import error",
			reason("dir not exists"),
			kv("scope", c.scope),
			kv("path", path),
		)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errW(err, "new local import error",
			reason("get abs-path error"),
			kv("scope", c.scope),
			kv("path", path),
		)
	}
	path = absPath
	if path == c.manifest.projectPath {
		return nil, nil
	}
	local := &projectImportLocal{
		RawDir: rawDir,
	}
	imp = newProjectImport(c.context, c.manifest, original, registry, path, local, nil)
	return imp, nil
}

func (c *projectImportContainer) makeGitImport(original *projectImport, registry *projectImportRegistry, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) (imp *projectImport, err error) {
	if parsedUrl == nil {
		if parsedUrl, err = url.Parse(rawUrl); err != nil {
			return nil, errW(err, "new git import error",
				reason("parse git url error"),
				kv("scope", c.scope),
				kv("url", rawUrl),
			)
		}
	}
	if rawRef == "" {
		rawRef = "main"
		parsedRef = parseGitRef(rawRef)
	}
	if parsedRef == nil {
		parsedRef = parseGitRef(rawRef)
	}
	path := c.context.workspace.getGitProjectPath(parsedUrl, parsedRef)
	if path == c.manifest.projectPath {
		return nil, nil
	}
	git := &projectImportGit{
		RawUrl:    rawUrl,
		parsedUrl: parsedUrl,
		RawRef:    rawRef,
		parsedRef: parsedRef,
	}
	imp = newProjectImport(c.context, c.manifest, original, registry, path, nil, git)
	return imp, nil
}

func (c *projectImportContainer) redirectImport(original *projectImport) (_ *projectImport, err error) {
	var resources []string
	if original.Local != nil {
		resources = []string{original.Local.RawDir, original.Path}
	} else if original.Git != nil {
		resources = []string{original.Git.RawUrl, original.Path}
	} else {
		impossible()
	}
	definition, path := c.context.Profile.getWorkspaceImportRedirectDefinition(resources)
	if definition != nil {
		if definition.Local != nil {
			localRawDir, err := c.evaluator.evalRedirect(definition.Local.Dir, path, original)
			if err != nil {
				return nil, errW(err, "redirect import error",
					reason("eval local dir template error"),
					kv("scope", c.scope),
					kv("definition", definition),
				)
			}
			return c.makeLocalImport(original, original.Registry, localRawDir)
		} else if definition.Git != nil {
			gitRawUrl, err := c.evaluator.evalRedirect(definition.Git.Url, path, original)
			if err != nil {
				return nil, errW(err, "redirect import error",
					reason("eval git url template error"),
					kv("scope", c.scope),
					kv("definition", definition),
				)
			}
			gitRawRef := t(original.Git != nil, original.Git.RawRef, definition.Git.Ref)
			gitRawRef = t(gitRawRef != "", gitRawRef, "main")
			return c.makeGitImport(original, original.Registry, gitRawUrl, nil, gitRawRef, nil)
		} else {
			impossible()
		}
	}
	return original, nil
}

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
