package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
	"strings"
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
	Imports       []*projectImport
	importsByPath map[string]*projectImport
	importsLoaded bool
}

func makeProjectImportContainer(context *appContext, manifest *projectManifest, scope projectImportScope) (container *projectImportContainer, err error) {
	var imports []*projectManifestImport
	if scope == projectImportScopeScript {
		imports = manifest.Script.Imports
	} else if scope == projectImportScopeConfig {
		imports = manifest.Config.Imports
	} else {
		impossible()
	}
	container = &projectImportContainer{
		context:       context,
		manifest:      manifest,
		scope:         scope,
		importsByPath: make(map[string]*projectImport),
	}
	for i := 0; i < len(imports); i++ {
		if err = container.addImport(imports[i]); err != nil {
			return nil, err
		}
	}
	return container, nil
}

func (c *projectImportContainer) addImport(manifestImport *projectManifestImport) (err error) {
	if manifestImport.match != nil {
		matched, err := c.context.Option.evalProjectMatchExpr(c.manifest, manifestImport.match)
		if err != nil {
			return errW(err, "add import error",
				reason("eval match error"),
				kv("scope", c.scope),
				kv("match", manifestImport.Match),
			)
		}
		if !matched {
			return nil
		}
	}
	var imp *projectImport
	if manifestImport.Registry != nil {
		if imp, err = c.makeRegistryImport(manifestImport.Registry); err != nil {
			return err
		}
	} else if manifestImport.Local != nil {
		if imp, err = c.makeLocalImport(nil, nil, manifestImport.Local.Dir); err != nil {
			return err
		}
	} else if manifestImport.Git != nil {
		if imp, err = c.makeGitImport(nil, nil, manifestImport.Git.Url, manifestImport.Git.url, manifestImport.Git.Ref, manifestImport.Git.ref); err != nil {
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

func (c *projectImportContainer) makeRegistryImport(importRegistry *projectManifestImportRegistry) (*projectImport, error) {
	registryDefinition := c.context.Profile.getImportRegistry(importRegistry.Name)
	// TODO: error info
	if registryDefinition == nil {
		return nil, errN("make registry import error",
			reason("registry not found"),
			kv("scope", c.scope),
			kv("import", importRegistry),
		)
	}
	if registryDefinition.Local != nil {
		localRawDir, err := executeStringTemplate(registryDefinition.Local.Dir, map[string]any{
			"path": importRegistry.Path,
			"ref":  importRegistry.Ref,
		}, nil)
		localRawDir = strings.TrimSpace(localRawDir)
		if err != nil {
			return nil, errW(err, "add registry import error",
				reason("execute local dir template error"),
				kv("scope", c.scope),
				kv("name", importRegistry.Name),
				kv("path", importRegistry.Path),
				kv("ref", importRegistry.Ref),
			)
		}
		return c.makeLocalImport(nil, &projectImportRegistry{
			Name: importRegistry.Name,
			Path: importRegistry.Path,
			Ref:  importRegistry.Ref,
		}, localRawDir)
	} else if registryDefinition.Git != nil {
		gitRawUrl, err := executeStringTemplate(registryDefinition.Git.Url, map[string]any{
			"path": importRegistry.Path,
			"ref":  importRegistry.Ref,
		}, nil)
		if err != nil {
			return nil, errW(err, "new registry import error",
				reason("execute git url template error"),
				kv("scope", c.scope),
				kv("import", importRegistry),
				kv("registry", registryDefinition),
			)
		}
		gitRawUrl = strings.TrimSpace(gitRawUrl)
		gitRawRef := t(importRegistry.Ref != "", importRegistry.Ref, registryDefinition.Git.Ref)
		return c.makeGitImport(nil, &projectImportRegistry{
			Name: importRegistry.Name,
			Path: importRegistry.Path,
			Ref:  importRegistry.Ref,
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

func (c *projectImportContainer) redirectImport(originalImport *projectImport) (redirectImport *projectImport, err error) {
	var redirect *workspaceManifestImportRedirect
	if originalImport.Local != nil {
		redirect = c.context.workspace.manifest.Import.getRedirect(originalImport.Local.RawDir)
	} else if originalImport.Git != nil {
		redirect = c.context.workspace.manifest.Import.getRedirect(originalImport.Git.RawUrl)
	} else {
		impossible()
	}
	if redirect != nil {
		original := make(map[string]any)
		redirectPath := ""
		if originalImport.Local != nil {
			redirectPath = originalImport.Local.RawDir[len(redirect.Prefix):]
			original["mode"] = "local"
			original["dir"] = originalImport.Local.RawDir
		} else if originalImport.Git != nil {
			redirectPath = originalImport.Git.RawUrl[len(redirect.Prefix):]
			original["mode"] = "git"
			original["url"] = originalImport.Git.RawUrl
			original["ref"] = originalImport.Git.RawRef
		}
		if redirect.Local != nil {
			// TODO: template data
			localRawDir, err := executeStringTemplate(redirect.Local.Dir, map[string]any{
				"path":     redirectPath,
				"original": original,
			}, nil)
			localRawDir = strings.TrimSpace(localRawDir)
			if err != nil {
				return nil, errW(err, "redirect import error",
					reason("execute local dir template error"),
					kv("scope", c.scope),
					kv("redirect.prefix", redirect.Prefix),
					kv("redirect.local.dir", redirect.Local.Dir),
				)
			}
			return c.makeLocalImport(originalImport, originalImport.Registry, localRawDir)
		} else if redirect.Git != nil {
			// TODO: template data
			gitRawUrl, err := executeStringTemplate(redirect.Git.Url, map[string]any{
				"path":     redirectPath,
				"original": original,
			}, nil)
			if err != nil {
				return nil, errW(err, "redirect import error",
					reason("execute git url template error"),
					kv("scope", c.scope),
					kv("redirect.prefix", redirect.Prefix),
					kv("redirect.git.url", redirect.Git.Url),
				)
			}
			gitRawUrl = strings.TrimSpace(gitRawUrl)
			gitRawRef := t(originalImport.Git != nil, originalImport.Git.RawRef, redirect.Git.Ref)
			gitRawRef = t(gitRawRef != "", gitRawRef, "main")
			return c.makeGitImport(originalImport, originalImport.Registry, gitRawUrl, nil, gitRawRef, nil)
		} else {
			impossible()
		}
	}
	return originalImport, nil
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
