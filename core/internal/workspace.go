package internal

import (
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"os"
	"path/filepath"
)

// region WorkspaceCore

type WorkspaceCore struct {
	Dir             string
	Logger          *Logger
	SystemInfo      *SystemInfo
	Evaluator       *Evaluator
	Setting         *WorkspaceSetting
	ProfileSettings []*ProfileSetting
}

func NewWorkspaceCore(dir string, logger *Logger, systemInfo *SystemInfo, variables map[string]any) (core *WorkspaceCore, err error) {
	if dir == "" {
		dir = getWorkspaceDirDefault(systemInfo.HomeDir)
	}

	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, ErrW(err, "make workspace error",
			Reason("get abs-path error"),
			KV("dir", dir),
		)
	}
	dir = absPath

	logger.InfoDesc("make workspace", KV("dir", dir))
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, ErrW(err, "make workspace error",
			Reason("make dir error"),
			KV("dir", dir),
		)
	}
	setting, err := LoadWorkspaceSetting(logger, dir)
	if err != nil {
		return nil, ErrW(err, "make workspace error",
			Reason("load setting error"),
			KV("dir", dir),
		)
	}

	evaluator := NewEvaluator().SetData("global", variables).SetData("local", map[string]any{
		"os":                   systemInfo.Os,
		"arch":                 systemInfo.Arch,
		"hostname":             systemInfo.Hostname,
		"username":             systemInfo.Username,
		"home_dir":             systemInfo.HomeDir,
		"working_dir":          systemInfo.WorkingDir,
		"workspace_dir":        dir,
		"runtime_version":      GetRuntimeVersion(),
		"runtime_version_code": GetRuntimeVersionCode(),
	})

	profiles, err := setting.Profile.GetFiles(evaluator)
	if err != nil {
		return nil, err
	}
	var profileSettings []*ProfileSetting
	for i := 0; i < len(profiles); i++ {
		profileSetting, err := LoadProfileSetting(logger, profiles[i])
		if err != nil {
			return nil, err
		}
		profileSettings = append(profileSettings, profileSetting)
	}

	core = &WorkspaceCore{
		Dir:             dir,
		Logger:          logger,
		SystemInfo:      systemInfo,
		Evaluator:       evaluator,
		Setting:         setting,
		ProfileSettings: profileSettings,
	}
	return core, nil
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

// endregion
