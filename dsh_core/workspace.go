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
	projectManifestPathMap map[string]*projectManifest
	projectManifestNameMap map[string]*projectManifest
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

func OpenWorkspace(path string, logger *dsh_utils.Logger) (workspace *Workspace, err error) {
	if path == "" {
		path = GetWorkspaceDefaultPath()
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, dsh_utils.WrapError(err, "workspace abs-path get failed", map[string]any{
			"path": path,
		})
	}
	logger.Info("open workspace: path=%s", path)
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, dsh_utils.WrapError(err, "workspace dir make failed", map[string]any{
			"path": path,
		})
	}
	workspace = &Workspace{
		path:                   path,
		logger:                 logger,
		projectManifestPathMap: make(map[string]*projectManifest),
		projectManifestNameMap: make(map[string]*projectManifest),
	}
	return workspace, nil
}

func (workspace *Workspace) GetPath() string {
	return workspace.path
}

func (workspace *Workspace) loadLocalProjectManifest(path string) (manifest *projectManifest, err error) {
	if !dsh_utils.IsDirExists(path) {
		return nil, dsh_utils.NewError("project dir is not exists", map[string]any{
			"path": path,
		})
	}

	path, err = filepath.Abs(path)
	if err != nil {
		return nil, dsh_utils.WrapError(err, "project abs-path get failed", map[string]any{
			"path": path,
		})
	}

	if manifest, exist := workspace.projectManifestPathMap[path]; exist {
		return manifest, nil
	}

	workspace.logger.Debug("load project manifest: path=%s", path)
	if manifest, err = loadProjectManifest(path); err != nil {
		return nil, err
	}

	if existManifest, exist := workspace.projectManifestNameMap[manifest.Name]; exist {
		if existManifest.projectPath != manifest.projectPath {
			return nil, dsh_utils.NewError("project name is duplicated", map[string]any{
				"projectName":  manifest.Name,
				"projectPath1": manifest.projectPath,
				"projectPath2": existManifest.projectPath,
			})
		}
	}

	workspace.projectManifestPathMap[manifest.projectPath] = manifest
	workspace.projectManifestNameMap[manifest.Name] = manifest

	return manifest, nil
}

func (workspace *Workspace) loadGitProjectManifest(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) (manifest *projectManifest, err error) {
	if parsedUrl == nil {
		if parsedUrl, err = url.Parse(rawUrl); err != nil {
			return nil, dsh_utils.WrapError(err, "project git url parse failed", map[string]any{
				"url": rawUrl,
			})
		}
	}
	if parsedRef == nil {
		parsedRef = parseGitRef(rawRef)
	}
	if path == "" {
		path = workspace.getGitProjectPath(parsedUrl, parsedRef)
	}
	if err = workspace.downloadGitProject(path, rawUrl, parsedUrl, rawRef, parsedRef); err != nil {
		return nil, err
	}
	return workspace.loadLocalProjectManifest(path)
}

func (workspace *Workspace) OpenLocalProject(path string, optionValues map[string]string) (*Project, error) {
	manifest, err := workspace.loadLocalProjectManifest(path)
	if err != nil {
		return nil, err
	}
	return openProject(workspace, manifest, optionValues)
}

func (workspace *Workspace) OpenGitProject(url string, ref string, optionValues map[string]string) (*Project, error) {
	manifest, err := workspace.loadGitProjectManifest("", url, nil, ref, nil)
	if err != nil {
		return nil, err
	}
	return openProject(workspace, manifest, optionValues)
}
