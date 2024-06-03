package main

import (
	"dsh/dsh_core"
	"dsh/dsh_utils"
	"os"
	"runtime"
	"testing"
)

func TestProject1(t *testing.T) {
	logger := dsh_utils.NewLogger(dsh_utils.LogLevelAll)
	err := os.Setenv("DSH_GLOBAL_VAR1", "global variable 1 in env")
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = os.Setenv("DSH_GLOBAL_VAR2", "global variable 2 in env")
	if err != nil {
		logger.Panic("%+v", err)
	}
	global, err := dsh_core.MakeGlobal(logger, map[string]string{
		"var2": "global variable 2 in map",
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	workspace, err := dsh_core.MakeWorkspace(global, "./.test_workspace")
	if err != nil {
		logger.Panic("%+v", err)
	}

	profileSettingBuilder := dsh_core.NewProfileSettingBuilder().
		AddOptionItemMap(map[string]string{
			"_os":  "linux",
			"test": "a",
		})

	maker := workspace.NewAppMaker()
	err = maker.AddProfileSettingBuilder(0, profileSettingBuilder)
	if err != nil {
		logger.Panic("%+v", err)
	}
	app, err := maker.Build("dir:./.test1/app1")
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
	err = workspace.Clean(dsh_core.WorkspaceCleanSetting{
		ExcludeOutputPath: artifact.OutputPath,
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
}

func TestProject2(t *testing.T) {
	logger := dsh_utils.NewLogger(dsh_utils.LogLevelAll)
	global, err := dsh_core.MakeGlobal(logger, nil)
	if err != nil {
		logger.Panic("%+v", err)
	}
	workspace, err := dsh_core.MakeWorkspace(global, "./.test_workspace")
	if err != nil {
		logger.Panic("%+v", err)
	}
	maker := workspace.NewAppMaker()

	profileSettingBuilder := dsh_core.NewProfileSettingBuilder().
		AddOptionItemMap(map[string]string{
			"option1": "value1",
		})

	err = maker.AddProfileSettingBuilder(0, profileSettingBuilder)
	if err != nil {
		logger.Panic("%+v", err)
	}
	_, err = maker.Build("git:https://github.com/orz-dsh/not-exist-project.git#ref=main")
	if err != nil {
		logger.Panic("%+v", err)
	}
}

func TestProject3(t *testing.T) {
	logger := dsh_utils.NewLogger(dsh_utils.LogLevelAll)
	global, err := dsh_core.MakeGlobal(logger, nil)
	if err != nil {
		logger.Panic("%+v", err)
	}
	workspace, err := dsh_core.MakeWorkspace(global, "./.test_workspace")
	if err != nil {
		logger.Panic("%+v", err)
	}
	maker := workspace.NewAppMaker()
	options := map[string]string{}
	if runtime.GOOS == "windows" {
		options["_shell"] = "powershell"
	}

	profileSettingBuilder := dsh_core.NewProfileSettingBuilder().
		AddOptionItemMap(options)

	err = maker.AddProfileSettingBuilder(0, profileSettingBuilder)
	if err != nil {
		logger.Panic("%+v", err)
	}
	app, err := maker.Build("dir:./.test2/app1")
	if err != nil {
		logger.Panic("%+v", err)
	}
	artifact, err := app.MakeScripts(dsh_core.AppMakeScriptsSettings{})
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = workspace.Clean(dsh_core.WorkspaceCleanSetting{
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
