package dsh_core

import (
	"dsh/dsh_utils"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Workspace struct {
	logger           *dsh_utils.Logger
	path             string
	manifest         *workspaceManifest
	evaluator        *Evaluator
	profileManifests []*AppProfileManifest
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

	workingDir, err := dsh_utils.GetWorkingDir()
	if err != nil {
		return nil, err
	}
	evaluator := dsh_utils.NewEvaluator().SetData("local", map[string]any{
		"working_dir":          workingDir,
		"workspace_dir":        path,
		"runtime_version":      dsh_utils.GetRuntimeVersion(),
		"runtime_version_code": dsh_utils.GetRuntimeVersionCode(),
		"os":                   strings.ToLower(runtime.GOOS),
	})

	profiles, err := manifest.Profile.definitions.getFiles(evaluator)
	if err != nil {
		return nil, err
	}
	var profileManifests []*AppProfileManifest
	for i := 0; i < len(profiles); i++ {
		profileManifest, err := loadAppProfileManifest(profiles[i])
		if err != nil {
			return nil, err
		}
		profileManifests = append(profileManifests, profileManifest)
	}

	workspace = &Workspace{
		logger:           logger,
		path:             path,
		manifest:         manifest,
		evaluator:        evaluator,
		profileManifests: profileManifests,
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

func (w *Workspace) NewAppFactory() *AppFactory {
	return newAppFactory(w)
}

func (w *Workspace) Clean(settings WorkspaceCleanSettings) error {
	return w.cleanOutputDir(settings.ExcludeOutputPath)
}
