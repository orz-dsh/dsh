package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"path/filepath"
)

type appContext struct {
	systemInfo             *SystemInfo
	logger                 *Logger
	evaluator              *Evaluator
	workspace              *Workspace
	manifest               *ProjectManifest
	profile                *appProfile
	Option                 *appOption
	projectsByName         map[string]*Project
	projectManifestsByPath map[string]*ProjectManifest
	projectManifestsByName map[string]*ProjectManifest
}

func makeAppContext(workspace *Workspace, profile *appProfile, link *projectResolvedLink) (*appContext, error) {
	context := &appContext{
		systemInfo:             workspace.systemInfo,
		logger:                 workspace.logger,
		workspace:              workspace,
		evaluator:              profile.evaluator,
		profile:                profile,
		projectsByName:         map[string]*Project{},
		projectManifestsByPath: map[string]*ProjectManifest{},
		projectManifestsByName: map[string]*ProjectManifest{},
	}

	manifest, err := context.loadProjectManifest(link)
	if err != nil {
		return nil, err
	}
	context.manifest = manifest

	optionSpecifyItems, err := profile.getProjectOptionSpecifyItems()
	if err != nil {
		return nil, err
	}
	context.Option = newAppOption(context, manifest, optionSpecifyItems)

	return context, nil
}

func (c *appContext) loadProject(manifest *ProjectManifest) (project *Project, err error) {
	if existProject, exist := c.projectsByName[manifest.projectName]; exist {
		return existProject, nil
	}
	if project, err = makeProject(c, manifest); err != nil {
		return nil, err
	}
	c.projectsByName[manifest.projectName] = project
	return project, nil
}

func (c *appContext) loadMainProject() (p *Project, err error) {
	return c.loadProject(c.manifest)
}

func (c *appContext) isMainProject(manifest *ProjectManifest) bool {
	return c.manifest.projectName == manifest.projectName
}

func (c *appContext) loadProjectManifest(link *projectResolvedLink) (manifest *ProjectManifest, err error) {
	if link.Git != nil {
		return c.loadGitProjectManifest(link.Path, link.Git.Url, link.Git.parsedUrl, link.Git.Ref, link.Git.parsedRef)
	} else {
		return c.loadDirProjectManifest(link.Path)
	}
}

func (c *appContext) loadDirProjectManifest(path string) (manifest *ProjectManifest, err error) {
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
	if existManifest, exist := c.projectManifestsByName[manifest.projectName]; exist {
		if existManifest.projectPath != manifest.projectPath {
			return nil, errN("load project manifest error",
				reason("project name duplicated"),
				kv("projectName", manifest.projectName),
				kv("projectPath1", manifest.projectPath),
				kv("projectPath2", existManifest.projectPath),
			)
		}
	}
	c.projectManifestsByPath[manifest.projectPath] = manifest
	c.projectManifestsByName[manifest.projectName] = manifest
	return manifest, nil
}

func (c *appContext) loadGitProjectManifest(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *ProjectLinkGitRef) (manifest *ProjectManifest, err error) {
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
