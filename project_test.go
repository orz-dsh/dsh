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
	app, err := workspace.OpenLocalApp("./.test1/app1", map[string]string{
		"_os":  "linux",
		"test": "a",
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	artifact, err := app.MakeScripts(dsh_core.AppMakeScriptsOptions{
		OutputPath: "./.test1/app1/output",
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = workspace.Clean(dsh_core.WorkspaceCleanOptions{
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
	_, err = workspace.OpenGitApp("https://github.com/orz-dsh/not-exist-project.git", "main", map[string]string{
		"option1": "value1",
	})
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
	options := make(map[string]string)
	if runtime.GOOS == "windows" {
		options["_shell"] = "powershell"
	}
	app, err := workspace.OpenLocalApp("./.test2/app1", options)
	if err != nil {
		logger.Panic("%+v", err)
	}
	artifact, err := app.MakeScripts(dsh_core.AppMakeScriptsOptions{
		OutputPath: "./.test2/app1/output",
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = workspace.Clean(dsh_core.WorkspaceCleanOptions{
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
