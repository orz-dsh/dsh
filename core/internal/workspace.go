package internal

import (
	. "github.com/orz-dsh/dsh/core/inspection"
	. "github.com/orz-dsh/dsh/core/internal/setting"
	. "github.com/orz-dsh/dsh/utils"
	"os"
	"path/filepath"
)

// region WorkspaceCore

type WorkspaceCore struct {
	Dir             string
	Environment     *EnvironmentCore
	Logger          *Logger
	Evaluator       *Evaluator
	Setting         *WorkspaceSetting
	ProfileSettings []*ProfileSetting
}

func NewWorkspaceCore(environment *EnvironmentCore, dir string) (core *WorkspaceCore, err error) {
	if dir == "" {
		dir = environment.GetWorkspaceDir()
	}

	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, ErrW(err, "make workspace error",
			Reason("get abs-path error"),
			KV("dir", dir),
		)
	}
	dir = absPath

	environment.Logger.InfoDesc("make workspace", KV("dir", dir))
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return nil, ErrW(err, "make workspace error",
			Reason("make dir error"),
			KV("dir", dir),
		)
	}
	setting, err := LoadWorkspaceSetting(environment.Logger, dir)
	if err != nil {
		return nil, ErrW(err, "make workspace error",
			Reason("load setting error"),
			KV("dir", dir),
		)
	}

	setting.Merge(environment.Setting.Workspace.GetWorkspaceSetting())
	setting.MergeDefault()

	evaluator := environment.Evaluator.MergeData("local", map[string]any{
		"workspace_dir": dir,
	})

	profiles, err := setting.Profile.GetFiles(evaluator)
	if err != nil {
		return nil, err
	}
	var profileSettings []*ProfileSetting
	for i := 0; i < len(profiles); i++ {
		profileSetting, err := LoadProfileSetting(environment.Logger, profiles[i])
		if err != nil {
			return nil, err
		}
		profileSettings = append(profileSettings, profileSetting)
	}

	core = &WorkspaceCore{
		Dir:             dir,
		Environment:     environment,
		Logger:          environment.Logger,
		Evaluator:       evaluator,
		Setting:         setting,
		ProfileSettings: profileSettings,
	}
	return core, nil
}

func (w *WorkspaceCore) Inspect() *WorkspaceInspection {
	return NewWorkspaceInspection(w.Dir, w.Setting.Inspect())
}

// endregion
