package main

import (
	"github.com/orz-dsh/dsh/core"
	"github.com/orz-dsh/dsh/utils"
	"os"
	"runtime"
	"testing"
)

func TestProject1(t *testing.T) {
	logger := utils.NewLogger(utils.LogLevelAll)
	err := os.Setenv("DSH_GLOBAL_VAR1", "global variable 1 in env")
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = os.Setenv("DSH_GLOBAL_VAR2", "global variable 2 in env")
	if err != nil {
		logger.Panic("%+v", err)
	}
	global, err := core.MakeGlobal(logger, map[string]string{
		"var2": "global variable 2 in map",
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	workspace, err := core.MakeWorkspace(global, "./.test_workspace")
	if err != nil {
		logger.Panic("%+v", err)
	}
	app, err := workspace.NewAppBuilder().
		AddProfileSetting(0).
		SetOptionSetting().
		AddItemMap(map[string]string{
			"_os":  "linux",
			"test": "a",
		}).
		CommitOptionSetting().
		CommitProfileSetting().
		Build("dir:./.test1/app1")
	if err != nil {
		logger.Panic("%+v", err)
	}
	artifact, err := app.MakeScripts(core.AppMakeScriptsSettings{
		OutputPath:      "./.test1/app1/output",
		OutputPathClear: true,
		UseHardLink:     true,
		Inspection:      true,
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = workspace.Clean(core.WorkspaceCleanSetting{
		ExcludeOutputPath: artifact.OutputPath,
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
}

func TestProject2(t *testing.T) {
	logger := utils.NewLogger(utils.LogLevelAll)
	global, err := core.MakeGlobal(logger, nil)
	if err != nil {
		logger.Panic("%+v", err)
	}
	workspace, err := core.MakeWorkspace(global, "./.test_workspace")
	if err != nil {
		logger.Panic("%+v", err)
	}
	_, err = workspace.NewAppBuilder().
		AddProfileSetting(0).
		SetOptionSetting().
		AddItemMap(map[string]string{
			"option1": "value1",
		}).
		CommitOptionSetting().
		CommitProfileSetting().
		Build("git:https://github.com/orz-dsh/not-exist-project.git#ref=main")
	if err != nil {
		logger.Panic("%+v", err)
	}
}

func TestProject3(t *testing.T) {
	logger := utils.NewLogger(utils.LogLevelAll)
	global, err := core.MakeGlobal(logger, nil)
	if err != nil {
		logger.Panic("%+v", err)
	}
	workspace, err := core.MakeWorkspace(global, "./.test_workspace")
	if err != nil {
		logger.Panic("%+v", err)
	}
	options := map[string]string{}
	if runtime.GOOS == "windows" {
		options[core.GenericOptionNameExecutor] = "powershell"
	}
	app, err := workspace.NewAppBuilder().
		AddProfileSetting(0).
		SetOptionSetting().
		AddItemMap(options).
		CommitOptionSetting().
		CommitProfileSetting().
		Build("dir:./.test2/app1")
	if err != nil {
		logger.Panic("%+v", err)
	}
	artifact, err := app.MakeScripts(core.AppMakeScriptsSettings{})
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = workspace.Clean(core.WorkspaceCleanSetting{
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
