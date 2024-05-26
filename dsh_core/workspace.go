package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"os"
	"path/filepath"
)

type Workspace struct {
	logger                 *dsh_utils.Logger
	path                   string
	manifest               *workspaceManifest
	projectManifestsByPath map[string]*projectManifest
	projectManifestsByName map[string]*projectManifest
}

type WorkspaceOpenAppSettings struct {
	ProfilePaths []string
	Options      map[string]string
}

type WorkspaceCleanSettings struct {
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
		logger:                 logger,
		path:                   path,
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

func (w *Workspace) PrepareLocalApp(projectPath string, profilePaths []string) (*AppProfile, error) {
	m, err := w.loadDirProjectManifest(projectPath)
	if err != nil {
		return nil, errW(err, "prepare local app error",
			reason("load project manifest error"),
			kv("projectPath", projectPath),
			kv("profilePaths", profilePaths),
		)
	}
	return loadAppProfile(w, m, profilePaths)
}

func (w *Workspace) PrepareGitApp(projectUrl string, projectRef string, profilePaths []string) (*AppProfile, error) {
	m, err := w.loadGitProjectManifest("", projectUrl, nil, projectRef, nil)
	if err != nil {
		return nil, errW(err, "prepare git app error",
			reason("load git project manifest error"),
			kv("projectUrl", projectUrl),
			kv("projectRef", projectRef),
			kv("profilePaths", profilePaths),
		)
	}
	return loadAppProfile(w, m, profilePaths)
}

func (w *Workspace) Clean(settings WorkspaceCleanSettings) error {
	return w.cleanOutputDir(settings.ExcludeOutputPath)
}

func (w *Workspace) loadProjectManifest(link *projectResolvedLink) (manifest *projectManifest, err error) {
	if link.Git != nil {
		return w.loadGitProjectManifest(link.Path, link.Git.RawUrl, link.Git.parsedUrl, link.Git.RawRef, link.Git.parsedRef)
	} else {
		return w.loadDirProjectManifest(link.Path)
	}
}

func (w *Workspace) loadDirProjectManifest(path string) (manifest *projectManifest, err error) {
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

func (w *Workspace) loadGitProjectManifest(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *ProjectLinkGitRef) (manifest *projectManifest, err error) {
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
		if parsedRef, err = parseProjectLinkGitRef(rawRef); err != nil {
			return nil, errW(err, "load git project manifest error",
				reason("parse ref error"),
				kv("url", rawUrl),
				kv("ref", rawRef),
			)
		}
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
	manifest, err = w.loadDirProjectManifest(path)
	if err != nil {
		return nil, errW(err, "load git project manifest error",
			reason("load manifest error"),
			kv("url", rawUrl),
			kv("ref", rawRef),
		)
	}
	return manifest, nil
}
