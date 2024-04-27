package main

import (
	"dsh/dsh_core"
	"dsh/dsh_utils"
	"testing"
)

func TestProject1(t *testing.T) {
	logger := dsh_utils.NewLogger(dsh_utils.LogLevelAll)
	workspace, err := dsh_core.LoadWorkspace("", logger)
	if err != nil {
		logger.Panic("%+v", err)
	}
	context := dsh_core.NewContext(workspace, logger)
	project, err := workspace.OpenLocalProject(context, "./.test/app1")
	if err != nil {
		logger.Panic("%+v", err)
	}
	err = project.Build("")
	if err != nil {
		logger.Panic("%+v", err)
	}
}

func TestProject2(t *testing.T) {
	logger := dsh_utils.NewLogger(dsh_utils.LogLevelAll)
	workspace, err := dsh_core.LoadWorkspace("", logger)
	if err != nil {
		logger.Panic("%+v", err)
	}
	context := dsh_core.NewContext(workspace, logger)
	_, err = workspace.OpenGitProject(context, "https://github.com/orz-dsh/not-exist-project.git", "main")
	if err != nil {
		logger.Panic("%+v", err)
	}
}
