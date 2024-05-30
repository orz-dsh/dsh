package dsh_core

import (
	"dsh/dsh_utils"
	"os"
	"path/filepath"
)

type Workspace struct {
	systemInfo       *SystemInfo
	logger           *Logger
	path             string
	manifest         *workspaceManifest
	evaluator        *Evaluator
	profileManifests []*ProfileManifest
}

type WorkspaceCleanSettings struct {
	ExcludeOutputPath string
}

func OpenWorkspace(path string, logger *Logger) (workspace *Workspace, err error) {
	systemInfo, err := dsh_utils.GetSystemInfo()
	if err != nil {
		return nil, errW(err, "open workspace error",
			reason("get system info error"),
			kv("path", path),
		)
	}
	if path == "" {
		path = getWorkspacePathDefault(systemInfo)
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

	evaluator := dsh_utils.NewEvaluator().SetData("local", map[string]any{
		"os":                   systemInfo.Os,
		"arch":                 systemInfo.Arch,
		"hostname":             systemInfo.Hostname,
		"username":             systemInfo.Username,
		"home_dir":             systemInfo.HomeDir,
		"working_dir":          systemInfo.WorkingDir,
		"workspace_dir":        path,
		"runtime_version":      dsh_utils.GetRuntimeVersion(),
		"runtime_version_code": dsh_utils.GetRuntimeVersionCode(),
	})

	profiles, err := manifest.Profile.entities.getFiles(evaluator)
	if err != nil {
		return nil, err
	}
	var profileManifests []*ProfileManifest
	for i := 0; i < len(profiles); i++ {
		profileManifest, err := loadProfileManifest(profiles[i])
		if err != nil {
			return nil, err
		}
		profileManifests = append(profileManifests, profileManifest)
	}

	workspace = &Workspace{
		systemInfo:       systemInfo,
		logger:           logger,
		path:             path,
		manifest:         manifest,
		evaluator:        evaluator,
		profileManifests: profileManifests,
	}
	return workspace, nil
}

func getWorkspacePathDefault(systemInfo *SystemInfo) string {
	if path, exist := os.LookupEnv("DSH_WORKSPACE"); exist {
		return path
	}
	if systemInfo.HomeDir != "" {
		return filepath.Join(systemInfo.HomeDir, "dsh")
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
