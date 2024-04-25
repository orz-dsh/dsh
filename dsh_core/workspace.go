package dsh_core

import (
	"dsh/dsh_utils"
	"net/url"
	"os"
	"path/filepath"
)

type Workspace struct {
	Path           string
	ProjectPathMap map[string]*Project
	ProjectNameMap map[string]*Project
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

func LoadWorkspace(path string) (workspace *Workspace, err error) {
	if path == "" {
		path = GetWorkspaceDefaultPath()
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return nil, dsh_utils.WrapError(err, "workspace abs-path get failed", map[string]interface{}{
			"path": path,
		})
	}
	err = os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, dsh_utils.WrapError(err, "workspace dir make failed", map[string]interface{}{
			"path": path,
		})
	}
	workspace = &Workspace{
		Path:           path,
		ProjectPathMap: make(map[string]*Project),
		ProjectNameMap: make(map[string]*Project),
	}
	return workspace, nil
}

func (w *Workspace) LoadLocalProject(path string) (project *Project, err error) {
	if !dsh_utils.IsDirExists(path) {
		return nil, dsh_utils.NewError("project dir is not exists", map[string]interface{}{
			"path": path,
		})
	}

	projectPath, err := filepath.Abs(path)
	if err != nil {
		return nil, dsh_utils.WrapError(err, "project abs-path get failed", map[string]interface{}{
			"path": path,
		})
	}

	if project, exist := w.ProjectPathMap[projectPath]; exist {
		return project, nil
	}

	project = NewProject(w, projectPath)

	var setups []func() error

	manifestYamlPath := filepath.Join(projectPath, "project.yml")
	if !dsh_utils.IsFileExists(manifestYamlPath) {
		manifestYamlPath = filepath.Join(projectPath, "project.yaml")
		if !dsh_utils.IsFileExists(manifestYamlPath) {
			manifestYamlPath = ""
		}
	}
	if manifestYamlPath != "" {
		manifest := &Manifest{}
		if err = manifest.LoadYaml(manifestYamlPath); err != nil {
			return nil, dsh_utils.WrapError(err, "project manifest load failed", map[string]interface{}{
				"projectPath":      projectPath,
				"manifestYamlPath": manifestYamlPath,
			})
		}
		if err = manifest.PreCheck(project); err != nil {
			return nil, dsh_utils.WrapError(err, "project manifest pre-check failed", map[string]interface{}{
				"projectPath":      projectPath,
				"manifestYamlPath": manifestYamlPath,
			})
		}
		setups = append(setups, func() error {
			return manifest.Setup(project)
		})
	} else {
		return nil, dsh_utils.NewError("project manifest file not found", map[string]interface{}{
			"projectPath": projectPath,
		})
	}

	if existProject, exist := w.ProjectNameMap[project.Name]; exist {
		if existProject.Path != project.Path {
			return nil, dsh_utils.NewError("project name is duplicated", map[string]interface{}{
				"projectName":  project.Name,
				"projectPath1": project.Path,
				"projectPath2": existProject.Path,
			})
		}
	}

	for i := 0; i < len(setups); i++ {
		if err = setups[i](); err != nil {
			return nil, err
		}
	}

	w.ProjectPathMap[project.Path] = project
	w.ProjectNameMap[project.Name] = project

	return project, nil
}

func (w *Workspace) LoadGitProject(projectPath string, rawUrl string, parsedUrl *url.URL, rawRef string, parsedRef *GitRef) (project *Project, err error) {
	if parsedUrl == nil {
		if parsedUrl, err = url.Parse(rawUrl); err != nil {
			return nil, dsh_utils.WrapError(err, "project git url parse failed", map[string]interface{}{
				"url": rawUrl,
			})
		}
	}
	if parsedRef == nil {
		parsedRef = ParseGitRef(rawRef)
	}
	if projectPath == "" {
		projectPath = w.GetGitProjectPath(parsedUrl, parsedRef)
	}
	if err = w.DownloadGitProject(projectPath, rawUrl, parsedUrl, rawRef, parsedRef); err != nil {
		return nil, err
	}
	return w.LoadLocalProject(projectPath)
}
