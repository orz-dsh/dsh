package main

import (
	. "github.com/orz-dsh/dsh/core"
	. "github.com/orz-dsh/dsh/core/common"
	. "github.com/orz-dsh/dsh/utils"
	"os"
	"runtime"
	"testing"
)

func TestProject1(t *testing.T) {
	logger := NewLogger(LogLevelAll)
	err := os.Setenv("DSH_ARGUMENT_ITEM_VAR1", "global variable 1 in env")
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = os.Setenv("DSH_ARGUMENT_ITEM_VAR2", "global variable 2 in env")
	if err != nil {
		logger.Panic("%+v", err)
	}
	environment, err := NewEnvironment(logger, map[string]string{
		"argument_item_var2":          "global variable 2 in map",
		"argument_item_test01":        "value01",
		"workspace_dir":               "./.test_workspace",
		"workspace_clean":             "{\"output\":{\"count\":30}}",
		"workspace_registry_item_002": "{\"name\":\"registry1\",\"link\":\"git:https://github.com\"}",
		"workspace_registry_item_001": "{\"name\":\"registry2\",\"link\":\"git:https://github.com\"}",
		"workspace_registry_item_1":   "{\"name\":\"registry3\",\"link\":\"git:https://github.com\"}",
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	workspace, err := NewWorkspace(environment, "")
	if err != nil {
		logger.Panic("%+v", err)
	}
	app, err := workspace.NewAppBuilder().
		AddProfileSetting("test code", 0).
		SetArgumentSetting().
		AddItemMap(map[string]string{
			"_os":  "linux",
			"test": "a",
		}).
		CommitArgumentSetting().
		CommitProfileSetting().
		Build("dir:./.test1/app1")
	if err != nil {
		logger.Panic("%+v", err)
	}

	inspection, err := app.Inspect()
	if err != nil {
		logger.Panic("%+v", err)
	}
	logger.InfoDesc("inspect app", KV("inspection", inspection))

	artifact, err := app.MakeArtifact(MakeArtifactOptions{
		OutputDir:         "./.test1/app1/output",
		OutputDirClear:    true,
		UseHardLink:       true,
		InspectSerializer: YamlSerializerDefault,
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = workspace.Clean(WorkspaceCleanOptions{
		ExcludeOutputDir: artifact.GetOutputDir(),
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
}

func TestProject2(t *testing.T) {
	logger := NewLogger(LogLevelAll)
	global, err := NewEnvironment(logger, nil)
	if err != nil {
		logger.Panic("%+v", err)
	}
	workspace, err := NewWorkspace(global, "./.test_workspace")
	if err != nil {
		logger.Panic("%+v", err)
	}
	_, err = workspace.NewAppBuilder().
		AddProfileSetting("test code", 0).
		SetArgumentSetting().
		AddItemMap(map[string]string{
			"option1": "value1",
		}).
		CommitArgumentSetting().
		CommitProfileSetting().
		Build("git:https://github.com/orz-dsh/not-exist-project.git#ref=main")
	if err != nil {
		logger.Panic("%+v", err)
	}
}

func TestProject3(t *testing.T) {
	logger := NewLogger(LogLevelAll)
	environment, err := NewEnvironment(logger, nil)
	if err != nil {
		logger.Panic("%+v", err)
	}
	workspace, err := NewWorkspace(environment, "./.test_workspace")
	if err != nil {
		logger.Panic("%+v", err)
	}
	options := map[string]string{}
	if runtime.GOOS == "windows" {
		options[OptionNameCommonExecutor] = "powershell"
	}
	app, err := workspace.NewAppBuilder().
		AddProfileSetting("test code", 0).
		SetArgumentSetting().
		AddItemMap(options).
		CommitArgumentSetting().
		CommitProfileSetting().
		Build("dir:./.test2/app1")
	if err != nil {
		logger.Panic("%+v", err)
	}
	artifact, err := app.MakeArtifact(MakeArtifactOptions{
		InspectSerializer: YamlSerializerDefault,
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = workspace.Clean(WorkspaceCleanOptions{
		ExcludeOutputDir: artifact.GetOutputDir(),
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

func TestEnvironment(t *testing.T) {
	logger := NewLogger(LogLevelAll)
	environment, err := NewEnvironment(logger, map[string]string{
		"workspace_dir":               "./.test_workspace",
		"workspace_clean":             "{\"output\":{\"count\":30}}",
		"workspace_registry_item_002": "{\"name\":\"registry1\",\"link\":\"git:https://github.com\"}",
		"workspace_registry_item_001": "{\"name\":\"registry2\",\"link\":\"git:https://github.com\"}",
		"workspace_registry_item_1":   "{\"name\":\"registry3\",\"link\":\"git:https://github.com\"}",
		"workspace_profile_item_001":  "{\"file\":\"profile1\"}",
		"argument_item_test01":        "value01",
	})
	if err != nil {
		logger.Panic("%+v", err)
	}
	logger.InfoDesc("test environment", KV("environment", environment))
}
