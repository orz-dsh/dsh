package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
	"strings"
)

type projectImport struct {
	context *appContext
	// TODO
	reference    *projectManifest
	mode         projectImportMode
	unique       string
	projectPath  string
	gitRawUrl    string
	gitParsedUrl *url.URL
	gitRawRef    string
	gitParsedRef *gitRef
	project      *project
}

type projectImportContainer struct {
	context         *appContext
	scope           projectImportScope
	imports         []*projectImport
	importsByUnique map[string]*projectImport
	importsLoaded   bool
}

type projectImportScope string

const (
	projectImportScopeScript projectImportScope = "script"
	projectImportScopeConfig projectImportScope = "config"
)

type projectImportMode string

const (
	projectImportModeLocal projectImportMode = "local"
	projectImportModeGit   projectImportMode = "git"
)

func newLocalProjectImport(context *appContext, reference *projectManifest, projectPath string) *projectImport {
	return &projectImport{
		context:     context,
		reference:   reference,
		mode:        projectImportModeLocal,
		unique:      projectPath,
		projectPath: projectPath,
	}
}

func newGitProjectImport(context *appContext, reference *projectManifest, projectPath string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) *projectImport {
	return &projectImport{
		context:      context,
		reference:    reference,
		mode:         projectImportModeGit,
		unique:       projectPath,
		projectPath:  projectPath,
		gitRawUrl:    rawUrl,
		gitParsedUrl: parsedUrl,
		gitRawRef:    rawRef,
		gitParsedRef: parsedRef,
	}
}

func (i *projectImport) loadProject() error {
	if i.project == nil {
		w := i.context.workspace
		if i.mode == projectImportModeLocal {
			pm, err := w.loadProjectManifest(i.projectPath)
			if err != nil {
				return errW(err, "load import project error",
					kv("reason", "load project manifest error"),
					kv("projectPath", i.projectPath),
				)
			}
			p, err := i.context.loadProject(pm)
			if err != nil {
				return errW(err, "load import project error",
					kv("reason", "load project error"),
					kv("projectPath", i.projectPath),
				)
			}
			i.project = p
		} else {
			pm, err := w.loadGitProjectManifest(i.projectPath, i.gitRawUrl, i.gitParsedUrl, i.gitRawRef, i.gitParsedRef)
			if err != nil {
				return errW(err, "load import project error",
					kv("reason", "load git project manifest error"),
					kv("projectPath", i.projectPath),
					kv("gitUrl", i.gitRawUrl),
					kv("gitRef", i.gitRawRef),
				)
			}
			p, err := i.context.loadProject(pm)
			if err != nil {
				return errW(err, "load import project error",
					kv("reason", "load project error"),
					kv("projectPath", i.projectPath),
					kv("gitUrl", i.gitRawUrl),
					kv("gitRef", i.gitRawRef),
				)
			}
			i.project = p
		}
	}
	return nil
}

func loadProjectImportContainer(context *appContext, manifest *projectManifest, scope projectImportScope) (container *projectImportContainer, err error) {
	var imports []*projectManifestImport
	if scope == projectImportScopeScript {
		imports = manifest.Script.Imports
	} else if scope == projectImportScopeConfig {
		imports = manifest.Config.Imports
	} else {
		panic(desc("invalid import scope", kv("scope", scope)))
	}
	container = &projectImportContainer{
		context:         context,
		scope:           scope,
		importsByUnique: make(map[string]*projectImport),
	}
	for i := 0; i < len(imports); i++ {
		imp := imports[i]
		if imp.Match != "" {
			matched, err := context.option.evalProjectMatchExpr(manifest, imp.match)
			if err != nil {
				return nil, err
			}
			if !matched {
				continue
			}
		}
		var pimp *projectImport
		if imp.Registry != nil {
			if pimp, err = container.newRegistryImport(imp.Registry, manifest); err != nil {
				return nil, err
			}
		} else if imp.Local != nil {
			if pimp, err = container.newLocalImport(imp.Local.Dir, manifest); err != nil {
				return nil, err
			}
		} else if imp.Git != nil {
			if pimp, err = container.newGitImport(manifest, imp.Git.Url, imp.Git.url, imp.Git.Ref, imp.Git.ref); err != nil {
				return nil, err
			}
		}
		if pimp != nil {
			pimp, err = container.redirectImport(pimp)
			if err != nil {
				return nil, err
			}
			if _, exist := container.importsByUnique[pimp.unique]; !exist {
				container.imports = append(container.imports, pimp)
				container.importsByUnique[pimp.unique] = pimp
			}
		}
	}
	return container, nil
}

