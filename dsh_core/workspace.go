package dsh_core

import (
	"dsh/dsh_utils"
	"os"
	"path/filepath"
)

type Workspace struct {
	logger   *dsh_utils.Logger
	path     string
	manifest *workspaceManifest
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
		logger:   logger,
		path:     path,
		manifest: manifest,
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

func (w *Workspace) MakeAppFactory() (*AppFactory, error) {
	return makeAppFactory(w)
}

func (w *Workspace) Clean(settings WorkspaceCleanSettings) error {
	return w.cleanOutputDir(settings.ExcludeOutputPath)
}
