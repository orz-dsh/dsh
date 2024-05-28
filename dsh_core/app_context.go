package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
)

type appContext struct {
	logger                 *dsh_utils.Logger
	evaluator              *Evaluator
	workspace              *Workspace
	manifest               *projectManifest
	profile                *appProfile
	Option                 *appOption
	projectsByName         map[string]*project
	projectManifestsByPath map[string]*projectManifest
	projectManifestsByName map[string]*projectManifest
}

func makeAppContext(workspace *Workspace, evaluator *Evaluator, profile *appProfile, link *projectResolvedLink) (*appContext, error) {
	context := &appContext{
		logger:                 workspace.logger,
		workspace:              workspace,
		evaluator:              evaluator,
		profile:                profile,
		projectsByName:         make(map[string]*project),
		projectManifestsByPath: make(map[string]*projectManifest),
		projectManifestsByName: make(map[string]*projectManifest),
	}

	manifest, err := context.loadProjectManifest(link)
	if err != nil {
		return nil, err
	}
	context.manifest = manifest

	options, err := profile.getProjectOptions()
	if err != nil {
		return nil, err
	}
	option, err := loadAppOption(context, manifest, options)
	if err != nil {
		return nil, err
	}
	context.Option = option

	return context, nil
}

func (c *appContext) loadProject(manifest *projectManifest) (p *project, err error) {
	if p, exist := c.projectsByName[manifest.Name]; exist {
		return p, nil
	}
	if p, err = loadProject(c, manifest); err != nil {
		return nil, err
	}
	c.projectsByName[manifest.Name] = p
	return p, nil
}

func (c *appContext) loadMainProject() (p *project, err error) {
	return c.loadProject(c.manifest)
}

func (c *appContext) isMainProject(manifest *projectManifest) bool {
	return c.manifest.Name == manifest.Name
}

func (c *appContext) loadProjectManifest(link *projectResolvedLink) (manifest *projectManifest, err error) {
	if link.Git != nil {
		return c.loadGitProjectManifest(link.Path, link.Git.Url, link.Git.parsedUrl, link.Git.Ref, link.Git.parsedRef)
	} else {
		return c.loadDirProjectManifest(link.Path)
	}
}

func (c *appContext) loadDirProjectManifest(path string) (manifest *projectManifest, err error) {
	if !dsh_utils.IsDirExists(path) {
		return nil, errN("load project manifest error",
			reason("project dir not exists"),
			kv("path", path),
		)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("get abs-path error"),
			kv("path", path),
		)
	}
	path = absPath
	if m, exist := c.projectManifestsByPath[path]; exist {
		return m, nil
	}
	c.logger.DebugDesc("load project manifest", kv("path", path))
	if manifest, err = loadProjectManifest(path); err != nil {
		return nil, err
	}
	if existManifest, exist := c.projectManifestsByName[manifest.Name]; exist {
		if existManifest.projectPath != manifest.projectPath {
			return nil, errN("load project manifest error",
				reason("project name duplicated"),
				kv("projectName", manifest.Name),
				kv("projectPath1", manifest.projectPath),
				kv("projectPath2", existManifest.projectPath),
			)
		}
	}
	c.projectManifestsByPath[manifest.projectPath] = manifest
	c.projectManifestsByName[manifest.Name] = manifest
	return manifest, nil
}

func (c *appContext) loadGitProjectManifest(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *ProjectLinkGitRef) (manifest *projectManifest, err error) {
	if parsedUrl == nil {
		if parsedUrl, err = url.Parse(rawUrl); err != nil {
			return nil, errW(err, "load project manifest error",
				reason("parse url error"),
				kv("url", rawUrl),
				kv("ref", rawRef),
			)
		}
	}
	if parsedRef == nil {
		if parsedRef, err = parseProjectLinkGitRef(rawRef); err != nil {
			return nil, errW(err, "load project manifest error",
				reason("parse ref error"),
				kv("url", rawUrl),
				kv("ref", rawRef),
			)
		}
	}
	if path == "" {
		path = c.workspace.getGitProjectPath(parsedUrl, parsedRef)
	}
	if err = c.workspace.downloadGitProject(path, rawUrl, parsedUrl, rawRef, parsedRef); err != nil {
		return nil, errW(err, "load project manifest error",
			reason("download project error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	manifest, err = c.loadDirProjectManifest(path)
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("load manifest error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	return manifest, nil
}
