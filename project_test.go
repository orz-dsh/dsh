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
	project, err := workspace.LoadLocalProject("./.test/app1")
	if err != nil {
		logger.Panic("%+v", err)
	}
	builder := project.NewBuilder()
	err = builder.Build("")
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
	_, err = workspace.LoadGitProject("", "https://github.com/orz-dsh/not-exist-project.git", nil, "main", nil)
	if err != nil {
		logger.Panic("%+v", err)
	}
}