func (c *projectImportContainer) newRegistryImport(imp *projectManifestImportRegistry, reference *projectManifest) (*projectImport, error) {
	// TODO: error info
	registry := c.context.workspace.manifest.getImportRegistry(imp.Name)
	if registry == nil {
		return nil, errN("new registry import error",
			reason("registry not found"),
			kv("scope", c.scope),
			kv("import", imp),
		)
	}
	if registry.Local != nil {
		localDir, err := executeStringTemplate(registry.Local.Dir, map[string]any{
			"path": imp.Path,
			"ref":  imp.Ref,
		}, nil)
		localDir = strings.TrimSpace(localDir)
		if err != nil {
			return nil, errW(err, "add registry import error",
				reason("execute local dir template error"),
				kv("scope", c.scope),
				kv("name", imp.Name),
				kv("path", imp.Path),
				kv("ref", imp.Ref),
			)
		}
		return c.newLocalImport(localDir, reference)
	} else if registry.Git != nil {
		gitRawUrl, err := executeStringTemplate(registry.Git.Url, map[string]any{
			"path": imp.Path,
			"ref":  imp.Ref,
		}, nil)
		gitRawUrl = strings.TrimSpace(gitRawUrl)
		if err != nil {
			return nil, errW(err, "new registry import error",
				reason("execute git url template error"),
				kv("scope", c.scope),
				kv("import", imp),
				kv("registry", registry),
			)
		}
		gitParsedUrl, err := url.Parse(gitRawUrl)
		if err != nil {
			return nil, errW(err, "new registry import error",
				reason("parse git url error"),
				kv("scope", c.scope),
				kv("import", imp),
				kv("registry", registry),
				kv("url", gitRawUrl),
			)
		}
		gitRawRef := imp.Ref
		if gitRawRef == "" {
			gitRawRef = registry.Git.Ref
		}
		gitParsedRef := parseGitRef(gitRawRef)
		return c.newGitImport(reference, gitRawUrl, gitParsedUrl, gitRawRef, gitParsedRef)
	} else {
		// impossible
		panic(desc("invalid registry import",
			kv("scope", c.scope),
			kv("name", imp.Name),
			kv("path", imp.Path),
			kv("ref", imp.Ref),
			kv("registry", registry),
		))
	}
}

func (c *projectImportContainer) newLocalImport(path string, reference *projectManifest) (imp *projectImport, err error) {
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
	if path == reference.projectPath {
		return nil, nil
	}
	imp = newLocalProjectImport(c.context, reference, path)
	return imp, nil
}

func (c *projectImportContainer) newGitImport(reference *projectManifest, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) (*projectImport, error) {
	path := c.context.workspace.getGitProjectPath(parsedUrl, parsedRef)
	if path == reference.projectPath {
		return nil, nil
	}
	imp := newGitProjectImport(c.context, reference, path, rawUrl, parsedUrl, rawRef, parsedRef)
	return imp, nil
}

func (c *projectImportContainer) redirectImport(oImp *projectImport) (rImp *projectImport, err error) {
	var redirect *workspaceManifestImportRedirect
	if oImp.mode == projectImportModeLocal {
		redirect = c.context.workspace.manifest.getImportRedirect(oImp.projectPath)
	} else {
		redirect = c.context.workspace.manifest.getImportRedirect(oImp.gitRawUrl)
	}
	if redirect != nil {
		origin := map[string]any{
			"mode": oImp.mode,
		}
		if oImp.mode == projectImportModeLocal {
			origin["path"] = oImp.projectPath
		} else if oImp.mode == projectImportModeGit {
			origin["url"] = oImp.gitRawUrl
			origin["ref"] = oImp.gitRawRef
		} else {
			// impossible
			panic(desc("invalid origin import mode",
				kv("scope", c.scope),
				kv("originImport.unique", oImp.unique),
				kv("originImport.mode", oImp.mode),
			))
		}
		if redirect.Local != nil {
			// TODO: template data
			localDir, err := executeStringTemplate(redirect.Local.Dir, map[string]any{
				"path":   oImp.projectPath[len(redirect.Prefix):],
				"origin": origin,
			}, nil)
			localDir = strings.TrimSpace(localDir)
			if err != nil {
				return nil, errW(err, "redirect import error",
					reason("execute local dir template error"),
					kv("scope", c.scope),
					kv("redirect.prefix", redirect.Prefix),
					kv("redirect.local.dir", redirect.Local.Dir),
				)
			}
			return c.newLocalImport(localDir, oImp.reference)
		} else if redirect.Git != nil {
			// TODO: template data
			gitRawUrl, err := executeStringTemplate(redirect.Git.Url, map[string]any{
				"path":   oImp.gitRawUrl[len(redirect.Prefix):],
				"origin": origin,
			}, nil)
			gitRawUrl = strings.TrimSpace(gitRawUrl)
			if err != nil {
				return nil, errW(err, "redirect import error",
					reason("execute git url template error"),
					kv("scope", c.scope),
					kv("redirect.prefix", redirect.Prefix),
					kv("redirect.git.url", redirect.Git.Url),
				)
			}
			gitParsedUrl, err := url.Parse(gitRawUrl)
			if err != nil {
				return nil, errW(err, "new registry import error",
					reason("parse git url error"),
					kv("scope", c.scope),
					kv("redirect.prefix", redirect.Prefix),
					kv("redirect.git.url", redirect.Git.Url),
					kv("url", gitRawUrl),
				)
			}
			gitRawRef := oImp.gitRawRef
			if redirect.Git.Ref != "" {
				gitRawRef = redirect.Git.Ref
			}
			gitParsedRef := parseGitRef(gitRawRef)
			return c.newGitImport(oImp.reference, gitRawUrl, gitParsedUrl, gitRawRef, gitParsedRef)
		} else {
			// impossible
			panic(desc("invalid redirect import",
				kv("scope", c.scope),
				kv("redirect", redirect),
			))
		}
	}
	return oImp, nil
}

func (c *projectImportContainer) loadImports() (err error) {
	if c.importsLoaded {
		return nil
	}
	for i := 0; i < len(c.imports); i++ {
		if err = c.imports[i].loadProject(); err != nil {
			return errW(err, "load imports error",
				reason("load import project error"),
				kv("scope", c.scope),
			)
		}
	}
	c.importsLoaded = true
	return nil
}
