package dsh_core

import (
	"dsh/dsh_utils"
	"os"
	"path/filepath"
)

type Workspace struct {
	global           *Global
	logger           *Logger
	dir              string
	setting          *workspaceSetting
	evaluator        *Evaluator
	profileManifests []*ProfilePref
}

type WorkspaceCleanSetting struct {
	ExcludeOutputPath string
}

func MakeWorkspace(global *Global, dir string) (workspace *Workspace, err error) {
	if dir == "" {
		dir = getWorkspaceDirDefault(global.systemInfo.HomeDir)
	}

	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, errW(err, "make workspace error",
			reason("get abs-path error"),
			kv("dir", dir),
		)
	}
	dir = absPath

	global.logger.InfoDesc("make workspace", kv("dir", dir))
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, errW(err, "make workspace error",
			reason("make dir error"),
			kv("dir", dir),
		)
	}
	setting, err := loadWorkspaceSetting(dir)
	if err != nil {
		return nil, errW(err, "make workspace error",
			reason("load setting error"),
			kv("dir", dir),
		)
	}

	evaluator := dsh_utils.NewEvaluator().SetData("global", global.variables).SetData("local", map[string]any{
		"os":                   global.systemInfo.Os,
		"arch":                 global.systemInfo.Arch,
		"hostname":             global.systemInfo.Hostname,
		"username":             global.systemInfo.Username,
		"home_dir":             global.systemInfo.HomeDir,
		"working_dir":          global.systemInfo.WorkingDir,
		"workspace_dir":        dir,
		"runtime_version":      dsh_utils.GetRuntimeVersion(),
		"runtime_version_code": dsh_utils.GetRuntimeVersionCode(),
	})

	profiles, err := setting.Profile.getFiles(evaluator)
	if err != nil {
		return nil, err
	}
	var profileManifests []*ProfilePref
	for i := 0; i < len(profiles); i++ {
		profileManifest, err := loadProfilePref(profiles[i])
		if err != nil {
			return nil, err
		}
		profileManifests = append(profileManifests, profileManifest)
	}

	workspace = &Workspace{
		global:           global,
		logger:           global.logger,
		dir:              dir,
		setting:          setting,
		evaluator:        evaluator,
		profileManifests: profileManifests,
	}
	return workspace, nil
}

func getWorkspaceDirDefault(homeDir string) string {
	if path, exist := os.LookupEnv("DSH_WORKSPACE"); exist {
		return path
	}
	if homeDir != "" {
		return filepath.Join(homeDir, "dsh")
	}
	return filepath.Join(os.TempDir(), "dsh")
}

func (w *Workspace) DescExtraKeyValues() KVS {
	return KVS{
		kv("global", w.global),
		kv("dir", w.dir),
		kv("setting", w.setting),
	}
}

func (w *Workspace) GetDir() string {
	return w.dir
}

func (w *Workspace) NewAppMaker() *AppMaker {
	return newAppMaker(w)
}

func (w *Workspace) Clean(setting WorkspaceCleanSetting) error {
	return w.cleanOutputDir(setting.ExcludeOutputPath)
}
