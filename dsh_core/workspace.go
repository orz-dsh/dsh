package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"os"
	"path/filepath"
)

type Workspace struct {
	Path               string
	Logger             *dsh_utils.Logger
	ProjectInfoPathMap map[string]*ProjectInfo
	ProjectInfoNameMap map[string]*ProjectInfo
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
		Path:               path,
		Logger:             logger,
		ProjectInfoPathMap: make(map[string]*ProjectInfo),
		ProjectInfoNameMap: make(map[string]*ProjectInfo),
	}
	return workspace, nil
}

func (workspace *Workspace) LoadLocalProjectInfo(path string) (info *ProjectInfo, err error) {
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

	if info, exist := workspace.ProjectInfoPathMap[path]; exist {
		return info, nil
	}

	workspace.Logger.Debug("load project info: path=%s", path)
	if info, err = LoadProjectInfo(workspace, path); err != nil {
		return nil, err
	}

	if existProject, exist := workspace.ProjectInfoNameMap[info.Name]; exist {
		if existProject.Path != info.Path {
			return nil, dsh_utils.NewError("project name is duplicated", map[string]any{
				"projectName":  info.Name,
				"projectPath1": info.Path,
				"projectPath2": existProject.Path,
			})
		}
	}

	workspace.ProjectInfoPathMap[info.Path] = info
	workspace.ProjectInfoNameMap[info.Name] = info

	return info, nil
}

func (workspace *Workspace) LoadGitProjectInfo(path string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *GitRef) (info *ProjectInfo, err error) {
	if parsedUrl == nil {
		if parsedUrl, err = url.Parse(rawUrl); err != nil {
			return nil, dsh_utils.WrapError(err, "project git url parse failed", map[string]any{
				"url": rawUrl,
			})
		}
	}
	if parsedRef == nil {
		parsedRef = ParseGitRef(rawRef)
	}
	if path == "" {
		path = workspace.GetGitProjectPath(parsedUrl, parsedRef)
	}
	if err = workspace.DownloadGitProject(path, rawUrl, parsedUrl, rawRef, parsedRef); err != nil {
		return nil, err
	}
	return workspace.LoadLocalProjectInfo(path)
}

func (workspace *Workspace) OpenLocalProject(context *Context, path string) (*Project, error) {
	info, err := workspace.LoadLocalProjectInfo(path)
	if err != nil {
		return nil, err
	}
	return OpenProject(context, info)
}

func (workspace *Workspace) OpenGitProject(context *Context, url string, ref string) (*Project, error) {
	info, err := workspace.LoadGitProjectInfo("", url, nil, ref, nil)
	if err != nil {
		return nil, err
	}
	return OpenProject(context, info)
}
