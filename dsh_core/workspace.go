package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"os"
	"path/filepath"
)

type Workspace struct {
	path                   string
	logger                 *dsh_utils.Logger
	manifest               *workspaceManifest
	projectManifestsByPath map[string]*projectManifest
	projectManifestsByName map[string]*projectManifest
}

func GetWorkspaceDefaultPath() string {
	if path, exist := os.LookupEnv("DSH_WORKSPACE"); exist {
		return path
	}
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, "dsh")
	}
	return filepath.Join(os.TempDir(), "dsh")
}

func OpenWorkspace(path string, logger *dsh_utils.Logger) (w *Workspace, err error) {
	if path == "" {
		path = GetWorkspaceDefaultPath()
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, errW(err, "open workspace error",
			reason("get abs-path error"),
			kv("path", path),
		)
	}
	logger.InfoDesc("open workspace", kv("path", path))
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, errW(err, "open workspace error",
			reason("make dir error"),
			kv("path", path),
		)
	}
	manifest, err := loadWorkspaceManifest(path)
	if err != nil {
		return nil, errW(err, "open workspace error",
			reason("load manifest error"),
			kv("path", path),
		)
	}
	w = &Workspace{
		path:                   path,
		logger:                 logger,
		manifest:               manifest,
		projectManifestsByPath: make(map[string]*projectManifest),
		projectManifestsByName: make(map[string]*projectManifest),
	}
	return w, nil
}

func (w *Workspace) GetPath() string {
	return w.path
}

func (w *Workspace) loadProjectManifest(path string) (pm *projectManifest, err error) {
	if !dsh_utils.IsDirExists(path) {
		return nil, errN("load project manifest error",
			reason("project dir not exists"),
			kv("path", path),
		)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, errW(err, "load project manifest error",
			reason("get abs-path error"),
			kv("path", path),
		)
	}
	if pm, exist := w.projectManifestsByPath[path]; exist {
		return pm, nil
	}
	w.logger.DebugDesc("load project manifest", kv("path", path))
	if pm, err = loadProjectManifest(path); err != nil {
		return nil, err
	}
	if existManifest, exist := w.projectManifestsByName[pm.Name]; exist {
		if existManifest.projectPath != pm.projectPath {
			return nil, errN("load project manifest error",
				reason("project name duplicated"),
				kv("projectName", pm.Name),
				kv("projectPath1", pm.projectPath),
				kv("projectPath2", existManifest.projectPath),
			)
		}
	}
	w.projectManifestsByPath[pm.projectPath] = pm
	w.projectManifestsByName[pm.Name] = pm
	return pm, nil
}

func (w *Workspace) loadGitProjectManifest(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) (pm *projectManifest, err error) {
	if parsedUrl == nil {
		if parsedUrl, err = url.Parse(rawUrl); err != nil {
			return nil, errW(err, "load git project manifest error",
				reason("parse url error"),
				kv("url", rawUrl),
				kv("ref", rawRef),
			)
		}
	}
	if parsedRef == nil {
		parsedRef = parseGitRef(rawRef)
	}
	if path == "" {
		path = w.getGitProjectPath(parsedUrl, parsedRef)
	}
	if err = w.downloadGitProject(path, rawUrl, parsedUrl, rawRef, parsedRef); err != nil {
		return nil, errW(err, "load git project manifest error",
			reason("download project error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	pm, err = w.loadProjectManifest(path)
	if err != nil {
		return nil, errW(err, "load git project manifest error",
			reason("load manifest error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	return pm, nil
}

func (w *Workspace) OpenLocalApp(path string, options map[string]string) (app *App, err error) {
	m, err := w.loadProjectManifest(path)
	if err != nil {
		return nil, errW(err, "open local app error",
			reason("load project manifest error"),
			kv("path", path),
			kv("options", options),
		)
	}
	app, err = loadApp(w, m, options)
	if err != nil {
		return nil, errW(err, "open local app error",
			reason("load app error"),
			kv("path", path),
			kv("options", options),
		)
	}
	return app, nil
}

func (w *Workspace) OpenGitApp(url string, ref string, options map[string]string) (app *App, err error) {
	m, err := w.loadGitProjectManifest("", url, nil, ref, nil)
	if err != nil {
		return nil, errW(err, "open git app error",
			reason("load git project manifest error"),
			kv("url", url),
			kv("ref", ref),
			kv("options", options),
		)
	}
	app, err = loadApp(w, m, options)
	if err != nil {
		return nil, errW(err, "open git app error",
			reason("load app error"),
			kv("url", url),
			kv("ref", ref),
			kv("options", options),
		)
	}
	return app, nil
}
