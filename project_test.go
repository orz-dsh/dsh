package main

import (
	"dsh/dsh_core"
	"dsh/dsh_utils"
	"runtime"
	"testing"
)

func TestProject1(t *testing.T) {
	logger := dsh_utils.NewLogger(dsh_utils.LogLevelAll)
	workspace, err := dsh_core.OpenWorkspace("./.test_workspace", logger)
	if err != nil {
		logger.Panic("%+v", err)
	}
	profile, err := workspace.PrepareLocalApp("./.test1/app1", nil)
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = profile.AddManifestOptionValues(-1, map[string]string{
		"_os":  "linux",
		"test": "a",
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	app, err := profile.MakeApp()
	if err != nil {
		logger.Panic("%+v", err)
	}
	artifact, err := app.MakeScripts(dsh_core.AppMakeScriptsSettings{
		OutputPath:      "./.test1/app1/output",
		OutputPathClear: true,
		UseHardLink:     true,
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = workspace.Clean(dsh_core.WorkspaceCleanSettings{
		ExcludeOutputPath: artifact.OutputPath,
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
}

func TestProject2(t *testing.T) {
	logger := dsh_utils.NewLogger(dsh_utils.LogLevelAll)
	workspace, err := dsh_core.OpenWorkspace("./.test_workspace", logger)
	if err != nil {
		logger.Panic("%+v", err)
	}
	profile, err := workspace.PrepareGitApp("https://github.com/orz-dsh/not-exist-project.git", "main", nil)
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = profile.AddManifestOptionValues(-1, map[string]string{
		"option1": "value1",
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	_, err = profile.MakeApp()
	if err != nil {
		logger.Panic("%+v", err)
	}
}

func TestProject3(t *testing.T) {
	logger := dsh_utils.NewLogger(dsh_utils.LogLevelAll)
	workspace, err := dsh_core.OpenWorkspace("./.test_workspace", logger)
	if err != nil {
		logger.Panic("%+v", err)
	}
	profile, err := workspace.PrepareLocalApp("./.test2/app1", nil)
	if err != nil {
		logger.Panic("%+v", err)
	}
	options := make(map[string]string)
	if runtime.GOOS == "windows" {
		options["_shell"] = "powershell"
	}
	err = profile.AddManifestOptionValues(-1, options)
	if err != nil {
		logger.Panic("%+v", err)
	}
	app, err := profile.MakeApp()
	if err != nil {
		logger.Panic("%+v", err)
	}
	artifact, err := app.MakeScripts(dsh_core.AppMakeScriptsSettings{})
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = workspace.Clean(dsh_core.WorkspaceCleanSettings{
		ExcludeOutputPath: artifact.OutputPath,
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	exit, err := artifact.ExecuteInChildProcess("app")
	if err != nil {
		logger.Panic("%+v", err)
	}
	logger.Info("exit code: %d", exit)
	if runtime.GOOS != "windows" {
		err = artifact.ExecuteInThisProcess("app")
		if err != nil {
			logger.Panic("%+v", err)
		}
	}
}
