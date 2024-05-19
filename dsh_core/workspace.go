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

type WorkspaceCleanOptions struct {
	ExcludeOutputPath string
}

func OpenWorkspace(path string, logger *dsh_utils.Logger) (workspace *Workspace, err error) {
	if path == "" {
		path = getWorkspaceDefaultPath()
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, errW(err, "open workspace error",
			reason("get abs-path error"),
			kv("path", path),
		)
	}
	path = absPath
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
	workspace = &Workspace{
		path:                   path,
		logger:                 logger,
		manifest:               manifest,
		projectManifestsByPath: make(map[string]*projectManifest),
		projectManifestsByName: make(map[string]*projectManifest),
	}
	return workspace, nil
}

func getWorkspaceDefaultPath() string {
	if path, exist := os.LookupEnv("DSH_WORKSPACE"); exist {
		return path
	}
	if homeDir, err := os.UserHomeDir(); err == nil {
		return filepath.Join(homeDir, "dsh")
	}
	return filepath.Join(os.TempDir(), "dsh")
}

func (w *Workspace) DescExtraKeyValues() KVS {
	return KVS{
		kv("path", w.path),
		kv("manifest", w.manifest),
	}
}

func (w *Workspace) GetPath() string {
	return w.path
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

func (w *Workspace) Clean(options WorkspaceCleanOptions) error {
	return w.cleanOutputDir(options.ExcludeOutputPath)
}

func (w *Workspace) loadProjectManifest(path string) (manifest *projectManifest, err error) {
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
	if m, exist := w.projectManifestsByPath[path]; exist {
		return m, nil
	}
	w.logger.DebugDesc("load project manifest", kv("path", path))
	if manifest, err = loadProjectManifest(path); err != nil {
		return nil, err
	}
	if existManifest, exist := w.projectManifestsByName[manifest.Name]; exist {
		if existManifest.projectPath != manifest.projectPath {
			return nil, errN("load project manifest error",
				reason("project name duplicated"),
				kv("projectName", manifest.Name),
				kv("projectPath1", manifest.projectPath),
				kv("projectPath2", existManifest.projectPath),
			)
		}
	}
	w.projectManifestsByPath[manifest.projectPath] = manifest
	w.projectManifestsByName[manifest.Name] = manifest
	return manifest, nil
}

func (w *Workspace) loadGitProjectManifest(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) (manifest *projectManifest, err error) {
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
	manifest, err = w.loadProjectManifest(path)
	if err != nil {
		return nil, errW(err, "load git project manifest error",
			reason("load manifest error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	return manifest, nil
}
