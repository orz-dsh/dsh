package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
)

type projectImport struct {
	context      *appContext
	reference    *projectManifest
	importType   projectImportType
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

type projectImportType string

const (
	projectImportTypeLocal projectImportType = "local"
	projectImportTypeGit   projectImportType = "git"
)

func newLocalProjectImport(context *appContext, reference *projectManifest, projectPath string) *projectImport {
	return &projectImport{
		context:     context,
		reference:   reference,
		importType:  projectImportTypeLocal,
		unique:      projectPath,
		projectPath: projectPath,
	}
}

func newGitProjectImport(context *appContext, reference *projectManifest, projectPath string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) *projectImport {
	return &projectImport{
		context:      context,
		reference:    reference,
		importType:   projectImportTypeGit,
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
		if i.importType == projectImportTypeLocal {
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

func loadProjectImportContainer(context *appContext, manifest *projectManifest, scope projectImportScope) (ic *projectImportContainer, err error) {
	var imports []*projectManifestImport
	if scope == projectImportScopeScript {
		imports = manifest.Script.Imports
	} else if scope == projectImportScopeConfig {
		imports = manifest.Config.Imports
	} else {
		panic(desc("invalid import scope", kv("scope", scope)))
	}
	ic = &projectImportContainer{
		context:         context,
		scope:           scope,
		importsByUnique: make(map[string]*projectImport),
	}
	for i := 0; i < len(imports); i++ {
		imp := imports[i]
		if imp.Local != nil && imp.Local.Dir != "" {
			if imp.Match != "" {
				matched, err := context.option.evalProjectMatchExpr(manifest, imp.match)
				if err != nil {
					return nil, err
				}
				if !matched {
					continue
				}
			}
			if err = ic.addLocalImport(context, imp.Local.Dir, manifest); err != nil {
				return nil, err
			}
		} else if imp.Git != nil && imp.Git.Url != "" && imp.Git.Ref != "" {
			if imp.Match != "" {
				matched, err := context.option.evalProjectMatchExpr(manifest, imp.match)
				if err != nil {
					return nil, err
				}
				if !matched {
					continue
				}
			}
			if err = ic.addGitImport(context, manifest, imp.Git.Url, imp.Git.url, imp.Git.Ref, imp.Git.ref); err != nil {
				return nil, err
			}
		}
	}
	return ic, nil
}

func (c *projectImportContainer) addLocalImport(context *appContext, path string, reference *projectManifest) (err error) {
	if !dsh_utils.IsDirExists(path) {
		return errN("add local import error",
			reason("dir not exists"),
			kv("scope", c.scope),
			kv("path", path),
		)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return errW(err, "add local import error",
			reason("get abs-path error"),
			kv("scope", c.scope),
			kv("path", path),
		)
	}
	if path == reference.projectPath {
		return nil
	}
	imp := newLocalProjectImport(context, reference, path)
	if _, exist := c.importsByUnique[imp.unique]; !exist {
		c.imports = append(c.imports, imp)
		c.importsByUnique[imp.unique] = imp
	}
	return nil
}

func (c *projectImportContainer) addGitImport(context *appContext, reference *projectManifest, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) error {
	path := context.workspace.getGitProjectPath(parsedUrl, parsedRef)
	if path == reference.projectPath {
		return nil
	}
	imp := newGitProjectImport(context, reference, path, rawUrl, parsedUrl, rawRef, parsedRef)
	if _, exist := c.importsByUnique[imp.unique]; !exist {
		c.imports = append(c.imports, imp)
		c.importsByUnique[imp.unique] = imp
	}
	return nil
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
