package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"os"
	"path/filepath"
)

type Workspace struct {
	path               string
	logger             *dsh_utils.Logger
	projectInfoPathMap map[string]*projectInfo
	projectInfoNameMap map[string]*projectInfo
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
		path:               path,
		logger:             logger,
		projectInfoPathMap: make(map[string]*projectInfo),
		projectInfoNameMap: make(map[string]*projectInfo),
	}
	return workspace, nil
}

func (workspace *Workspace) GetPath() string {
	return workspace.path
}

func (workspace *Workspace) loadLocalProjectInfo(path string) (info *projectInfo, err error) {
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

	if info, exist := workspace.projectInfoPathMap[path]; exist {
		return info, nil
	}

	workspace.logger.Debug("load project info: path=%s", path)
	if info, err = loadProjectInfo(workspace, path); err != nil {
		return nil, err
	}

	if existProject, exist := workspace.projectInfoNameMap[info.name]; exist {
		if existProject.path != info.path {
			return nil, dsh_utils.NewError("project name is duplicated", map[string]any{
				"projectName":  info.name,
				"projectPath1": info.path,
				"projectPath2": existProject.path,
			})
		}
	}

	workspace.projectInfoPathMap[info.path] = info
	workspace.projectInfoNameMap[info.name] = info

	return info, nil
}

func (workspace *Workspace) loadGitProjectInfo(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *gitRef) (info *projectInfo, err error) {
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
	return workspace.loadLocalProjectInfo(path)
}

func (workspace *Workspace) OpenLocalProject(context *Context, path string) (*Project, error) {
	info, err := workspace.loadLocalProjectInfo(path)
	if err != nil {
		return nil, err
	}
	return openProject(context, info)
}

func (workspace *Workspace) OpenGitProject(context *Context, url string, ref string) (*Project, error) {
	info, err := workspace.loadGitProjectInfo("", url, nil, ref, nil)
	if err != nil {
		return nil, err
	}
	return openProject(context, info)
}
